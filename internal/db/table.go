package db

// Table represents a table meta-data in the db
type Table struct {
	Name          string
	AproxRowCount int64
}
