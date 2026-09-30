package vergeos

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const sampleVolumeKey = "c3437883534918dcf2abbb3e9b9622b865226c68"

func TestVMExportUnmarshal(t *testing.T) {
	var row VMExport
	raw := `{"$key":"2","volume":null,"quiesced":"1","status":"idle","max_exports":"5","create_current":false}`
	if err := json.Unmarshal([]byte(raw), &row); err != nil {
		t.Fatal(err)
	}
	if row.Key.Int() != 2 || row.Volume != "" || !row.Quiesced || row.CreateCurrent || row.MaxExports != 5 {
		t.Fatalf("export = %+v", row)
	}

	var stat VMExportStat
	if err := json.Unmarshal([]byte(`{"$key":1,"errors":null,"virtual_machines":2,"file_name":"nightly","timestamp":"20"}`), &stat); err != nil {
		t.Fatal(err)
	}
	if stat.Errors != 0 || stat.VirtualMachines != 2 || stat.Timestamp != 20 || stat.FileName != "nightly" {
		t.Fatalf("stat = %+v", stat)
	}
}

func TestVMExportServiceCRUD(t *testing.T) {
	var created map[string]any
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_vm_exports": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("fields"); got == "" || got == "most" {
				t.Errorf("fields = %q", got)
			}
			filter := r.URL.Query().Get("filter")
			switch {
			case filter == "":
				jsonResponse(w, 200, []VMExport{{Key: 2, Volume: sampleVolumeKey, VolumeName: "backups", Status: VMExportStatusIdle, MaxExports: 3}})
			case filter == "volume eq '"+sampleVolumeKey+"'":
				jsonResponse(w, 200, []VMExport{{Key: 2, Volume: sampleVolumeKey, VolumeName: "backups", Status: VMExportStatusIdle}})
			case strings.Contains(filter, "missing"):
				jsonResponse(w, 200, []VMExport{})
			case strings.Contains(filter, "dup"):
				jsonResponse(w, 200, []VMExport{{Key: 2, Volume: "dup"}, {Key: 3, Volume: "dup"}})
			default:
				t.Errorf("filter = %q", filter)
				jsonResponse(w, 200, []VMExport{})
			}
		},
		"GET /api/v4/volume_vm_exports/2": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMExport{Key: 2, Volume: sampleVolumeKey, VolumeName: "backups", Status: VMExportStatusIdle, Quiesced: true})
		},
		"POST /api/v4/volume_vm_exports": func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(&created); err != nil {
				t.Fatal(err)
			}
			jsonResponse(w, 200, map[string]any{"$key": 2})
		},
		"PUT /api/v4/volume_vm_exports/2": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
		"DELETE /api/v4/volume_vm_exports/2": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	}))

	rows, err := client.VMExports.List(context.Background())
	if err != nil || len(rows) != 1 || rows[0].VolumeName != "backups" {
		t.Fatalf("list = %+v %v", rows, err)
	}
	got, err := client.VMExports.GetByVolume(context.Background(), sampleVolumeKey)
	if err != nil || got.Key.Int() != 2 || !got.Quiesced {
		t.Fatalf("by volume = %+v %v", got, err)
	}
	_, err = client.VMExports.GetByVolume(context.Background(), "missing")
	if !IsNotFoundError(err) {
		t.Fatalf("missing: %v", err)
	}
	_, err = client.VMExports.GetByVolume(context.Background(), "dup")
	if !IsAmbiguousNameError(err) {
		t.Fatalf("dup: %v", err)
	}

	maxExports := 5
	quiesced := true
	createdRow, err := client.VMExports.Create(context.Background(), &VMExportCreateRequest{
		Volume: sampleVolumeKey, Quiesced: &quiesced, MaxExports: &maxExports,
	})
	if err != nil || createdRow.Key.Int() != 2 {
		t.Fatalf("create = %+v %v", createdRow, err)
	}
	if created["volume"] != sampleVolumeKey || created["quiesced"] != true || created["max_exports"] != float64(5) {
		t.Fatalf("body = %+v", created)
	}
	if _, err := client.VMExports.Update(context.Background(), 2, &VMExportUpdateRequest{MaxExports: &maxExports}); err != nil {
		t.Fatal(err)
	}
	if err := client.VMExports.Delete(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	bad := 0
	if _, err := client.VMExports.Create(context.Background(), &VMExportCreateRequest{Volume: sampleVolumeKey, MaxExports: &bad}); !IsValidationError(err) {
		t.Fatalf("max exports: %v", err)
	}
}

func TestVMExportServiceRunAndActions(t *testing.T) {
	var created atomic.Int32
	var action map[string]any
	quiesced := false
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_vm_exports": func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/4") {
				jsonResponse(w, 200, VMExport{Key: 4, Volume: sampleVolumeKey, Status: VMExportStatusIdle, Quiesced: false, MaxExports: 3})
				return
			}
			jsonResponse(w, 200, []VMExport{})
		},
		"POST /api/v4/volume_vm_exports": func(w http.ResponseWriter, r *http.Request) {
			created.Add(1)
			jsonResponse(w, 200, map[string]any{"$key": 4})
		},
		"GET /api/v4/volume_vm_exports/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMExport{Key: 4, Volume: sampleVolumeKey, Status: VMExportStatusIdle, Quiesced: false, MaxExports: 3})
		},
		"POST /api/v4/volume_vm_export_actions": func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusOK)
		},
	}))

	got, err := client.VMExports.Run(context.Background(), &VMExportRunRequest{
		Volume:   sampleVolumeKey,
		VMs:      []int{7, 8},
		Name:     "nightly",
		Quiesced: &quiesced,
	})
	if err != nil || got.Key.Int() != 4 || created.Load() != 1 {
		t.Fatalf("run = %+v %v created=%d", got, err, created.Load())
	}
	if action["action"] != "start_export" || action["volume_vm_export"] != float64(4) {
		t.Fatalf("action = %+v", action)
	}
	params, _ := action["params"].(map[string]any)
	if params["name"] != "nightly" {
		t.Fatalf("params = %+v", params)
	}
	vms, _ := params["vms"].([]any)
	if len(vms) != 2 || vms[0] != float64(7) {
		t.Fatalf("vms = %#v", params["vms"])
	}

	if err := client.VMExports.Stop(context.Background(), 4); err != nil {
		t.Fatal(err)
	}
	if action["action"] != "stop_export" {
		t.Fatalf("stop = %+v", action)
	}
	if err := client.VMExports.Cleanup(context.Background(), 4); err != nil {
		t.Fatal(err)
	}
	if action["action"] != "cleanup" {
		t.Fatalf("cleanup = %+v", action)
	}
}

