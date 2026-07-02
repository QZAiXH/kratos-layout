package decimalx

import "github.com/shopspring/decimal"

func YuanFromFen(fen int64) decimal.Decimal {
	return decimal.NewFromInt(fen).Div(decimal.NewFromInt(100))
}

func FenFromYuan(yuan decimal.Decimal) int64 {
	return yuan.Mul(decimal.NewFromInt(100)).Round(0).IntPart()
}
