package writer

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/config"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"strings"
	"time"
)

// Record - struct for write to redis.
type Record struct {
	Rcpt string
	TTL  time.Duration
	Info analyzer.RecordInfo
}

// PutRecord puts Record to Redis db.
func PutRecord(rec Record, conf config.RedisConfig) (err error) {
	client, err := rClient(conf)

	if err != nil {
		return err
	}

	defer closeConnect(client)

	if rec.TTL > 0 {
		err = client.Set(rec.Rcpt, marshalToJSON(rec.Info), rec.TTL).Err()
	}

	return err
}

// IsPresent checks address existence in redis db.
func IsPresent(addr string, conf config.RedisConfig) bool {
	client, err := rClient(conf)
	if err != nil {
		return false
	}

	defer closeConnect(client)

	_, err = client.Get(strings.ToLower(addr)).Result()

	//no record - Good. Proceed to another router
	//got record - Bad. Kill message
	return err != redis.Nil
}

func rClient(conf config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return client, nil
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
