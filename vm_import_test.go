package vergeos

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const sampleImportKey = "8f73f8bcc9c9f1aaba32f733bfc295acaf548554"

func TestVMImportUnmarshalNullsAndStrings(t *testing.T) {
	var row VMImport
	raw := `{
		"$key":"abc","id":"abc","name":"imported-vm","vm":null,"file":"12",
		"volume":null,"importing":"true","aborted":0,"failed_drive_count":null,
		"preserve_macs":1,"status":"importing"
	}`
	if err := json.Unmarshal([]byte(raw), &row); err != nil {
		t.Fatal(err)
	}
	if row.Key != "abc" || row.VM != nil || row.Volume != "" || !row.Importing || row.Aborted {
		t.Fatalf("row = %+v", row)
	}
	if row.File == nil || row.File.Int() != 12 || !row.PreserveMACs || row.FailedDriveCount != 0 {
		t.Fatalf("row = %+v", row)
	}
}

func TestVMImportLogUnmarshalUserNumber(t *testing.T) {
	var row VMImportLog
	if err := json.Unmarshal([]byte(`{"$key":"4","vm_import":"abc","level":"error","text":"disk failed","user":7}`), &row); err != nil {
		t.Fatal(err)
	}
	if row.Key.Int() != 4 || row.User != "7" || row.Text != "disk failed" {
		t.Fatalf("log = %+v", row)
	}
}

func TestVMImportServiceListGetAndName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("fields"); got == "" || got == "most" {
				t.Errorf("fields = %q", got)
			}
			switch r.URL.Query().Get("filter") {
			case "":
				jsonResponse(w, 200, []VMImport{{Key: sampleImportKey, ID: sampleImportKey, Name: "imported-vm", Status: VMImportStatusComplete}})
			case "id eq '" + sampleImportKey + "'":
				jsonResponse(w, 200, []VMImport{{Key: sampleImportKey, ID: sampleImportKey, Name: "imported-vm", Status: VMImportStatusComplete}})
			case "name eq 'imported-vm'":
				jsonResponse(w, 200, []VMImport{{Key: sampleImportKey, ID: sampleImportKey, Name: "imported-vm"}})
			case "name eq 'O\\'Brien'":
				jsonResponse(w, 200, []VMImport{})
			case "name eq 'dup'":
				jsonResponse(w, 200, []VMImport{
					{Key: "a", Name: "dup"},
					{Key: "b", Name: "dup"},
				})
			default:
				t.Errorf("filter = %q", r.URL.Query().Get("filter"))
				jsonResponse(w, 200, []VMImport{})
			}
		},
	}))

	rows, err := client.VMImports.List(context.Background())
	if err != nil || len(rows) != 1 || rows[0].Name != "imported-vm" {
		t.Fatalf("list = %+v %v", rows, err)
	}
	got, err := client.VMImports.Get(context.Background(), sampleImportKey)
	if err != nil || got.Status != VMImportStatusComplete {
		t.Fatalf("get = %+v %v", got, err)
	}
	byName, err := client.VMImports.GetByName(context.Background(), "imported-vm")
	if err != nil || byName.Key != sampleImportKey {
		t.Fatalf("by name = %+v %v", byName, err)
	}
	_, err = client.VMImports.GetByName(context.Background(), "dup")
	if !IsAmbiguousNameError(err) {
		t.Fatalf("ambiguous: %v", err)
	}
	_, err = client.VMImports.GetByName(context.Background(), "O'Brien")
	if !IsNotFoundError(err) {
		t.Fatalf("missing escaped name: %v", err)
	}
	_, err = client.VMImports.Get(context.Background(), "")
	if !IsValidationError(err) {
		t.Fatalf("empty id: %v", err)
	}
}

