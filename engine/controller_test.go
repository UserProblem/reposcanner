package engine_test

import (
	"log"
	"testing"
	"time"

	"github.com/UserProblem/reposcanner/engine"
	"github.com/UserProblem/reposcanner/models"
)

type DummyScanner struct {
	Started chan *engine.Job
	Stopped chan *engine.Job
}

func (o *DummyScanner) StartScan(j *engine.Job) {
	log.Printf("Dummy scanner start received job %v\n", j.Id)
	o.Started <- j
}

func (o *DummyScanner) StopScan(j *engine.Job) {
	log.Printf("Dummy scanner stop received job %v\n", j.Id)
	o.Stopped <- j
}

func initializeDummyScanner() *DummyScanner {
	return &DummyScanner{
		Started: make(chan *engine.Job),
		Stopped: make(chan *engine.Job),
	}
}

func setupControllerTests() (*engine.Controller, *DummyScanner) {
	var c engine.Controller

	o := initializeDummyScanner()
	c.Initialize(o)

	return &c, o
}

func TestAddJobAddsToIncomingChannel(t *testing.T) {
	c, _ := setupControllerTests()

	ri := models.DefaultRepositoryInfo()
	job := c.AddJob(ri)
	if job == nil {
		t.Fatalf("Could not add job to the queue.")
	}

	timeout := time.NewTimer(1 * time.Second)

	select {
	case queuedJob := <-c.Incoming:
		if job.Id != queuedJob.Id {
			t.Errorf("Job expected id %v. Got %v\n", job.Id, queuedJob.Id)
		}
	case <-timeout.C:
		t.Errorf("Job was not added to the queue.\n")
	}
}

func TestRunOnceHandlesJobFromIncomingChannel(t *testing.T) {
	c, o := setupControllerTests()

	ri := models.DefaultRepositoryInfo()
	job := c.AddJob(ri)

	c.RunOnce()

	timeout := time.NewTimer(1 * time.Second)

	select {
	case handledJob := <-o.Started:
		if job.Id != handledJob.Id {
			t.Errorf("Job expected id %v. Got %v\n", job.Id, handledJob.Id)
		}
	case <-timeout.C:
		t.Errorf("Job was not handled when RunOnce called.\n")
	}
}

func TestRemoveJobAddsToCancellingChannel(t *testing.T) {
	c, _ := setupControllerTests()

	ri := models.DefaultRepositoryInfo()
	job := c.AddJob(ri)
	if job == nil {
		t.Fatalf("Could not add job to the queue.")
	}

	c.RemoveJob(job)

	timeout := time.NewTimer(1 * time.Second)

	select {
	case queuedJob := <-c.Cancelling:
		if job.Id != queuedJob.Id {
			t.Errorf("Job expected id %v. Got %v\n", job.Id, queuedJob.Id)
		}
	case <-timeout.C:
		t.Errorf("Job was not added to the queue.\n")
	}
}

func TestRunOnceHandlesJobFromCancellingChannel(t *testing.T) {
	c, o := setupControllerTests()

	ri := models.DefaultRepositoryInfo()
	job := c.AddJob(ri)

	c.RunOnce()
	c.RemoveJob(job)
	c.RunOnce()

	timeout := time.NewTimer(1 * time.Second)

	select {
	case handledJob := <-o.Stopped:
		if job.Id != handledJob.Id {
			t.Errorf("Job expected id %v. Got %v\n", job.Id, handledJob.Id)
		}
	case <-timeout.C:
		t.Errorf("Job was not handled when RunOnce called.\n")
	}
}

func TestCallingStopStopsController(t *testing.T) {
	c, _ := setupControllerTests()

	if c.QuitFlag != false {
		t.Fatalf("Expected controller to be running.\n")
	}

	c.Stop()
	c.RunOnce()

	if c.QuitFlag != true {
		t.Fatalf("Expected controller to be stopped.\n")
	}
}
