package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"surf-backend/internal/config"
	"surf-backend/internal/control"
	"surf-backend/internal/web"
)

const dashboardTokenHeader = "X-Surf-Dashboard-Token"

func dockerDashboardHandler(cfg *config.Config, token string, backend http.Handler) http.Handler {
	root := bytes.Replace(managementHTML, []byte("<body>"), []byte(`<body data-runtime="docker">`), 1)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; connect-src 'self'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(root)
			return
		case "/icon.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(surfIconPNG)
			return
		}
		if !dashboardAuthorized(r, token) {
			http.Error(w, "dashboard authorization required; run make dashboard", http.StatusForbidden)
			return
		}
		if r.URL.Path == "/api/config" && r.Method == http.MethodGet {
			endpoint := strings.TrimSpace(cfg.PublicAddress)
			if endpoint != "" && !strings.Contains(endpoint, "://") {
				endpoint = "https://" + endpoint
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"runtime": "docker", "serverName": cfg.ServerName,
				"publicAddress": cfg.PublicAddress, "baseURL": endpoint,
				"lanURL": endpoint, "port": cfg.Port, "running": true,
				"startAtLogin": false, "startupChoiceMade": true,
			})
			return
		}
		request := r.Clone(r.Context())
		request.URL = cloneURL(r.URL)
		query := request.URL.Query()
		query.Del("dashboardToken")
		request.URL.RawQuery = query.Encode()
		switch {
		case r.URL.Path == "/api/status" && r.Method == http.MethodGet:
			request.URL.Path = web.APIRoot + "/health"
			request.URL.RawQuery = "stats=1"
		case strings.HasPrefix(r.URL.Path, "/api/backend/"):
			request.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/backend")
		default:
			http.NotFound(w, r)
			return
		}
		backend.ServeHTTP(w, request)
	})
}

func cloneURL(source *url.URL) *url.URL {
	clone := *source
	return &clone
}

func dashboardAuthorized(r *http.Request, token string) bool {
	provided := r.Header.Get(dashboardTokenHeader)
	if provided == "" && r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/admin/logs/stream") {
		provided = r.URL.Query().Get("dashboardToken")
	}
	return token != "" && len(provided) == len(token) &&
		subtle.ConstantTimeCompare([]byte(provided), []byte(token)) == 1
}

func dockerDashboardProxy(home string) http.Handler {
	return &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Scheme = "https"
			request.URL.Host = "surf-control"
			request.Host = "surf-control"
		},
		Transport: &serverControlTransport{home: home},
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(w, "Surf backend is unavailable", http.StatusBadGateway)
		},
	}
}

func startDockerDashboard(ctx context.Context, cfg *config.Config, token string) (<-chan error, error) {
	listener, err := net.Listen("tcp", cfg.DashboardAddr)
	if err != nil {
		return nil, fmt.Errorf("dashboard listen on %s: %w", cfg.DashboardAddr, err)
	}
	home, err := surfHome()
	if err != nil {
		_ = listener.Close()
		return nil, err
	}
	server := &http.Server{
		Handler:           dockerDashboardHandler(cfg, token, dockerDashboardProxy(home)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	done := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		done <- err
	}()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	return done, nil
}

func dockerDashboardURL(base, token string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(base))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil {
		return "", fmt.Errorf("invalid SURF_DASHBOARD_PUBLIC_URL")
	}
	parsed.RawQuery = ""
	parsed.Fragment = url.Values{"token": {token}}.Encode()
	return parsed.String(), nil
}

func runDashboardURLCommand() error {
	home, err := surfHome()
	if err != nil {
		return err
	}
	descriptor, err := control.Load(home)
	if err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.DashboardPublicURL == "" {
		return fmt.Errorf("Docker dashboard is disabled")
	}
	value, err := dockerDashboardURL(cfg.DashboardPublicURL, descriptor.AdminToken)
	if err != nil {
		return err
	}
	fmt.Println(value)
	return nil
}
