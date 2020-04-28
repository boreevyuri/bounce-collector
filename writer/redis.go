package writer

import (
	"fmt"
	"github.com/go-redis/redis"
)

//func redisWrite(res analyzer.Result) {
//
//}

func RedisClient() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	err := client.Set("key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get("key").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("key", val)

	val2, err := client.Get("key2").Result()

	switch err {
	case redis.Nil:
		fmt.Println("key2 does not exist")
	case nil:
		fmt.Println("key2", val2)
	default:
		panic(err)
	}
	return
}
