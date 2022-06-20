package swagger_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	sw "github.com/UserProblem/reposcanner/go"
)

const api_version string = "/v0"

var app sw.App

func TestMain(m *testing.M) {
	app.Initialize()
	os.Exit(m.Run())
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func addDummyRepoRecords(t *testing.T, n int) {
	for i := 1; i <= n; i++ {
		ri := sw.RepositoryInfo{
			Name:   "repo name " + strconv.Itoa(i),
			Url:    "repo url " + strconv.Itoa(i),
			Branch: "main",
		}

		if _, err := app.RepoStore.Insert(&ri); err != nil {
			t.Fatalf("Failed to add record to the repo store.\n")
		}
	}
}

func TestGetNoRepositories(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("GET", api_version+"/repositories", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var body sw.RepositoryList
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

func TestGetRepositoryList(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 10)

	pp := sw.PaginationParams{Offset: 2, PageSize: 5}
	reqBody, _ := json.Marshal(pp)
	req, _ := http.NewRequest("GET", api_version+"/repositories", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var body sw.RepositoryList
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
		expectedId := int64(i) + int64(pp.Offset)
		if body.Items[i].Id != expectedId {
			t.Errorf("Expected id to be %v. Got %v\n", expectedId, body.Items[i].Id)
		}
	}
}

func TestGetRepositoryListInvalidBody(t *testing.T) {
	app.ClearStores()

	reqBody := []byte("invalid body")
	req, _ := http.NewRequest("GET", api_version+"/repositories", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestGetRepositoryListInvalidOffset(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 5)

	pp := sw.PaginationParams{Offset: 6, PageSize: 5}
	reqBody, _ := json.Marshal(pp)
	req, _ := http.NewRequest("GET", api_version+"/repositories", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestGetRepositoryListInvalidPageSize(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 5)

	pp := sw.PaginationParams{Offset: 0, PageSize: 0}
	reqBody, _ := json.Marshal(pp)
	req, _ := http.NewRequest("GET", api_version+"/repositories", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestPostNewRepository(t *testing.T) {
	app.ClearStores()

	newRepo := sw.DefaultRepositoryInfo()
	reqBody, _ := json.Marshal(newRepo)

	req, _ := http.NewRequest("POST", api_version+"/repository", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var body sw.ApiResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON received as response body.")
	}

	if body.Id != 1 {
		t.Errorf("Expected new repo Id to be 1. Got %v\n", body.Id)
	}

	expMsg := "repository created successfully"
	if body.Message != expMsg {
		t.Errorf("Expected response message to be '%v'. Got '%v'\n", expMsg, body.Message)
	}
}

func TestPostNewRepositoryUnspecifiedBranch(t *testing.T) {
	app.ClearStores()

	newRepo := &sw.RepositoryInfo{
		Name: "repo name",
		Url:  "repo url",
	}
	reqBody, _ := json.Marshal(newRepo)

	req, _ := http.NewRequest("POST", api_version+"/repository", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var body sw.ApiResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON received as response body.")
	}

	rr, err := app.RepoStore.Retrieve(body.Id)
	if err != nil {
		t.Fatalf("Failed to retrieve newly created repository record.\n")
	}

	if rr.Info.Branch != "main" {
		t.Errorf("Expected branch to be 'main'. Got '%v'\n", rr.Info.Branch)
	}
}

func TestPostNewRepositoryNoBody(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("POST", api_version+"/repository", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestPostNewRepositoryInvalidBody(t *testing.T) {
	app.ClearStores()

	reqBody := []byte("invalid body")
	req, _ := http.NewRequest("POST", api_version+"/repository", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestGetRepository(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 2)

	req, _ := http.NewRequest("GET", api_version+"/repository/2", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var body sw.RepositoryRecord
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON received as response body.")
	}

	if body.Id != 2 {
		t.Errorf("Expected id to be 2. Got %v\n", body.Id)
	}

	if body.Info.Name != "repo name 2" {
		t.Errorf("Expected name to be 'repo name 2'. Got '%v'\n", body.Info.Name)
	}

	if body.Info.Branch != "main" {
		t.Errorf("Expected branch to be 'main'. Got '%v'\n", body.Info.Branch)
	}
}

func TestGetRepositoryInvalidId(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("GET", api_version+"/repository/invalid", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestGetNonExistentRepository(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("GET", api_version+"/repository/42", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestModifyRepository(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 2)

	rr, err := app.RepoStore.Retrieve(2)
	if err != nil {
		t.Fatalf("Failed to retrieve repo record: %v\n", err.Error())
	}

	if rr.Info.Name != "repo name 2" {
		t.Errorf("Expected name to be 'repo name 2'. Got '%v'\n", rr.Info.Name)
	}

	if rr.Info.Url != "repo url 2" {
		t.Errorf("Expected url to be 'repo url 2'. Got '%v'\n", rr.Info.Url)
	}

	if rr.Info.Branch != "main" {
		t.Errorf("Expected branch to be 'main'. Got '%v'\n", rr.Info.Branch)
	}

	modifiedRepo := sw.RepositoryInfo{
		Name:   "modified repo name",
		Url:    "modified repo url",
		Branch: "modified",
	}

	reqBody, _ := json.Marshal(modifiedRepo)
	req, _ := http.NewRequest("PUT", api_version+"/repository/2", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var body sw.ApiResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON received as response body.")
	}

	if body.Id != 2 {
		t.Errorf("Expected new repo Id to be 2. Got %v\n", body.Id)
	}

	expMsg := "repository modified successfully"
	if body.Message != expMsg {
		t.Errorf("Expected response message to be '%v'. Got '%v'\n", expMsg, body.Message)
	}

	rr, err = app.RepoStore.Retrieve(2)
	if err != nil {
		t.Fatalf("Failed to retrieve repo record: %v\n", err.Error())
	}

	if rr.Info.Name != modifiedRepo.Name {
		t.Errorf("Expected name to be '%v'. Got '%v'\n", modifiedRepo.Name, rr.Info.Name)
	}

	if rr.Info.Url != modifiedRepo.Url {
		t.Errorf("Expected url to be '%v'. Got '%v'\n", modifiedRepo.Url, rr.Info.Url)
	}

	if rr.Info.Branch != modifiedRepo.Branch {
		t.Errorf("Expected branch to be '%v'. Got '%v'\n", modifiedRepo.Branch, rr.Info.Branch)
	}
}

func TestModifyRepositoryInvalidId(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 2)

	modifiedRepo := sw.RepositoryInfo{
		Name:   "modified repo name",
		Url:    "modified repo url",
		Branch: "modified",
	}

	reqBody, _ := json.Marshal(modifiedRepo)
	req, _ := http.NewRequest("PUT", api_version+"/repository/invalid", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestModifyRepositoryNoBody(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 2)

	req, _ := http.NewRequest("PUT", api_version+"/repository/2", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestModifyRepositoryInvalidBody(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 2)

	reqBody := []byte("invalid body")
	req, _ := http.NewRequest("PUT", api_version+"/repository/2", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestModifyRepositoryNonExistentId(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 2)

	modifiedRepo := sw.RepositoryInfo{
		Name:   "modified repo name",
		Url:    "modified repo url",
		Branch: "modified",
	}

	reqBody, _ := json.Marshal(modifiedRepo)
	req, _ := http.NewRequest("PUT", api_version+"/repository/42", bytes.NewBuffer(reqBody))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestDeleteRepository(t *testing.T) {
	app.ClearStores()
	addDummyRepoRecords(t, 3)

	req, _ := http.NewRequest("DELETE", api_version+"/repository/2", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if response.Body != nil && response.Body.Len() != 0 {
		t.Errorf("Expected no body in response\n")
	}

	rl, err := app.RepoStore.List(&sw.PaginationParams{Offset: 0, PageSize: 10})
	if err != nil {
		t.Fatalf("Could not retrieve repository list\n")
	}

	if rl.Total != int32(2) {
		t.Errorf("Expected total records to be 2. Got %v\n", rl.Total)
	}

	if rl.Items[0].Id != 1 || rl.Items[1].Id != 3 {
		t.Errorf("Expected ids 1 and 3 to be present. Got %v and %v\n", rl.Items[0].Id, rl.Items[1].Id)
	}
}

func TestDeleteRepositoryInvalidId(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("DELETE", api_version+"/repository/invalid", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestDeleteRepositoryNonExistentId(t *testing.T) {
	app.ClearStores()

	req, _ := http.NewRequest("DELETE", api_version+"/repository/42", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}
