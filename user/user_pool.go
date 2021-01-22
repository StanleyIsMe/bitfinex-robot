package user

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"robot/logger"
	"strconv"
	"sync"

	//"robot/utils"
	"robot/utils/redis"
)

const (
	UserKey = "user"
)

var UserPool *Pool
var poolSingleton sync.Once

type Pool struct {
	sync.RWMutex
	UserNumLimit int
	UserList     map[int64]*UserInfo
}

func NewPool(userNumLimit int) *Pool {
	poolSingleton.Do(func() {
		UserPool = &Pool{
			UserList:     make(map[int64]*UserInfo, 0),
			UserNumLimit: userNumLimit,
		}
		UserPool.SetManageUser()

	})
	return UserPool
}

func GetInstance() *Pool {
	if UserPool == nil {
		NewPool(5)
	}
	return UserPool
}

func (pool *Pool) RegisterUser(telegramId int64, key, sec string) error {
	pool.Lock()
	defer func() {
		pool.Unlock()
		if r := recover(); r != nil {
			logger.LOG.WithFields(logrus.Fields{
				"userId": telegramId,
				"key":    key,
				"sec":    sec,
			}).Error(r)
		}
	}()

	if len(pool.UserList) >= pool.UserNumLimit {
		return errors.New("註冊人數已滿")
	}

	telegramIdStr := strconv.Itoa(int(telegramId))
	result, err := redis.HGET(UserKey, telegramIdStr)
	if err != nil {
		return err
	}

	if result != "" {
		return errors.New("已註冊過")
	}

	//var user UserInfo
	//if err := user.BinaryUnmarshaler([]byte(result)); err != nil {
	//	return err
	//}

	newuser := NewUser(telegramId, key, sec)
	//fmt.Println(newuser.MarshalBinary())
	mp := newuser.Config.ConvertToMap()
	fmt.Println(mp,"!!!!!!!!!!!!")
	if err := redis.HSET(fmt.Sprintf("%s:%s",UserKey, telegramIdStr), mp); err != nil {
		return err
	}
	pool.UserList[telegramId] = newuser
	return nil
}

//func (pool *Pool) GetAllUser() map[int64]*UserInfo {
//	result, err := redis.HGetAll(UserKey)
//
//	if err != nil {
//		return nil
//	}
//	response := make(map[int64]*UserInfo, 0)
//
//	for _, val := range result {
//		var user UserInfo
//		//msgpack.Unmarshal([]byte(val), &user)
//		//user.BinaryUnmarshaler([]byte(val))
//		//user.TelegramId = key
//		response[user.TelegramId] = &user
//		//pool.UserList = append(pool.UserList, &user)
//	}
//
//	return response
//}
func (pool *Pool) GetUserById(telegramId int64) *UserInfo {
	pool.Lock()
	defer pool.Unlock()

	if user, ok := pool.UserList[telegramId]; ok {
		return user
	}
	return nil
}

func (pool *Pool) SetManageUser() {
	userId, _ := strconv.ParseInt(os.Getenv("TELEGRAM_MANAGE_ID"), 10, 64)
	pool.RegisterUser(userId, os.Getenv("API_KEY"), os.Getenv("API_SEC"))
}
