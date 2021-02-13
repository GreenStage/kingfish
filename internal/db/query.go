package db

// ColInfo contains a columns information
type ColInfo struct {
	Name string
	Type string
}

// QueryResult defines an output from a successful query operation
type QueryResult struct {
	Columns []ColInfo
	Rows    [][]interface{}
}
