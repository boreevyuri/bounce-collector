package writer

import (
	"encoding/json"
	"github.com/boreevyuri/bounce-collector/analyzer"
	"github.com/go-redis/redis"
	"time"
)

// Структура записи в БД.
type Record struct {
	Rcpt string
	TTL  time.Duration
	Info analyzer.RecordInfo
}

type Config struct {
	Addr     string `yaml:"address"`
	Password string `yaml:"password"`
}

func PutRecord(rec Record, config Config) (err error) {
	client := rClient(config)

	if rec.TTL > 0 {
		err = client.Set(rec.Rcpt, marshalToJSON(rec.Info), rec.TTL).Err()
	}

	_ = client.Close()

	return err
}

func rClient(config Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       0,
	})

	return client
}

func marshalToJSON(r analyzer.RecordInfo) string {
	e, err := json.Marshal(r)
	if err != nil {
		return ""
	}

	return string(e)
}
