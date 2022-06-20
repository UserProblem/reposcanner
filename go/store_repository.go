package swagger

import (
	"errors"
	"fmt"

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
		return nil, errors.New(fmt.Sprintf("Cannot initialize db: %s", err.Error()))
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
func (rs *RepoStore) Insert(ri *RepositoryInfo) (*RepositoryRecord, error) {
	txn := rs.DB.Txn(true)

	rr := RepositoryRecord{
		Id:   rs.NextId(),
		Info: ri.Clone(),
	}

	if err := txn.Insert("repositories", rr); err != nil {
		txn.Abort()
		return nil, errors.New(fmt.Sprintf("Error inserting data to the DB: %v\n", rr))
	}
	txn.Commit()
	rs.total++

	return &rr, nil
}

// Retrieve an existing repository record from the data store.
// Returns a pointer to a copy of the retrieved repository record
// or nil and an error on failure.
func (rs *RepoStore) Retrieve(id int64) (*RepositoryRecord, error) {
	var rr RepositoryRecord

	txn := rs.DB.Txn(false)
	raw, err := txn.First("repositories", "id", id)
	if err != nil {
		txn.Abort()
		return nil, errors.New(fmt.Sprintf("Error retrieving data from the DB. Id %v", id))
	}

	if raw == nil {
		return nil, errors.New(fmt.Sprintf("Id %v does not exist.", id))
	}

	rr = raw.(RepositoryRecord)
	return rr.Clone(), nil
}

// Delete an existing repository record from the data store.
// Returns nil on success or an error on failure.
func (rs *RepoStore) Delete(id int64) error {
	txn := rs.DB.Txn(false)

	rr, err := rs.Retrieve(id)
	if err != nil {
		return errors.New(fmt.Sprintf("Id not found: %v", err.Error()))
	}

	txn = rs.DB.Txn(true)
	err = txn.Delete("repositories", rr)
	if err != nil {
		txn.Abort()
		return errors.New(fmt.Sprintf("Failed to delete record: %v", err.Error()))
	}

	txn.Commit()
	rs.total--

	return nil
}

// Update an existing repository record in the data store.
// Returns nil on success or an error on failure.
func (rs *RepoStore) Update(rr *RepositoryRecord) error {
	txn := rs.DB.Txn(false)
	if _, err := rs.Retrieve(rr.Id); err != nil {
		return errors.New(fmt.Sprintf("Id not found: %v", err.Error()))
	}

	txn = rs.DB.Txn(true)
	if err := txn.Insert("repositories", *(rr.Clone())); err != nil {
		txn.Abort()
		return errors.New("Update failed")
	}
	txn.Commit()

	return nil
}

// List returns a repository list based on the provided pagination
// parameters. It will return a maximum of page size repository
// records while skipping offset-1 records from the start of the
// data store.
func (rs *RepoStore) List(pp *PaginationParams) (*RepositoryList, error) {
	if pp.Offset > rs.total {
		return nil, errors.New("Invalid offset")
	}

	if pp.PageSize < 1 {
		return nil, errors.New("Invalid page size")
	}

	rl := RepositoryList{
		Total:      rs.total,
		Pagination: pp.Clone(),
		Items:      make([]RepositoryRecord, 0),
	}

	if rs.total == 0 {
		return &rl, nil
	}

	txn := rs.DB.Txn(false)
	it, err := txn.Get("repositories", "id")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Cannot retrieve repository list: %v", err.Error()))
	}

	for i := int32(1); i < pp.Offset; i++ {
		_ = it.Next()
	}

	limit := rs.total - pp.Offset + 1
	if limit > pp.PageSize {
		limit = pp.PageSize
	}

	for i := int32(0); i < limit; i++ {
		rr := it.Next().(RepositoryRecord)
		rl.Items = append(rl.Items, *rr.Clone())
	}

	return &rl, nil
}
