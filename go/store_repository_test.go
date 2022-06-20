package swagger_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	sw "github.com/UserProblem/reposcanner/go"
)

func initializeRepoStore(t *testing.T) *sw.RepoStore {
	rs, err := sw.NewRepoStore()
	if err != nil {
		t.Errorf("Failed to initialize repository store.")
		t.FailNow()
	}
	return rs
}

func TestNextIdIncrements(t *testing.T) {
	rs := initializeRepoStore(t)

	first := rs.NextId()
	second := rs.NextId()

	if first != 1 {
		t.Errorf("Expected first id is 1. Got %v\n", first)
	}

	if second != first+1 {
		t.Errorf("Expected second id is one more than first id. Got second=%v, first=%v\n", second, first)
	}
}

func TestStoreNewRepositoryInfo(t *testing.T) {
	rs := initializeRepoStore(t)

	ri := sw.DefaultRepositoryInfo()
	rr, err := rs.Insert(ri)

	if err != nil {
		t.Errorf("Failed to insert repository info into the database.\n")
		t.FailNow()
	}

	if rr.Id != 1 {
		t.Errorf("Expected id is 1. Got %v\n", rr.Id)
	}

	if rr.Info.Name != ri.Name {
		t.Errorf("Expected name is '%v'. Got '%v'\n", ri.Name, rr.Info.Name)
	}

	if rr.Info.Url != ri.Url {
		t.Errorf("Expected branch is '%v'. Got '%v'\n", ri.Url, rr.Info.Url)
	}

	if rr.Info.Branch != ri.Branch {
		t.Errorf("Expected branch is %v. Got %v\n", ri.Branch, rr.Info.Branch)
	}
}

func TestRetrieveRepositoryRecord(t *testing.T) {
	rs := initializeRepoStore(t)

	ri := sw.DefaultRepositoryInfo()
	rr, err := rs.Insert(ri)

	if err != nil {
		t.Errorf("Failed to insert repository info into the database.\n")
		t.FailNow()
	}

	rr, err = rs.Retrieve(1)
	if err != nil {
		t.Errorf("Failed to retrieve repository record from the database.\n")
		t.FailNow()
	}

	if rr.Id != 1 {
		t.Errorf("Expected id is 1. Got %v\n", rr.Id)
	}

	if rr.Info.Name != ri.Name {
		t.Errorf("Expected name is '%v'. Got '%v'\n", ri.Name, rr.Info.Name)
	}

	if rr.Info.Url != ri.Url {
		t.Errorf("Expected branch is '%v'. Got '%v'\n", ri.Url, rr.Info.Url)
	}

	if rr.Info.Branch != ri.Branch {
		t.Errorf("Expected branch is %v. Got %v\n", ri.Branch, rr.Info.Branch)
	}
}

func TestRetrieveInvalidRepositoryRecord(t *testing.T) {
	rs := initializeRepoStore(t)

	if _, err := rs.Retrieve(1); err == nil {
		t.Errorf("Expected error but got successful result.\n")
	}
}

func TestDeleteRepositoryRecord(t *testing.T) {
	rs := initializeRepoStore(t)

	ri := sw.DefaultRepositoryInfo()
	_, err := rs.Insert(ri)

	if err != nil {
		t.Errorf("Failed to insert repository info into the database.\n")
		t.FailNow()
	}

	if err := rs.Delete(1); err != nil {
		t.Errorf("Failed to delete repository record from the database (%s)\n", err.Error())
		t.FailNow()
	}

	err = rs.Delete(1)
	if err == nil {
		t.Errorf("Expected error result but got a valid record.\n")
		t.FailNow()
	}

	if !strings.HasPrefix(err.Error(), "Id not found") {
		t.Errorf("Expected error message to be 'Id does not exist'. Got '%v'\n", err.Error())
	}
}

func TestDeleteInvalidRepositoryRecord(t *testing.T) {
	rs := initializeRepoStore(t)

	if err := rs.Delete(1); err == nil {
		t.Errorf("Expected error but got successful result.\n")
	}
}

