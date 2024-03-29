package engine_test

import (
	"log"
	"testing"
	"time"

	"github.com/UserProblem/reposcanner/engine"
	"github.com/UserProblem/reposcanner/models"
)

func setupScannerTests(limit int) *engine.Scanner {
	var s engine.Scanner
	s.Initialize(limit, true)
	return &s
}

func TestScannerWorksOnJob(t *testing.T) {
	s := setupScannerTests(1)

	results := make(chan *engine.JobUpdate)

	j := &engine.Job{
		Id:     "A",
		Repo:   models.DefaultRepositoryInfo(),
		Result: results,
	}

	s.StartScan(j)

	timeout := time.NewTimer(3 * time.Second)
	defer timeout.Stop()

	select {
	case r := <-results:
		if r.Status != "ONGOING" {
			t.Fatalf("Expected job status to be ONGOING. Got %v\n", r.Status)
		}
	case <-timeout.C:
		t.Fatalf("Expected to receive job status change, but timed out.\n")
	}

	select {
	case r := <-results:
		if r.Status != "SUCCESS" {
			t.Fatalf("Expected job status to be SUCCESS. Got %v\n", r.Status)
		}
	case <-timeout.C:
		t.Fatalf("Expected to receive job status change, but timed out.\n")
	}
}

func TestStartScanRejectsDuplicateJobIds(t *testing.T) {
	s := setupScannerTests(5)

	j := &engine.Job{
		Id:     "A",
		Repo:   models.DefaultRepositoryInfo(),
		Result: make(chan *engine.JobUpdate),
	}

	log.Printf("Starting first job\n")
	s.StartScan(j)

	results := make(chan *engine.JobUpdate)
	j = &engine.Job{
		Id:     "A",
		Repo:   models.DefaultRepositoryInfo(),
		Result: results,
	}

	log.Printf("Starting second job\n")
	s.StartScan(j)

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	log.Printf("Waiting for second job to fail\n")
	for {
		select {
		case r := <-results:
			if r.Status == "FAILURE" {
				return
			}
		case <-timeout.C:
			t.Fatalf("Expected FAILURE response but timed out.")
			return
		}
	}
}

func TestScannerWorkHandlesCancellation(t *testing.T) {
	s := setupScannerTests(1)

	results := make(chan *engine.JobUpdate)

	j := &engine.Job{
		Id:     "A",
		Repo:   models.DefaultRepositoryInfo(),
		Result: results,
	}

	s.StartScan(j)
	s.StopScan(j)

	timeout := time.NewTimer(3 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case r := <-results:
			if r.Status == "SUCCESS" {
				t.Fatalf("Received success result but cancellation expected.\n")
			}
		case <-timeout.C:
			return
		}
	}
}

func TestScannerWorksOnRealJob(t *testing.T) {
	var s engine.Scanner
	s.Initialize(1, false)

	results := make(chan *engine.JobUpdate)

	j := &engine.Job{
		Id: "A",
		Repo: &models.RepositoryInfo{
			Name:   "testdata",
			Url:    "https://github.com/UserProblem/testdata.git",
			Branch: "master",
		},
		Result: results,
	}

	s.StartScan(j)

	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()

	select {
	case r := <-results:
		if r.Status != "ONGOING" {
			t.Fatalf("Expected job status to be ONGOING. Got %v\n", r.Status)
		}
	case <-timeout.C:
		t.Fatalf("Expected to receive job status change, but timed out.\n")
	}

	select {
	case r := <-results:
		if r.Status != "SUCCESS" {
			t.Fatalf("Expected job status to be SUCCESS. Got %v\n", r.Status)
		}
		if len(r.Findings) != 6 {
			t.Errorf("Expected number of findings to be 6. Got %v\n", len(r.Findings))
		}
	case <-timeout.C:
		t.Fatalf("Expected to receive job status change, but timed out.\n")
	}
}

func TestScannerWorksOnRealJobNotFoundUrl(t *testing.T) {
	var s engine.Scanner
	s.Initialize(1, false)

	results := make(chan *engine.JobUpdate)

	j := &engine.Job{
		Id: "A",
		Repo: &models.RepositoryInfo{
			Name:   "testdata",
			Url:    "https://not.a/real/repo",
			Branch: "master",
		},
		Result: results,
	}

	s.StartScan(j)

	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()

	select {
	case r := <-results:
		if r.Status != "ONGOING" {
			t.Fatalf("Expected job status to be ONGOING. Got %v\n", r.Status)
		}
	case <-timeout.C:
		t.Fatalf("Expected to receive job status change, but timed out.\n")
	}

	select {
	case r := <-results:
		if r.Status != "FAILURE" {
			t.Fatalf("Expected job status to be FAILURE. Got %v\n", r.Status)
		}
	case <-timeout.C:
		t.Fatalf("Expected to receive job status change, but timed out.\n")
	}
}
