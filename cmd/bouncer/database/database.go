package database

import (
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

// RecordPayload - struct for passing data to command channel.
type RecordPayload struct {
	Key   string
	Value []byte
	TTL   time.Duration
}

// command - struct for passing command to command channel.
type command struct {
	action commandAction
	data   RecordPayload
	result chan bool
}

// DB - database interface.
type DB interface {
	Insert(payload RecordPayload) bool
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
