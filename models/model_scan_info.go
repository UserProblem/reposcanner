/*
 * Repository Secrets Scanner
 *
 * This is a simple backend API to allow a user to configure repositories for scanning, trigger a scan of those repositories, and retrieve the results.
 *
 * API version: 0.0.1
 * Contact: sean.critica@gmail.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package models

type ScanInfo struct {

	// id of the target repository for this scan
	RepoId int64 `json:"repoId"`

	// timestamp when this scan was started
	QueuedAt string `json:"queuedAt"`

	// timestamp when this scan was started
	ScanningAt string `json:"scanningAt"`

	// timestamp when this scan was finished
	FinishedAt string `json:"finishedAt"`

	// the current execution status of this scan
	Status string `json:"status"`
}

func DefaultScanInfo() *ScanInfo {
	return &ScanInfo{
		RepoId:     1,
		QueuedAt:   "1970-01-01T00:00:00+01:00",
		ScanningAt: "",
		FinishedAt: "",
		Status:     "QUEUED",
	}
}

func (si *ScanInfo) Clone() *ScanInfo {
	return &ScanInfo{
		RepoId:     si.RepoId,
		QueuedAt:   si.QueuedAt,
		ScanningAt: si.ScanningAt,
		FinishedAt: si.FinishedAt,
		Status:     si.Status,
	}
}
