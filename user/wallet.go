package user

import "sync"

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
