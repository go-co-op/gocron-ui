package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-co-op/gocron-ui/server"
	"github.com/go-co-op/gocron/v2"
)

func TestServerWithGorillaDefaults(t *testing.T) {
	t.Run("get index page", func(t *testing.T) {
		res := sendGetRequest(t, "/", "")

		expectStatusCode(t, res, http.StatusOK)
		expectHeader(t, res, "Content-Type", "text/html; charset=utf-8")
	})

	t.Run("get api config", func(t *testing.T) {
		res := sendGetRequest(t, "/api/config", "")

		expectStatusCode(t, res, http.StatusOK)
		expectHeader(t, res, "Content-Type", "application/json")
		expectBody(t, res, `{"title":"GoCron UI","api_enabled":true,"websocket_enabled":true}`)
	})
}

func TestServerWithGorillaDefaultsBehindHTTPDefaultMux(t *testing.T) {
	t.Run("get index page", func(t *testing.T) {
		res := sendGetRequest(t, "/", "/")

		expectStatusCode(t, res, http.StatusOK)
		expectHeader(t, res, "Content-Type", "text/html; charset=utf-8")
	})

	t.Run("get api config", func(t *testing.T) {
		res := sendGetRequest(t, "/api/config", "/")

		expectStatusCode(t, res, http.StatusOK)
		expectHeader(t, res, "Content-Type", "application/json")
		expectBody(t, res, `{"title":"GoCron UI","api_enabled":true,"websocket_enabled":true}`)
	})
}

func TestServerBehindPathMux(t *testing.T) {
	t.Run("get index page", func(t *testing.T) {
		res := sendGetRequest(t, "/admin/", "/admin/", server.WithBasePath("/admin/"))

		expectStatusCode(t, res, http.StatusOK)
		expectHeader(t, res, "Content-Type", "text/html; charset=utf-8")
	})

	t.Run("get api config", func(t *testing.T) {
		res := sendGetRequest(t, "/admin/api/config", "/admin/", server.WithBasePath("/admin/"))

		expectStatusCode(t, res, http.StatusOK)
		expectHeader(t, res, "Content-Type", "application/json")
		expectBody(t, res, `{"title":"GoCron UI","api_enabled":true,"websocket_enabled":true}`)
	})
}

// normal usage using gorilla

func sendGetRequest(t *testing.T, reqPath, muxPath string, opts ...server.Option) *http.Response {
	t.Helper()

	scheduler, _ := gocron.NewScheduler()
	srv := server.NewServer(scheduler, 8080, opts...)
	if muxPath != "" {
		mux := http.NewServeMux()
		mux.Handle(muxPath, srv.Router)
		return sendGetRequestToServer(t, reqPath, mux)
	}

	return sendGetRequestToServer(t, reqPath, srv.Router)
}

func sendGetRequestToServer(t *testing.T, path string, srv http.Handler) *http.Response {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), "GET", "http://example.com"+path, nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)
	return w.Result()
}

func expectStatusCode(t *testing.T, res *http.Response, expectedStatusCode int) {
	t.Helper()
	if res.StatusCode != expectedStatusCode {
		t.Fatalf("expected status code %d; got %d", expectedStatusCode, res.StatusCode)
	}
}

func expectHeader(t *testing.T, res *http.Response, key, expectedValue string) {
	t.Helper()
	actualValue := res.Header.Get(key)
	if actualValue != expectedValue {
		t.Fatalf("expected header %q to be %q; got %q", key, expectedValue, actualValue)
	}
}

func expectBody(t *testing.T, res *http.Response, expectedBody string) {
	t.Helper()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading response body: %v", err)
	}
	res.Body.Close()

	actualBody := strings.TrimSpace(string(data))
	if actualBody != expectedBody {
		t.Fatalf("expected body %q; got %q", expectedBody, actualBody)
	}
}
