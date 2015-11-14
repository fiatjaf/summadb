package handle

type Error struct {
	Error  string `json:"error"`
	Reason string `json:"reason"`
	code   int
}

type AllDocs struct {
	TotalRows int   `json:"total_rows"`
	Offset    int   `json:"offset"`
	Rows      []Row `json:"rows"`
}

type Row struct {
	Id    string      `json:"id"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type Success struct {
	Id  string `json:"id"`
	Ok  bool   `json:"ok"`
	Rev string `json:"rev"`
}
