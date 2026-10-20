package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type MediaCheckJob struct {
	ID          string            `json:"id"`
	NodeID      string            `json:"node_id,omitempty"`
	ConfigID    string            `json:"config_id,omitempty"`
	Status      string            `json:"status"`
	Source      string            `json:"source"`
	Network     string            `json:"network"`
	Region      string            `json:"region"`
	RequestedAt time.Time         `json:"requested_at"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at"`
	FinishedAt  *time.Time        `json:"finished_at,omitempty"`
	Results     map[string]string `json:"results"`
	Error       string            `json:"error,omitempty"`
	Sequence    int               `json:"sequence"`
	Generation  string            `json:"generation,omitempty"`
}

type MediaCheckStore struct {
	mu   sync.Mutex
	path string
	jobs map[string]MediaCheckJob
}

var errMediaConflict = errors.New("media check is finished or report is out of order")

func NewMediaCheckStore(path string) (*MediaCheckStore, error) {
	store := &MediaCheckStore{path: path, jobs: make(map[string]MediaCheckJob)}
	if path == "" {
		return store, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &store.jobs); err != nil {
		return nil, err
	}
	for id, job := range store.jobs {
		if job.Status == "running" {
			now := time.Now().UTC()
			job.Status, job.Error, job.FinishedAt = "failed", "Panel restarted during the check; run again", &now
			store.jobs[id] = job
		}
	}
	return store, store.saveLocked()
}

func (store *MediaCheckStore) saveLocked() error {
	if store.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(store.path), 0o750); err != nil {
		return err
	}
	data, err := json.Marshal(store.jobs)
	if err != nil {
		return err
	}
	if err := os.WriteFile(store.path+".tmp", data, 0o600); err != nil {
		return err
	}
	return os.Rename(store.path+".tmp", store.path)
}

func mediaJobActive(job MediaCheckJob) bool { return job.Status == "queued" || job.Status == "running" }

func (store *MediaCheckStore) expireLocked() {
	now := time.Now().UTC()
	for id, job := range store.jobs {
		if mediaJobActive(job) && now.Sub(job.UpdatedAt) > 90*time.Second {
			job.Status, job.Error, job.FinishedAt = "failed", "Worker has not responded for 90 seconds; check its connection and timer", &now
			store.jobs[id] = job
		}
		if !mediaJobActive(job) && now.Sub(job.UpdatedAt) > 48*time.Hour {
			delete(store.jobs, id)
		}
	}
}

func (store *MediaCheckStore) Latest(nodeID, configID string) (MediaCheckJob, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.expireLocked()
	return store.latestLocked(nodeID, configID)
}

func (store *MediaCheckStore) latestLocked(nodeID, configID string) (MediaCheckJob, bool) {
	var latest MediaCheckJob
	for _, job := range store.jobs {
		if job.NodeID == nodeID && job.ConfigID == configID && job.RequestedAt.After(latest.RequestedAt) {
			latest = job
		}
	}
	latest.Results = cloneStringMap(latest.Results)
	if latest.Results == nil {
		latest.Results = map[string]string{}
	}
	return latest, latest.ID != ""
}

func (store *MediaCheckStore) Create(nodeID, configID, generation string) (MediaCheckJob, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.expireLocked()
	if existing, ok := store.latestLocked(nodeID, configID); ok && mediaJobActive(existing) && existing.Generation == generation {
		return existing, nil
	}
	id, err := randomToken()
	if err != nil {
		return MediaCheckJob{}, err
	}
	now := time.Now().UTC()
	for previousID, previous := range store.jobs {
		if previous.NodeID == nodeID && previous.ConfigID == configID && mediaJobActive(previous) {
			previous.Status, previous.Error, previous.FinishedAt = "failed", "Configuration changed; a new check was queued", &now
			store.jobs[previousID] = previous
		}
	}
	job := MediaCheckJob{ID: id, NodeID: nodeID, ConfigID: configID, Generation: generation, Status: "queued", Source: "check.unlock.media", Network: "ipv4", Region: "0", RequestedAt: now, UpdatedAt: now, Results: map[string]string{}}
	store.jobs[id] = job
	return job, store.saveLocked()
}

func (store *MediaCheckStore) Get(id string) (MediaCheckJob, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.expireLocked()
	job, ok := store.jobs[id]
	job.Results = cloneStringMap(job.Results)
	if job.Results == nil {
		job.Results = map[string]string{}
	}
	return job, ok
}

func (store *MediaCheckStore) Claim(nodeID, configID string) (*MediaCheckJob, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.expireLocked()
	job, ok := store.latestLocked(nodeID, configID)
	if !ok || job.Status != "queued" {
		return nil, nil
	}
	now := time.Now().UTC()
	job.Status, job.StartedAt, job.UpdatedAt = "running", &now, now
	store.jobs[job.ID] = job
	return &job, store.saveLocked()
}

