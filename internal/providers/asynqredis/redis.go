package asynqredis

import (
	"fmt"

	"lfiber/configs"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

const defaultDialerRetries = 2

type clientOpt struct {
	addr     string
	password string
	db       int
}

func NewClientOpt(cfg *configs.Config, dbOffset int) asynq.RedisConnOpt {
	return clientOpt{
		addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		password: cfg.Redis.Password,
		db:       cfg.Redis.DB + dbOffset,
	}
}

func (opt clientOpt) MakeRedisClient() interface{} {
	return redis.NewClient(&redis.Options{
		Addr:          opt.addr,
		Password:      opt.password,
		DB:            opt.db,
		DialerRetries: defaultDialerRetries,
	})
}
