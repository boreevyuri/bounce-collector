package writer

import (
	"bounce-collector/cmd/bouncer/config"
	"errors"
	"github.com/go-redis/redis"
	"log"
	"strings"
)

type commandAction int

const (
	find commandAction = iota
	insert
)

type commandData struct {
	action commandAction
	value  interface{}
	result chan<- bool
}

type processRedis chan commandData

type ProcessRedis interface {
	Insert(value Record) bool
	Find(value string) bool
}

func New(conf config.RedisConfig) ProcessRedis {
	pr := make(processRedis)
	go pr.run(conf)
	return pr
}

func (pr processRedis) Insert(value Record) bool {
	reply := make(chan bool)
	pr <- commandData{action: insert, value: value, result: reply}
	return <-reply
}

func (pr processRedis) Find(value string) bool {
	reply := make(chan bool)
	pr <- commandData{action: find, value: value, result: reply}
	return <-reply
}

func (pr processRedis) run(conf config.RedisConfig) {
	rc := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       0,
	})
	defer closeConnect(rc)
	//store = make(map[string]interface{})
	for command := range pr {
		switch command.action {
		case find:
			found := findRecord(rc, command.value)
			command.result <- found
		case insert:
			err := insertRecord(rc, command.value)
			if err != nil {
				log.Fatal("Unable to insert record:", err)
				//os.Exit(13)
			}
			command.result <- true
		}
	}
}

func insertRecord(client *redis.Client, rec interface{}) (err error) {
	if rec, ok := rec.(Record); ok {
		if rec.TTL > 0 {
			err = client.Set(rec.Rcpt, marshalToJSON(rec.Info), rec.TTL).Err()
		}
		return err
	}
	return errors.New("unknown record format")
}

func findRecord(client *redis.Client, rec interface{}) bool {
	if rec, ok := rec.(string); ok {
		_, err := client.Get(strings.ToLower(rec)).Result()
		return err != redis.Nil
	}
	return false
}
