package swagger_test

import (
	"os"
	"strings"
	"testing"
	"time"

	sw "github.com/UserProblem/reposcanner/go"
	"github.com/UserProblem/reposcanner/models"
)

func DropScanTable(t *testing.T) {
	PDB := sw.GetPsqlDBInstance()
	if _, err := PDB.DB.Exec("DROP TABLE IF EXISTS scans CASCADE"); err != nil {
		t.Fatalf("Failed to drop scans table: %v\n", err.Error())
	}
	if _, err := PDB.DB.Exec("DROP TABLE IF EXISTS findings CASCADE"); err != nil {
		t.Fatalf("Failed to drop findings table: %v\n", err.Error())
	}
}

func initializeScanStore(t *testing.T) sw.ScanStore {
	dbtype := os.Getenv("DATABASE_TYPE")
	if dbtype == "postgresql" {
		DropScanTable(t)
		DropRepoTable(t)
	}

	if _, err := sw.NewRepoStore(dbtype); err != nil {
		t.Fatalf("Failed to initialize repo store: %v\n", err.Error())
	}

	ss, err := sw.NewScanStore(dbtype)
	if err != nil {
		t.Fatalf("Failed to initialize scan store: %v\n", err.Error())
	}

	return ss
}

func addDummyRepo(t *testing.T) {
	dbtype := os.Getenv("DATABASE_TYPE")
	rs, err := sw.NewRepoStore(dbtype)

	if err != nil {
		t.Fatalf("Failed to initialize repo store: %v\n", err.Error())
	}

	_, err = rs.Insert(&models.RepositoryInfo{
		Name:   "test repo",
		Url:    "http://example.com/test/repo",
		Branch: "main",
	})

	if err != nil {
		t.Fatalf("Failed to insert repo to the repo store: %v\n", err.Error())
	}
}

func addDummyScan(t *testing.T, ss sw.ScanStore) string {
	sr, err := ss.Insert(&models.ScanInfo{
		RepoId:     int64(1),
		QueuedAt:   "1970-01-01T00:00:01Z",
		ScanningAt: "",
		FinishedAt: "",
		Status:     "QUEUED",
	})

	if err != nil {
		t.Fatalf("Failed to insert scan to the scan store: %v\n", err.Error())
	}

	return sr.Id
}

func checkTimestampsEquivalent(first, second string) bool {
	var t1, t2 time.Time
	var err error

	if first == "" && second == "" {
		return true
	}

	if t1, err = time.Parse("2006-01-02T15:04:05Z07:00", first); err != nil {
		return false
	}

	if t2, err = time.Parse("2006-01-02T15:04:05Z07:00", second); err != nil {
		return false
	}

	return t1.Equal(t2)
}

func TestStoreNewScanInfo(t *testing.T) {
	ss := initializeScanStore(t)
	addDummyRepo(t)

	si := models.DefaultScanInfo()
	sr, err := ss.Insert(si)

	if err != nil {
		t.Fatalf("Failed to insert scan info into the database: %v\n", err.Error())
	}

	if sr.Id != sw.EncodeScanId(1) {
		t.Errorf("Expected id is %v. Got %v\n", sw.EncodeScanId(1), sr.Id)
	}

	if sr.Info.RepoId != si.RepoId {
		t.Errorf("Expected repo id is %v. Got %v\n", si.RepoId, sr.Info.RepoId)
	}

	if !checkTimestampsEquivalent(sr.Info.QueuedAt, si.QueuedAt) {
		t.Errorf("Expected queued at is '%v'. Got '%v'\n", si.QueuedAt, sr.Info.QueuedAt)
	}

	if !checkTimestampsEquivalent(sr.Info.ScanningAt, si.ScanningAt) {
		t.Errorf("Expected scanning at is '%v'. Got '%v'\n", si.ScanningAt, sr.Info.ScanningAt)
	}

	if !checkTimestampsEquivalent(sr.Info.FinishedAt, si.FinishedAt) {
		t.Errorf("Expected finished at is '%v'. Got '%v'\n", si.FinishedAt, sr.Info.FinishedAt)
	}

	if sr.Info.Status != si.Status {
		t.Errorf("Expected status is %v. Got %v\n", si.Status, sr.Info.Status)
	}
}

