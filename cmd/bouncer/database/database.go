package database

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/config"
	"errors"
	"time"
)

// commandAction - enum for command action - doFind or doUpsert.
type commandAction int

const (
	doFind commandAction = iota
	doUpsert
	doClose
)

// recordPayload - struct for passing data to command channel.
type recordPayload struct {
	key   string
	value analyzer.RecordInfo
	ttl   time.Duration
}

// command - struct for passing command to command channel.
type command struct {
	action commandAction
	data   recordPayload
	result chan bool
}

// DB - database interface.
type DB interface {
	Insert(payload recordPayload) bool
	Find(key string) bool
	Close()
}

// NewDB - create new database client.
func NewDB(conf *config.Conf) (DB, error) {
	// if config.Redis exists, use redis
	// if config.Postgres exists, use postgres
	// if config.MySQL exists, use mysql
	// if config.SQLite exists, use sqlite
	if conf.Redis.Addr != "" {
		client, err := NewRedisDB(conf.Redis)
		if err != nil {
			return nil, err
		}

		return client, nil
	}

	return nil, errors.New("no database config found")
}
