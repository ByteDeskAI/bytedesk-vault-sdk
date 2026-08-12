package vaultsdk

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidateAndPackVaultTarget(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(`{
		"id":"vaultplug","version":"0.1.0","spawn":true,"binary":"vaultplug",
		"targets":["vault"],"minCoreVersion":"0.1.0"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "vaultplug"), []byte("#!/bin/true\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateDir(dir); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	res, err := PackDir(dir, out)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(res.Archive); err != nil {
		t.Fatal(err)
	}
	if res.ID != "vaultplug" || !res.Unsigned {
		t.Fatalf("got %+v", res)
	}
}

func TestValidateRejectsGatewayOnly(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(`{"id":"g","version":"1.0.0","targets":["gateway"]}`), 0o644)
	if _, err := ValidateDir(dir); err == nil {
		t.Fatal("expected gateway-only plugin rejected on vault SDK")
	}
}

func TestValidateRejectsLegacyEmptyTargets(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(`{"id":"legacy","version":"1.0.0"}`), 0o644)
	if _, err := ValidateDir(dir); err == nil {
		t.Fatal("empty targets default to gateway; vault SDK must reject")
	}
}

func TestValidateAcceptsBothTargets(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(`{"id":"both","version":"1.0.0","targets":["gateway","vault"]}`), 0o644)
	m, err := ValidateDir(dir)
	if err != nil || m.ID != "both" {
		t.Fatalf("got %+v err=%v", m, err)
	}
}

func TestServeHealthz(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "plugin.sock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- ServeContext(ctx, Config{Socket: sock, ID: "vaultplug"})
	}()
	var conn net.Conn
	var err error
	for i := 0; i < 50; i++ {
		conn, err = net.Dial("unix", sock)
		if err == nil {
			_ = conn.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", sock)
		},
	}}
	resp, err := client.Get("http://plugin/healthz")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || string(body) != "ok" {
		t.Fatalf("status=%d body=%q", resp.StatusCode, body)
	}
	cancel()
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("serve did not exit")
	}
}
