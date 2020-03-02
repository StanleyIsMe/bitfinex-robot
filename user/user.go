package user

import (
	"sync"

	"github.com/vmihailenco/msgpack"

	//"robot/utils"
	"robot/utils/redis"
)

const (
	UserKey = "robot:user"
)

var pool *Pool
var poolSingleton sync.Once

type Pool struct {
	UserList []*UserInfo
}
type UserInfo struct {
	//utils.Marshal
	TelegramId string
	Config     *ConfigManage
	Key        string // bitfinex api key
	Sec        string // bitfinex sec
	Wallet     *Wallet
}

func GetUserList() {

}

func (pool *Pool) RegisterUser() {
	user := &UserInfo{
		TelegramId: "938162785",
		Config:     NewConfig(),
		Key:        "111",
		Sec:        "222",
		Wallet:     NewWallet(),
	}

	redis.HSET(UserKey, "938162785", user)
}

func (pool *Pool) GetAllUser() []*UserInfo {
	result, err := redis.HGetAll(UserKey)

	if err != nil {
		return nil
	}


	for key, val := range result {
		var user UserInfo
		//msgpack.Unmarshal([]byte(val), &user)
		user.BinaryUnmarshaler([]byte(val))
		user.TelegramId = key
		pool.UserList = append(pool.UserList, &user)
	}

	return pool.UserList
}

func NewPool() *Pool {
	poolSingleton.Do(func() {
		pool = &Pool{}
	})
	return pool
}

func (t *UserInfo) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(t)
}

func (t *UserInfo) BinaryUnmarshaler(data []byte) error {

	return msgpack.Unmarshal(data, t)
}
