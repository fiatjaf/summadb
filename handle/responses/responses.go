package responses

import db "github.com/fiatjaf/summadb/database"

type Error struct {
	Error  string `json:"error"`
	Reason string `json:"reason"`
	Code   int    `json:"-"`
}

type Success struct {
	Id  string `json:"id"`
	Ok  bool   `json:"ok"`
	Rev string `json:"rev"`
}

type DatabaseInfo struct {
	DBName            string `json:"db_name"`
	UpdateSeq         uint64 `json:"update_seq"`
	InstanceStartTime int64  `json:"instance_start_time"`
	DocCount          int    `json:"doc_count"`
	DocDelCount       int    `json:"doc_del_count"`
	DiskSize          int    `json:"disk_size"`
	DataSize          int    `json:"data_size"`
	PurgeSeq          int    `json:"purge_seq"`
	CompactRunning    bool   `json:"compact_running"`
}

type Revisions struct {
	Start int      `json:"start"`
	Ids   []string `json:"ids"`
}

type RevInfo struct {
	Rev    string `json:"rev"`
	Status string `json:"status"`
}

type Changes struct {
	LastSeq uint64      `json:"last_seq"`
	Results []db.Change `json:"results"`
}

type AllDocs struct {
	TotalRows int   `json:"total_rows"`
	Offset    int   `json:"offset"`
	Rows      []Row `json:"rows"`
}

type Row struct {
	Id    string                 `json:"id"`
	Key   string                 `json:"key"`
	Value interface{}            `json:"value"`
	Doc   map[string]interface{} `json:"doc"`
	Error string                 `json:"error,omitempty"`
}

type BulkGet struct {
	Results []BulkGetResult `json:"results"`
}

type BulkGetResult struct {
	Id   string       `json:"id"`
	Docs []DocOrError `json:"docs"`
}

type DocOrError struct {
	Ok    *map[string]interface{} `json:"ok,omitempty"`
	Error *Error                  `json:"error,omitempty"`
}

type BulkDocsResult struct {
	Id     string `json:"id"`
	Ok     bool   `json:"ok,omitempty"`
	Rev    string `json:"rev,omitempty"`
	Error  string `json:"error,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type RevsDiffResult struct {
	Missing []string `json:"missing"`
}
