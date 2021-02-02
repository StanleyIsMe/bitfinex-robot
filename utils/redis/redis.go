package redis

import (
	"os"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
	"robot/logger"
)

type RedisClient struct {
	conn *redis.Client
	log  *logrus.Logger
}

var redisClient *RedisClient

func Init() {
	if redisClient == nil {
		redisClient = &RedisClient{
			log: logger.LOG,
		}

		redisClient.connect()
	}
}

func (client *RedisClient) connect() {
	conn := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
		PoolSize: 10, // default size
	})

	err := conn.Ping().Err()
	if err != nil {
		client.log.Errorf("Redis Connect Error : %v", err)
	}

	client.conn = conn
}

func (client *RedisClient) close() {
	client.conn.Close()
}

func HSET(key, field string, values interface{}) error {
	err := redisClient.conn.HSet(key, field, values).Err()
	if err != nil {
		redisClient.log.Errorf("Redis HSET Error : %v", err)
		return err
	}
	return nil
}

func HGET(key, field string) (string, error) {
	result, err := redisClient.conn.HGet(key, field).Result()
	if err == redis.Nil {
		return "", nil
	}

	if err != nil {
		redisClient.log.Errorf("Redis HSET Error : %v", err)
		return "", err
	}
	return result, nil
}

func HGetAll(key string) (map[string]string, error) {
	result, err := redisClient.conn.HGetAll(key).Result()

	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		redisClient.log.Errorf("Redis HGetAll Error : %v", err)
		return nil, err
	}
	return result, nil
}

func SetNX(key string, value interface{}, expiration time.Duration) bool {
	result, err := redisClient.conn.SetNX(key, value, expiration).Result()
	if err != nil {
		redisClient.log.Errorf("Redis SetNX Error : %v", err)
		return false
	}
	return result
}

func Get(key string) (string, error) {
	result, err := redisClient.conn.Get(key).Result()
	if err == redis.Nil {
		return "", nil
	}

	if err != nil {
		redisClient.log.Errorf("Redis Get Error : %v", err)
		return "", err
	}

	return result, nil
}

//
//func Encode(data interface{}) ([]byte, error) {
//	buf := bytes.NewBuffer(nil)
//	enc := gob.NewEncoder(buf)
//	err := enc.Encode(data)
//	if err != nil {
//		return nil, errors.New("redis cache value encode error")
//	}
//	return buf.Bytes(), nil
//}
//
//func Decode(data []byte, to interface{}) error {
//	buf := bytes.NewBuffer(data)
//	dec := gob.NewDecoder(buf)
//	err := dec.Decode(to)
//	if err != nil {
//		return errors.New("redis cache value decode error")
//	}
//	return nil
//}