func TestVMImportServiceCreateFromFile(t *testing.T) {
	fileID := 41
	preserve := false
	var posted map[string]any
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VM{})
		},
		"POST /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatal(err)
			}
			jsonResponse(w, 200, map[string]any{"$key": sampleImportKey})
		},
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMImport{{
				Key: sampleImportKey, ID: sampleImportKey, Name: "imported-vm", Status: VMImportStatusImporting,
			}})
		},
	}))

	got, err := client.VMImports.Create(context.Background(), &VMImportCreateRequest{
		Name:          "imported-vm",
		File:          &fileID,
		PreserveMACs:  &preserve,
		PreferredTier: "3",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Key != sampleImportKey || got.AlreadyExisted {
		t.Fatalf("created = %+v", got)
	}
	if posted["name"] != "imported-vm" || posted["importing"] != true || posted["preserve_macs"] != false || posted["preferred_tier"] != "3" {
		t.Fatalf("body = %+v", posted)
	}
	if posted["file"] != float64(41) {
		t.Fatalf("file = %#v", posted["file"])
	}
}

func TestVMImportServiceCreateFromURL(t *testing.T) {
	var fileBody map[string]any
	var importBody map[string]any
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VM{})
		},
		"POST /api/v4/files": func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(&fileBody); err != nil {
				t.Fatal(err)
			}
			jsonResponse(w, 200, map[string]any{"$key": 9})
		},
		"GET /api/v4/files": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, File{Key: 9, Name: "disk.qcow2"})
		},
		"POST /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(&importBody); err != nil {
				t.Fatal(err)
			}
			jsonResponse(w, 200, map[string]any{"$key": sampleImportKey})
		},
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMImport{{Key: sampleImportKey, ID: sampleImportKey, Name: "cloud-image", Status: VMImportStatusImporting}})
		},
	}))

	got, err := client.VMImports.Create(context.Background(), &VMImportCreateRequest{
		Name: "cloud-image",
		URL:  "https://example.com/images/disk.qcow2?token=1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "cloud-image" {
		t.Fatalf("created = %+v", got)
	}
	if fileBody["name"] != "disk.qcow2" || fileBody["url"] != "https://example.com/images/disk.qcow2?token=1" {
		t.Fatalf("file body = %+v", fileBody)
	}
	if importBody["file"] != float64(9) || importBody["importing"] != true {
		t.Fatalf("import body = %+v", importBody)
	}
}

func TestVMImportServiceCreateValidation(t *testing.T) {
	var hits atomic.Int32
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.NotFound(w, r)
	}))
	fileID := 1
	_, err := client.VMImports.Create(context.Background(), &VMImportCreateRequest{Name: "n", File: &fileID, URL: "https://example.com/a.ova"})
	if !IsValidationError(err) {
		t.Fatalf("two sources: %v", err)
	}
	_, err = client.VMImports.Create(context.Background(), &VMImportCreateRequest{Name: "n", URL: "ftp://example.com/a.ova"})
	if !IsValidationError(err) {
		t.Fatalf("scheme: %v", err)
	}
	_, err = client.VMImports.Create(context.Background(), &VMImportCreateRequest{Name: "n", URL: "https://example.com/images/"})
	if !IsValidationError(err) {
		t.Fatalf("directory url: %v", err)
	}
	_, err = client.VMImports.Create(context.Background(), nil)
	if !IsValidationError(err) {
		t.Fatalf("nil: %v", err)
	}
	if hits.Load() != 0 {
		t.Fatalf("validation made %d requests", hits.Load())
	}
}

func TestVMImportServiceCreateExistingVM(t *testing.T) {
	var posts atomic.Int32
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/vms/") {
				jsonResponse(w, 200, VM{Key: 7, Name: "imported-vm"})
				return
			}
			jsonResponse(w, 200, []VM{{Key: 7, Name: "imported-vm"}})
		},
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if strings.Contains(filter, "name eq") {
				jsonResponse(w, 200, []VMImport{{Key: sampleImportKey, ID: sampleImportKey, Name: "imported-vm", Status: VMImportStatusComplete}})
				return
			}
			jsonResponse(w, 200, []VMImport{{Key: sampleImportKey, ID: sampleImportKey, Name: "imported-vm", Status: VMImportStatusComplete, VM: flexPtr(7)}})
		},
		"POST /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			posts.Add(1)
			http.Error(w, "posted", http.StatusInternalServerError)
		},
	}))

	got, err := client.VMImports.Create(context.Background(), &VMImportCreateRequest{Name: "imported-vm", URL: "https://example.com/a.ova"})
	if err != nil {
		t.Fatal(err)
	}
	if !got.AlreadyExisted || got.Key != sampleImportKey || posts.Load() != 0 {
		t.Fatalf("existing = %+v posts=%d", got, posts.Load())
	}
}

func TestVMImportServiceCreateDuplicateName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/vms/") {
				jsonResponse(w, 200, VM{Key: 7, Name: "qa-p-dup"})
				return
			}
			jsonResponse(w, 200, []VM{{Key: 7, Name: "qa-p-dup"}})
		},
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMImport{
				{Key: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "qa-p-dup"},
				{Key: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Name: "qa-p-dup"},
			})
		},
	}))

	got, err := client.VMImports.Create(context.Background(), &VMImportCreateRequest{Name: "qa-p-dup", File: intPtr(3)})
	if err != nil {
		t.Fatal(err)
	}
	if !got.AlreadyExisted || got.Key != "" || got.VM == nil || got.VM.Int() != 7 {
		t.Fatalf("duplicate = %+v", got)
	}
}

