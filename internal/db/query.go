package db

type QueryResult struct {
	Header []string
	Rows   [][]interface{}
}

func (q QueryResult) GetColumnIndex(colName string) int {
	for i, h := range q.Header {
		if h == colName {
			return i
		}
	}
	return -1
}
