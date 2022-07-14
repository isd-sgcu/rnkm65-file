package database

import (
	"github.com/go-redis/redis/v8"
	"github.com/isd-sgcu/rnkm65-file/src/config"
	"github.com/pkg/errors"
)

func InitRedisConnect(conf *config.Redis) (cache *redis.Client, err error) {
	cache = redis.NewClient(&redis.Options{
		Addr: conf.Host,
		DB:   1,
	})

	if cache == nil {
		return nil, errors.New("Cannot connect to redis server")
	}

	return
}
