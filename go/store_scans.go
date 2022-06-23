package swagger

import (
	"encoding/base64"
	"encoding/binary"

	"github.com/UserProblem/reposcanner/models"
)

type ScanStore interface {
	Insert(si *models.ScanInfo) (*models.ScanRecord, error)
	Retrieve(id string) (*models.ScanRecord, error)
	Delete(id string) error
	Update(sr *models.ScanRecord) error
	List(pp *models.PaginationParams) (*models.ScanList, error)
	InsertFindings(scanId string, findings []*models.FindingsInfo) error
	ListFindings(scanId string) ([]*models.FindingsInfo, error)
	DeleteFindings(scanId string) (int, error)
}

// Create and return a pointer to a new scan data store.
// Returns nil and an error on failure
func NewScanStore(dbtype string) (ScanStore, error) {
	if dbtype == "postgresql" {
		return NewScanStorePsqlDB()
	} else {
		return NewScanStoreMemDB()
	}
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

// Helper function to continuously generate an incrementing
// uint64 value on the given channel
func generateObjIds(ch chan<- uint64) {
	nextId := uint64(1)
	for {
		ch <- nextId
		nextId++
	}
}