func TestRetrieveScanRecord(t *testing.T) {
	ss := initializeScanStore(t)
	addDummyRepo(t)

	si := models.DefaultScanInfo()
	_, err := ss.Insert(si)

	if err != nil {
		t.Errorf("Failed to insert scan info into the database: %v\n", err.Error())
		t.FailNow()
	}

	var sr *models.ScanRecord
	sr, err = ss.Retrieve(sw.EncodeScanId(1))
	if err != nil {
		t.Errorf("Failed to retrieve scan record from the database: %v\n", err.Error())
		t.FailNow()
	}

	if sr.Id != sw.EncodeScanId(1) {
		t.Errorf("Expected id is %v. Got %v\n", sw.EncodeScanId(1), sr.Id)
	}

	if sr.Info.RepoId != si.RepoId {
		t.Errorf("Expected repo id is %v. Got %v\n", si.RepoId, sr.Info.RepoId)
	}

	if !checkTimestampsEquivalent(sr.Info.QueuedAt, si.QueuedAt) {
		t.Errorf("Expected queued at is '%v'. Got '%v'\n", si.QueuedAt, sr.Info.QueuedAt)
	}

	if !checkTimestampsEquivalent(sr.Info.ScanningAt, si.ScanningAt) {
		t.Errorf("Expected scanning at is '%v'. Got '%v'\n", si.ScanningAt, sr.Info.ScanningAt)
	}

	if !checkTimestampsEquivalent(sr.Info.FinishedAt, si.FinishedAt) {
		t.Errorf("Expected finished at is '%v'. Got '%v'\n", si.FinishedAt, sr.Info.FinishedAt)
	}

	if sr.Info.Status != si.Status {
		t.Errorf("Expected status is %v. Got %v\n", si.Status, sr.Info.Status)
	}
}

func TestRetrieveInvalidScanRecord(t *testing.T) {
	ss := initializeScanStore(t)

	if _, err := ss.Retrieve(sw.EncodeScanId(1)); err == nil {
		t.Errorf("Expected error but got successful result.\n")
	}
}

func TestDeleteScanRecord(t *testing.T) {
	ss := initializeScanStore(t)
	addDummyRepo(t)

	si := models.DefaultScanInfo()
	_, err := ss.Insert(si)

	if err != nil {
		t.Errorf("Failed to insert scan info into the database.\n")
		t.FailNow()
	}

	if err := ss.Delete(sw.EncodeScanId(1)); err != nil {
		t.Fatalf("Failed to delete scan record from the database (%s)\n", err.Error())
	}

	err = ss.Delete(sw.EncodeScanId(1))
	if err == nil {
		t.Fatalf("Expected error result but got a valid record.\n")
	}

	if !strings.HasPrefix(err.Error(), "id not found") {
		t.Errorf("Expected error message to be 'Id not found'. Got '%v'\n", err.Error())
	}
}

func TestDeleteInvalidScanRecord(t *testing.T) {
	ss := initializeScanStore(t)

	if err := ss.Delete(sw.EncodeScanId(1)); err == nil {
		t.Errorf("Expected error but got successful result.\n")
	}
}

