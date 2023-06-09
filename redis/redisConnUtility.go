package redis

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

type RedisPoolMap struct {
	PoolMap map[string]redis.Pool
}


func GetRedisConnPool(rate int,connectionUrl string, dbIndex int, connectionName string) *redis.Pool {

		client := &redis.Pool{
			MaxIdle: rate * 3,
			MaxActive: rate * 3,
			IdleTimeout: 300 * time.Second,
			Wait: false,
			MaxConnLifetime: 0,
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}
				_, err := c.Do("PING")
				return err
			},
			Dial: func() (redis.Conn,error) {
				c,err := redis.Dial("tcp",connectionUrl,redis.DialReadTimeout(60 * time.Second),
					redis.DialWriteTimeout(60 * time.Second),redis.DialConnectTimeout(60 * time.Second))
				if c != nil && err != nil {
					c.Do("SELECT", dbIndex)
				}
				return c, err
			},
		}
	return client
}