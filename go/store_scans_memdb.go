package swagger

import (
	"errors"
	"fmt"

	"github.com/UserProblem/reposcanner/models"
	"github.com/hashicorp/go-memdb"
)

type ScanStoreMemDB struct {
	DB             *memdb.MemDB
	nextId         chan uint64
	total          int32
	nextFindingsId chan uint64
}

func NewScanStoreMemDB() (ScanStore, error) {
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
			"findings": {
				Name: "findings",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.IntFieldIndex{Field: "Id"},
					},
					"scanid": {
						Name:    "scanid",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "ScanId", Lowercase: false},
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize db: %s", err.Error())
	}

	chS, chF := make(chan uint64), make(chan uint64)
	go generateObjIds(chS)
	go generateObjIds(chF)

	return &ScanStoreMemDB{
		DB:             db,
		nextId:         chS,
		total:          0,
		nextFindingsId: chF,
	}, nil
}

// Helper function to auto-generate the next unique id value
// that can be used for new scan records.
func (ss *ScanStoreMemDB) NextId() string {
	return EncodeScanId(<-ss.nextId)
}

// Add a new scan record to the data store. Returns a pointer
// to the newly added scan record or nil and an error on failure.
func (ss *ScanStoreMemDB) Insert(si *models.ScanInfo) (*models.ScanRecord, error) {
	txn := ss.DB.Txn(true)

	sr := models.ScanRecord{
		Id:   ss.NextId(),
		Info: si.Clone(),
	}

	if err := txn.Insert("scans", sr); err != nil {
		txn.Abort()
		return nil, fmt.Errorf("error inserting data to the DB: %v", sr)
	}
	txn.Commit()
	ss.total++

	return &sr, nil
}

// Retrieve an existing scan record from the data store.
// Returns a pointer to a copy of the retrieved scan record
// or nil and an error on failure.
func (ss *ScanStoreMemDB) Retrieve(id string) (*models.ScanRecord, error) {
	var sr models.ScanRecord

	txn := ss.DB.Txn(false)
	raw, err := txn.First("scans", "id", id)
	if err != nil {
		txn.Abort()
		return nil, fmt.Errorf("error retrieving data from the DB. Id %v", id)
	}

	if raw == nil {
		return nil, fmt.Errorf("id %v does not exist", id)
	}

	sr = raw.(models.ScanRecord)
	return sr.Clone(), nil
}

// Delete an existing scan record from the data store.
// Returns nil on success or an error on failure.
func (ss *ScanStoreMemDB) Delete(id string) error {
	sr, err := ss.Retrieve(id)
	if err != nil {
		return fmt.Errorf("id not found: %v", err.Error())
	}

	txn := ss.DB.Txn(true)
	err = txn.Delete("scans", sr)
	if err != nil {
		txn.Abort()
		return fmt.Errorf("failed to delete record: %v", err.Error())
	}

	txn.Commit()
	ss.total--

	return nil
}

// Update an existing scan record in the data store.
// Returns nil on success or an error on failure.
func (ss *ScanStoreMemDB) Update(sr *models.ScanRecord) error {
	if _, err := ss.Retrieve(sr.Id); err != nil {
		return fmt.Errorf("id not found: %v", err.Error())
	}

	txn := ss.DB.Txn(true)
	if err := txn.Insert("scans", *(sr.Clone())); err != nil {
		txn.Abort()
		return errors.New("update failed")
	}
	txn.Commit()

	return nil
}

// List returns a scan list based on the provided pagination
// parameters. It will return a maximum of page size repository
// records while skipping offset-1 records from the start of the
// data store.
func (ss *ScanStoreMemDB) List(pp *models.PaginationParams) (*models.ScanList, error) {
	if pp.Offset > ss.total {
		return nil, errors.New("invalid offset")
	}

	if pp.PageSize < 1 {
		return nil, errors.New("invalid page size")
	}

	sl := models.ScanList{
		Total:      ss.total,
		Pagination: pp.Clone(),
		Items:      make([]models.ScanRecord, 0),
	}

	if ss.total == 0 {
		return &sl, nil
	}

	txn := ss.DB.Txn(false)
	it, err := txn.Get("scans", "id")
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve scan list: %v", err.Error())
	}

	for i := int32(0); i < pp.Offset; i++ {
		_ = it.Next()
	}

	limit := ss.total - pp.Offset
	if limit > pp.PageSize {
		limit = pp.PageSize
	}

	for i := int32(0); i < limit; i++ {
		sr := it.Next().(models.ScanRecord)
		sl.Items = append(sl.Items, *sr.Clone())
	}

	return &sl, nil
}

// Helper function to auto-generate the next unique id value
// that can be used for new findings records.
func (ss *ScanStoreMemDB) NextFindingsId() int {
	return int(<-ss.nextFindingsId)
}

// InsertFindings stores all of the contents of the findings
// list into the data store, indexed by scanId. All operations
// related to findings are done in bulk. It returns nil on success
// or an error on failure, at which point none of the findings
// will be stored.
func (ss *ScanStoreMemDB) InsertFindings(scanId string, findings []*models.FindingsInfo) error {
	if len(findings) == 0 {
		return nil
	}

	txn := ss.DB.Txn(true)

	for _, fi := range findings {
		fr := models.FindingsRecord{
			Id:      ss.NextFindingsId(),
			ScanId:  scanId,
			Finding: fi,
		}

		if err := txn.Insert("findings", fr); err != nil {
			txn.Abort()
			return fmt.Errorf("error inserting data to the DB: %v", fr)
		}
	}

	txn.Commit()
	return nil
}

// ListFindings retrieves all of the findings from the data store
// indexed by scanId. All operations related to findings are done
// in bulk. It returns a list of findings on success, or nil and
// an error on failure.
func (ss *ScanStoreMemDB) ListFindings(scanId string) ([]*models.FindingsInfo, error) {
	txn := ss.DB.Txn(false)

	it, err := txn.Get("findings", "scanid", scanId)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve findings list: %v", err.Error())
	}

	findings := make([]*models.FindingsInfo, 0)
	for fr := it.Next(); fr != nil; fr = it.Next() {
		fi := fr.(models.FindingsRecord).Finding
		findings = append(findings, fi)
	}

	return findings, nil
}

// DeleteFindings deletes all of the findings from the data store
// indexed by scanId. All operations related to findings are done
// in bulk. It returns the number of deleted records on success,
// or zero and an error on failure, at which point none of the findings
// will be deleted.
func (ss *ScanStoreMemDB) DeleteFindings(scanId string) (int, error) {
	txn := ss.DB.Txn(true)

	count, err := txn.DeleteAll("findings", "scanid", scanId)
	if err != nil {
		txn.Abort()
		return 0, fmt.Errorf("cannot delete all findings: %v", err.Error())
	}

	txn.Commit()
	return count, nil
}
