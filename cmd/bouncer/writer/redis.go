package writer

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

// Record - struct for write to redis.
type Record struct {
	Rcpt string
	TTL  time.Duration
	Info analyzer.RecordInfo
}

//// Config - struct for redis config.
// type Config struct {
//	Addr     string `yaml:"address"`
//	Password string `yaml:"password"`
// }

// PutRecord puts Record to Redis db.
func PutRecord(ctx context.Context, rec Record, config config.RedisConfig) (err error) {
	client, err := rClient(ctx, config)

	if err != nil {
		return err
	}

	defer closeConnect(client)

	if rec.TTL > 0 {
		err = client.Set(ctx, rec.Rcpt, marshalToJSON(rec.Info), rec.TTL).Err()
	}

	return err
}

// IsPresent checks address existence in redis db.
func IsPresent(ctx context.Context, addr string, config config.RedisConfig) bool {
	client, err := rClient(ctx, config)
	if err != nil {
		return false
	}

	defer closeConnect(client)

	_, err = client.Get(ctx, strings.ToLower(addr)).Result()

	// no record - Good. Proceed to another router
	// got record - Bad. Kill message
	return !errors.Is(err, redis.Nil)
}

func rClient(ctx context.Context, config config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()
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
