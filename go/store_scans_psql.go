package swagger

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/UserProblem/reposcanner/models"
)

type ScanStorePsqlDB struct {
	DB             *sql.DB
	nextId         chan uint64
	nextFindingsId chan uint64
}

func NewScanStorePsqlDB() (ScanStore, error) {
	actualDB := GetPsqlDBInstance().DB
	if actualDB == nil {
		return nil, fmt.Errorf("cannot retrieve psql instance")
	}

	createEnumStatusQuery := `CREATE TYPE enum_status AS ENUM ( 'QUEUED', 'IN PROGRESS', 'SUCCESS', 'FAILURE' )`

	createScanTableQuery := `CREATE TABLE IF NOT EXISTS scans 
	(
		id TEXT PRIMARY KEY,
		repoId INTEGER NOT NULL REFERENCES repositories(id),
		queuedAt TIMESTAMPTZ NOT NULL,
		scanningAt TIMESTAMPTZ,
		finishedAt TIMESTAMPTZ,
		status enum_status NOT NULL
	)`

	createFindingsTableQuery := `CREATE TABLE IF NOT EXISTS findings
	(
		id SERIAL PRIMARY KEY,
		scanId TEXT NOT NULL REFERENCES scans(id),
		finding JSONB NOT NULL
	)`

	if _, err := actualDB.Exec(createEnumStatusQuery); err != nil {
		if !strings.HasSuffix(err.Error(), "already exists") {
			return nil, fmt.Errorf("could not create enum 'enum_status': %v", err.Error())
		}
	}

	if _, err := actualDB.Exec(createScanTableQuery); err != nil {
		return nil, fmt.Errorf("could not create table 'scans': %v", err.Error())
	}

	if _, err := actualDB.Exec(createFindingsTableQuery); err != nil {
		return nil, fmt.Errorf("could not create table 'findings': %v", err.Error())
	}

	chS, chF := make(chan uint64), make(chan uint64)
	go generateObjIds(chS)
	go generateObjIds(chF)

	return &ScanStorePsqlDB{
		DB:             actualDB,
		nextId:         chS,
		nextFindingsId: chF,
	}, nil
}

// Helper function to auto-generate the next unique id value
// that can be used for new scan records.
func (ss *ScanStorePsqlDB) NextId() string {
	for {
		tmpId := EncodeScanId(<-ss.nextId)
		if _, err := ss.Retrieve(tmpId); err != nil {
			if strings.HasSuffix(err.Error(), "does not exist") {
				return tmpId
			}
		}
	}
}

