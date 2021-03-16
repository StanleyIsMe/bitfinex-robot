package user

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"robot/logger"
	"robot/model"
	"robot/utils/s2c"
	"strconv"
	"sync"

	"robot/utils/redis"
)

const (
	UserKey = "user"
	NotifyKey = "notify"
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

func (pool *Pool) RegisterUser(registerInput model.RegisterRequest) error {
	pool.Lock()
	defer func() {
		pool.Unlock()
		if r := recover(); r != nil {
			logger.LOG.WithFields(logrus.Fields{
				"req": registerInput,
			}).Error(r)
		}
	}()

	if len(pool.UserList) >= pool.UserNumLimit {
		return errors.New("註冊人數已滿")
	}

	if _, ok := pool.UserList[registerInput.UserId]; ok {
		return errors.New("已註冊過")
	}

	telegramIdStr := strconv.Itoa(int(registerInput.UserId))
	result, err := redis.HGET(UserKey, telegramIdStr)
	if err != nil {
		return err
	}

	if result != "" {
		return errors.New("已註冊過")
	}

	newuser := NewUser(registerInput.UserId, registerInput.Token, registerInput.Sec, registerInput.Name)
	//utils.PrintWithStruct(newuser.API.Wallets(telegramId), "!!!!!!!!")
	if err := redis.HSET(UserKey, telegramIdStr, newuser); err != nil {
		return err
	}

	pool.UserList[registerInput.UserId] = newuser
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

		if user.Idle == 0 {
			user.Idle = 1
			user.StartActive()
		}

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

//func (pool *Pool) SetManageUser() {
//	userId, _ := strconv.ParseInt(os.Getenv("TELEGRAM_MANAGE_ID"), 10, 64)
//	pool.RegisterUser(userId, os.Getenv("API_KEY"), os.Getenv("API_SEC"))
//}

func (pool *Pool) UpdateById(telegramId int64) error {
	telegramIdStr := strconv.Itoa(int(telegramId))
	user := pool.GetUserById(telegramId)
	if user == nil {
		return errors.New(fmt.Sprintf("User Id [%d] Not Found", telegramId))
	}

	if err := redis.HSET(UserKey, telegramIdStr, user); err != nil {
		logger.LOG.Errorf("UpdateById Error %v", err)
		return err
	}
	return nil
}

func (pool *Pool) KillUser(telegramId int64) error {
	telegramIdStr := strconv.Itoa(int(telegramId))
	if err := redis.HDel(UserKey, telegramIdStr); err != nil {
		logger.LOG.Errorf("KillUser Error %v", err)
		return err
	}
	if member := pool.GetUserById(telegramId); member != nil {
		member.StopActive()
		delete(pool.UserList, telegramId)
	}
	return nil
}

func (pool *Pool) BroadcastMsg(message string) {
	for userId, _ :=  range pool.UserList {
		s2c.SendMessage(userId, message)
	}
}