func TestUpdateScanRecord(t *testing.T) {
	ss := initializeScanStore(t)
	addDummyRepo(t)

	si := models.DefaultScanInfo()

	ssOrig, err := ss.Insert(si)
	if err != nil {
		t.Fatalf("Failed to insert scan info into the database.\n")
	}

	ssMod := &models.ScanRecord{
		Id: ssOrig.Id,
		Info: &models.ScanInfo{
			RepoId:     ssOrig.Info.RepoId,
			QueuedAt:   ssOrig.Info.QueuedAt,
			ScanningAt: "1970-01-01T00:00:01Z",
			FinishedAt: "",
			Status:     "IN PROGRESS",
		},
	}

	if err := ss.Update(ssMod); err != nil {
		t.Fatalf("Failed to update scan info into the database (%v).\n", err.Error())
	}

	ssMod, err = ss.Retrieve(ssOrig.Id)
	if err != nil {
		t.Fatalf("Failed to retrieve scan record from the database.\n")
	}

	if !checkTimestampsEquivalent(ssMod.Info.QueuedAt, ssOrig.Info.QueuedAt) {
		t.Errorf("Expected queued at is '%v'. Got '%v'\n", ssOrig.Info.QueuedAt, ssMod.Info.QueuedAt)
	}

	if !checkTimestampsEquivalent(ssMod.Info.ScanningAt, "1970-01-01T00:00:01Z") {
		t.Errorf("Expected scanning at to be '1970-01-01T00:00:01Z'. Got '%v'\n", ssMod.Info.ScanningAt)
	}

	if ssMod.Info.Status != "IN PROGRESS" {
		t.Errorf("Expected status to be 'IN PROGRESS'. Got '%v'\n", ssMod.Info.Status)
	}
}

func TestUpdateInvalidScanRecord(t *testing.T) {
	ss := initializeScanStore(t)
	addDummyRepo(t)

	sr := models.ScanRecord{
		Id:   sw.EncodeScanId(1),
		Info: models.DefaultScanInfo(),
	}

	if err := ss.Update(&sr); err == nil {
		t.Errorf("Expected error but got successful result.\n")
	}
}

func TestRetrieveScanList(t *testing.T) {
	ss := initializeScanStore(t)

	var si *models.ScanInfo
	var err error

	var totalRecords, i int32 = 10, 1
	for ; i <= totalRecords; i++ {
		addDummyRepo(t)
		si = models.DefaultScanInfo()
		si.RepoId = int64(i)

		if _, err = ss.Insert(si); err != nil {
			t.Fatalf(err.Error())
		}
	}

	sl, err := ss.List(&models.PaginationParams{Offset: 2, PageSize: 5})
	if err != nil {
		t.Fatalf("Failed to retrieve scan list: %v", err.Error())
	}

	if sl.Total != totalRecords {
		t.Errorf("Expected total to be %v. Got %v\n", totalRecords, sl.Total)
	}

	if sl.Pagination.Offset != 2 {
		t.Errorf("Expected offset to be 2. Got %v\n", sl.Pagination.Offset)
	}

	if sl.Pagination.PageSize != 5 {
		t.Errorf("Expected pagesize to be 5. Got %v\n", sl.Pagination.PageSize)
	}

	if len(sl.Items) != 5 {
		t.Errorf("Expected total number of items to be 5. Got %v\n", len(sl.Items))
	}

	var j int64 = 3
	for k := range sl.Items {
		sr := sl.Items[k]
		if sr.Info.RepoId != j {
			t.Errorf("Expected repo id to be %v. Got %v\n", j, sr.Info.RepoId)
		}
		j++
	}
}

func TestRetrieveEmptyScanList(t *testing.T) {
	ss := initializeScanStore(t)

	sl, err := ss.List(&models.PaginationParams{Offset: 0, PageSize: 20})
	if err != nil {
		t.Fatalf("Expected successful operation. Got error: %v\n", err.Error())
	}

	if sl.Total != 0 {
		t.Errorf("Expected total to be 0. Got %v\n", sl.Total)
	}

	if len(sl.Items) != 0 {
		t.Errorf("Expected items to be empty. Got %v\n", sl.Items)
	}
}

func TestRetrieveScanListInvalidOffset(t *testing.T) {
	ss := initializeScanStore(t)

	_, err := ss.List(&models.PaginationParams{Offset: 2, PageSize: 5})
	if err.Error() != "invalid offset" {
		t.Errorf("Expected error 'invalid offset'. Got '%v'\n", err.Error())
	}
}