func TestVMImportServiceCreateConflictBecomesExisting(t *testing.T) {
	var vmLists atomic.Int32
	var posts atomic.Int32
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/vms/") {
				jsonResponse(w, 200, VM{Key: 7, Name: "imported-vm"})
				return
			}
			if vmLists.Add(1) == 1 {
				jsonResponse(w, 200, []VM{})
				return
			}
			jsonResponse(w, 200, []VM{{Key: 7, Name: "imported-vm"}})
		},
		"POST /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			posts.Add(1)
			w.WriteHeader(http.StatusConflict)
			_, _ = io.WriteString(w, `{"err":"This name is already in use"}`)
		},
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMImport{{Key: sampleImportKey, ID: sampleImportKey, Name: "imported-vm", Status: VMImportStatusComplete}})
		},
	}))

	got, err := client.VMImports.Create(context.Background(), &VMImportCreateRequest{Name: "imported-vm", File: intPtr(4)})
	if err != nil {
		t.Fatal(err)
	}
	if !got.AlreadyExisted || got.Key != sampleImportKey || posts.Load() != 1 {
		t.Fatalf("conflict = %+v posts=%d", got, posts.Load())
	}
}

func TestVMImportServiceDeleteByName(t *testing.T) {
	var deleted atomic.Int32
	rows := []VMImport{
		{Key: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "qa-p-dup", Status: VMImportStatusComplete},
		{Key: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Name: "qa-p-dup", Status: VMImportStatusError},
	}
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, rows)
		},
		"DELETE /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			deleted.Add(1)
			w.WriteHeader(http.StatusOK)
		},
	}))

	if err := client.VMImports.DeleteByName(context.Background(), "qa-p-dup"); err != nil {
		t.Fatal(err)
	}
	if deleted.Load() != 2 {
		t.Fatalf("deleted %d", deleted.Load())
	}

	rows = []VMImport{
		{Key: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "qa-p-dup", Status: VMImportStatusComplete},
		{Key: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Name: "qa-p-dup", Status: VMImportStatusImporting, Importing: true},
	}
	deleted.Store(0)
	err := client.VMImports.DeleteByName(context.Background(), "qa-p-dup")
	if !IsVMImportInProgress(err) || deleted.Load() != 0 {
		t.Fatalf("in progress: %v deleted=%d", err, deleted.Load())
	}

	rows = []VMImport{{Key: sampleImportKey, Name: "live", Status: VMImportStatusImporting, Importing: true}}
	if err := client.VMImports.DeleteByName(context.Background(), "live"); err != nil {
		t.Fatal(err)
	}
	if deleted.Load() != 1 {
		t.Fatalf("single live deleted %d", deleted.Load())
	}

	rows = nil
	deleted.Store(0)
	if err := client.VMImports.DeleteByName(context.Background(), "live"); err != nil {
		t.Fatal(err)
	}
	if deleted.Load() != 0 {
		t.Fatalf("missing deleted %d", deleted.Load())
	}
}

func TestVMImportServiceActions(t *testing.T) {
	var method, path string
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			method, path = r.Method, r.URL.RequestURI()
			w.WriteHeader(http.StatusOK)
		},
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMImport{{Key: sampleImportKey, ID: sampleImportKey, Name: "renamed"}})
		},
	}))
	if err := client.VMImports.Start(context.Background(), sampleImportKey); err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPut || !strings.Contains(path, "action=import") {
		t.Fatalf("start %s %s", method, path)
	}
	if err := client.VMImports.Abort(context.Background(), sampleImportKey); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(path, "action=abort") {
		t.Fatalf("abort %s", path)
	}
	newName := "renamed"
	if _, err := client.VMImports.Update(context.Background(), sampleImportKey, &VMImportUpdateRequest{Name: &newName}); err != nil {
		t.Fatal(err)
	}
}

