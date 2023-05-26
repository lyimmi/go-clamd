package clamd

import (
	"context"
	"os"
	"testing"
)

func TestPing(t *testing.T) {
	clamd := NewClamd()
	got, err := clamd.Ping(context.Background())
	if err != nil {
		t.Errorf("%v", err)
	}
	if !got {
		t.Errorf("clamd.Ping() = %v; want true", got)
	}
}

func TestVersion(t *testing.T) {
	clamd := NewClamd()
	got, err := clamd.Version(context.Background())
	if err != nil {
		t.Errorf("%v", err)
	}
	if got == "" {
		t.Errorf("clamd.Version() = %s; want version string", got)
	}
}

func TestReload(t *testing.T) {
	clamd := NewClamd()
	got, err := clamd.Reload(context.Background())
	if err != nil {
		t.Errorf("%v", err)
	}
	if !got {
		t.Errorf("clamd.Reload() = %v; want true", got)
	}
}

func TestScan(t *testing.T) {
	f, err := os.CreateTemp("", "go-clamd-test")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.Remove(f.Name())
	_, err = f.WriteString("this is a test file for go-clamd")
	if err != nil {
		t.Errorf("%v", err)
	}

	clamd := NewClamd()
	got, err := clamd.Scan(context.Background(), f.Name())
	if err != nil {
		t.Errorf("%v", err)
	}
	if !got {
		t.Errorf("clamd.Scan() = %v; want true", got)
	}
}

func TestStream(t *testing.T) {
	f, err := os.CreateTemp("", "go-clamd-test-stream")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer os.Remove(f.Name())

	for i := 0; i < 10; i++ {
		_, err = f.WriteString("this is a test file for go-clamd\n")
		if err != nil {
			t.Errorf("%v", err)
		}
	}

	clamd := NewClamd()
	got, err := clamd.ScanStream(context.Background(), f)
	if err != nil {
		t.Errorf("%v", err)
	}
	if !got {
		t.Errorf("clamd.Scan() = %v; want true", got)
	}
}

func TestScanAll(t *testing.T) {
	clamd := NewClamd()
	got, err := clamd.ScanAll(context.Background(), "/tmp")
	if err != nil {
		t.Errorf("%v", err)
	}
	if !got {
		t.Errorf("clamd.ScanAll() = %v; want true", got)
	}
}

//func TestShutdown(t *testing.T) {
//	clamd := NewClamd()
//	got, err := clamd.Shutdown(context.Background())
//	if err != nil {
//		t.Errorf("%v", err)
//	}
//	if !got {
//		t.Errorf("clamd.Shutdown() = %v; want true", got)
//	}
//}
