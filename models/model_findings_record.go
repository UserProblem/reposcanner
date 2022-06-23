package models

type FindingsRecord struct {
	Id      int           `json:"id"`
	ScanId  string        `json:"scanid"`
	Finding *FindingsInfo `json:"finding"`
}
