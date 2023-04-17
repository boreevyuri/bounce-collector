package db

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/config"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"strings"
)

// processRedis - channel for commandData.
type processRedis chan commandData

func (pr processRedis) run(conf config.RedisConfig) {
	rc := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       0,
	})
	defer closeConnect(rc)

	_, err := rc.Ping().Result()
	if err != nil {
		log.Fatal("unable to connect redis:", err)
	}

	// store = make(map[string]interface{})
	for command := range pr {
		switch command.action {
		case doFind:
			found := findRecord(rc, command.value)
			command.result <- found
		case doUpsert:
			err := insertRecord(rc, command.value)
			if err != nil {
				log.Fatal("Unable to doUpsert record:", err)
				// os.Exit(13)
			}
			command.result <- true
		}
	}
}

func (pr processRedis) Insert(value Record) bool {
	reply := make(chan bool)
	pr <- commandData{action: doUpsert, value: value, result: reply}
	return <-reply
}

func (pr processRedis) Find(value string) bool {
	reply := make(chan bool)
	pr <- commandData{action: doFind, value: value, result: reply}
	return <-reply
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

func marshalToJSON(r analyzer.RecordInfo) string {
	e, err := json.Marshal(r)
	if err != nil {
		return ""
	}

	return string(e)
}

func closeConnect(client *redis.Client) {
	if err := client.Close(); err != nil {
		fmt.Println("Connection already closed")
	}
}
