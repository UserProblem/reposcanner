package swagger_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sw "github.com/UserProblem/reposcanner/go"
	"github.com/UserProblem/reposcanner/models"
)

func addDummyScanRecords(t *testing.T, n int) {
	for i := 1; i <= n; i++ {
		si := models.ScanInfo{
			RepoId:     int64(i),
			QueuedAt:   fmt.Sprintf("1970-01-01 00:%02d:00+0", i),
			ScanningAt: "",
			FinishedAt: "",
			Status:     "QUEUED",
		}

		if _, err := app.ScanStore.Insert(&si); err != nil {
			t.Fatalf("Failed to add record to the scan store.\n")
		}
	}
}

func TestAddScan(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 1)

	req, _ := http.NewRequest("POST", api_version+"/repository/1/startScan", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var body models.ApiResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON received as response body.")
	}

	if body.Id != 0 {
		t.Errorf("Expected Id to be 0. Got %v\n", body.Id)
	}

	expMsg := sw.EncodeScanId(1)
	if body.Message != expMsg {
		t.Errorf("Expected response message to be '%v'. Got '%v'\n", expMsg, body.Message)
	}

	if sr, err := app.ScanStore.Retrieve(sw.EncodeScanId(1)); err != nil {
		t.Fatalf("Could not retrieve scan record\n")
	} else {
		if sr.Info.RepoId != 1 {
			t.Errorf("Expected repo id to be 1. Got %v\n", sr.Info.RepoId)
		}

		if sr.Info.QueuedAt == "" {
			t.Errorf("Expected queued at to not be empty. Got '%v'\n", sr.Info.QueuedAt)
		}

		if sr.Info.ScanningAt != "" {
			t.Errorf("Expected scanning at to be empty. Got '%v'\n", sr.Info.ScanningAt)
		}

		if sr.Info.FinishedAt != "" {
			t.Errorf("Expected finished at to be empty. Got '%v'\n", sr.Info.FinishedAt)
		}

		if sr.Info.Status != "QUEUED" {
			t.Errorf("Expected status to be QUEUED. Got %v\n", sr.Info.Status)
		}
	}

}

