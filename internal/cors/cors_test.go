package cors

import (
	"net/http"
	"testing"
)

func TestParseOrigins_Default(t *testing.T) {
	got := ParseOrigins("")
	if len(got) != 2 {
		t.Fatalf("expected 2 defaults, got %v", got)
	}
}

func TestAllows(t *testing.T) {
	cfg := Config{Origins: []string{"http://localhost:4200"}}
	if !cfg.Allows("http://localhost:4200") {
		t.Fatal("expected allow")
	}
	if cfg.Allows("http://evil.example") {
		t.Fatal("expected deny")
	}
	if !(Config{Origins: []string{"*"}}).AllowAll() {
		t.Fatal("expected allow all")
	}
}

func TestOriginPatterns(t *testing.T) {
	cfg := Config{Origins: []string{"http://localhost:4200", "https://mateusfmfm.github.io"}}
	p := cfg.OriginPatterns()
	if len(p) != 2 || p[0] != "localhost:4200" || p[1] != "mateusfmfm.github.io" {
		t.Fatalf("unexpected patterns: %v", p)
	}
}

func TestMiddleware_Preflight(t *testing.T) {
	h := Middleware(Config{Origins: []string{"http://localhost:4200"}}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req, _ := http.NewRequest(http.MethodOptions, "/query", nil)
	req.Header.Set("Origin", "http://localhost:4200")
	rr := newRecorder()
	h.ServeHTTP(rr, req)
	if rr.code != http.StatusNoContent {
		t.Fatalf("status %d", rr.code)
	}
	if rr.header.Get("Access-Control-Allow-Origin") != "http://localhost:4200" {
		t.Fatalf("missing CORS header: %v", rr.header)
	}
}

type recorder struct {
	code   int
	header http.Header
}

func newRecorder() *recorder {
	return &recorder{header: make(http.Header), code: 200}
}

func (r *recorder) Header() http.Header         { return r.header }
func (r *recorder) Write(b []byte) (int, error) { return len(b), nil }
func (r *recorder) WriteHeader(statusCode int)  { r.code = statusCode }
