package swagger_test

import (
	"strings"
	"testing"

	sw "github.com/UserProblem/reposcanner/go"
)

func initializeScanStore(t *testing.T) *sw.ScanStore {
	ss, err := sw.NewScanStore()
	if err != nil {
		t.Errorf("Failed to initialize scan store.")
		t.FailNow()
	}
	return ss
}

func TestNextScanIdIncrements(t *testing.T) {
	ss := initializeScanStore(t)

	first := ss.NextId()
	second := ss.NextId()

	if first != sw.EncodeScanId(1) {
		t.Errorf("Expected first id is %v. Got %v\n", sw.EncodeScanId(1), first)
	}

	if second != sw.EncodeScanId(2) {
		t.Errorf("Expected second id is %v. Got %v\n", sw.EncodeScanId(2), second)
	}
}

func TestStoreNewScanInfo(t *testing.T) {
	ss := initializeScanStore(t)

	si := sw.DefaultScanInfo()
	sr, err := ss.Insert(si)

	if err != nil {
		t.Fatalf("Failed to insert scan info into the database.\n")
	}

	if sr.Id != sw.EncodeScanId(1) {
		t.Errorf("Expected id is %v. Got %v\n", sw.EncodeScanId(1), sr.Id)
	}

	if sr.Info.RepoId != si.RepoId {
		t.Errorf("Expected repo id is %v. Got %v\n", si.RepoId, sr.Info.RepoId)
	}

	if sr.Info.QueuedAt != si.QueuedAt {
		t.Errorf("Expected queued at is '%v'. Got '%v'\n", si.QueuedAt, sr.Info.QueuedAt)
	}

	if sr.Info.ScanningAt != si.ScanningAt {
		t.Errorf("Expected scanning at is '%v'. Got '%v'\n", si.ScanningAt, sr.Info.ScanningAt)
	}

	if sr.Info.FinishedAt != si.FinishedAt {
		t.Errorf("Expected finished at is '%v'. Got '%v'\n", si.FinishedAt, sr.Info.FinishedAt)
	}

	if sr.Info.Status != si.Status {
		t.Errorf("Expected status is %v. Got %v\n", si.Status, sr.Info.Status)
	}
}

func TestRetrieveScanRecord(t *testing.T) {
	ss := initializeScanStore(t)

	si := sw.DefaultScanInfo()
	sr, err := ss.Insert(si)

	if err != nil {
		t.Errorf("Failed to insert scan info into the database.\n")
		t.FailNow()
	}

	sr, err = ss.Retrieve(sw.EncodeScanId(1))
	if err != nil {
		t.Errorf("Failed to retrieve scan record from the database.\n")
		t.FailNow()
	}

	if sr.Id != sw.EncodeScanId(1) {
		t.Errorf("Expected id is %v. Got %v\n", sw.EncodeScanId(1), sr.Id)
	}

	if sr.Info.RepoId != si.RepoId {
		t.Errorf("Expected repo id is %v. Got %v\n", si.RepoId, sr.Info.RepoId)
	}

	if sr.Info.QueuedAt != si.QueuedAt {
		t.Errorf("Expected queued at is '%v'. Got '%v'\n", si.QueuedAt, sr.Info.QueuedAt)
	}

	if sr.Info.ScanningAt != si.ScanningAt {
		t.Errorf("Expected scanning at is '%v'. Got '%v'\n", si.ScanningAt, sr.Info.ScanningAt)
	}

	if sr.Info.FinishedAt != si.FinishedAt {
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

	si := sw.DefaultScanInfo()
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

	if !strings.HasPrefix(err.Error(), "Id not found") {
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

	si := sw.DefaultScanInfo()

	ssOrig, err := ss.Insert(si)
	if err != nil {
		t.Fatalf("Failed to insert scan info into the database.\n")
	}

	ssMod := &sw.ScanRecord{
		Id: ssOrig.Id,
		Info: &sw.ScanInfo{
			RepoId:     ssOrig.Info.RepoId,
			QueuedAt:   ssOrig.Info.QueuedAt,
			ScanningAt: "1970-01-01 00:00:01+0",
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

	if ssMod.Info.QueuedAt != ssOrig.Info.QueuedAt {
		t.Errorf("Expected queued at to be '%v'. Got '%v'\n", ssOrig.Info.QueuedAt, ssMod.Info.QueuedAt)
	}

	if ssMod.Info.ScanningAt != "1970-01-01 00:00:01+0" {
		t.Errorf("Expected scanning at to be '1970-01-01 00:00:01+0'. Got '%v'\n", ssMod.Info.ScanningAt)
	}

	if ssMod.Info.Status != "IN PROGRESS" {
		t.Errorf("Expected status to be 'IN PROGRESS'. Got '%v'\n", ssMod.Info.Status)
	}
}

func TestUpdateInvalidScanRecord(t *testing.T) {
	ss := initializeScanStore(t)

	sr := sw.ScanRecord{
		Id:   sw.EncodeScanId(1),
		Info: sw.DefaultScanInfo(),
	}

	if err := ss.Update(&sr); err == nil {
		t.Errorf("Expected error but got successful result.\n")
	}
}

func TestRetrieveScanList(t *testing.T) {
	ss := initializeScanStore(t)

	var si *sw.ScanInfo
	var err error

	var totalRecords, i int32 = 10, 1
	for ; i <= totalRecords; i++ {
		si = sw.DefaultScanInfo()
		si.RepoId = int64(i)

		if _, err = ss.Insert(si); err != nil {
			t.Fatalf(err.Error())
		}
	}

	sl, err := ss.List(&sw.PaginationParams{Offset: 2, PageSize: 5})
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
			t.Errorf("Expected repo id to be %v. Got %v\n", j, sr.Id)
		}
		j++
	}
}

func TestRetrieveEmptyScanList(t *testing.T) {
	ss := initializeScanStore(t)

	sl, err := ss.List(&sw.PaginationParams{Offset: 0, PageSize: 20})
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

	_, err := ss.List(&sw.PaginationParams{Offset: 2, PageSize: 5})
	if err.Error() != "Invalid offset" {
		t.Errorf("Expected error 'Invalid offset'. Got '%v'\n", err.Error())
	}
}

func TestRetrieveScanListInvalidPageSize(t *testing.T) {
	ss := initializeScanStore(t)

	_, err := ss.List(&sw.PaginationParams{Offset: 0, PageSize: 0})
	if err.Error() != "Invalid page size" {
		t.Errorf("Expected error 'Invalid page size'. Got '%v'\n", err.Error())
	}
}