func TestAddScanInvalidRepoId(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("POST", api_version+"/repository/invalid/startScan", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestAddScanNonExistentRepoId(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("POST", api_version+"/repository/42/startScan", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetScan(t *testing.T) {
	app.ClearStores()
	addDummyScanRecords(t, 2)

	req, _ := http.NewRequest("GET", api_version+"/scan/"+sw.EncodeScanId(2), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var body models.ScanResults
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON received as response body.")
	}

	if body.Id != sw.EncodeScanId(2) {
		t.Errorf("Expected id to be %v. Got %v\n", sw.EncodeScanId(2), body.Id)
	}

	if body.Info.RepoId != 2 {
		t.Errorf("Expected repo id to be 2. Got %v\n", body.Info.RepoId)
	}

	if body.Info.QueuedAt == "" {
		t.Errorf("Expected queued at to not be empty. Got '%v'\n", body.Info.QueuedAt)
	}

	if body.Info.ScanningAt != "" {
		t.Errorf("Expected scanning at to be empty. Got '%v'\n", body.Info.ScanningAt)
	}

	if body.Info.FinishedAt != "" {
		t.Errorf("Expected finished at to be empty. Got '%v'\n", body.Info.FinishedAt)
	}

	if body.Info.Status != "QUEUED" {
		t.Errorf("Expected status to be QUEUED. Got %v\n", body.Info.Status)
	}

	if len(body.Findings) != 0 {
		t.Errorf("Expected empty findings list. Got '%v'\n", body.Findings)
	}
}

func TestGetScanInvalidId(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("GET", api_version+"/scan/invalid", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestGetNonExistentScan(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("GET", api_version+"/scan/"+sw.EncodeScanId(42), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestDeleteScan(t *testing.T) {
	app.ClearStores()
	addDummyScanRecords(t, 3)

	req, _ := http.NewRequest("DELETE", api_version+"/scan/"+sw.EncodeScanId(2), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if response.Body != nil && response.Body.Len() != 0 {
		t.Errorf("Expected no body in response\n")
	}

	sl, err := app.ScanStore.List(&models.PaginationParams{Offset: 0, PageSize: 10})
	if err != nil {
		t.Fatalf("Could not retrieve repository list\n")
	}

	if sl.Total != int32(2) {
		t.Errorf("Expected total records to be 2. Got %v\n", sl.Total)
	}

	if sl.Items[0].Id != sw.EncodeScanId(1) || sl.Items[1].Id != sw.EncodeScanId(3) {
		t.Errorf("Expected ids %v and %v to be present. Got %v and %v\n",
			sw.EncodeScanId(1), sw.EncodeScanId(3), sl.Items[0].Id, sl.Items[1].Id)
	}
}

func TestDeleteScanInvalidId(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("DELETE", api_version+"/scan/invalid", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestDeleteScanNonExistentId(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("DELETE", api_version+"/scan/"+sw.EncodeScanId(42), nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetNoScans(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("GET", api_version+"/scans", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var body models.ScanList
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON received as response body.")
	}

	if body.Total != 0 {
		t.Errorf("Expected total count to be 0. Got %v\n", body.Total)
	}

	if body.Pagination.Offset != 0 {
		t.Errorf("Expected pagination offset to be 0. Got %v\n", body.Pagination.Offset)
	}

	if body.Pagination.PageSize != 20 {
		t.Errorf("Expected pagination pagesize to be 20. Got %v\n", body.Pagination.PageSize)
	}

	if len(body.Items) != 0 {
		t.Errorf("Expected items to have 0 length. Got %v (%v)\n", len(body.Items), body.Items)
	}
}

func TestGetScanList(t *testing.T) {
	app.ClearStores()
	addDummyScanRecords(t, 10)

	pp := models.PaginationParams{Offset: 2, PageSize: 5}
	reqBody, _ := json.Marshal(pp)
	req, _ := http.NewRequest("GET", api_version+"/scans", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var body models.ScanList
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON received as response body.")
	}

	if body.Total != 10 {
		t.Errorf("Expected total count to be 10. Got %v\n", body.Total)
	}

	if body.Pagination.Offset != pp.Offset {
		t.Errorf("Expected pagination offset to be %v. Got %v\n", pp.Offset, body.Pagination.Offset)
	}

	if body.Pagination.PageSize != pp.PageSize {
		t.Errorf("Expected pagination pagesize to be %v. Got %v\n", pp.PageSize, body.Pagination.PageSize)
	}

	if len(body.Items) != int(pp.PageSize) {
		t.Fatalf("Expected items to have %v length. Got %v (%v)\n", pp.PageSize, body.Items, body.Items)
	}

	for i := 0; i < int(pp.PageSize); i++ {
		expectedId := sw.EncodeScanId(uint64(i) + uint64(pp.Offset) + 1)
		if body.Items[i].Id != expectedId {
			t.Errorf("Expected id to be %v. Got %v\n", expectedId, body.Items[i].Id)
		}
	}
}

func TestGetScanListInvalidBody(t *testing.T) {
	app.ClearStores()

	reqBody := []byte("invalid body")
	req, _ := http.NewRequest("GET", api_version+"/scans", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestGetScanListInvalidOffset(t *testing.T) {
	app.ClearStores()
	addDummyScanRecords(t, 5)

	pp := models.PaginationParams{Offset: 6, PageSize: 5}
	reqBody, _ := json.Marshal(pp)
	req, _ := http.NewRequest("GET", api_version+"/scans", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetScanListInvalidPageSize(t *testing.T) {
	app.ClearStores()
	addDummyScanRecords(t, 5)

	pp := models.PaginationParams{Offset: 0, PageSize: 0}
	reqBody, _ := json.Marshal(pp)
	req, _ := http.NewRequest("GET", api_version+"/scans", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestAddScanAndReceiveFindings(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 1)

	req, _ := http.NewRequest("POST", api_version+"/repository/1/startScan", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var addBody models.ApiResponse
	if err := json.Unmarshal(response.Body.Bytes(), &addBody); err != nil {
		t.Fatalf("Invalid JSON received as response body.")
	}

	// Trigger the scan to start
	app.EngineController.RunOnce()

	waitSeconds(1)

	scanId := addBody.Message

	req, _ = http.NewRequest("GET", api_version+"/scan/"+scanId, nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	checkScanStatus(t, response, true, false, "IN PROGRESS")

	waitSeconds(3)

	req, _ = http.NewRequest("GET", api_version+"/scan/"+scanId, nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	checkScanStatus(t, response, true, true, "SUCCESS")

	var sr models.ScanResults
	_ = json.Unmarshal(response.Body.Bytes(), &sr)

	if len(sr.Findings) == 0 {
		t.Errorf("Expected findings to not be empty.\n")
	}
}

func waitSeconds(n int) {
	timer := time.NewTimer(time.Duration(n) * time.Second)
	<-timer.C
}

func checkScanStatus(t *testing.T, rsp *httptest.ResponseRecorder, scanningAt, finishedAt bool, status string) {
	var sr models.ScanResults
	if err := json.Unmarshal(rsp.Body.Bytes(), &sr); err != nil {
		t.Fatalf("Invalid JSON received as response body.")
	}

	if scanningAt && sr.Info.ScanningAt == "" {
		t.Fatalf("Expected scanning at to be set but it is still empty.\n")
	}

	if finishedAt && sr.Info.FinishedAt == "" {
		t.Fatalf("Expected finished at to be set but it is still empty.\n")
	}

	if sr.Info.Status != status {
		t.Fatalf("Expected status to be %v. Got %v\n", status, sr.Info.Status)
	}
}
