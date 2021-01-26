package user

import (
	"errors"
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
		//UserPool.SetManageUser()
		UserPool.InitAllUser()
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

	if _, ok := pool.UserList[telegramId]; ok {
		return errors.New("已註冊過")
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

	if err := redis.HSET(UserKey, telegramIdStr, newuser); err != nil {
		return err
	}

	//result2, err := redis.HGET(UserKey, telegramIdStr)
	//if err != nil {
	//	fmt.Println(err, "!!!")
	//	return err
	//}
	//
	//us := &UserInfo{}
	//if err := us.UnmarshalBinary([]byte(result2)); err != nil {
	//	fmt.Println(err, "!!!")
	//	return err
	//}
	//
	//utils.PrintWithStruct(us)
	//fmt.Println("==========================")
	pool.UserList[telegramId] = newuser
	return nil
}

func (pool *Pool) InitAllUser() {
	result, err := redis.HGetAll(UserKey)

	if err != nil {
		logger.LOG.Errorf("redis.HGetAll Error %v", err)
		return
	}

	for _, val := range result {
		user := &UserInfo{}
		err := user.UnmarshalBinary([]byte(val))
		if err != nil {
			logger.LOG.Errorf("InitAllUser Error %v", err)
			continue
		}
		user.StartActive()
		pool.UserList[user.TelegramId] = user
	}
}

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