func TestUpdateRepositoryRecord(t *testing.T) {
	rs := initializeRepoStore(t)

	ri := sw.DefaultRepositoryInfo()

	rrOrig, err := rs.Insert(ri)
	if err != nil {
		t.Errorf("Failed to insert repository info into the database.\n")
		t.FailNow()
	}

	rrMod := &sw.RepositoryRecord{
		Id: rrOrig.Id,
		Info: &sw.RepositoryInfo{
			Name:   "modified repo",
			Url:    "modified url",
			Branch: "modified",
		},
	}

	if err := rs.Update(rrMod); err != nil {
		t.Errorf("Failed to update repository info into the database (%v).\n", err.Error())
		t.FailNow()
	}

	rrMod, err = rs.Retrieve(rrOrig.Id)
	if err != nil {
		t.Errorf("Failed to retrieve repository record from the database.\n")
		t.FailNow()
	}

	if rrMod.Info.Name != "modified repo" {
		t.Errorf("Expected name to be 'mmodified repo'. Got '%v'\n", rrMod.Info.Name)
	}

	if rrMod.Info.Url != "modified url" {
		t.Errorf("Expected url to be 'mmodified url'. Got '%v'\n", rrMod.Info.Url)
	}

	if rrMod.Info.Branch != "modified" {
		t.Errorf("Expected branch to be 'mmodified'. Got '%v'\n", rrMod.Info.Branch)
	}
}

func TestUpdateInvalidRepositoryRecord(t *testing.T) {
	rs := initializeRepoStore(t)

	rr := sw.RepositoryRecord{
		Id: 1,
		Info: sw.DefaultRepositoryInfo(),
	}

	if err := rs.Update(&rr); err == nil {
		t.Errorf("Expected error but got successful result.\n")
	}
}

func TestRetrieveRepositoryList(t *testing.T) {
	rs := initializeRepoStore(t)

	var ri *sw.RepositoryInfo
	var err error

	var totalRecords, i int32 = 10, 1
	for ; i <= totalRecords; i++ {
		ri = sw.DefaultRepositoryInfo()
		ri.Name = ri.Name + fmt.Sprintf(" %v", i)
		ri.Url = ri.Url + fmt.Sprintf(" %v", i)

		if _, err = rs.Insert(ri); err != nil {
			t.Errorf(err.Error())
			t.FailNow()
		}
	}

	rl, err := rs.List(&sw.PaginationParams{Offset: 2, PageSize: 5})
	if err != nil {
		t.Errorf("Failed to retrieve repository list: %v", err.Error())
		t.FailNow()
	}

	if rl.Total != totalRecords {
		t.Errorf("Expected total to be %v. Got %v\n", totalRecords, rl.Total)
	}

	if rl.Pagination.Offset != 2 {
		t.Errorf("Expected offset to be 2. Got %v\n", rl.Pagination.Offset)
	}

	if rl.Pagination.PageSize != 5 {
		t.Errorf("Expected pagesize to be 5. Got %v\n", rl.Pagination.PageSize)
	}

	if len(rl.Items) != 5 {
		t.Errorf("Expected total number of items to be 5. Got %v\n", len(rl.Items))
	}

	var j int64 = 2
	for k := range rl.Items {
		rr := rl.Items[k]
		if rr.Id != j {
			t.Errorf("Expected id to be %v. Got %v\n", j, rr.Id)
		}

		if !strings.HasSuffix(rr.Info.Name, strconv.FormatInt(j, 10)) {
			t.Errorf("Expected name to end in %v. Got %v\n", j, rr.Info.Name)
		}

		if !strings.HasSuffix(rr.Info.Url, strconv.FormatInt(j, 10)) {
			t.Errorf("Expected url to end in %v. Got %v\n", j, rr.Info.Url)
		}

		j++
	}
}

func TestRetrieveEmptyRepositoryList(t *testing.T) {
	rs := initializeRepoStore(t)

	rl, err := rs.List(&sw.PaginationParams{Offset: 0, PageSize: 20})
	if err != nil {
		t.Errorf("Expected successful operation. Got error: %v\n", err.Error())
		t.FailNow()
	}

	if rl.Total != 0 {
		t.Errorf("Expected total to be 0. Got %v\n", rl.Total)
	}

	if len(rl.Items) != 0 {
		t.Errorf("Expected items to be empty. Got %v\n", rl.Items)
	}
}

func TestRetrieveRepositoryListInvalidOffset(t *testing.T) {
	rs := initializeRepoStore(t)
	
	_, err := rs.List(&sw.PaginationParams{Offset: 2, PageSize: 5})
	if err.Error() != "Invalid offset" {
		t.Errorf("Expected error 'Invalid offset'. Got '%v'\n", err.Error())
	}
}

func TestRetrieveRepositoryListInvalidPageSize(t *testing.T) {
	rs := initializeRepoStore(t)

	_, err := rs.List(&sw.PaginationParams{Offset: 0, PageSize: 0})
	if err.Error() != "Invalid page size" {
		t.Errorf("Expected error 'Invalid page size'. Got '%v'\n", err.Error())
	}
}
