package db

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/config"
	"time"
)

// commandAction - enum for command action - doFind or doUpsert.
type commandAction int

const (
	doFind commandAction = iota
	doUpsert
)

// Record - struct for write to database
type Record struct {
	Rcpt string
	TTL  time.Duration
	Info analyzer.RecordInfo
}

// commandData - struct for command data - action, value and result.
type commandData struct {
	action commandAction
	value  interface{}
	result chan<- bool
}

// DBInterface - interface for database
type DBInterface interface {
	Insert(value Record) bool
	Find(value string) bool
}

// New - create new database connection
func New(conf config.Conf) DBInterface {
	pr := make(processRedis)
	go pr.run(conf.Redis)
	return pr
}
