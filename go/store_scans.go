package swagger

import (
	"errors"

	"github.com/UserProblem/reposcanner/models"
)

type ScanStore interface {
	NextId() string
	Insert(si *models.ScanInfo) (*models.ScanRecord, error)
	Retrieve(id string) (*models.ScanRecord, error)
	Delete(id string) error
	Update(sr *models.ScanRecord) error
	List(pp *models.PaginationParams) (*models.ScanList, error)
	NextFindingsId() int
	InsertFindings(scanId string, findings []*models.FindingsInfo) error
	ListFindings(scanId string) ([]*models.FindingsInfo, error)
	DeleteFindings(scanId string) (int, error)
}

// Create and return a pointer to a new scan data store.
// Returns nil and an error on failure
func NewScanStore(dbtype string) (ScanStore, error) {
	if dbtype == "postgresql" {
		return nil, errors.New("not implemented")
	} else {
		return NewScanStoreMemDB()
	}
}

/*
CREATE TYPE enum_status (
	"QUEUED",
	"IN PROGRESS",
	"SUCCESS",
	"FAILURE"
)
CREATE TABLE IF NOT EXISTS scans
(
    id TEXT,
	repoId INTEGER NOT NULL REFERENCES repositories(id),
    queuedAt TIMESTAMPTZ NOT NULL,
	scanningAt TIMESTAMPTZ,
	finishedAt TIMESTAMPTZ,
	status enum_status NOT NULL,
    CONSTRAINT scans_pkey PRIMARY KEY (id)
)

CREATE TABLE IF NOT EXISTS findings
(
	id SERIAL,
	scanId TEXT NOT NULL REFERENCES scans(id)
	finding JSONB NOT NULL
	CONSTRAINT findings_pkey PRIMARY KEY (id)
)
*/
