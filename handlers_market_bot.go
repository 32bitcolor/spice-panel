package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// botProxy forwards a request to the market bot API and returns the raw body.
// Returns (body, statusCode, error).
func botProxy(method, path string, body io.Reader) ([]byte, int, error) {
	if marketBotAddr == "" {
		return nil, 503, fmt.Errorf("market_bot_addr not configured")
	}
	req, err := http.NewRequestWithContext(context.Background(), method, marketBotAddr+path, body)
	if err != nil {
		return nil, 500, err
	}
	if marketBotToken != "" {
		req.Header.Set("Authorization", "Bearer "+marketBotToken)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 503, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

// handleMarketBotStatus proxies GET /status from the bot API.
func handleMarketBotStatus(w http.ResponseWriter, r *http.Request) {
	data, code, err := botProxy(http.MethodGet, "/status", nil)
	if err != nil {
		jsonErr(w, err, code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data) //nolint:errcheck
}

// handleMarketBotConfig proxies GET/PUT /config from/to the bot API.
func handleMarketBotConfig(w http.ResponseWriter, r *http.Request) {
	var body io.Reader
	if r.Method == http.MethodPut {
		body = r.Body
	}
	data, code, err := botProxy(r.Method, "/config", body)
	if err != nil {
		jsonErr(w, err, code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data) //nolint:errcheck
}

var botCmdAllowlist = map[string]bool{
	"start": true, "stop": true, "restart": true,
}

// handleMarketBotExec runs a lifecycle command (start/stop/restart) on the bot
// using the configured control plane.
func handleMarketBotExec(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Cmd string `json:"cmd"`
	}
	if err := decode(r, &req); err != nil {
		jsonErr(w, err, 400)
		return
	}
	if !botCmdAllowlist[req.Cmd] {
		jsonErr(w, fmt.Errorf("unknown command %q", req.Cmd), 400)
		return
	}
	if globalExecutor == nil {
		jsonErr(w, fmt.Errorf("not connected"), 503)
		return
	}
	if marketBotContainer == "" {
		jsonErr(w, fmt.Errorf("market_bot_container not configured"), 503)
		return
	}

	out, err := execBotCommand(r.Context(), req.Cmd)
	if err != nil {
		jsonErr(w, fmt.Errorf("exec: %w — output: %s", err, out), 500)
		return
	}
	jsonOK(w, map[string]string{"output": out})
}

// execBotCommand runs start/stop/restart on the bot container/deployment.
func execBotCommand(ctx context.Context, cmd string) (string, error) {
	if globalControl == nil {
		return "", fmt.Errorf("not connected")
	}
	switch globalControl.Name() {
	case "kubectl":
		ns := marketBotNamespace
		if ns == "" {
			ns = "default"
		}
		switch cmd {
		case "start":
			return globalExecutor.Exec(fmt.Sprintf("kubectl scale deployment/%s -n %s --replicas=1 2>&1", marketBotContainer, ns))
		case "stop":
			return globalExecutor.Exec(fmt.Sprintf("kubectl scale deployment/%s -n %s --replicas=0 2>&1", marketBotContainer, ns))
		case "restart":
			return globalExecutor.Exec(fmt.Sprintf("kubectl rollout restart deployment/%s -n %s 2>&1", marketBotContainer, ns))
		}
	case "docker":
		return globalExecutor.Exec(fmt.Sprintf("docker %s %s 2>&1", cmd, marketBotContainer))
	}
	return "", fmt.Errorf("lifecycle commands not supported for %s control plane", globalControl.Name())
}

// handleMarketBotLogs streams bot container logs over WebSocket.
// It discovers the bot pod (kubectl) or uses the container name directly (docker).
func handleMarketBotLogs(w http.ResponseWriter, r *http.Request) {
	if globalControl == nil || globalExecutor == nil {
		http.Error(w, "not connected", 503)
		return
	}
	if marketBotContainer == "" {
		http.Error(w, "market_bot_container not configured", 503)
		return
	}

	ns, name, err := botLogSource(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 503)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Time{})

	ch, cancel, err := globalControl.StreamLog(r.Context(), globalExecutor, ns, name)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error: "+err.Error())) //nolint:errcheck
		return
	}
	defer cancel()

	for line := range ch {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			return
		}
	}
}

// botLogSource returns the namespace and pod/container name for log streaming.
func botLogSource(ctx context.Context) (ns, name string, err error) {
	switch globalControl.Name() {
	case "kubectl":
		botNS := marketBotNamespace
		if botNS == "" {
			botNS = "default"
		}
		out, execErr := globalExecutor.Exec(fmt.Sprintf(
			"kubectl get pods -n %s --field-selector=status.phase=Running -o jsonpath='{.items[0].metadata.name}' 2>/dev/null",
			botNS))
		pod := strings.TrimSpace(strings.Trim(out, "'"))
		if execErr != nil || pod == "" {
			// Retry without Running filter — pod might be initialising.
			out, _ = globalExecutor.Exec(fmt.Sprintf(
				"kubectl get pods -n %s -o jsonpath='{.items[0].metadata.name}' 2>/dev/null",
				botNS))
			pod = strings.TrimSpace(strings.Trim(out, "'"))
		}
		if pod == "" {
			return "", "", fmt.Errorf("no pod found in namespace %s", botNS)
		}
		return botNS, pod, nil
	default:
		// docker/local: container name is used directly.
		return "", marketBotContainer, nil
	}
}
