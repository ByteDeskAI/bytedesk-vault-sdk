package vaultsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPasswordLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/begin":
			_, _ = w.Write([]byte(`{"challengeId":"c1"}`))
		case "/v1/auth/finish":
			_, _ = w.Write([]byte(`{"assertion":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := &Client{BaseURL: srv.URL, GatewayID: "gw1", HTTP: srv.Client()}
	ok, user, err := c.PasswordLogin(context.Background(), "Ryan", "pw")
	if err != nil || !ok || user != "ryan" {
		t.Fatalf("ok=%v user=%q err=%v", ok, user, err)
	}
}

func TestConfigured(t *testing.T) {
	if (&Client{}).Configured() {
		t.Fatal("empty should be false")
	}
}
