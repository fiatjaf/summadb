// +build levelupjs

package database

import "honnef.co/go/js/console"

func init() {
	log = ConsoleLogger{}
}

type ConsoleLogger struct{}

func (l ConsoleLogger) Info(msg string, args ...interface{}) {
	console.Log(append([]interface{}{"INFO", msg}, args...)...)
}

func (l ConsoleLogger) Error(msg string, args ...interface{}) {
	console.Error(append([]interface{}{"ERRR", msg}, args...)...)
}

func (l ConsoleLogger) Warn(msg string, args ...interface{}) {
	console.Warn(append([]interface{}{"WARN", msg}, args...)...)
}

func (l ConsoleLogger) Debug(msg string, args ...interface{}) {
	console.Log(append([]interface{}{"DEBG ", msg}, args...)...)
}
