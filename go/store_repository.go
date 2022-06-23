package swagger

import (
	"github.com/UserProblem/reposcanner/models"
)

type RepoStore interface {
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
		return NewRepoStorePsqlDB()
	} else {
		return NewRepoStoreMemDB()
	}
}
