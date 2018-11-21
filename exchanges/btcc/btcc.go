package btcc

import (
	"github.com/gorilla/websocket"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
)

const (
	btccAuthRate   = 0
	btccUnauthRate = 0
)

// BTCC is the main overaching type across the BTCC package
// NOTE this package is websocket connection dependant, the REST endpoints have
// been dropped
type BTCC struct {
	exchange.Base
	Conn *websocket.Conn
}

// GetFee returns an estimate of fee based on type of transaction
func (b *BTCC) GetFee(feeBuilder exchange.FeeBuilder) (float64, error) {
	var fee float64

	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyWithdrawalFee:
		fee = getCryptocurrencyWithdrawalFee(feeBuilder.FirstCurrency)
	case exchange.InternationalBankWithdrawalFee:
		fee = getInternationalBankWithdrawalFee(feeBuilder.CurrencyItem, feeBuilder.Amount)
	}
	if fee < 0 {
		fee = 0
	}
	return fee, nil
}

func getCryptocurrencyWithdrawalFee(currency string) float64 {
	return WithdrawalFees[currency]
}

func getInternationalBankWithdrawalFee(currency string, amount float64) float64 {
	var fee float64

	fee = WithdrawalFees[currency] * amount
	return fee
}
