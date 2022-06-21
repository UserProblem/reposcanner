package swagger

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/hashicorp/go-memdb"
)

type ScanStore struct {
	DB     *memdb.MemDB
	nextId uint64
	total  int32
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
    id SERIAL,
	repoId INTEGER NOT NULL REFERENCES repositories(id),
    queuedAt TIMESTAMPTZ NOT NULL,
	scanningAt TIMESTAMPTZ,
	finishedAt TIMESTAMPTZ,
	status enum_status NOT NULL,
    CONSTRAINT scans_pkey PRIMARY KEY (id)
)
*/

// Create and return a pointer to a new scan data store.
// Returns nil and an error on failure
func NewScanStore() (*ScanStore, error) {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"scans": {
				Name: "scans",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Id", Lowercase: false},
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Cannot initialize db: %s", err.Error()))
	}

	return &ScanStore{DB: db, nextId: 1, total: 0}, nil
}

// Helper function to auto-generate the next unique id value
// that can be used for new scan records.
func (rs *ScanStore) NextId() string {
	defer func() { rs.nextId++ }()
	return EncodeScanId(rs.nextId)
}

// Helper function to convert a numeric value into a base64
// string value that can be used as an id
func EncodeScanId(v uint64) string {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Helper function to validate that the given string is a
// proper base64 string value
func ValidScanId(s string) bool {
	if len(s) != 11 {
		return false
	}
	if _, err := base64.RawURLEncoding.DecodeString(s); err != nil {
		return false
	}
	return true
}

// Add a new scan record to the data store. Returns a pointer
// to the newly added scan record or nil and an error on failure.
func (ss *ScanStore) Insert(si *ScanInfo) (*ScanRecord, error) {
	txn := ss.DB.Txn(true)

	sr := ScanRecord{
		Id:   ss.NextId(),
		Info: si.Clone(),
	}

	if err := txn.Insert("scans", sr); err != nil {
		txn.Abort()
		return nil, errors.New(fmt.Sprintf("Error inserting data to the DB: %v\n", sr))
	}
	txn.Commit()
	ss.total++

	return &sr, nil
}

// Retrieve an existing scan record from the data store.
// Returns a pointer to a copy of the retrieved scan record
// or nil and an error on failure.
func (ss *ScanStore) Retrieve(id string) (*ScanRecord, error) {
	var sr ScanRecord

	txn := ss.DB.Txn(false)
	raw, err := txn.First("scans", "id", id)
	if err != nil {
		txn.Abort()
		return nil, errors.New(fmt.Sprintf("Error retrieving data from the DB. Id %v", id))
	}

	if raw == nil {
		return nil, errors.New(fmt.Sprintf("Id %v does not exist.", id))
	}

	sr = raw.(ScanRecord)
	return sr.Clone(), nil
}

// Delete an existing scan record from the data store.
// Returns nil on success or an error on failure.
func (ss *ScanStore) Delete(id string) error {
	txn := ss.DB.Txn(false)

	sr, err := ss.Retrieve(id)
	if err != nil {
		return errors.New(fmt.Sprintf("Id not found: %v", err.Error()))
	}

	txn = ss.DB.Txn(true)
	err = txn.Delete("scans", sr)
	if err != nil {
		txn.Abort()
		return errors.New(fmt.Sprintf("Failed to delete record: %v", err.Error()))
	}

	txn.Commit()
	ss.total--

	return nil
}

// Update an existing scan record in the data store.
// Returns nil on success or an error on failure.
func (ss *ScanStore) Update(sr *ScanRecord) error {
	txn := ss.DB.Txn(false)
	if _, err := ss.Retrieve(sr.Id); err != nil {
		return errors.New(fmt.Sprintf("Id not found: %v", err.Error()))
	}

	txn = ss.DB.Txn(true)
	if err := txn.Insert("scans", *(sr.Clone())); err != nil {
		txn.Abort()
		return errors.New("Update failed")
	}
	txn.Commit()

	return nil
}

// List returns a scan list based on the provided pagination
// parameters. It will return a maximum of page size repository
// records while skipping offset-1 records from the start of the
// data store.
func (ss *ScanStore) List(pp *PaginationParams) (*ScanList, error) {
	if pp.Offset > ss.total {
		return nil, errors.New("Invalid offset")
	}

	if pp.PageSize < 1 {
		return nil, errors.New("Invalid page size")
	}

	sl := ScanList{
		Total:      ss.total,
		Pagination: pp.Clone(),
		Items:      make([]ScanRecord, 0),
	}

	if ss.total == 0 {
		return &sl, nil
	}

	txn := ss.DB.Txn(false)
	it, err := txn.Get("scans", "id")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Cannot retrieve scan list: %v", err.Error()))
	}

	for i := int32(0); i < pp.Offset; i++ {
		_ = it.Next()
	}

	limit := ss.total - pp.Offset
	if limit > pp.PageSize {
		limit = pp.PageSize
	}

	for i := int32(0); i < limit; i++ {
		sr := it.Next().(ScanRecord)
		sl.Items = append(sl.Items, *sr.Clone())
	}

	return &sl, nil
}
