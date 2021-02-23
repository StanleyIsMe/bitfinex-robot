package common

const (
	LowFloatLowRate StrategyType = iota + 1
	ExpectHighRate
	HighFloatHighRate
)

type StrategyType int32

func (st StrategyType) String() string {
	switch st {
	case LowFloatLowRate:
		return "低利率低波動"
	case ExpectHighRate:
		return "預期高利率"
	case HighFloatHighRate:
		return "高利率高波動"
	}
	return "無此策略"
}
