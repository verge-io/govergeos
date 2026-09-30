//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	vergeos "github.com/verge-io/govergeos"
)

// TestVMImportsAndExports lists VM imports, the logs of the first import,
// and VM export configurations. It does not create, start, or delete anything.
func TestVMImportsAndExports(t *testing.T) {
	client := setupTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	imports, err := client.VMImports.List(ctx, vergeos.WithLimit(5))
	if err != nil {
		t.Fatalf("Failed to list VM imports: %v", err)
	}
	t.Logf("Found %d VM import(s) in the first page", len(imports))
	for _, imp := range imports {
		t.Logf("Import: Key=%s Name=%q Status=%s FailedDrives=%d", imp.Key, imp.Name, imp.Status, imp.FailedDriveCount)
	}
	if len(imports) > 0 && imports[0].Key != "" {
		got, err := client.VMImports.Get(ctx, imports[0].Key)
		if err != nil {
			t.Fatalf("Failed to get import %s: %v", imports[0].Key, err)
		}
		if got.Name != imports[0].Name {
			t.Fatalf("Get name = %q, list name = %q", got.Name, imports[0].Name)
		}

		logs, err := client.VMImports.Logs(ctx, imports[0].Key, vergeos.WithLimit(5))
		if err != nil {
			t.Fatalf("Failed to list logs for %s: %v", imports[0].Key, err)
		}
		t.Logf("Found %d log line(s) on %s", len(logs), imports[0].Name)
		for _, line := range logs {
			// Level only. The text can carry a download URL.
			t.Logf("Log: Key=%d Level=%s", line.Key.Int(), line.Level)
		}
	}

	exports, err := client.VMExports.List(ctx, vergeos.WithLimit(5))
	if err != nil {
		t.Fatalf("Failed to list VM exports: %v", err)
	}
	t.Logf("Found %d VM export(s) in the first page", len(exports))
	for _, exp := range exports {
		t.Logf("Export: Key=%d Volume=%q Status=%s MaxExports=%d", exp.Key.Int(), exp.VolumeName, exp.Status, exp.MaxExports)
	}
	if len(exports) == 0 || exports[0].Key.Int() == 0 {
		return
	}
	got, err := client.VMExports.Get(ctx, exports[0].Key.Int())
	if err != nil {
		t.Fatalf("Failed to get export %d: %v", exports[0].Key.Int(), err)
	}
	if got.Volume != exports[0].Volume {
		t.Fatalf("Get volume = %q, list volume = %q", got.Volume, exports[0].Volume)
	}
	stats, err := client.VMExports.Stats(ctx, got.Key.Int(), vergeos.WithLimit(5))
	if err != nil {
		t.Fatalf("Failed to list export stats: %v", err)
	}
	t.Logf("Found %d export stat row(s)", len(stats))
}
