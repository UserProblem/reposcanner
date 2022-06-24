package engine

import (
	"errors"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/UserProblem/reposcanner/models"
)

type Scanner struct {
	tokens       chan struct{}
	jobBoard     map[string]*Job
	jobBoardLock sync.RWMutex
	jobBoardOpen bool
	noop         bool
}

func (s *Scanner) Initialize(limit int, noop bool) {
	// Number of concurrent running jobs
	s.tokens = make(chan struct{}, limit)

	// Concurrent job board access
	s.jobBoard = make(map[string]*Job)
	s.jobBoardLock = sync.RWMutex{}
	s.jobBoardOpen = true
	s.noop = noop
}

func (s *Scanner) CleanUp() {
	// Close and clear the job board
	s.jobBoardOpen = false

	s.jobBoardLock.Lock()
	defer s.jobBoardLock.Unlock()

	for k := range s.jobBoard {
		delete(s.jobBoard, k)
	}
}

func (s *Scanner) addToJobBoard(j *Job) error {
	if !s.jobBoardOpen {
		return errors.New("job board is closed")
	}

	s.jobBoardLock.Lock()
	defer s.jobBoardLock.Unlock()

	if _, ok := s.jobBoard[j.Id]; !ok {
		s.jobBoard[j.Id] = j
		return nil
	}

	return errors.New("job already exists")
}

func (s *Scanner) removeFromJobBoard(id string) {
	s.jobBoardLock.Lock()
	defer s.jobBoardLock.Unlock()
	delete(s.jobBoard, id)
}

func (s *Scanner) getFromJobBoard(id string) *Job {
	s.jobBoardLock.RLock()
	defer s.jobBoardLock.RUnlock()

	if j, ok := s.jobBoard[id]; ok {
		return j
	}
	return nil
}

func (s *Scanner) StartScan(j *Job) {
	if err := s.addToJobBoard(j); err != nil {
		go func() {
			j.Result <- &JobUpdate{
				Status:   "FAILURE",
				Findings: nil,
			}
		}()
	}

	go s.Work(j.Id)
}

func (s *Scanner) StopScan(j *Job) {
	s.removeFromJobBoard(j.Id)
}

func (s *Scanner) Work(id string) {
	// Reserve work token to limit parallel job execution
	s.tokens <- struct{}{}
	defer func() { <-s.tokens }()

	j := s.getFromJobBoard(id)

	// Check for cancellation
	if j == nil {
		return
	}

	// Update job to ongoing
	j.Result <- &JobUpdate{
		Status:   "ONGOING",
		Findings: nil,
	}

	log.Println("Scanner starting repository download from url.")
	var checkoutDir string
	if s.noop {
		<-time.NewTimer(1 * time.Second).C
	} else {
		// Make temp directory
		if path, err := ioutil.TempDir("", "reposcanner"); err != nil {
			log.Printf("failed to create checkout directory: %v", err.Error())
			s.endJobWithFailure(j)
		} else {
			checkoutDir = path
		}
		defer DeleteTmpDirectory(checkoutDir)

		// Download url
		if err := CloneRepository(j.Repo.Url, j.Repo.Branch, checkoutDir); err != nil {
			log.Printf("failed to download repository: %v", err.Error())
			s.endJobWithFailure(j)
		}
	}

	// Check for cancellation
	if x := s.getFromJobBoard(id); x == nil {
		return
	}

	log.Println("Scanner starting repository scan.")
	var findings []*models.FindingsInfo
	if s.noop {
		<-time.NewTimer(1 * time.Second).C

		// Send dummy results for testing purposes
		findings = []*models.FindingsInfo{
			{
				Type_:  "sast",
				RuleId: "G001",
				Location: &models.FindingsLocation{
					Path: "hello.go",
					Positions: &models.FileLocation{
						Begin: &models.LineLocation{Line: 21},
					},
				},
				Metadata: &models.FindingsMetadata{
					Description: "Hard-coded secret - public key",
					Severity:    "HIGH",
				},
			},
			{
				Type_:  "sast",
				RuleId: "G002",
				Location: &models.FindingsLocation{
					Path: "world.go",
					Positions: &models.FileLocation{
						Begin: &models.LineLocation{Line: 41},
						End:   &models.LineLocation{Line: 43},
					},
				},
				Metadata: &models.FindingsMetadata{
					Description: "Hard-coded secret - private key",
					Severity:    "HIGH",
				},
			},
		}
	} else {
		var sf SecretFinder
		sf.Initialize()
		findings = sf.FindSecrets(checkoutDir)
	}

	// Check for cancellation
	if x := s.getFromJobBoard(id); x == nil {
		return
	}

	// Send results
	j.Result <- &JobUpdate{
		Status:   "SUCCESS",
		Findings: findings,
	}

	s.removeFromJobBoard(id)
}

func (s *Scanner) endJobWithFailure(j *Job) {
	j.Result <- &JobUpdate{
		Status:   "FAILURE",
		Findings: nil,
	}

	s.removeFromJobBoard(j.Id)
}
