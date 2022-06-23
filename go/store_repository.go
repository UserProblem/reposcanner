package swagger

import (
	"errors"

	"github.com/UserProblem/reposcanner/models"
)

type RepoStore interface {
	NextId() int64
	Insert(ri *models.RepositoryInfo) (*models.RepositoryRecord, error)
	Retrieve(id int64) (*models.RepositoryRecord, error)
	Delete(id int64) error
	Update(rr *models.RepositoryRecord) error
	List(pp *models.PaginationParams) (*models.RepositoryList, error)
}

// Create and return a pointer to a new repository data store.
// Returns nil and an error on failure
func NewRepoStore(dbtype string) (RepoStore, error) {
	if dbtype == "postgresql" {
		return nil, errors.New("not implemented")
	} else {
		return NewRepoStoreMemDB()
	}
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
