package server

import "github.com/summadb/summadb/utils"

func jsonError(errString string) []byte {
	b := append([]byte(`{"error":`), utils.JSONString(errString)...)
	return append(b, '}')
}

func jsonSuccess() []byte {
	return []byte(`{"success":true}`)
}
