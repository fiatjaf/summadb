package handle

import (
	db "github.com/fiatjaf/summadb/database"
)

type Error struct {
	Error  string `json:"error"`
	Reason string `json:"reason"`
	code   int
}

type AllDocs struct {
	TotalRows int   `json:"total_rows"`
	Offset    int   `json:"offset"`
	Rows      []row `json:"rows"`
}

type row struct {
	Id    string      `json:"id"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type Success struct {
	Id  string `json:"id"`
	Ok  bool   `json:"ok"`
	Rev string `json:"rev"`
}

type ChangesResponse struct {
	LastSeq uint64      `json:"last_seq"`
	Results []db.Change `json:"results"`
}
