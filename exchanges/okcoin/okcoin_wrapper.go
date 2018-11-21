package okcoin

import (
	"errors"
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
func (o *OKCoin) GetDefaultConfig() (*config.ExchangeConfig, error) {
	o.SetDefaults()
	exchCfg := new(config.ExchangeConfig)
	exchCfg.Name = o.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = common.JoinStrings(o.BaseCurrencies, ",")

	err := o.SetupDefaults(exchCfg)
	if err != nil {
		return nil, err
	}

	if o.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err = o.UpdateTradablePairs(true)
		if err != nil {
			return nil, err
		}
	}

	return exchCfg, nil
}

// SetDefaults sets current default values for this package
func (o *OKCoin) SetDefaults() {
	o.SetErrorDefaults()
	o.SetWebsocketErrorDefaults()
	o.Name = "OKCOIN International"
	o.Enabled = true
	o.Verbose = true
	o.APIWithdrawPermissions = exchange.AutoWithdrawCrypto | exchange.WithdrawFiatViaWebsiteOnly
	o.API.CredentialsValidator.RequiresKey = true
	o.API.CredentialsValidator.RequiresSecret = true

	o.CurrencyPairs = exchange.CurrencyPairs{
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

	o.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,

			Trading: exchange.TradingSupported{
				Spot:   true,
				Margin: true,
			},

			RESTCapabilities: exchange.ProtocolFeatures{
				AutoPairUpdates: true,
				TickerBatching:  false,
			},
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}

	o.Requester = request.New(o.Name,
		request.NewRateLimit(time.Second, okcoinAuthRate),
		request.NewRateLimit(time.Second, okcoinUnauthRate),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))

	o.API.Endpoints.URLDefault = okcoinAPIURL
	o.API.Endpoints.URL = o.API.Endpoints.URLDefault
	o.API.Endpoints.WebsocketURL = okcoinWebsocketURL
	o.WebsocketInit()
}

// Setup sets exchange configuration parameters
func (o *OKCoin) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		o.SetEnabled(false)
		return nil
	}

	err := o.SetupDefaults(exch)
	if err != nil {
		return err
	}

	return o.WebsocketSetup(o.WsConnect,
		exch.Name,
		exch.Features.Enabled.Websocket,
		okcoinWebsocketURL,
		o.API.Endpoints.WebsocketURL)
}

// Start starts the OKCoin go routine
func (o *OKCoin) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		o.Run()
		wg.Done()
	}()
}