func (store *MediaCheckStore) Report(id, nodeID, configID, status string, sequence int, results map[string]string, message string) (MediaCheckJob, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.expireLocked()
	job, ok := store.jobs[id]
	if !ok || job.NodeID != nodeID || job.ConfigID != configID {
		return MediaCheckJob{}, os.ErrNotExist
	}
	if job.Status != "running" || sequence <= job.Sequence {
		return MediaCheckJob{}, errMediaConflict
	}
	if status != "running" && status != "completed" && status != "failed" {
		return MediaCheckJob{}, errors.New("invalid media check status")
	}
	for label, value := range results {
		if !globalMediaLabel(label) || len(value) > 1024 || strings.ContainsAny(value, "\x1b\r\n") {
			return MediaCheckJob{}, errors.New("invalid IPv4 media result")
		}
	}
	if len(message) > 1024 {
		return MediaCheckJob{}, errors.New("media error is too long")
	}
	job.Results = cloneStringMap(job.Results)
	if job.Results == nil {
		job.Results = map[string]string{}
	}
	for label, value := range results {
		if strings.TrimSpace(value) != "" {
			job.Results[label] = strings.TrimSpace(value)
		}
	}
	job.Status, job.Sequence, job.Error, job.UpdatedAt = status, sequence, message, time.Now().UTC()
	if status != "running" {
		now := job.UpdatedAt
		job.FinishedAt = &now
	}
	if status == "completed" && len(job.Results) == 0 {
		job.Status, job.Error = "failed", "Official script returned no IPv4 platform results"
	}
	store.jobs[id] = job
	err := store.saveLocked()
	job.Results = cloneStringMap(job.Results)
	return job, err
}

var globalMediaLabels = []string{"Dazn", "Disney+", "Netflix", "YouTube Premium", "Amazon Prime Video", "TVBAnywhere+", "Spotify Registration", "OneTrust Region", "iQyi Oversea Region", "Bing Region", "Apple Region", "YouTube CDN", "Netflix Preferred CDN", "ChatGPT", "Google Gemini", "Claude", "Wikipedia Editability", "Google Play Store", "Google Search CAPTCHA Free", "Steam Currency", "Reddit", "SD Gundam G Generation Eternal"}

func (app *App) controllerNodeRole(id string, role *string) error {
	database, err := sql.Open("sqlite", app.controllerDB)
	if err != nil {
		return err
	}
	defer database.Close()
	return database.QueryRow(`SELECT role FROM nodes WHERE id = ?`, id).Scan(role)
}

func globalMediaLabel(label string) bool {
	for _, allowed := range globalMediaLabels {
		if label == allowed {
			return true
		}
	}
	return false
}

func mediaGeneration(record ipConfigRecord) string {
	if record.ServiceAuditRequestedAt == nil {
		return ""
	}
	return record.ServiceAuditRequestedAt.Format(time.RFC3339Nano)
}

func (app *App) handleMediaChecks(writer http.ResponseWriter, request *http.Request) {
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, 401, map[string]string{"error": "登录已失效"})
		return
	}
	var payload struct {
		NodeID   string `json:"node_id"`
		ConfigID string `json:"config_id"`
	}
	if request.Method == http.MethodGet {
		payload.NodeID, payload.ConfigID = request.URL.Query().Get("node_id"), request.URL.Query().Get("config_id")
	} else if request.Method == http.MethodPost {
		if err := decodeJSON(request, &payload); err != nil {
			writeJSON(writer, 400, map[string]string{"error": err.Error()})
			return
		}
	} else {
		methodNotAllowed(writer, http.MethodGet, http.MethodPost)
		return
	}
	if (payload.NodeID == "") == (payload.ConfigID == "") {
		writeJSON(writer, 400, map[string]string{"error": "Specify exactly one node_id or config_id"})
		return
	}
	generation := ""
	if payload.ConfigID != "" {
		record, ok := app.ipStore.Record(payload.ConfigID)
		if !ok {
			writeJSON(writer, 404, map[string]string{"error": "IP 配置不存在"})
			return
		}
		generation = mediaGeneration(record)
	} else {
		var role string
		if err := app.controllerNodeRole(payload.NodeID, &role); err != nil || role != "proxy" {
			writeJSON(writer, 404, map[string]string{"error": "解锁机不存在"})
			return
		}
	}
	if request.Method == http.MethodGet {
		job, ok := app.mediaChecks.Latest(payload.NodeID, payload.ConfigID)
		if !ok {
			writeJSON(writer, 200, nil)
			return
		}
		writeJSON(writer, 200, job)
		return
	}
	job, err := app.mediaChecks.Create(payload.NodeID, payload.ConfigID, generation)
	if err != nil {
		writeJSON(writer, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusAccepted, job)
}

