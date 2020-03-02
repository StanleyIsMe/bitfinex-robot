package redis

import (
	"os"

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
			log:  logger.LOG,
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

func HSET(key, field string, value interface{}) {
	err := redisClient.conn.HSet(key, field, value).Err()
	if err !=nil {
		redisClient.log.Errorf("Redis HSET Error : %v", err)
	}
	//err = client.HMSet("htest", aa).Err()
	////client.HGetAll("htest").Result()
	//fmt.Println(err,11111111)
	//
	//
	////ree := make(map[string]*Test, 0)
	//var ts Test
	//result, err := client.HMGet("htest", "11", "22").Result()
	////
	//////err = utils.Decode([]byte(v), userModel)

	//fmt.Println(err, result)

	//msgpack.Unmarshal(result, &ts)
	//fmt.Println(ts)
}

func HGetAll(key string) (map[string]string, error){
	result, err := redisClient.conn.HGetAll(key).Result()

	if err !=nil {
		redisClient.log.Errorf("Redis HGetAll Error : %v", err)
		return nil, err
	}
	return result, nil
	//for _, val := range result {
	//	msgpack.Unmarshal([]byte(val.(string)), &ts)
	//	fmt.Println(ts.Id, ts.Name)
	//}
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
