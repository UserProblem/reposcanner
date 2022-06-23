package swagger

import (
	"log"
	"sync"

	"github.com/UserProblem/reposcanner/engine"
	"github.com/UserProblem/reposcanner/models"
)

type App struct {
	Router           HttpRouter
	DBType           string
	DB               *PsqlDB
	RepoStore        RepoStore
	ScanStore        ScanStore
	EngineController engine.Controller
	EngineScanner    engine.Scanner
	ActiveJobs       map[string]*ScanJob
	ActiveJobsLock   sync.RWMutex
}

type ScanJob struct {
	Job        *engine.Job
	CancelFlag chan bool
}

const scannerLimit int = 5

func (a *App) Initialize() {
	a.Router = a.NewRouter()
	a.ClearStores()
	a.EngineScanner.Initialize(scannerLimit)
	a.EngineController.Initialize(&a.EngineScanner)
	a.ActiveJobs = make(map[string]*ScanJob)
	a.ActiveJobsLock = sync.RWMutex{}
}

func (a *App) ClearStores() {
	var err error

	var rs RepoStore
	if rs, err = NewRepoStore(a.DBType); err != nil {
		log.Fatal("Cannot initialize repository data store.\n")
	}
	a.RepoStore = rs

	var ss ScanStore
	if ss, err = NewScanStore(a.DBType); err != nil {
		log.Fatal("Cannot initialize scan data store.\n")
	}
	a.ScanStore = ss
}

func (a *App) Run() {
	go a.EngineController.Run()
}

func (a *App) CleanUp() {
	a.ActiveJobsLock.RLock()
	for _, v := range a.ActiveJobs {
		ch := v.CancelFlag
		go func() { ch <- true }()
	}
	a.ActiveJobsLock.RUnlock()

	a.EngineController.Stop()
	a.EngineScanner.CleanUp()
}

func (a *App) AddScanRequest(ri *models.RepositoryInfo, sr *models.ScanRecord) {
	job := a.EngineController.AddJob(ri)
	sj := ScanJob{
		Job:        job,
		CancelFlag: make(chan bool),
	}

	a.ActiveJobsLock.Lock()
	defer a.ActiveJobsLock.Unlock()
	a.ActiveJobs[sr.Id] = &sj

	go a.ScanRequestHandler(sr.Id)
}

func (a *App) RemoveScanRequest(id string) {
	a.ActiveJobsLock.RLock()
	sj, ok := a.ActiveJobs[id]
	a.ActiveJobsLock.RUnlock()

	if !ok {
		log.Printf("Scan %v is not in the active list.\n", id)
		return
	}

	go func() { sj.CancelFlag <- true }()
}

func (a *App) ScanRequestHandler(id string) {
	a.ActiveJobsLock.RLock()
	sj, ok := a.ActiveJobs[id]
	a.ActiveJobsLock.RUnlock()

	if !ok {
		log.Printf("Scan %v is not in the active list.\n", id)
		return
	}

	active := true
	for active {
		select {
		case jupd := <-sj.Job.Result:
			if sr, err := a.ScanStore.Retrieve(id); err != nil {
				log.Printf("Error retrieving scan record: %v\n", err.Error())
				active = false
			} else {
				var newsr *models.ScanRecord

				switch jupd.Status {
				case "ONGOING":
					newsr = sr.Clone()
					newsr.Info.ScanningAt = currentTimestamptz()
					newsr.Info.Status = "IN PROGRESS"
				case "FAILURE":
					newsr = sr.Clone()
					newsr.Info.FinishedAt = currentTimestamptz()
					newsr.Info.Status = "FAILURE"
					active = false
				case "SUCCESS":
					newsr = sr.Clone()
					newsr.Info.FinishedAt = currentTimestamptz()
					newsr.Info.Status = "SUCCESS"
					active = false

					// Save findings to the data store
					if err := a.ScanStore.InsertFindings(id, jupd.Findings); err != nil {
						log.Printf("Error storing findings: %v\n", err.Error())
					}
				}

				if err = a.ScanStore.Update(newsr); err != nil {
					log.Printf("Error updating scan record: %v\n", err.Error())
					active = false
				}
			}
		case <-sj.CancelFlag:
			active = false
		}
	}

	a.ActiveJobsLock.Lock()
	defer a.ActiveJobsLock.Unlock()
	delete(a.ActiveJobs, id)
}
