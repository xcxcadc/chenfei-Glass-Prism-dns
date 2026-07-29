package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"sort"
	"strings"
	"time"
)

const failoverAttemptWindow = 30 * time.Minute

var errNoFailover = errors.New("no automatic failover")

type FailoverEvent struct {
	Status          string    `json:"status"`
	Reason          string    `json:"reason"`
	FromProxyID     string    `json:"from_proxy_id,omitempty"`
	FromProxyName   string    `json:"from_proxy_name,omitempty"`
	ToProxyID       string    `json:"to_proxy_id,omitempty"`
	ToProxyName     string    `json:"to_proxy_name,omitempty"`
	TriedProxyIDs   []string  `json:"tried_proxy_ids,omitempty"`
	TriggeredAt     time.Time `json:"triggered_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	RecoveredResult string    `json:"recovered_result,omitempty"`
}

type failoverCandidate struct {
	ID        string
	Name      string
	IPv4s     []string
	Priority  int
	Latency   int
	Heartbeat time.Time
}

func auditResultPassed(result string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(result))
	return strings.HasPrefix(normalized, "YES") ||
		strings.HasPrefix(normalized, "PASS") ||
		strings.HasPrefix(normalized, "AVAILABLE")
}

func controllerUnlockPassed(result string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(result))
	return strings.HasPrefix(normalized, "YES") ||
		strings.HasPrefix(normalized, "PASS") ||
		strings.HasPrefix(normalized, "AVAILABLE")
}

func (app *App) reconcileAutomaticFailover(config IPConfig, results map[string]string) (IPConfig, error) {
	services := make(map[string]Service)
	catalogServices := app.catalog.Snapshot(context.Background(), false).Services
	for _, service := range catalogServices {
		services[service.ID] = service
	}
	current := config
	processed := make(map[string]struct{})
	for serviceID, result := range results {
		if _, skip := processed[serviceID]; skip {
			continue
		}
		if auditResultPassed(result) {
			updated, err := app.ipStore.MarkFailoverRecovered(current.ID, serviceID, result)
			if err == nil {
				current = updated
			} else if !errors.Is(err, errNoFailover) {
				return current, err
			}
			continue
		}
		if _, ok := services[serviceID]; !ok {
			continue
		}
		linkedServiceIDs := conflictingServiceIDs(serviceID, current.Routes, catalogServices)
		for _, linkedServiceID := range linkedServiceIDs {
			processed[linkedServiceID] = struct{}{}
		}
		candidate, ok, err := app.selectFailoverCandidate(current, linkedServiceIDs, services)
		if err != nil {
			return current, err
		}
		if !ok {
			updated, updateErr := app.ipStore.RecordFailoverUnavailableGroup(current.ID, linkedServiceIDs, result)
			if updateErr != nil {
				return current, updateErr
			}
			current = updated
			continue
		}
		updated, updateErr := app.ipStore.ApplyAutomaticFailoverGroup(
			current.ID,
			linkedServiceIDs,
			candidate.ID,
			candidate.Name,
			candidate.IPv4s,
			result,
		)
		if updateErr != nil {
			return current, updateErr
		}
		current = updated
	}
	return current, nil
}

func (app *App) selectFailoverCandidate(config IPConfig, serviceIDs []string, services map[string]Service) (failoverCandidate, bool, error) {
	database, err := sql.Open("sqlite", app.controllerDB)
	if err != nil {
		return failoverCandidate{}, false, err
	}
	defer database.Close()

	rows, err := database.Query(`
		SELECT id, COALESCE(name, ''), COALESCE(public_ip, ''), COALESCE(address, ''),
		       COALESCE(priority, 0), COALESCE(latency, 0), COALESCE(last_heartbeat, ''),
		       COALESCE(unlock_json, '')
		FROM nodes
		WHERE role = 'proxy'
	`)
	if err != nil {
		return failoverCandidate{}, false, err
	}
	defer rows.Close()

	excluded := make(map[string]struct{})
	for _, serviceID := range serviceIDs {
		excluded[config.Routes[serviceID]] = struct{}{}
		if event, exists := config.Failovers[serviceID]; exists && time.Since(event.TriggeredAt) <= failoverAttemptWindow {
			for _, proxyID := range event.TriedProxyIDs {
				excluded[proxyID] = struct{}{}
			}
		}
	}
	now := time.Now().UTC()
	candidates := make([]failoverCandidate, 0)
	for rows.Next() {
		var candidate failoverCandidate
		var publicIP, address, heartbeatValue, unlockValue string
		if err := rows.Scan(
			&candidate.ID,
			&candidate.Name,
			&publicIP,
			&address,
			&candidate.Priority,
			&candidate.Latency,
			&heartbeatValue,
			&unlockValue,
		); err != nil {
			return failoverCandidate{}, false, err
		}
		if _, skip := excluded[candidate.ID]; skip {
			continue
		}
		candidate.IPv4s = ipv4Values(publicIP, address)
		if len(candidate.IPv4s) == 0 {
			continue
		}
		candidate.Heartbeat = parseControllerTime(heartbeatValue)
		if candidate.Heartbeat.IsZero() || now.Sub(candidate.Heartbeat) > 90*time.Second {
			continue
		}
		var unlock map[string]string
		if json.Unmarshal([]byte(unlockValue), &unlock) != nil {
			continue
		}
		available := true
		for _, serviceID := range serviceIDs {
			providers := unlockTestProviders(services[serviceID])
			if len(providers) == 0 {
				continue
			}
			serviceAvailable := false
			for _, provider := range providers {
				if controllerUnlockPassed(unlock[provider]) {
					serviceAvailable = true
					break
				}
			}
			if !serviceAvailable {
				available = false
				break
			}
		}
		if !available {
			continue
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return failoverCandidate{}, false, err
	}
	if len(candidates) == 0 {
		return failoverCandidate{}, false, nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		leftLatency := candidates[i].Latency
		rightLatency := candidates[j].Latency
		if leftLatency <= 0 {
			leftLatency = int(^uint(0) >> 1)
		}
		if rightLatency <= 0 {
			rightLatency = int(^uint(0) >> 1)
		}
		if leftLatency != rightLatency {
			return leftLatency < rightLatency
		}
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority > candidates[j].Priority
		}
		return candidates[i].Name < candidates[j].Name
	})
	return candidates[0], true, nil
}

func ipv4Values(values ...string) []string {
	seen := make(map[string]struct{})
	for _, value := range values {
		for _, candidate := range strings.FieldsFunc(value, func(character rune) bool {
			return character == ',' || character == ';' || character == ' ' || character == '\t'
		}) {
			candidate = strings.Trim(candidate, "[]")
			if host, _, err := net.SplitHostPort(candidate); err == nil {
				candidate = strings.Trim(host, "[]")
			}
			parsed := net.ParseIP(candidate)
			if parsed != nil && parsed.To4() != nil {
				seen[parsed.String()] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func parseControllerTime(value string) time.Time {
	value = normalizeControllerTime(value)
	for _, layout := range []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05-07:00",
	} {
		if parsed, err := time.Parse(layout, strings.TrimSpace(value)); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}

func normalizeControllerTime(value string) string {
	value = strings.TrimSpace(value)
	dotIndex := strings.IndexByte(value, '.')
	if dotIndex < 0 {
		return value
	}
	timezoneOffset := strings.IndexAny(value[dotIndex+1:], "+-Z")
	if timezoneOffset < 0 {
		return value
	}
	timezoneIndex := dotIndex + 1 + timezoneOffset
	fraction := value[dotIndex+1 : timezoneIndex]
	if len(fraction) <= 9 {
		return value
	}
	return value[:dotIndex+1] + fraction[:9] + value[timezoneIndex:]
}
