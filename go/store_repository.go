package swagger

import (
	"errors"
	"fmt"

	"github.com/UserProblem/reposcanner/models"
	"github.com/hashicorp/go-memdb"
)

type RepoStore struct {
	DB     *memdb.MemDB
	nextId int64
	total  int32
}

/*
CREATE TABLE IF NOT EXISTS repositories
(
    id SERIAL,
    name TEXT NOT NULL,
	url TEXT NOT NULL,
	branch TEXT NOT NULL
    CONSTRAINT repositories_pkey PRIMARY KEY (id)
)
*/

// Create and return a pointer to a new repository data store.
// Returns nil and an error on failure
func NewRepoStore() (*RepoStore, error) {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"repositories": {
				Name: "repositories",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.IntFieldIndex{Field: "Id"},
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize db: %s", err.Error())
	}

	return &RepoStore{DB: db, nextId: 1, total: 0}, nil
}

// Helper function to auto-generate the next unique id value
// that can be used for new repository records.
func (rs *RepoStore) NextId() int64 {
	defer func() { rs.nextId++ }()
	return rs.nextId
}

// Add a new repository record to the data store. Returns a pointer
// to the newly added repository record or nil and an error on failure.
func (rs *RepoStore) Insert(ri *models.RepositoryInfo) (*models.RepositoryRecord, error) {
	txn := rs.DB.Txn(true)

	rr := models.RepositoryRecord{
		Id:   rs.NextId(),
		Info: ri.Clone(),
	}

	if err := txn.Insert("repositories", rr); err != nil {
		txn.Abort()
		return nil, fmt.Errorf("error inserting data to the DB: %v", rr)
	}
	txn.Commit()
	rs.total++

	return &rr, nil
}

// Retrieve an existing repository record from the data store.
// Returns a pointer to a copy of the retrieved repository record
// or nil and an error on failure.
func (rs *RepoStore) Retrieve(id int64) (*models.RepositoryRecord, error) {
	var rr models.RepositoryRecord

	txn := rs.DB.Txn(false)
	raw, err := txn.First("repositories", "id", id)
	if err != nil {
		txn.Abort()
		return nil, fmt.Errorf("error retrieving data from the DB. Id %v", id)
	}

	if raw == nil {
		return nil, fmt.Errorf("id %v does not exist", id)
	}

	rr = raw.(models.RepositoryRecord)
	return rr.Clone(), nil
}

// Delete an existing repository record from the data store.
// Returns nil on success or an error on failure.
func (rs *RepoStore) Delete(id int64) error {
	rr, err := rs.Retrieve(id)
	if err != nil {
		return fmt.Errorf("id not found: %v", err.Error())
	}

	txn := rs.DB.Txn(true)
	err = txn.Delete("repositories", rr)
	if err != nil {
		txn.Abort()
		return fmt.Errorf("failed to delete record: %v", err.Error())
	}

	txn.Commit()
	rs.total--

	return nil
}

// Update an existing repository record in the data store.
// Returns nil on success or an error on failure.
func (rs *RepoStore) Update(rr *models.RepositoryRecord) error {
	if _, err := rs.Retrieve(rr.Id); err != nil {
		return fmt.Errorf("id not found: %v", err.Error())
	}

	txn := rs.DB.Txn(true)
	if err := txn.Insert("repositories", *(rr.Clone())); err != nil {
		txn.Abort()
		return errors.New("update failed")
	}
	txn.Commit()

	return nil
}

// List returns a repository list based on the provided pagination
// parameters. It will return a maximum of page size repository
// records while skipping offset-1 records from the start of the
// data store.
func (rs *RepoStore) List(pp *models.PaginationParams) (*models.RepositoryList, error) {
	if pp.Offset > rs.total {
		return nil, errors.New("invalid offset")
	}

	if pp.PageSize < 1 {
		return nil, errors.New("invalid page size")
	}

	rl := models.RepositoryList{
		Total:      rs.total,
		Pagination: pp.Clone(),
		Items:      make([]models.RepositoryRecord, 0),
	}

	if rs.total == 0 {
		return &rl, nil
	}

	txn := rs.DB.Txn(false)
	it, err := txn.Get("repositories", "id")
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve repository list: %v", err.Error())
	}

	for i := int32(0); i < pp.Offset; i++ {
		_ = it.Next()
	}

	limit := rs.total - pp.Offset
	if limit > pp.PageSize {
		limit = pp.PageSize
	}

	for i := int32(0); i < limit; i++ {
		rr := it.Next().(models.RepositoryRecord)
		rl.Items = append(rl.Items, *rr.Clone())
	}

	return &rl, nil
}
