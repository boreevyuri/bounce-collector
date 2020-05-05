package writer

import (
	"github.com/go-redis/redis"
	"time"
)

// Структура записи в БД
type Record struct {
	Rcpt string
	TTL  time.Duration
	Info string
}

func rClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	return client
}

func PutRecord(rec *Record) (err error) {
	client := rClient()

	if rec.TTL > 0 {
		err = client.Set(rec.Rcpt, rec.Info, rec.TTL).Err()
	}

	_ = client.Close()

	return err
}
