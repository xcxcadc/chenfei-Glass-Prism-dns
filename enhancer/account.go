package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var accountNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,20}$`)

type accountUpdateRequest struct {
	OldUsername string `json:"old_username"`
	OldPassword string `json:"old_password"`
	NewUsername string `json:"new_username"`
	NewPassword string `json:"new_password"`
}

func (app *App) handleAccountUpdate(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(writer, http.MethodPost)
		return
	}
	if !app.authorize(request.Context(), request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "登录已失效"})
		return
	}
	var payload accountUpdateRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	payload.OldUsername = strings.TrimSpace(payload.OldUsername)
	payload.NewUsername = strings.TrimSpace(payload.NewUsername)
	if !accountNamePattern.MatchString(payload.NewUsername) {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "用户名必须为 3-20 位字母、数字或下划线"})
		return
	}
	if len(payload.NewPassword) < 6 || len(payload.NewPassword) > 72 {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "密码长度必须为 6-72 位"})
		return
	}
	if err := app.verifyCredentials(request.Context(), payload.OldUsername, payload.OldPassword); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "旧用户名或旧密码不正确"})
		return
	}
	if err := updateControllerAccount(app.controllerDB, payload); err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]bool{"updated": true})
}

func (app *App) verifyCredentials(ctx context.Context, username, password string) error {
	body, err := json.Marshal(map[string]string{"username": username, "password": password})
	if err != nil {
		return err
	}
	loginURL := *app.upstream
	loginURL.Path = "/api/login"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := app.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return errors.New("invalid credentials")
	}
	return nil
}

func updateControllerAccount(databasePath string, payload accountUpdateRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("生成密码摘要失败: %w", err)
	}
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return fmt.Errorf("打开 Controller 数据库失败: %w", err)
	}
	defer database.Close()
	transaction, err := database.Begin()
	if err != nil {
		return fmt.Errorf("开始账户更新失败: %w", err)
	}
	defer transaction.Rollback()
	var duplicate int
	if err := transaction.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ? AND username <> ?`, payload.NewUsername, payload.OldUsername).Scan(&duplicate); err != nil {
		return fmt.Errorf("检查用户名失败: %w", err)
	}
	if duplicate > 0 {
		return errors.New("新用户名已存在")
	}
	result, err := transaction.Exec(`UPDATE users SET username = ?, password = ? WHERE username = ?`, payload.NewUsername, string(hash), payload.OldUsername)
	if err != nil {
		return fmt.Errorf("更新账户失败: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil || updated != 1 {
		return errors.New("旧账户不存在或账户数据异常")
	}
	if _, err := transaction.Exec(`DELETE FROM sessions`); err != nil {
		return fmt.Errorf("清理旧登录会话失败: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("保存账户失败: %w", err)
	}
	return nil
}
