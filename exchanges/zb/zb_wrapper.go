package zb

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/assets"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/request"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
	log "github.com/thrasher-/gocryptotrader/logger"
)

// GetDefaultConfig returns a default exchange config
func (z *ZB) GetDefaultConfig() (*config.ExchangeConfig, error) {
	z.SetDefaults()
	exchCfg := new(config.ExchangeConfig)
	exchCfg.Name = z.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = common.JoinStrings(z.BaseCurrencies, ",")

	err := z.SetupDefaults(exchCfg)
	if err != nil {
		return nil, err
	}

	if z.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err = z.UpdateTradablePairs(true)
		if err != nil {
			return nil, err
		}
	}

	return exchCfg, nil
}

// SetDefaults sets default values for the exchange
func (z *ZB) SetDefaults() {
	z.Name = "ZB"
	z.Enabled = true
	z.Verbose = true
	z.APIWithdrawPermissions = exchange.AutoWithdrawCrypto | exchange.NoFiatWithdrawals
	z.API.CredentialsValidator.RequiresKey = true
	z.API.CredentialsValidator.RequiresSecret = true

	z.CurrencyPairs = exchange.CurrencyPairs{
		AssetTypes: assets.AssetTypes{
			assets.AssetTypeSpot,
		},

		UseGlobalPairFormat: true,
		RequestFormat: config.CurrencyPairFormatConfig{
			Delimiter: "_",
		},
		ConfigFormat: config.CurrencyPairFormatConfig{
			Delimiter: "_",
			Uppercase: true,
		},
	}

	z.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: false,

			Trading: exchange.TradingSupported{
				Spot: true,
			},

			RESTCapabilities: exchange.ProtocolFeatures{
				AutoPairUpdates: true,
				TickerBatching:  true,
			},
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}

	z.Requester = request.New(z.Name,
		request.NewRateLimit(time.Second*10, zbAuthRate),
		request.NewRateLimit(time.Second*10, zbUnauthRate),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))

	z.API.Endpoints.URLDefault = zbTradeURL
	z.API.Endpoints.URL = z.API.Endpoints.URLDefault
	z.API.Endpoints.URLSecondaryDefault = zbMarketURL
	z.API.Endpoints.URLSecondary = z.API.Endpoints.URLSecondaryDefault
}

// Setup sets user configuration
func (z *ZB) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		z.SetEnabled(false)
		return nil
	}

	return z.SetupDefaults(exch)
}

// Start starts the OKEX go routine
func (z *ZB) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		z.Run()
		wg.Done()
	}()
}

