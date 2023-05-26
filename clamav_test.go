package clamav

import (
	"context"
	"os"
	"testing"
)

func TestPing(t *testing.T) {
	clamav := NewClamAV(WithUnix())
	got, err := clamav.Ping(context.Background())
	if err != nil {
		t.Errorf("%v", err)
	}
	if !got {
		t.Errorf("clamav.Ping() = %v; want true", got)
	}
}

func TestVersion(t *testing.T) {
	clamav := NewClamAV(WithUnix())
	got, err := clamav.Version(context.Background())
	if err != nil {
		t.Errorf("%v", err)
	}
	if got == "" {
		t.Errorf("clamav.Version() = %s; want version string", got)
	}
}

func TestReload(t *testing.T) {
	clamav := NewClamAV(WithUnix())
	got, err := clamav.Reload(context.Background())
	if err != nil {
		t.Errorf("%v", err)
	}
	if !got {
		t.Errorf("clamav.Reload() = %v; want true", got)
	}
}

func TestScan(t *testing.T) {
	f, err := os.CreateTemp("", "go-clamav-test")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.Remove(f.Name())
	_, err = f.WriteString("this is a test file for go-clamav")
	if err != nil {
		t.Errorf("%v", err)
	}

	clamav := NewClamAV(WithUnix())
	got, err := clamav.Scan(context.Background(), f.Name())
	if err != nil {
		t.Errorf("%v", err)
	}
	if !got {
		t.Errorf("clamav.Scan() = %v; want true", got)
	}
}

func TestScanAll(t *testing.T) {
	clamav := NewClamAV(WithUnix())
	got, err := clamav.ScanAll(context.Background(), "/tmp")
	if err != nil {
		t.Errorf("%v", err)
	}
	if !got {
		t.Errorf("clamav.Scan() = %v; want true", got)
	}
}
