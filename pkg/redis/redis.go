package redis

import (
	"context"
	"github.com/gomodule/redigo/redis"
	"time"
	"timer/common/conf"
	"timer/pkg/logger"
)

type Client struct {
	pool *redis.Pool
}

func GetClient(config *conf.RedisConfig) *Client {
	return &Client{
		pool: getRedisPool(config),
	}
}

func getRedisPool(config *conf.RedisConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     config.MaxIdle,
		IdleTimeout: time.Duration(config.IdleTimeoutSeconds) * time.Second,
		MaxActive:   config.MaxActive,
		Wait:        config.Wait,
		Dial: func() (redis.Conn, error) {
			c, err := newRedisConn(config)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				logger.Errorf("Failed to ping redis server, cased by %s", err)
			}
			return err
		},
	}
}

func newRedisConn(config *conf.RedisConfig) (redis.Conn, error) {
	conn, err := redis.Dial(config.Network, config.Address, redis.DialPassword(config.Password))
	if err != nil {
		logger.Errorf("Failed to connect to redis, cased by %s", err)
		return nil, err
	}
	return conn, nil
}

func (c *Client) GetConn(ctx context.Context) (redis.Conn, error) {
	return c.pool.GetContext(ctx)
}