// Run implements the OKEX wrapper
func (z *ZB) Run() {
	if z.Verbose {
		log.Debugf("%s %d currencies enabled: %s.\n", z.GetName(), len(z.CurrencyPairs.Spot.Enabled), z.CurrencyPairs.Spot.Enabled)
	}

	if !z.GetEnabledFeatures().AutoPairUpdates {
		return
	}

	err := z.UpdateTradablePairs(false)
	if err != nil {
		log.Printf("%s failed to update tradable pairs. Err: %s", z.Name, err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (z *ZB) FetchTradablePairs(asset assets.AssetType) ([]string, error) {
	markets, err := z.GetMarkets()
	if err != nil {
		return nil, err
	}

	var currencies []string
	for x := range markets {
		currencies = append(currencies, x)
	}

	return currencies, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (z *ZB) UpdateTradablePairs(forceUpdate bool) error {
	pairs, err := z.FetchTradablePairs(assets.AssetTypeSpot)
	if err != nil {
		return nil
	}

	return z.UpdatePairs(pairs, assets.AssetTypeSpot, false, forceUpdate)
}

// UpdateTicker updates and returns the ticker for a currency pair
func (z *ZB) UpdateTicker(p pair.CurrencyPair, assetType assets.AssetType) (ticker.Price, error) {
	var tickerPrice ticker.Price

	result, err := z.GetTickers()
	if err != nil {
		return tickerPrice, err
	}

	for _, x := range z.GetEnabledPairs(assetType) {
		currencySplit := common.SplitStrings(z.FormatExchangeCurrency(x, assetType).String(), "_")
		currency := currencySplit[0] + currencySplit[1]
		var tp ticker.Price
		tp.Pair = x
		tp.High = result[currency].High
		tp.Last = result[currency].Last
		tp.Ask = result[currency].Sell
		tp.Bid = result[currency].Buy
		tp.Last = result[currency].Last
		tp.Low = result[currency].Low
		tp.Volume = result[currency].Vol
		ticker.ProcessTicker(z.Name, x, tp, assetType)
	}

	return ticker.GetTicker(z.Name, p, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (z *ZB) FetchTicker(p pair.CurrencyPair, assetType assets.AssetType) (ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(z.GetName(), p, assetType)
	if err != nil {
		return z.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (z *ZB) FetchOrderbook(currency pair.CurrencyPair, assetType assets.AssetType) (orderbook.Base, error) {
	ob, err := orderbook.GetOrderbook(z.GetName(), currency, assetType)
	if err != nil {
		return z.UpdateOrderbook(currency, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (z *ZB) UpdateOrderbook(p pair.CurrencyPair, assetType assets.AssetType) (orderbook.Base, error) {
	var orderBook orderbook.Base
	currency := z.FormatExchangeCurrency(p, assetType).String()

	orderbookNew, err := z.GetOrderbook(currency)
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		orderBook.Bids = append(orderBook.Bids, orderbook.Item{Amount: data[1], Price: data[0]})
	}

	for x := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		orderBook.Asks = append(orderBook.Asks, orderbook.Item{Amount: data[1], Price: data[0]})
	}

	orderbook.ProcessOrderbook(z.GetName(), p, orderBook, assetType)
	return orderbook.GetOrderbook(z.Name, p, assetType)
}

// GetAccountInfo retrieves balances for all enabled currencies for the
// ZB exchange
func (z *ZB) GetAccountInfo() (exchange.AccountInfo, error) {
	var info exchange.AccountInfo
	bal, err := z.GetAccountInformation()
	if err != nil {
		return info, err
	}

	var balances []exchange.AccountCurrencyInfo
	for _, data := range bal.Result.Coins {
		hold, err := strconv.ParseFloat(data.Freez, 64)
		if err != nil {
			return info, err
		}

		avail, err := strconv.ParseFloat(data.Available, 64)
		if err != nil {
			return info, err
		}

		balances = append(balances, exchange.AccountCurrencyInfo{
			CurrencyName: data.EnName,
			TotalValue:   hold + avail,
			Hold:         hold,
		})
	}

	info.ExchangeName = z.GetName()
	info.Currencies = balances
	return info, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (z *ZB) GetFundingHistory() ([]exchange.FundHistory, error) {
	var fundHistory []exchange.FundHistory
	return fundHistory, common.ErrFunctionNotSupported
}

// GetExchangeHistory returns historic trade data since exchange opening.
func (z *ZB) GetExchangeHistory(p pair.CurrencyPair, assetType assets.AssetType) ([]exchange.TradeHistory, error) {
	var resp []exchange.TradeHistory

	return resp, common.ErrNotYetImplemented
}

// SubmitOrder submits a new order
func (z *ZB) SubmitOrder(p pair.CurrencyPair, side exchange.OrderSide, orderType exchange.OrderType, amount, price float64, clientID string) (exchange.SubmitOrderResponse, error) {
	var submitOrderResponse exchange.SubmitOrderResponse
	var oT SpotNewOrderRequestParamsType

	if side == exchange.Buy {
		oT = SpotNewOrderRequestParamsTypeBuy
	} else {
		oT = SpotNewOrderRequestParamsTypeSell
	}

	var params = SpotNewOrderRequestParams{
		Amount: amount,
		Price:  price,
		Symbol: common.StringToLower(p.Pair().String()),
		Type:   oT,
	}
	response, err := z.SpotNewOrder(params)

	if response > 0 {
		submitOrderResponse.OrderID = fmt.Sprintf("%v", response)
	}

	if err == nil {
		submitOrderResponse.IsOrderPlaced = true
	}

	return submitOrderResponse, err
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (z *ZB) ModifyOrder(action exchange.ModifyOrder) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (z *ZB) CancelOrder(order exchange.OrderCancellation) error {
	orderIDInt, err := strconv.ParseInt(order.OrderID, 10, 64)

	if err != nil {
		return err
	}

	return z.CancelExistingOrder(orderIDInt, z.FormatExchangeCurrency(order.CurrencyPair,
		order.AssetType).String())
}

// CancelAllOrders cancels all orders associated with a currency pair
func (z *ZB) CancelAllOrders(orderCancellation exchange.OrderCancellation) (exchange.CancelAllOrdersResponse, error) {
	cancelAllOrdersResponse := exchange.CancelAllOrdersResponse{
		OrderStatus: make(map[string]string),
	}
	var allOpenOrders []UnfinishedOpenOrder
	for _, currency := range z.GetEnabledPairs(assets.AssetTypeSpot) {
		openOrders, err := z.GetUnfinishedOrdersIgnoreTradeType(z.FormatExchangeCurrency(currency,
			assets.AssetTypeSpot).String(), "1", "10")
		if err != nil {
			return cancelAllOrdersResponse, err
		}

		for _, openOrder := range openOrders {
			allOpenOrders = append(allOpenOrders, openOrder)
		}
	}

	for _, openOrder := range allOpenOrders {
		err := z.CancelExistingOrder(openOrder.ID, openOrder.Currency)
		if err != nil {
			cancelAllOrdersResponse.OrderStatus[strconv.FormatInt(openOrder.ID, 10)] = err.Error()
		}
	}

	return cancelAllOrdersResponse, nil
}

// GetOrderInfo returns information on a current open order
func (z *ZB) GetOrderInfo(orderID int64) (exchange.OrderDetail, error) {
	var orderDetail exchange.OrderDetail
	return orderDetail, common.ErrNotYetImplemented
}

// GetDepositAddress returns a deposit address for a specified currency
func (z *ZB) GetDepositAddress(cryptocurrency pair.CurrencyItem) (string, error) {
	return "", common.ErrNotYetImplemented
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (z *ZB) WithdrawCryptocurrencyFunds(withdrawRequest exchange.WithdrawRequest) (string, error) {
	return z.Withdraw(withdrawRequest.Currency.Lower().String(), withdrawRequest.Address, withdrawRequest.TradePassword, withdrawRequest.Amount, withdrawRequest.FeeAmount, false)
}

// WithdrawFiatFunds returns a withdrawal ID when a
// withdrawal is submitted
func (z *ZB) WithdrawFiatFunds(withdrawRequest exchange.WithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a
// withdrawal is submitted
func (z *ZB) WithdrawFiatFundsToInternationalBank(withdrawRequest exchange.WithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// GetWebsocket returns a pointer to the exchange websocket
func (z *ZB) GetWebsocket() (*exchange.Websocket, error) {
	return nil, common.ErrNotYetImplemented
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (z *ZB) GetFeeByType(feeBuilder exchange.FeeBuilder) (float64, error) {
	return z.GetFee(feeBuilder)
}

// GetWithdrawCapabilities returns the types of withdrawal methods permitted by the exchange
func (z *ZB) GetWithdrawCapabilities() uint32 {
	return z.GetWithdrawPermissions()
}
