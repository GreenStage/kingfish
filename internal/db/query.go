package db

type ColInfo struct {
	Name string
	Type string
}

type QueryResult struct {
	Columns []ColInfo
	Rows    [][]interface{}
}
