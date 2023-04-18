package database

import (
	"bounce-collector/cmd/bouncer/config"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"log"
)

type redisConnection chan command

func NewRedisDB(conf config.RedisConfig) (DB, error) {
	rc := make(redisConnection)
	go rc.run(conf)
	return &rc, nil
}

func (r *redisConnection) run(conf config.RedisConfig) {
	ctx := context.Background()
	rc := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       0,
	})

	_, err := rc.Ping(ctx).Result()
	if err != nil {
		log.Fatal("unable to connect redis:", err)
	}

	defer r.Close()

	for cmd := range *r {
		switch cmd.action {
		case doFind:
			_, err := rc.Get(ctx, cmd.data.Key).Result()
			switch {
			case errors.Is(err, redis.Nil):
				cmd.result <- false
			case err != nil:
				cmd.result <- false
			}
			cmd.result <- true
		case doUpsert:
			err := rc.Set(ctx, cmd.data.Key, cmd.data.Value, cmd.data.TTL).Err()
			if err != nil {
				cmd.result <- false
			}
			cmd.result <- true
		case doClose:
			err := rc.Close()
			if err != nil {
				log.Panic("unable to close redis connection:", err)
			}
			return
		}
	}
}

// Insert inserts Key-Value pair into redis.
func (r *redisConnection) Insert(payload RecordPayload) bool {
	cmd := command{
		action: doUpsert,
		data:   payload,
		result: make(chan bool),
	}
	*r <- cmd
	return <-cmd.result
}

// Find checks if Key exists in redis.
func (r *redisConnection) Find(key string) bool {
	cmd := command{
		action: doFind,
		data: RecordPayload{
			Key: key,
		},
		result: make(chan bool),
	}
	*r <- cmd
	return <-cmd.result
}

// Close closes the redis connection.
func (r *redisConnection) Close() {
	// close connection to redis
	*r <- command{action: doClose}
	// close channel
	close(*r)
}
