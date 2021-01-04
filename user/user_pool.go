package user

import (
	"errors"
	"strconv"
	"sync"

	//"robot/utils"
	"robot/utils/redis"
)

const (
	UserKey = "robot:user"
)

var pool *Pool
var poolSingleton sync.Once

type Pool struct {
	sync.RWMutex
	UserNumLimit int32
	UserList     map[int64]*UserInfo
}

func NewPool(userNumLimit int32) *Pool {
	poolSingleton.Do(func() {
		pool = &Pool{
			UserList:     make(map[int64]*UserInfo, 0),
			UserNumLimit: userNumLimit,
		}
	})
	return pool
}

func (pool *Pool) RegisterUser(telegramId int64, key, sec string) error {
	pool.Lock()
	defer pool.Unlock()

	if pool.UserNumLimit >= 5 {
		return errors.New("註冊人數已滿")
	}

	telegramIdStr := strconv.Itoa(int(telegramId))
	result, err := redis.HGET(UserKey, telegramIdStr)
	if err != nil {
		return err
	}

	var user UserInfo
	if err := user.BinaryUnmarshaler([]byte(result)); err != nil {
		return err
	}

	newuser := NewUser(telegramId, key, sec)

	if err := redis.HSET(UserKey, telegramIdStr, newuser); err != nil {
		return err
	}
	pool.UserList[telegramId] = newuser
	return nil
}

func (pool *Pool) GetAllUser() map[int64]*UserInfo {
	result, err := redis.HGetAll(UserKey)

	if err != nil {
		return nil
	}
	response := make(map[int64]*UserInfo, 0)

	for _, val := range result {
		var user UserInfo
		//msgpack.Unmarshal([]byte(val), &user)
		user.BinaryUnmarshaler([]byte(val))
		//user.TelegramId = key
		response[user.TelegramId] = &user
		//pool.UserList = append(pool.UserList, &user)
	}

	return response
}
func (pool *Pool) GetUserById(telegramId int64) *UserInfo {
	pool.Lock()
	defer pool.Unlock()

	if user, ok := pool.UserList[telegramId]; ok {
		return user
	}
	return nil
}