// Run implements the OKCoin wrapper
func (o *OKCoin) Run() {
	if o.Verbose {
		log.Debugf("%s Websocket: %s. (url: %s).\n", o.GetName(), common.IsEnabled(o.Websocket.IsEnabled()), o.API.Endpoints.WebsocketURL)
		log.Debugf("%s %d currencies enabled: %s.\n", o.GetName(), len(o.CurrencyPairs.Spot.Enabled), o.CurrencyPairs.Spot.Enabled)
	}

	forceUpdate := false
	if !common.StringDataContains(o.CurrencyPairs.Spot.Enabled, "_") || !common.StringDataContains(o.CurrencyPairs.Spot.Available, "_") {
		forceUpdate = true
		enabledPairs := []string{"btc_usd"}
		log.Warn("WARNING: Available pairs for OKCoin International reset due to config upgrade, please enable the pairs you would like again.")

		err := o.UpdatePairs(enabledPairs, assets.AssetTypeSpot, true, true)
		if err != nil {
			log.Errorf("%s failed to update enabled currencies. Err: %s", o.Name, err)
		}
	}

	if !o.GetEnabledFeatures().AutoPairUpdates && !forceUpdate {
		return
	}

	err := o.UpdateTradablePairs(forceUpdate)
	if err != nil {
		log.Errorf("%s failed to update tradable pairs. Err: %s", o.Name, err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (o *OKCoin) FetchTradablePairs(asset assets.AssetType) ([]string, error) {
	prods, err := o.GetSpotInstruments()
	if err != nil {
		return nil, err
	}

	var pairs []string
	for x := range prods {
		pairs = append(pairs, prods[x].BaseCurrency+"_"+prods[x].QuoteCurrency)
	}

	return pairs, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (o *OKCoin) UpdateTradablePairs(forceUpdate bool) error {
	pairs, err := o.FetchTradablePairs(assets.AssetTypeSpot)
	if err != nil {
		return err
	}

	return o.UpdatePairs(pairs, assets.AssetTypeSpot, false, forceUpdate)
}

// UpdateTicker updates and returns the ticker for a currency pair
func (o *OKCoin) UpdateTicker(p pair.CurrencyPair, assetType assets.AssetType) (ticker.Price, error) {
	currency := o.FormatExchangeCurrency(p, assetType).String()
	var tickerPrice ticker.Price

	if assetType != assets.AssetTypeSpot && o.API.Endpoints.URL == okcoinAPIURL {
		tick, err := o.GetFuturesTicker(currency, assetType.String())
		if err != nil {
			return tickerPrice, err
		}
		tickerPrice.Pair = p
		tickerPrice.Ask = tick.Sell
		tickerPrice.Bid = tick.Buy
		tickerPrice.Low = tick.Low
		tickerPrice.Last = tick.Last
		tickerPrice.Volume = tick.Vol
		tickerPrice.High = tick.High
		ticker.ProcessTicker(o.GetName(), p, tickerPrice, assetType)
	} else {
		tick, err := o.GetTicker(currency)
		if err != nil {
			return tickerPrice, err
		}
		tickerPrice.Pair = p
		tickerPrice.Ask = tick.Sell
		tickerPrice.Bid = tick.Buy
		tickerPrice.Low = tick.Low
		tickerPrice.Last = tick.Last
		tickerPrice.Volume = tick.Vol
		tickerPrice.High = tick.High
		ticker.ProcessTicker(o.GetName(), p, tickerPrice, assets.AssetTypeSpot)

	}
	return ticker.GetTicker(o.Name, p, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (o *OKCoin) FetchTicker(p pair.CurrencyPair, assetType assets.AssetType) (ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(o.GetName(), p, assetType)
	if err != nil {
		return o.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (o *OKCoin) FetchOrderbook(currency pair.CurrencyPair, assetType assets.AssetType) (orderbook.Base, error) {
	ob, err := orderbook.GetOrderbook(o.GetName(), currency, assetType)
	if err != nil {
		return o.UpdateOrderbook(currency, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (o *OKCoin) UpdateOrderbook(currency pair.CurrencyPair, assetType assets.AssetType) (orderbook.Base, error) {
	var orderBook orderbook.Base
	orderbookNew, err := o.GetOrderBook(o.FormatExchangeCurrency(currency, assetType).String(), 200, false)
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

	orderbook.ProcessOrderbook(o.GetName(), currency, orderBook, assetType)
	return orderbook.GetOrderbook(o.Name, currency, assetType)
}

// GetAccountInfo retrieves balances for all enabled currencies for the
// OKCoin exchange
func (o *OKCoin) GetAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.ExchangeName = o.GetName()
	assets, err := o.GetUserInfo()
	if err != nil {
		return response, err
	}

	response.Currencies = append(response.Currencies, exchange.AccountCurrencyInfo{
		CurrencyName: "BTC",
		TotalValue:   assets.Info.Funds.Free.BTC,
		Hold:         assets.Info.Funds.Freezed.BTC,
	})

	response.Currencies = append(response.Currencies, exchange.AccountCurrencyInfo{
		CurrencyName: "LTC",
		TotalValue:   assets.Info.Funds.Free.LTC,
		Hold:         assets.Info.Funds.Freezed.LTC,
	})

	response.Currencies = append(response.Currencies, exchange.AccountCurrencyInfo{
		CurrencyName: "USD",
		TotalValue:   assets.Info.Funds.Free.USD,
		Hold:         assets.Info.Funds.Freezed.USD,
	})

	response.Currencies = append(response.Currencies, exchange.AccountCurrencyInfo{
		CurrencyName: "CNY",
		TotalValue:   assets.Info.Funds.Free.CNY,
		Hold:         assets.Info.Funds.Freezed.CNY,
	})

	return response, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (o *OKCoin) GetFundingHistory() ([]exchange.FundHistory, error) {
	var fundHistory []exchange.FundHistory
	return fundHistory, common.ErrFunctionNotSupported
}

// GetExchangeHistory returns historic trade data since exchange opening.
func (o *OKCoin) GetExchangeHistory(p pair.CurrencyPair, assetType assets.AssetType) ([]exchange.TradeHistory, error) {
	var resp []exchange.TradeHistory

	return resp, common.ErrNotYetImplemented
}

// SubmitOrder submits a new order
func (o *OKCoin) SubmitOrder(p pair.CurrencyPair, side exchange.OrderSide, orderType exchange.OrderType, amount, price float64, clientID string) (exchange.SubmitOrderResponse, error) {
	var submitOrderResponse exchange.SubmitOrderResponse
	var oT string
	if orderType == exchange.Limit {
		if side == exchange.Buy {
			oT = "buy"
		} else {
			oT = "sell"
		}
	} else if orderType == exchange.Market {
		if side == exchange.Buy {
			oT = "buy_market"
		} else {
			oT = "sell_market"
		}
	} else {
		return submitOrderResponse, errors.New("Unsupported order type")
	}

	response, err := o.Trade(amount, price, p.Pair().String(), oT)

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
func (o *OKCoin) ModifyOrder(action exchange.ModifyOrder) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (o *OKCoin) CancelOrder(order exchange.OrderCancellation) error {
	orderIDInt, err := strconv.ParseInt(order.OrderID, 10, 64)
	orders := []int64{orderIDInt}

	if err != nil {
		return err
	}

	resp, err := o.CancelExistingOrder(orders, o.FormatExchangeCurrency(order.CurrencyPair,
		assets.AssetTypeSpot).String())
	if !resp.Result {
		return errors.New(resp.ErrorCode)
	}
	return err
}

// CancelAllOrders cancels all orders associated with a currency pair
func (o *OKCoin) CancelAllOrders(orderCancellation exchange.OrderCancellation) (exchange.CancelAllOrdersResponse, error) {
	cancelAllOrdersResponse := exchange.CancelAllOrdersResponse{
		OrderStatus: make(map[string]string),
	}
	orderInfo, err := o.GetOrderInformation(-1, o.FormatExchangeCurrency(orderCancellation.CurrencyPair,
		assets.AssetTypeSpot).String())
	if err != nil {
		return cancelAllOrdersResponse, err
	}

	var ordersToCancel []int64
	for _, order := range orderInfo {
		ordersToCancel = append(ordersToCancel, order.OrderID)
	}

	if len(ordersToCancel) > 0 {
		resp, err := o.CancelExistingOrder(ordersToCancel, o.FormatExchangeCurrency(orderCancellation.CurrencyPair,
			assets.AssetTypeSpot).String())
		if err != nil {
			return cancelAllOrdersResponse, err
		}

		for _, order := range common.SplitStrings(resp.ErrorCode, ",") {
			if err != nil {
				cancelAllOrdersResponse.OrderStatus[order] = "Order could not be cancelled"
			}
		}
	}

	return cancelAllOrdersResponse, nil
}

// GetOrderInfo returns information on a current open order
func (o *OKCoin) GetOrderInfo(orderID int64) (exchange.OrderDetail, error) {
	var orderDetail exchange.OrderDetail
	return orderDetail, common.ErrNotYetImplemented
}

// GetDepositAddress returns a deposit address for a specified currency
func (o *OKCoin) GetDepositAddress(cryptocurrency pair.CurrencyItem) (string, error) {
	return "", common.ErrNotYetImplemented
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (o *OKCoin) WithdrawCryptocurrencyFunds(withdrawRequest exchange.WithdrawRequest) (string, error) {
	resp, err := o.Withdrawal(withdrawRequest.Currency.String(), withdrawRequest.FeeAmount, withdrawRequest.TradePassword, withdrawRequest.Address, withdrawRequest.Amount)
	return fmt.Sprintf("%v", resp), err
}

// WithdrawFiatFunds returns a withdrawal ID when a
// withdrawal is submitted
func (o *OKCoin) WithdrawFiatFunds(withdrawRequest exchange.WithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a
// withdrawal is submitted
func (o *OKCoin) WithdrawFiatFundsToInternationalBank(withdrawRequest exchange.WithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// GetWebsocket returns a pointer to the exchange websocket
func (o *OKCoin) GetWebsocket() (*exchange.Websocket, error) {
	return o.Websocket, nil
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (o *OKCoin) GetFeeByType(feeBuilder exchange.FeeBuilder) (float64, error) {
	return o.GetFee(feeBuilder)
}

// GetWithdrawCapabilities returns the types of withdrawal methods permitted by the exchange
func (o *OKCoin) GetWithdrawCapabilities() uint32 {
	return o.GetWithdrawPermissions()
}
