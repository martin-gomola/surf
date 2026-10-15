package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"surf-backend/internal/config"
)

func TestDockerDashboardRequiresTokenAndProxiesAdmin(t *testing.T) {
	var proxiedPath, proxiedQuery string
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxiedPath, proxiedQuery = r.URL.Path, r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	})
	handler := dockerDashboardHandler(&config.Config{
		ServerName: "Home Surf", PublicAddress: "surf.example.test:443", Port: 8080,
	}, "secret-token", backend)

	root := httptest.NewRecorder()
	handler.ServeHTTP(root, httptest.NewRequest(http.MethodGet, "/", nil))
	if root.Code != http.StatusOK || !strings.Contains(root.Body.String(), `data-runtime="docker"`) || strings.Contains(root.Body.String(), "secret-token") {
		t.Fatalf("dashboard root = %d token=%t docker=%t", root.Code, strings.Contains(root.Body.String(), "secret-token"), strings.Contains(root.Body.String(), `data-runtime="docker"`))
	}

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/config", nil))
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("tokenless config = %d", unauthorized.Code)
	}

	configured := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	request.Header.Set(dashboardTokenHeader, "secret-token")
	handler.ServeHTTP(configured, request)
	var payload map[string]any
	if err := json.Unmarshal(configured.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if configured.Code != http.StatusOK || payload["runtime"] != "docker" || payload["serverName"] != "Home Surf" {
		t.Fatalf("config = %d %+v", configured.Code, payload)
	}

	stream := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/backend/api/v1/admin/logs/stream?source=server&dashboardToken=secret-token", nil)
	handler.ServeHTTP(stream, request)
	if stream.Code != http.StatusOK || proxiedPath != "/api/v1/admin/logs/stream" || proxiedQuery != "source=server" {
		t.Fatalf("proxy = %d path=%q query=%q", stream.Code, proxiedPath, proxiedQuery)
	}
}

func TestDockerDashboardURLKeepsTokenInFragment(t *testing.T) {
	got, err := dockerDashboardURL("http://127.0.0.1:18081/", "secret token")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "token=secret+token" {
		t.Fatalf("dashboard URL = %q", got)
	}
}