func TestRetrieveScanListInvalidPageSize(t *testing.T) {
	ss := initializeScanStore(t)

	_, err := ss.List(&models.PaginationParams{Offset: 0, PageSize: 0})
	if err.Error() != "invalid page size" {
		t.Errorf("Expected error 'invalid page size'. Got '%v'\n", err.Error())
	}
}

func TestAddEmptyFindingsList(t *testing.T) {
	ss := initializeScanStore(t)

	if err := ss.InsertFindings("id", nil); err != nil {
		t.Fatalf("Expected insert nil findings is successful. Got error %v\n", err.Error())
	}

	if err := ss.InsertFindings("id", make([]*models.FindingsInfo, 0)); err != nil {
		t.Fatalf("Expected insert empty findings is successful. Got error %v\n", err.Error())
	}
}

func makeFindingsList(n int) []*models.FindingsInfo {
	findings := make([]*models.FindingsInfo, n)

	for i := 0; i < n; i++ {
		findings[i] = &models.FindingsInfo{
			Type_:  "sast",
			RuleId: "G001",
			Location: &models.FindingsLocation{
				Path: "hello.go",
				Positions: &models.FileLocation{
					Begin: &models.LineLocation{Line: int32(i)},
				},
			},
			Metadata: &models.FindingsMetadata{
				Description: "Hard-coded secret - public key",
				Severity:    "HIGH",
			},
		}
	}

	return findings
}

func TestAddFindingsList(t *testing.T) {
	ss := initializeScanStore(t)
	addDummyRepo(t)
	scanId := addDummyScan(t, ss)

	findings := makeFindingsList(10)

	if err := ss.InsertFindings(scanId, findings); err != nil {
		t.Errorf("Failed to insert findings list: %v\n", err.Error())
	}
}

func TestDeleteFindingsNonExistentScanId(t *testing.T) {
	ss := initializeScanStore(t)

	if n, err := ss.DeleteFindings("42"); err != nil {
		t.Fatalf("Expected no error. Got %v\n", err)
	} else {
		if n != 0 {
			t.Errorf("Expected 0 deleted count. Got %v\n", n)
		}
	}
}

func TestDeleteFindings(t *testing.T) {
	ss := initializeScanStore(t)
	addDummyRepo(t)
	scanId := addDummyScan(t, ss)

	findings := makeFindingsList(10)
	if err := ss.InsertFindings(scanId, findings); err != nil {
		t.Errorf("Failed to insert findings list: %v\n", err.Error())
	}

	if n, err := ss.DeleteFindings(scanId); err != nil {
		t.Fatalf("Expected no error. Got %v\n", err.Error())
	} else {
		if n != 10 {
			t.Errorf("Expected 10 items deleted. Got %v\n", n)
		}
	}
}

func TestListFindingsNonExistentScanId(t *testing.T) {
	ss := initializeScanStore(t)

	if findings, err := ss.ListFindings("42"); err != nil {
		t.Fatalf("Expected no error. Got %v\n", err.Error())
	} else {
		if len(findings) != 0 {
			t.Errorf("Expected empty results. Got %v\n", len(findings))
		}
	}
}

func TestListFindings(t *testing.T) {
	ss := initializeScanStore(t)
	addDummyRepo(t)
	scanId := addDummyScan(t, ss)

	findings := makeFindingsList(10)
	if err := ss.InsertFindings(scanId, findings); err != nil {
		t.Errorf("Failed to insert findings list: %v\n", err.Error())
	}

	if results, err := ss.ListFindings(scanId); err != nil {
		t.Fatalf("Expected no error. Got %v\n", err.Error())
	} else {
		for i := range findings {
			expected := findings[i].Location.Positions.Begin.Line
			actual := results[i].Location.Positions.Begin.Line

			if expected != actual {
				t.Errorf("Expected %v. Got %v\n", expected, actual)
			}
		}
	}
}
