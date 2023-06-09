package redis

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gomodule/redigo/redis"
	utils "goranger/genericUtilities"
	"strconv"
	"time"
)

func LGetRedisValue(redisPool *redis.Pool, key string, isPoP bool) (string, error) {

	var redisValue string
	var redisErr error

	conn := redisPool.Get()

	if isPoP {
		redisValue, redisErr = redis.String(conn.Do("LPOP", key))
	} else {
		length, err := GetLength(redisPool, key)
		if err == nil && length > 0 {
			redisValue, redisErr = redis.String(conn.Do("LINDEX", key, utils.GetRandomIndex(length)))
		}
	}

	conn.Close()

	return redisValue, redisErr
}

func LPushRedis(redisPool *redis.Pool, key string, value string) error {

	var redisErr error

	conn := redisPool.Get()

	_, redisErr = redis.String(conn.Do("LPUSH", key, value))

	conn.Close()

	return redisErr

}

func SGetRedisValue(redisPool *redis.Pool, key string, isPoP bool) (string, error) {

	var redisValue string
	var redisErr error

	conn := redisPool.Get()

	if isPoP {
		redisValue, redisErr = redis.String(conn.Do("SPOP", key))
	} else {
		redisValue, redisErr = redis.String(conn.Do("SRandMember", key))
	}

	conn.Close()

	return redisValue, redisErr
}

func SAddRedis(redisPool *redis.Pool, key string, value string) error {

	var redisErr error

	conn := redisPool.Get()

	_, redisErr = redis.Int64(conn.Do("SADD", key, value))

	conn.Close()

	return redisErr

}

func HMSETRedis(redisPool *redis.Pool, key string, hashkey string, hashvalue string) error {

	var redisErr error

	conn := redisPool.Get()

	_, redisErr = redis.String(conn.Do("HMSET", key, hashkey, hashvalue))

	conn.Close()

	return redisErr

}

func HMGETRedis(redisPool *redis.Pool, key string, hashkey string) (string, error) {

	var redisErr error

	conn := redisPool.Get()

	redisValue, redisErr := redis.Strings(conn.Do("HMGET", key, hashkey))

	if redisErr != nil {
		fmt.Println(redisErr.Error())
	}

	conn.Close()

	return redisValue[0], redisErr

}

func SetSessionData(redisPool *redis.Pool, value string) {

	conn := redisPool.Get()

	key := time.Now().String()

	_, redisErr := redis.String(conn.Do("SET", key, value))
	_, redisErrExpire := redis.Int(conn.Do("EXPIRE", key, 6000))

	conn.Close()

	if redisErr != nil {
		color.Red("Error setting values to session redis ----> ")
	}

	if redisErrExpire != nil {
		color.Red("Error setting TTL to session redis ----> ")
	}

}

func GetSessionData(redisPool *redis.Pool) (string, error) {

	var redisErr error

	conn := redisPool.Get()

	key, redisErr := redis.String(conn.Do("RANDOMKEY"))
	redisValue, redisErr := redis.String(conn.Do("GET", key))

	conn.Close()

	if redisErr != nil {
		color.Red("Error Getting values from session redis ----> ")
	}

	return redisValue, redisErr

}

func SetRedisData(redisPool *redis.Pool, key string, value string) (error, error) {

	conn := redisPool.Get()

	_, redisErr := redis.String(conn.Do("SET", key, value))
	_, redisErrExpire := redis.Int(conn.Do("EXPIRE", key, 6000))

	conn.Close()

	return redisErr, redisErrExpire

}

func GetRedisData(redisPool *redis.Pool, key string) (string, error) {

	var redisErr error

	conn := redisPool.Get()

	redisValue, redisErr := redis.String(conn.Do("GET", key))

	conn.Close()

	return redisValue, redisErr

}

func GetLength(redisPool *redis.Pool, key string) (int, error) {

	conn := redisPool.Get()
	keyType, _ := GetKeyType(conn, key)

	length := 0
	redisErr := error(nil)

	if keyType == "list" {
		length, redisErr = redis.Int(conn.Do("LLEN", key))

	} else if keyType == "set" {
		length, redisErr = redis.Int(conn.Do("SCARD", key))
	}

	if redisErr != nil {
		color.Red("Error getting length for the key from redis ---->" + key)
	}
	conn.Close()

	return length, redisErr
}

func GetKeyType(conn redis.Conn, key string) (string, error) {

	redisValue, redisErr := redis.String(conn.Do("TYPE", key))

	return redisValue, redisErr
}

func IsKeyAvailable(conn redis.Conn, key string) bool {

	redisValue, _ := redis.Int(conn.Do("EXISTS", key))

	return redisValue > 0
}

func WaitForRedisKeyToBeAvailable(conn redis.Conn, key string, ttl int) bool {

	IsAvailable := false
	for i := 1; i < ttl; i++ {
		if IsKeyAvailable(conn, key) {
			IsAvailable = true
			break
		}
		color.Red("Waiting for redis key to be available --->" + key + "---> " + strconv.Itoa(i) + " second(s)")
		time.Sleep(time.Second)
	}
	if !IsAvailable {
		color.Red("I am done waiting for key to be available --->" + key + " for " + strconv.Itoa(ttl) + " seconds")
	}

	return IsAvailable
}
