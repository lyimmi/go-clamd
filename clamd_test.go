package clamd

import (
	"context"
	"os"
	"testing"
)

var eicar = []byte(`X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`)

func writeTestFile(t testing.TB) string {
	f, err := os.CreateTemp("", "go-clamd-test-stream")
	if err != nil {
		t.Fatalf("%v", err)
	}

	_, err = f.Write(eicar)
	if err != nil {
		t.Errorf("%v", err)
	}
	return f.Name()
}

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
	clamd := NewClamd()

	tf := writeTestFile(t)
	defer os.Remove(tf)

	got, err := clamd.Scan(context.Background(), tf)
	if err != nil {
		t.Errorf("%v", err)
	}
	if got {
		t.Errorf("clamd.Scan() = %v; want false", got)
	}
}

func TestStream(t *testing.T) {
	clamd := NewClamd()

	tf := writeTestFile(t)
	defer os.Remove(tf)

	f, err := os.Open(tf)
	if err != nil {
		t.Fatal(err)
	}

	got, err := clamd.ScanStream(context.Background(), f)
	if err != nil {
		t.Errorf("%v", err)
	}
	if got {
		t.Errorf("clamd.Scan() = %v; want false", got)
	}
}

func TestScanAll(t *testing.T) {
	clamd := NewClamd()

	tf := writeTestFile(t)
	defer os.Remove(tf)

	got, err := clamd.ScanAll(context.Background(), "/tmp")
	if err != nil {
		t.Errorf("%v", err)
	}
	if got {
		t.Errorf("clamd.Scan() = %v; want false", got)
	}
}

//func TestShutdown(t *testing.T) {
//	clamd, teardown := setupTest(t)
//	defer teardown(t)
//	got, err := clamd.Shutdown(context.Background())
//	if err != nil {
//		t.Errorf("%v", err)
//	}
//	if !got {
//		t.Errorf("clamd.Shutdown() = %v; want true", got)
//	}
//}

func TestStats(t *testing.T) {
	clamd := NewClamd()

	got, err := clamd.Stats(context.Background())
	if err != nil {
		t.Errorf("%v", err)
	}
	if got == nil {
		t.Errorf("clamd.ScanAll() = %v; want Stats", got)
	}
}