func TestVMImportServiceWaitSurfacesError(t *testing.T) {
	var reads atomic.Int32
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			reads.Add(1)
			jsonResponse(w, 200, []VMImport{{
				Key: sampleImportKey, ID: sampleImportKey, Name: "imported-vm",
				Status: VMImportStatusError, StatusInfo: "Unable to download https://cloud.centos.org/image.qcow2",
			}})
		},
		"GET /api/v4/vm_import_logs": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if !strings.Contains(filter, sampleImportKey) || !strings.Contains(filter, "level eq 'error'") {
				t.Errorf("log filter = %q", filter)
			}
			jsonResponse(w, 200, []VMImportLog{{Key: 1, Level: VMImportLogLevelError, Text: "download failed"}})
		},
	}))

	start := time.Now()
	_, err := client.VMImports.Wait(context.Background(), sampleImportKey, &VMImportWaitOptions{Timeout: 30 * time.Second})
	elapsed := time.Since(start)
	if !IsVMImportFailed(err) {
		t.Fatalf("wait: %v", err)
	}
	if !strings.Contains(err.Error(), "Unable to download") || !strings.Contains(err.Error(), "download failed") {
		t.Fatalf("error = %v", err)
	}
	if reads.Load() != 1 {
		t.Fatalf("reads = %d", reads.Load())
	}
	if elapsed > time.Second {
		t.Fatalf("failed import waited %s", elapsed)
	}
}

func TestVMImportServiceWaitFailedDrives(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMImport{{
				Key: sampleImportKey, ID: sampleImportKey, Status: VMImportStatusImporting, FailedDriveCount: 1,
				StatusInfo: "drive OS failed",
			}})
		},
		"GET /api/v4/vm_import_logs": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VMImportLog{})
		},
	}))
	start := time.Now()
	_, err := client.VMImports.Wait(context.Background(), sampleImportKey, &VMImportWaitOptions{Timeout: 30 * time.Second})
	if !IsVMImportFailed(err) || !strings.Contains(err.Error(), "1 failed drive") {
		t.Fatalf("wait: %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatalf("failed drives waited %s", time.Since(start))
	}
}

func TestVMImportServiceWaitCompleteAndTimeout(t *testing.T) {
	var reads atomic.Int32
	status := VMImportStatusComplete
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_imports": func(w http.ResponseWriter, r *http.Request) {
			reads.Add(1)
			jsonResponse(w, 200, []VMImport{{Key: sampleImportKey, ID: sampleImportKey, Name: "imported-vm", Status: status}})
		},
	}))
	got, err := client.VMImports.Wait(context.Background(), sampleImportKey, nil)
	if err != nil || got.Status != VMImportStatusComplete || reads.Load() != 1 {
		t.Fatalf("complete = %+v %v reads=%d", got, err, reads.Load())
	}

	status = VMImportStatusImporting
	reads.Store(0)
	start := time.Now()
	_, err = client.VMImports.Wait(context.Background(), sampleImportKey, &VMImportWaitOptions{Timeout: 40 * time.Millisecond, PollInterval: 10 * time.Millisecond})
	if !IsTimeoutError(err) {
		t.Fatalf("timeout: %v", err)
	}
	if reads.Load() < 2 {
		t.Fatalf("reads = %d", reads.Load())
	}
	if time.Since(start) > time.Second {
		t.Fatalf("timeout took %s", time.Since(start))
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.VMImports.Wait(ctx, sampleImportKey, &VMImportWaitOptions{Timeout: 30 * time.Second})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
}

func TestVMImportLogServiceListByImport(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vm_import_logs": func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Query().Get("filter"), "vm_import eq '"+sampleImportKey+"'") {
				t.Errorf("filter = %q", r.URL.Query().Get("filter"))
			}
			jsonResponse(w, 200, []VMImportLog{{Key: 3, VMImport: sampleImportKey, Level: "message", Text: "started"}})
		},
		"GET /api/v4/vm_import_logs/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VMImportLog{Key: 3, Text: "started"})
		},
	}))
	logs, err := client.VMImports.Logs(context.Background(), sampleImportKey)
	if err != nil || len(logs) != 1 || logs[0].Text != "started" {
		t.Fatalf("logs = %+v %v", logs, err)
	}
	got, err := client.VMImportLogs.Get(context.Background(), 3)
	if err != nil || got.Text != "started" {
		t.Fatalf("get = %+v %v", got, err)
	}
	if _, err := client.VMImportLogs.ListByImport(context.Background(), ""); !IsValidationError(err) {
		t.Fatalf("empty import: %v", err)
	}
}

func flexPtr(n int) *FlexInt {
	v := FlexInt(n)
	return &v
}

func intPtr(n int) *int {
	return &n
}
