package liqui

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	log "github.com/thrasher-/gocryptotrader/logger"
)

const (
	liquiAPIPublicURL      = "https://api.Liqui.io/api"
	liquiAPIPrivateURL     = "https://api.Liqui.io/tapi"
	liquiAPIPublicVersion  = "3"
	liquiAPIPrivateVersion = "1"
	liquiInfo              = "info"
	liquiTicker            = "ticker"
	liquiDepth             = "depth"
	liquiTrades            = "trades"
	liquiAccountInfo       = "getInfo"
	liquiTrade             = "Trade"
	liquiActiveOrders      = "ActiveOrders"
	liquiOrderInfo         = "OrderInfo"
	liquiCancelOrder       = "CancelOrder"
	liquiTradeHistory      = "TradeHistory"
	liquiWithdrawCoin      = "WithdrawCoin"

	liquiAuthRate   = 0
	liquiUnauthRate = 1
)

// Liqui is the overarching type across the liqui package
type Liqui struct {
	exchange.Base
}

// GetTradablePairs returns all available pairs (hidden or not)
func (l *Liqui) GetTradablePairs(nonHidden bool) ([]string, error) {
	info, err := l.GetInfo()
	if err != nil {
		return nil, err
	}

	var pairs []string
	for x, y := range info.Pairs {
		if nonHidden && y.Hidden == 1 || x == "" {
			continue
		}
		pairs = append(pairs, common.StringToUpper(x))
	}
	return pairs, nil
}

// GetInfo provides all the information about currently active pairs, such as
// the maximum number of digits after the decimal point, the minimum price, the
// maximum price, the minimum transaction size, whether the pair is hidden, the
// commission for each pair.
func (l *Liqui) GetInfo() (Info, error) {
	resp := Info{}
	req := fmt.Sprintf("%s/%s/%s/", l.API.Endpoints.URL, liquiAPIPublicVersion, liquiInfo)

	return resp, l.SendHTTPRequest(req, &resp)
}

// GetTicker returns information about currently active pairs, such as: the
// maximum price, the minimum price, average price, trade volume, trade volume
// in currency, the last trade, Buy and Sell price. All information is provided
// over the past 24 hours.
//
// currencyPair - example "eth_btc"
func (l *Liqui) GetTicker(currencyPair string) (map[string]Ticker, error) {
	type Response struct {
		Data    map[string]Ticker
		Success int    `json:"success"`
		Error   string `json:"error"`
	}

	response := Response{Data: make(map[string]Ticker)}
	req := fmt.Sprintf("%s/%s/%s/%s", l.API.Endpoints.URL, liquiAPIPublicVersion, liquiTicker, currencyPair)

	return response.Data, l.SendHTTPRequest(req, &response.Data)
}

// GetDepth information about active orders on the pair. Additionally it accepts
// an optional GET-parameter limit, which indicates how many orders should be
// displayed (150 by default). Is set to less than 2000.
func (l *Liqui) GetDepth(currencyPair string) (Orderbook, error) {
	type Response struct {
		Data    map[string]Orderbook
		Success int    `json:"success"`
		Error   string `json:"error"`
	}

	response := Response{Data: make(map[string]Orderbook)}
	req := fmt.Sprintf("%s/%s/%s/%s", l.API.Endpoints.URL, liquiAPIPublicVersion, liquiDepth, currencyPair)

	return response.Data[currencyPair], l.SendHTTPRequest(req, &response.Data)
}

// GetTrades returns information about the last trades. Additionally it accepts
// an optional GET-parameter limit, which indicates how many orders should be
// displayed (150 by default). The maximum allowable value is 2000.
func (l *Liqui) GetTrades(currencyPair string) ([]Trades, error) {
	type Response struct {
		Data    map[string][]Trades
		Success int    `json:"success"`
		Error   string `json:"error"`
	}

	response := Response{Data: make(map[string][]Trades)}
	req := fmt.Sprintf("%s/%s/%s/%s", l.API.Endpoints.URL, liquiAPIPublicVersion, liquiTrades, currencyPair)

	return response.Data[currencyPair], l.SendHTTPRequest(req, &response.Data)
}

// GetAccountInformation returns information about the user’s current balance, API-key
// privileges, the number of open orders and Server Time. To use this method you
// need a privilege of the key info.
func (l *Liqui) GetAccountInformation() (AccountInfo, error) {
	var result AccountInfo

	return result,
		l.SendAuthenticatedHTTPRequest(liquiAccountInfo, url.Values{}, &result)
}

