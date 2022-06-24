package swagger_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	sw "github.com/UserProblem/reposcanner/go"
	"github.com/joho/godotenv"
)

const api_version string = "/v0"

var app sw.App

func TestMain(m *testing.M) {
	godotenv.Load(".env_test")
	prepareDb()
	app.Initialize()
	os.Exit(m.Run())
}

func prepareDb() {
	dbtype := os.Getenv("DATABASE_TYPE")
	log.Printf("Running with database type '%v'\n", dbtype)
	if dbtype == "postgresql" {
		PDB := sw.GetPsqlDBInstance()
		PDB.Host = os.Getenv("DATABASE_HOST")
		if pnum, err := strconv.Atoi(os.Getenv("DATABASE_PORT")); err != nil {
			log.Fatalf(err.Error())
		} else {
			PDB.Port = pnum
		}
		PDB.User = os.Getenv("DATABASE_USER")
		PDB.Password = os.Getenv("DATABASE_PASSWORD")
		PDB.DBname = os.Getenv("DATABASE_NAME")

		PDB.Initialize()
	}
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
