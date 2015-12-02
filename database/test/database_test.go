package db_test

import (
	"testing"

	db "github.com/fiatjaf/summadb/database"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCRUD(t *testing.T) {
	db.Start()
	defer db.End()
	RegisterFailHandler(Fail)
	RunSpecs(t, "CRUD Suite")
}

func value(v interface{}) map[string]interface{} {
	return map[string]interface{}{"_val": v}
}