// Trade creates orders on the exchange.
// to-do: convert orderid to int64
func (l *Liqui) Trade(pair, orderType string, amount, price float64) (float64, error) {
	req := url.Values{}
	req.Add("pair", pair)
	req.Add("type", orderType)
	req.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	req.Add("rate", strconv.FormatFloat(price, 'f', -1, 64))
	var result Trade

	err := l.SendAuthenticatedHTTPRequest(liquiTrade, req, &result)
	if result.Success == 0 {
		return -1, errors.New(result.Error)
	}
	return result.OrderID, err
}

// GetActiveOrders returns the list of your active orders.
func (l *Liqui) GetActiveOrders(pair string) (map[string]ActiveOrders, error) {
	result := make(map[string]ActiveOrders)

	req := url.Values{}
	req.Add("pair", pair)

	return result, l.SendAuthenticatedHTTPRequest(liquiActiveOrders, req, &result)
}

// GetOrderInfoByID returns the information on particular order.
func (l *Liqui) GetOrderInfoByID(OrderID int64) (map[string]OrderInfo, error) {
	result := make(map[string]OrderInfo)

	req := url.Values{}
	req.Add("order_id", strconv.FormatInt(OrderID, 10))

	return result, l.SendAuthenticatedHTTPRequest(liquiOrderInfo, req, &result)
}

// CancelExistingOrder method is used for order cancelation.
func (l *Liqui) CancelExistingOrder(OrderID int64) error {
	req := url.Values{}
	req.Add("order_id", strconv.FormatInt(OrderID, 10))
	var result CancelOrder

	err := l.SendAuthenticatedHTTPRequest(liquiCancelOrder, req, &result)
	if result.Success == 0 {
		return errors.New(result.Error)
	}
	return err
}

// GetTradeHistory returns trade history
func (l *Liqui) GetTradeHistory(vals url.Values, pair string) (map[string]TradeHistory, error) {
	result := make(map[string]TradeHistory)

	if pair != "" {
		vals.Add("pair", pair)
	}

	return result, l.SendAuthenticatedHTTPRequest(liquiTradeHistory, vals, &result)
}

// WithdrawCoins is designed for cryptocurrency withdrawals.
// API mentions that this isn't active now, but will be soon - you must provide the first 8 characters of the key
// in your ticket to support.
func (l *Liqui) WithdrawCoins(coin string, amount float64, address string) (WithdrawCoins, error) {
	req := url.Values{}
	req.Add("coinName", coin)
	req.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	req.Add("address", address)

	var result WithdrawCoins
	err := l.SendAuthenticatedHTTPRequest(liquiWithdrawCoin, req, &result)
	if err != nil {
		return WithdrawCoins{}, err
	}
	if len(result.Error) > 0 {
		return result, errors.New(result.Error)
	}

	return result, nil
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (l *Liqui) SendHTTPRequest(path string, result interface{}) error {
	return l.SendPayload("GET", path, nil, nil, result, false, l.Verbose)
}

// SendAuthenticatedHTTPRequest sends an authenticated http request to liqui
func (l *Liqui) SendAuthenticatedHTTPRequest(method string, values url.Values, result interface{}) (err error) {
	if !l.AllowAuthenticatedRequest() {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, l.Name)
	}

	if l.Nonce.Get() == 0 {
		l.Nonce.Set(time.Now().Unix())
	} else {
		l.Nonce.Inc()
	}
	values.Set("nonce", l.Nonce.String())
	values.Set("method", method)

	encoded := values.Encode()
	hmac := common.GetHMAC(common.HashSHA512, []byte(encoded), []byte(l.API.Credentials.Secret))

	if l.Verbose {
		log.Debugf("Sending POST request to %s calling method %s with params %s\n",
			l.API.Endpoints.URLSecondary, method, encoded)
	}

	headers := make(map[string]string)
	headers["Key"] = l.API.Credentials.Key
	headers["Sign"] = common.HexEncodeToString(hmac)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	return l.SendPayload("POST",
		l.API.Endpoints.URLSecondary, headers,
		strings.NewReader(encoded),
		result,
		true,
		l.Verbose)
}

// GetFee returns an estimate of fee based on type of transaction
func (l *Liqui) GetFee(feeBuilder exchange.FeeBuilder) (float64, error) {
	var fee float64
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		fee = calculateTradingFee(feeBuilder.PurchasePrice, feeBuilder.Amount, feeBuilder.IsMaker)
	case exchange.CryptocurrencyWithdrawalFee:
		fee = getCryptocurrencyWithdrawalFee(feeBuilder.FirstCurrency)
	}

	if fee < 0 {
		fee = 0
	}

	return fee, nil
}

func getCryptocurrencyWithdrawalFee(currency string) float64 {
	return WithdrawalFees[currency]
}

func calculateTradingFee(purchasePrice, amount float64, isMaker bool) (fee float64) {
	if isMaker {
		fee = 0.001
	} else {
		fee = 0.0025
	}
	return fee * purchasePrice * amount
}