func TestVMExportServiceRunUpdatesExisting(t *testing.T) {
	var created atomic.Int32
	var updated atomic.Int32
	wantQuiesced := true
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_vm_exports": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v4/volume_vm_exports/4" {
				row := VMExport{Key: 4, Volume: sampleVolumeKey, Status: VMExportStatusIdle, Quiesced: false, MaxExports: 3}
				if updated.Load() > 0 {
					row.Quiesced = true
					row.MaxExports = 5
				}
				jsonResponse(w, 200, row)
				return
			}
			jsonResponse(w, 200, []VMExport{{Key: 4, Volume: sampleVolumeKey, Quiesced: false, MaxExports: 3}})
		},
		"POST /api/v4/volume_vm_exports": func(w http.ResponseWriter, r *http.Request) {
			created.Add(1)
			http.Error(w, "created", http.StatusInternalServerError)
		},
		"PUT /api/v4/volume_vm_exports/4": func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			updated.Add(1)
			if !strings.Contains(string(body), `"quiesced":true`) || !strings.Contains(string(body), `"max_exports":5`) {
				t.Errorf("update body = %s", body)
			}
			w.WriteHeader(http.StatusOK)
		},
		"POST /api/v4/volume_vm_export_actions": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	}))

	maxExports := 5
	got, err := client.VMExports.Run(context.Background(), &VMExportRunRequest{
		Volume: sampleVolumeKey, Quiesced: &wantQuiesced, MaxExports: &maxExports, VMs: []int{7},
	})
	if err != nil || got.Key.Int() != 4 || created.Load() != 0 || updated.Load() != 1 {
		t.Fatalf("run = %+v %v created=%d updated=%d", got, err, created.Load(), updated.Load())
	}
}

func TestVMExportServiceWait(t *testing.T) {
	var reads atomic.Int32
	status := VMExportStatusError
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_vm_exports/4": func(w http.ResponseWriter, r *http.Request) {
			n := reads.Add(1)
			current := status
			if current == VMExportStatusBuilding && n > 1 {
				current = VMExportStatusIdle
			}
			jsonResponse(w, 200, VMExport{Key: 4, Volume: sampleVolumeKey, Status: current, StatusInfo: "disk full"})
		},
		"GET /api/v4/volume_vm_export_stats": func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Query().Get("filter"), "volume_vm_exports eq 4") {
				t.Errorf("filter = %q", r.URL.Query().Get("filter"))
			}
			jsonResponse(w, 200, []VMExportStat{
				{Key: 1, Errors: 0, Timestamp: 10, FileName: "old"},
				{Key: 2, Errors: 2, VirtualMachines: 3, Timestamp: 20, FileName: "nightly"},
			})
		},
	}))

	start := time.Now()
	_, err := client.VMExports.Wait(context.Background(), 4, &VMExportWaitOptions{Timeout: 30 * time.Second})
	if !IsVMExportFailed(err) || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("error status: %v", err)
	}
	if reads.Load() != 1 || time.Since(start) > time.Second {
		t.Fatalf("reads=%d elapsed=%s", reads.Load(), time.Since(start))
	}

	status = VMExportStatusIdle
	reads.Store(0)
	_, err = client.VMExports.Wait(context.Background(), 4, nil)
	if !IsVMExportFailed(err) || !strings.Contains(err.Error(), "2 errors") || !strings.Contains(err.Error(), "nightly") {
		t.Fatalf("stat errors: %v", err)
	}

	status = VMExportStatusBuilding
	reads.Store(0)
	// The stats handler still reports errors. Swap it by using a status path that
	// returns idle on the second read, then replace stats via a fresh client below.
	_, err = client.VMExports.Wait(context.Background(), 4, &VMExportWaitOptions{Timeout: time.Second, PollInterval: 5 * time.Millisecond})
	if !IsVMExportFailed(err) {
		t.Fatalf("building then errors: %v", err)
	}
	if reads.Load() < 2 {
		t.Fatalf("reads = %d", reads.Load())
	}
}

func TestVMExportServiceWaitIdleWithoutErrors(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/volume_vm_exports/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMExport{Key: 4, Status: VMExportStatusIdle})
		},
		"GET /api/v4/volume_vm_export_stats": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMExportStat{{Key: 1, Errors: 0, ExportSuccess: 1, Timestamp: 5}})
		},
	}))
	got, err := client.VMExports.Wait(context.Background(), 4, nil)
	if err != nil || got.Status != VMExportStatusIdle {
		t.Fatalf("idle = %+v %v", got, err)
	}
	stats, err := client.VMExports.Stats(context.Background(), 4)
	if err != nil || len(stats) != 1 || stats[0].ExportSuccess != 1 {
		t.Fatalf("stats = %+v %v", stats, err)
	}
}
