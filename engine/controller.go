package engine

import (
	"encoding/base64"
	"encoding/binary"
	"log"

	sw "github.com/UserProblem/reposcanner/go"
)

type Controller struct {
	Incoming       chan *Job
	Cancelling     chan *Job
	Quit           chan bool
	initializeFlag bool
	QuitFlag       bool
	nextJobId      chan string
	scanHandler    ScanHandler
}

type Job struct {
	Id     string
	Repo   *sw.RepositoryInfo
	Result chan *JobUpdate
}

type JobUpdate struct {
	Status   string
	Findings *[]sw.FindingsInfo
}

type ScanHandler interface {
	StartScan(*Job)
	StopScan(*Job)
}

// Setup the controller for use
func (c *Controller) Initialize(scanner ScanHandler) {
	c.Incoming = make(chan *Job)
	c.Cancelling = make(chan *Job)
	c.Quit = make(chan bool)
	c.QuitFlag = false
	c.nextJobId = make(chan string)
	go c.generateJobIds()

	c.scanHandler = scanner

	c.initializeFlag = true
}

// Generator to auto-generate the next unique id value that
// can be used for new jobs.
func (c *Controller) generateJobIds() {
	nextId := uint64(0)
	for {
		c.nextJobId <- encodeJobId(uint64(nextId))
	}
}

// Helper function to convert a numeric value into a base64
// string value that can be used as an id
func encodeJobId(v uint64) string {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Starts the controller. It will continue running until a
// message is received from the Quit channel. Initialize() must be
// called before calling this function.
func (c *Controller) Run() {
	for !c.QuitFlag {
		c.RunOnce()
	}
}

// The controller will execute for one loop. Initialize() must be
// called before calling this function.
func (c *Controller) RunOnce() {
	if !c.initializeFlag {
		log.Fatalf("Controller started before initialization.\n")
	}

	select {
	case job := <-c.Incoming:
		log.Printf("Handling incoming job. Id: %v\n", job.Id)
		go c.scanHandler.StartScan(job)
	case job := <-c.Cancelling:
		log.Printf("Handling cancellation of job. Id: %v\n", job.Id)
		go c.scanHandler.StopScan(job)
	case <-c.Quit:
		log.Println("Stopping controller.")
		c.QuitFlag = true
	}
}

// Notifies the controller to stop running. This is not guaranteed
// to be immediate, and pending jobs may still be processed before
// execution stops.
func (c *Controller) Stop() {
	go func() { c.Quit <- true }()
}

// API to allow users to add jobs to the controller queue. Returns a
// Job struct containing the identifier for the queued job, as well as
// the results channel where the output will be sent.
func (c *Controller) AddJob(ri *sw.RepositoryInfo) *Job {
	log.Printf("Received request to scan '%v'\n", ri.Name)

	job := Job{
		Id:     <-c.nextJobId,
		Repo:   ri.Clone(),
		Result: make(chan *JobUpdate),
	}
	go func() { c.Incoming <- &job }()
	return &job
}

// API to allow users to cancel jobs that are in the controller queue.
// When a job is cancelled by the user, the user should no longer perform
// any additional processing for the job. No further data should be expected
// from the results channel.
func (c *Controller) RemoveJob(job *Job) {
	log.Printf("Received request to cancel job '%v'\n", job.Id)
	go func() { c.Cancelling <- job }()
}
