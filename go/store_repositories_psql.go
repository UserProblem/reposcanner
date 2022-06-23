package swagger

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/UserProblem/reposcanner/models"
)

type RepoStorePsql struct {
	DB *sql.DB
}

func NewRepoStorePsqlDB() (RepoStore, error) {
	actualDB := GetPsqlDBInstance().DB
	if actualDB == nil {
		return nil, fmt.Errorf("cannot retrieve psql instance")
	}

	createTableQuery := `CREATE TABLE IF NOT EXISTS repositories 
	(
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		branch TEXT NOT NULL
	)`

	if _, err := actualDB.Exec(createTableQuery); err != nil {
		return nil, fmt.Errorf("could not create table 'repositories': %v", err.Error())
	}

	return &RepoStorePsql{
		DB: actualDB,
	}, nil
}

// Add a new repository record to the data store. Returns a pointer
// to the newly added repository record or nil and an error on failure.
func (rs *RepoStorePsql) Insert(ri *models.RepositoryInfo) (*models.RepositoryRecord, error) {
	var id int

	err := rs.DB.QueryRow(
		"INSERT INTO repositories(name, url, branch) VALUES ($1, $2, $3) RETURNING id",
		ri.Name, ri.Url, ri.Branch).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("error inserting data to the DB")
	}

	return &models.RepositoryRecord{
		Id:   int64(id),
		Info: ri.Clone(),
	}, nil
}

// Retrieve an existing repository record from the data store.
// Returns a pointer to a copy of the retrieved repository record
// or nil and an error on failure.
func (rs *RepoStorePsql) Retrieve(id int64) (*models.RepositoryRecord, error) {
	var ri models.RepositoryInfo

	err := rs.DB.QueryRow("SELECT name, url, branch FROM repositories WHERE id=$1",
		int(id)).Scan(&ri.Name, &ri.Url, &ri.Branch)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("id %v does not exist", id)
		} else {
			return nil, fmt.Errorf("error retrieving data from the DB. Id %v", id)
		}
	}

	return &models.RepositoryRecord{
		Id:   id,
		Info: ri.Clone(),
	}, nil
}

// Delete an existing repository record from the data store.
// Returns nil on success or an error on failure.
func (rs *RepoStorePsql) Delete(id int64) error {
	res, err := rs.DB.Exec("DELETE FROM repositories WHERE id=$1", int(id))

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

// Update an existing repository record in the data store.
// Returns nil on success or an error on failure.
func (rs *RepoStorePsql) Update(rr *models.RepositoryRecord) error {
	res, err := rs.DB.Exec("UPDATE repositories SET name=$1, url=$2, branch=$3 WHERE id=$4",
		rr.Info.Name, rr.Info.Url, rr.Info.Branch, rr.Id)

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

// List returns a repository list based on the provided pagination
// parameters. It will return a maximum of page size repository
// records while skipping offset-1 records from the start of the
// data store.
func (rs *RepoStorePsql) List(pp *models.PaginationParams) (*models.RepositoryList, error) {
	var total int
	err := rs.DB.QueryRow("SELECT count(*) AS row_count FROM repositories").Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve repository list: %v", err.Error())
	}

	if pp.Offset > int32(total) {
		return nil, errors.New("invalid offset")
	}

	if pp.PageSize < 1 {
		return nil, errors.New("invalid page size")
	}

	rows, err := rs.DB.Query(
		"SELECT id, name, url, branch FROM repositories LIMIT $1 OFFSET $2",
		int(pp.PageSize), int(pp.Offset))

	if err != nil {
		return nil, fmt.Errorf("cannot retrieve repository list: %v", err.Error())
	}

	defer rows.Close()

	rl := models.RepositoryList{
		Total:      int32(total),
		Pagination: pp.Clone(),
		Items:      make([]models.RepositoryRecord, 0),
	}

	for rows.Next() {
		var rr models.RepositoryRecord
		var ri models.RepositoryInfo

		if err := rows.Scan(&rr.Id, &ri.Name, &ri.Url, &ri.Branch); err != nil {
			return nil, fmt.Errorf("cannot retrieve repository list: %v", err.Error())
		}

		rr.Info = &ri
		rl.Items = append(rl.Items, rr)
	}

	return &rl, nil
}
