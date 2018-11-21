package exchange

import (
	"sync"

	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges/assets"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

// IBotExchange enforces standard functions for all exchanges supported in
// GoCryptoTrader
type IBotExchange interface {
	Setup(exch *config.ExchangeConfig) error
	Start(wg *sync.WaitGroup)
	SetDefaults()
	GetName() string
	IsEnabled() bool
	SetEnabled(bool)
	FetchTicker(currency pair.CurrencyPair, assetType assets.AssetType) (ticker.Price, error)
	UpdateTicker(currency pair.CurrencyPair, assetType assets.AssetType) (ticker.Price, error)
	FetchOrderbook(currency pair.CurrencyPair, assetType assets.AssetType) (orderbook.Base, error)
	UpdateOrderbook(currency pair.CurrencyPair, assetType assets.AssetType) (orderbook.Base, error)
	FetchTradablePairs(assetType assets.AssetType) ([]string, error)
	UpdateTradablePairs(forceUpdate bool) error
	GetEnabledPairs(assetType assets.AssetType) []pair.CurrencyPair
	GetAvailablePairs(assetType assets.AssetType) []pair.CurrencyPair
	GetAccountInfo() (AccountInfo, error)
	GetAuthenticatedAPISupport() bool
	SetPairs(pairs []pair.CurrencyPair, assetType assets.AssetType, enabled bool) error
	GetAssetTypes() assets.AssetTypes
	GetExchangeHistory(currencyPair pair.CurrencyPair, assetType assets.AssetType) ([]TradeHistory, error)
	SupportsAutoPairUpdates() bool
	SupportsRESTTickerBatchUpdates() bool
	GetLastPairsUpdateTime() int64

	GetWithdrawPermissions() uint32
	FormatWithdrawPermissions() string
	SupportsWithdrawPermissions(permissions uint32) bool

	GetFundingHistory() ([]FundHistory, error)
	SubmitOrder(p pair.CurrencyPair, side OrderSide, orderType OrderType, amount, price float64, clientID string) (SubmitOrderResponse, error)
	ModifyOrder(action ModifyOrder) (string, error)
	CancelOrder(order OrderCancellation) error
	CancelAllOrders(orders OrderCancellation) (CancelAllOrdersResponse, error)
	GetOrderInfo(orderID int64) (OrderDetail, error)
	GetDepositAddress(cryptocurrency pair.CurrencyItem) (string, error)

	WithdrawCryptocurrencyFunds(wtihdrawRequest WithdrawRequest) (string, error)
	WithdrawFiatFunds(wtihdrawRequest WithdrawRequest) (string, error)
	WithdrawFiatFundsToInternationalBank(wtihdrawRequest WithdrawRequest) (string, error)

	SetHTTPClientUserAgent(ua string)
	GetHTTPClientUserAgent() string
	SetClientProxyAddress(addr string) error

	SupportsWebsocket() bool
	SupportsREST() bool
	IsWebsocketEnabled() bool
	GetWebsocket() (*Websocket, error)
	GetDefaultConfig() (*config.ExchangeConfig, error)
}
