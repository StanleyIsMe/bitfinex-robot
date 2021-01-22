package user

import (
	"os"
	"strconv"
	"sync"
)

type Wallet struct {
	sync.RWMutex

	Balance          float64
	BalanceAvailable float64
	wg               *sync.WaitGroup
	IsUsed           int32
}

var myWallet *Wallet

func NewWallet() *Wallet {

	myWallet = &Wallet{
		Balance:          0,
		BalanceAvailable: 0,
		wg:               &sync.WaitGroup{},
		IsUsed:           0,
	}

	return myWallet
}

func (object *Wallet) Update(balance, balanceAvailable float64) {
	object.Lock()
	object.Balance = balance
	object.BalanceAvailable = balanceAvailable
	object.Unlock()
}

func (object *Wallet) GetAmount(basicAmount float64) float64 {
	minimumAmount,_ := strconv.ParseFloat(os.Getenv("OFFICIAL_MIN_FUNDING_MONEY"), 64)
	object.Lock()
	defer object.Unlock()

	if ((object.BalanceAvailable - basicAmount) < minimumAmount) || (object.BalanceAvailable <= basicAmount) {
		temp := object.BalanceAvailable
		object.BalanceAvailable = 0
		return temp
	}
	object.BalanceAvailable -= basicAmount
	return basicAmount
}