// Add a new scan record to the data store. Returns a pointer
// to the newly added scan record or nil and an error on failure.
func (ss *ScanStorePsqlDB) Insert(si *models.ScanInfo) (*models.ScanRecord, error) {
	id := ss.NextId()

	var scanningAt, finishedAt *string

	if si.ScanningAt == "" {
		scanningAt = nil
	} else {
		scanningAt = &si.ScanningAt
	}

	if si.FinishedAt == "" {
		finishedAt = nil
	} else {
		finishedAt = &si.FinishedAt
	}

	var res string
	err := ss.DB.QueryRow(
		`INSERT INTO scans(id, repoId, queuedAt, scanningAt, finishedAt, status)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		id, si.RepoId, si.QueuedAt, scanningAt, finishedAt, si.Status).Scan(&res)

	if err != nil {
		return nil, fmt.Errorf("error inserting data to the DB: %v", err.Error())
	}

	if res != id {
		return nil, fmt.Errorf("mismatch in returned id from insert")
	}

	return &models.ScanRecord{
		Id:   id,
		Info: si.Clone(),
	}, nil
}

// Retrieve an existing scan record from the data store.
// Returns a pointer to a copy of the retrieved scan record
// or nil and an error on failure.
func (ss *ScanStorePsqlDB) Retrieve(id string) (*models.ScanRecord, error) {
	var si models.ScanInfo

	var scanningAt, finishedAt *string

	err := ss.DB.QueryRow("SELECT repoId, queuedAt, scanningAt, finishedAt, status FROM scans WHERE id=$1",
		id).Scan(&si.RepoId, &si.QueuedAt, &scanningAt, &finishedAt, &si.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("id %v does not exist", id)
		} else {
			return nil, fmt.Errorf("error retrieving data from the DB. Id %v %v", id, err.Error())
		}
	}

	if scanningAt == nil {
		si.ScanningAt = ""
	} else {
		si.ScanningAt = *scanningAt
	}

	if finishedAt == nil {
		si.FinishedAt = ""
	} else {
		si.FinishedAt = *finishedAt
	}

	return &models.ScanRecord{
		Id:   id,
		Info: si.Clone(),
	}, nil
}

// Delete an existing scan record from the data store.
// Returns nil on success or an error on failure.
func (ss *ScanStorePsqlDB) Delete(id string) error {
	if sr, err := ss.Retrieve(id); err == nil {
		if sr.Info.Status == "SUCCESS" {
			if _, err = ss.DeleteFindings(id); err != nil {
				return fmt.Errorf("failed to delete related findings: %v", err.Error())
			}
		}
	}

	res, err := ss.DB.Exec("DELETE FROM scans WHERE id=$1", id)

	if err != nil {
		return fmt.Errorf("failed to delete record: %v", err.Error())
	}

	if count, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("failed to delete record: %v", err.Error())
	} else if count == 0 {
		return fmt.Errorf("id not found")
	}

	return nil
}

// Update an existing scan record in the data store.
// Returns nil on success or an error on failure.
func (ss *ScanStorePsqlDB) Update(sr *models.ScanRecord) error {
	var scanningAt, finishedAt *string

	if sr.Info.ScanningAt == "" {
		scanningAt = nil
	} else {
		scanningAt = &sr.Info.ScanningAt
	}

	if sr.Info.FinishedAt == "" {
		finishedAt = nil
	} else {
		finishedAt = &sr.Info.FinishedAt
	}

	res, err := ss.DB.Exec("UPDATE scans SET repoId=$1, queuedAt=$2, scanningAt=$3, finishedAt=$4, status=$5 WHERE id=$6",
		sr.Info.RepoId, sr.Info.QueuedAt, scanningAt, finishedAt, sr.Info.Status, sr.Id)

	if err != nil {
		return fmt.Errorf("failed to update record: %v", err.Error())
	}

	if count, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("failed to update record: %v", err.Error())
	} else if count == 0 {
		return fmt.Errorf("id not found")
	}

	return nil
}

// List returns a scan list based on the provided pagination
// parameters. It will return a maximum of page size repository
// records while skipping offset-1 records from the start of the
// data store.
func (ss *ScanStorePsqlDB) List(pp *models.PaginationParams) (*models.ScanList, error) {
	var total int
	err := ss.DB.QueryRow("SELECT count(*) AS row_count FROM scans").Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve scan list: %v", err.Error())
	}

	if pp.Offset > int32(total) {
		return nil, errors.New("invalid offset")
	}

	if pp.PageSize < 1 {
		return nil, errors.New("invalid page size")
	}

	rows, err := ss.DB.Query(
		"SELECT id, repoId, queuedAt, scanningAt, finishedAt, status FROM scans LIMIT $1 OFFSET $2",
		int(pp.PageSize), int(pp.Offset))

	if err != nil {
		return nil, fmt.Errorf("cannot retrieve scan list: %v", err.Error())
	}

	defer rows.Close()

	sl := models.ScanList{
		Total:      int32(total),
		Pagination: pp.Clone(),
		Items:      make([]models.ScanRecord, 0),
	}

	for rows.Next() {
		var sr models.ScanRecord
		var si models.ScanInfo
		var scanningAt, finishedAt *string

		if err := rows.Scan(&sr.Id, &si.RepoId, &si.QueuedAt, &scanningAt, &finishedAt, &si.Status); err != nil {
			return nil, fmt.Errorf("cannot retrieve scan list: %v", err.Error())
		}

		if scanningAt == nil {
			si.ScanningAt = ""
		} else {
			si.ScanningAt = *scanningAt
		}

		if finishedAt == nil {
			si.FinishedAt = ""
		} else {
			si.FinishedAt = *finishedAt
		}

		sr.Info = &si
		sl.Items = append(sl.Items, sr)
	}

	return &sl, nil
}

// InsertFindings stores all of the contents of the findings
// list into the data store, indexed by scanId. All operations
// related to findings are done in bulk. It returns nil on success
// or an error on failure, at which point none of the findings
// will be stored.
func (ss *ScanStorePsqlDB) InsertFindings(scanId string, findings []*models.FindingsInfo) error {
	if len(findings) == 0 {
		return nil
	}

	ctx := context.Background()
	txn, err := ss.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create DB transaction: %v", err.Error())
	}

	for _, fi := range findings {
		var res int
		fistr, _ := json.Marshal(fi)
		err := txn.QueryRow("INSERT INTO findings(scanId, finding) VALUES ($1, $2) RETURNING id", scanId, fistr).Scan(&res)
		if err != nil {
			txn.Rollback()
			return fmt.Errorf("error inserting data to the DB: %v", err.Error())
		}
	}

	txn.Commit()
	return nil
}

// ListFindings retrieves all of the findings from the data store
// indexed by scanId. All operations related to findings are done
// in bulk. It returns a list of findings on success, or nil and
// an error on failure.
func (ss *ScanStorePsqlDB) ListFindings(scanId string) ([]*models.FindingsInfo, error) {
	rows, err := ss.DB.Query("SELECT finding FROM findings WHERE scanId=$1", scanId)

	if err != nil {
		return nil, fmt.Errorf("cannot retrieve findings list: %v", err.Error())
	}

	defer rows.Close()

	findings := make([]*models.FindingsInfo, 0)
	for rows.Next() {
		var buffer []byte
		if err := rows.Scan(&buffer); err != nil {
			return nil, fmt.Errorf("cannot retrieve findings list: %v", err.Error())
		}

		var fi models.FindingsInfo
		decoder := json.NewDecoder(bytes.NewReader(buffer))
		if err = decoder.Decode(&fi); err != nil {
			return nil, fmt.Errorf("cannot retrieve findings list: %v", err.Error())
		}

		findings = append(findings, &fi)
	}

	return findings, nil
}

// DeleteFindings deletes all of the findings from the data store
// indexed by scanId. All operations related to findings are done
// in bulk. It returns the number of deleted records on success,
// or zero and an error on failure, at which point none of the findings
// will be deleted.
func (ss *ScanStorePsqlDB) DeleteFindings(scanId string) (int, error) {
	ctx := context.Background()
	txn, err := ss.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create DB transaction: %v", err.Error())
	}

	res, err := txn.ExecContext(ctx, "DELETE FROM findings WHERE scanId=$1", scanId)
	if err != nil {
		txn.Rollback()
		return 0, fmt.Errorf("cannot delete all findings: %v", err.Error())
	}

	count, err := res.RowsAffected()
	if err != nil {
		txn.Rollback()
		return 0, fmt.Errorf("cannot delete all findings: %v", err.Error())
	}

	txn.Commit()
	return int(count), nil
}