func (app *App) handleMediaCheck(writer http.ResponseWriter, request *http.Request) {
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, 401, map[string]string{"error": "登录已失效"})
		return
	}
	if request.Method != http.MethodGet {
		methodNotAllowed(writer, http.MethodGet)
		return
	}
	job, ok := app.mediaChecks.Get(strings.TrimPrefix(request.URL.Path, "/enhancer/api/media-checks/"))
	if !ok {
		writeJSON(writer, 404, map[string]string{"error": "检测任务不存在"})
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writeJSON(writer, 200, job)
}

func (app *App) handleMediaCheckWorker(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	var payload struct {
		Role       string            `json:"role"`
		Credential string            `json:"credential"`
		Action     string            `json:"action"`
		JobID      string            `json:"job_id"`
		Sequence   int               `json:"sequence"`
		Status     string            `json:"status"`
		Results    map[string]string `json:"results"`
		Error      string            `json:"error"`
	}
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, 400, map[string]string{"error": err.Error()})
		return
	}
	nodeID, configID, generation := "", "", ""
	var record ipConfigRecord
	if payload.Role == "client" {
		var ok bool
		record, ok = app.ipStore.GetByToken(payload.Credential)
		if !ok || payload.Credential == "" {
			writeJSON(writer, 401, map[string]string{"error": "invalid client credential"})
			return
		}
		configID, generation = record.ID, mediaGeneration(record)
	} else if payload.Role == "proxy" {
		node, err := controllerNodeBySecret(app.controllerDB, payload.Credential)
		if err != nil || node.Role != "proxy" {
			writeJSON(writer, 401, map[string]string{"error": "invalid proxy credential"})
			return
		}
		nodeID = node.ID
	} else {
		writeJSON(writer, 400, map[string]string{"error": "invalid role"})
		return
	}
	if payload.Action == "claim" {
		latest, exists := app.mediaChecks.Latest(nodeID, configID)
		if !exists || latest.Generation != generation || (!mediaJobActive(latest) && time.Since(latest.RequestedAt) > 6*time.Hour) {
			if _, err := app.mediaChecks.Create(nodeID, configID, generation); err != nil {
				writeJSON(writer, 500, map[string]string{"error": err.Error()})
				return
			}
		}
		job, err := app.mediaChecks.Claim(nodeID, configID)
		if err != nil {
			writeJSON(writer, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(writer, 200, map[string]any{"job": job})
		return
	}
	if payload.Action != "report" {
		writeJSON(writer, 400, map[string]string{"error": "invalid action"})
		return
	}
	previous, exists := app.mediaChecks.Get(payload.JobID)
	if !exists || previous.NodeID != nodeID || previous.ConfigID != configID {
		writeJSON(writer, 404, map[string]string{"error": "job not found"})
		return
	}
	if configID != "" && previous.Generation != generation {
		writeJSON(writer, 409, map[string]string{"error": "configuration changed; result rejected"})
		return
	}
	job, err := app.mediaChecks.Report(payload.JobID, nodeID, configID, payload.Status, payload.Sequence, payload.Results, payload.Error)
	if err != nil {
		writeJSON(writer, 409, map[string]string{"error": err.Error()})
		return
	}
	if configID != "" && len(job.Results) > 0 {
		results := make(map[string]string)
		for _, service := range app.catalog.Snapshot(request.Context(), false).Services {
			if _, selected := record.Routes[service.ID]; !selected {
				continue
			}
			results[service.ID] = mediaServiceResult(service, job)
		}
		if _, err := app.ipStore.UpdateServiceAudit(record.EnrollmentToken, results, generation); err != nil {
			writeJSON(writer, 409, map[string]string{"error": err.Error()})
			return
		}
	}
	writeJSON(writer, 200, job)
}

func mediaServiceResult(service Service, job MediaCheckJob) string {
	spec := mediaTestSpecForService(service)
	labels := append(append([]string(nil), spec.Required...), spec.Any...)
	covered := false
	for _, label := range labels {
		covered = covered || globalMediaLabel(label)
	}
	if !covered {
		return "NOT COVERED (check.unlock.media IPv4 / menu 0)"
	}
	var values []string
	for _, label := range labels {
		if value := job.Results[label]; value != "" {
			values = append(values, value)
		}
	}
	if len(values) == 0 {
		if mediaJobActive(job) {
			return "NOT TESTED (check.unlock.media IPv4 running)"
		}
		return "NOT TESTED (script did not return this IPv4 platform result)"
	}
	return strings.Join(values, "; ")
}
