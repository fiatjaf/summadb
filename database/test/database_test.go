package db_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCRUD(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CRUD Suite")
}

func value(v interface{}) map[string]interface{} {
	return map[string]interface{}{"_val": v}
}
