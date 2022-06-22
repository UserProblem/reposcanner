package swagger_test

import (
	"net/http"
	"net/http/httptest"
	"os"
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
