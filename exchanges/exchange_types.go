package exchange

import (
	"fmt"
	"time"

	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges/assets"
	"github.com/thrasher-/gocryptotrader/exchanges/nonce"
	"github.com/thrasher-/gocryptotrader/exchanges/request"
)

// FeeType custom type for calculating fees based on method
type FeeType string

// InternationalBankTransactionType custom type for calculating fees based on fiat transaction types
type InternationalBankTransactionType string

// Const vars for the exchange package
const (
	BankFee                        FeeType = "bankFee"
	InternationalBankDepositFee    FeeType = "internationalBankDepositFee"
	InternationalBankWithdrawalFee FeeType = "internationalBankWithdrawalFee"
	CryptocurrencyTradeFee         FeeType = "cryptocurrencyTradeFee"
	CyptocurrencyDepositFee        FeeType = "cyptocurrencyDepositFee"
	CryptocurrencyWithdrawalFee    FeeType = "cryptocurrencyWithdrawalFee"

	WireTransfer    InternationalBankTransactionType = "wireTransfer"
	PerfectMoney    InternationalBankTransactionType = "perfectMoney"
	Neteller        InternationalBankTransactionType = "neteller"
	AdvCash         InternationalBankTransactionType = "advCash"
	Payeer          InternationalBankTransactionType = "payeer"
	Skrill          InternationalBankTransactionType = "skrill"
	Simplex         InternationalBankTransactionType = "simplex"
	SEPA            InternationalBankTransactionType = "sepa"
	Swift           InternationalBankTransactionType = "swift"
	RapidTransfer   InternationalBankTransactionType = "rapidTransfer"
	MisterTangoSEPA InternationalBankTransactionType = "misterTangoSepa"
	Qiwi            InternationalBankTransactionType = "qiwi"
	VisaMastercard  InternationalBankTransactionType = "visaMastercard"
	WebMoney        InternationalBankTransactionType = "webMoney"
	Capitalist      InternationalBankTransactionType = "capitalist"
	WesternUnion    InternationalBankTransactionType = "westernUnion"
	MoneyGram       InternationalBankTransactionType = "moneyGram"
	Contact         InternationalBankTransactionType = "contact"
)

// SubmitOrderResponse is what is returned after submitting an order to an exchange
type SubmitOrderResponse struct {
	IsOrderPlaced bool
	OrderID       string
}

// FeeBuilder is the type which holds all parameters required to calculate a fee for an exchange
type FeeBuilder struct {
	FeeType FeeType
	//Used for calculating crypto trading fees, deposits & withdrawals
	FirstCurrency  string
	SecondCurrency string
	Delimiter      string
	IsMaker        bool
	// Fiat currency used for bank deposits & withdrawals
	CurrencyItem        string
	BankTransactionType InternationalBankTransactionType
	// Used to multiply for fee calculations
	PurchasePrice float64
	Amount        float64
}

// Definitions for each type of withdrawal method for a given exchange
const (
	// No withdraw
	NoAPIWithdrawalMethods                  uint32 = 0
	NoAPIWithdrawalMethodsText              string = "NONE, WEBSITE ONLY"
	AutoWithdrawCrypto                      uint32 = (1 << 0)
	AutoWithdrawCryptoWithAPIPermission     uint32 = (1 << 1)
	AutoWithdrawCryptoWithSetup             uint32 = (1 << 2)
	AutoWithdrawCryptoText                  string = "AUTO WITHDRAW CRYPTO"
	AutoWithdrawCryptoWithAPIPermissionText string = "AUTO WITHDRAW CRYPTO WITH API PERMISSION"
	AutoWithdrawCryptoWithSetupText         string = "AUTO WITHDRAW CRYPTO WITH SETUP"
	WithdrawCryptoWith2FA                   uint32 = (1 << 3)
	WithdrawCryptoWithSMS                   uint32 = (1 << 4)
	WithdrawCryptoWithEmail                 uint32 = (1 << 5)
	WithdrawCryptoWithWebsiteApproval       uint32 = (1 << 6)
	WithdrawCryptoWithAPIPermission         uint32 = (1 << 7)
	WithdrawCryptoWith2FAText               string = "WITHDRAW CRYPTO WITH 2FA"
	WithdrawCryptoWithSMSText               string = "WITHDRAW CRYPTO WITH SMS"
	WithdrawCryptoWithEmailText             string = "WITHDRAW CRYPTO WITH EMAIL"
	WithdrawCryptoWithWebsiteApprovalText   string = "WITHDRAW CRYPTO WITH WEBSITE APPROVAL"
	WithdrawCryptoWithAPIPermissionText     string = "WITHDRAW CRYPTO WITH API PERMISSION"
	AutoWithdrawFiat                        uint32 = (1 << 8)
	AutoWithdrawFiatWithAPIPermission       uint32 = (1 << 9)
	AutoWithdrawFiatWithSetup               uint32 = (1 << 10)
	AutoWithdrawFiatText                    string = "AUTO WITHDRAW FIAT"
	AutoWithdrawFiatWithAPIPermissionText   string = "AUTO WITHDRAW FIAT WITH API PERMISSION"
	AutoWithdrawFiatWithSetupText           string = "AUTO WITHDRAW FIAT WITH SETUP"
	WithdrawFiatWith2FA                     uint32 = (1 << 11)
	WithdrawFiatWithSMS                     uint32 = (1 << 12)
	WithdrawFiatWithEmail                   uint32 = (1 << 13)
	WithdrawFiatWithWebsiteApproval         uint32 = (1 << 14)
	WithdrawFiatWithAPIPermission           uint32 = (1 << 15)
	WithdrawFiatWith2FAText                 string = "WITHDRAW FIAT WITH 2FA"
	WithdrawFiatWithSMSText                 string = "WITHDRAW FIAT WITH SMS"
	WithdrawFiatWithEmailText               string = "WITHDRAW FIAT WITH EMAIL"
	WithdrawFiatWithWebsiteApprovalText     string = "WITHDRAW FIAT WITH WEBSITE APPROVAL"
	WithdrawFiatWithAPIPermissionText       string = "WITHDRAW FIAT WITH API PERMISSION"
	WithdrawCryptoViaWebsiteOnly            uint32 = (1 << 16)
	WithdrawFiatViaWebsiteOnly              uint32 = (1 << 17)
	WithdrawCryptoViaWebsiteOnlyText        string = "WITHDRAW CRYPTO VIA WEBSITE ONLY"
	WithdrawFiatViaWebsiteOnlyText          string = "WITHDRAW FIAT VIA WEBSITE ONLY"
	NoFiatWithdrawals                       uint32 = (1 << 18)
	NoFiatWithdrawalsText                   string = "NO FIAT WITHDRAWAL"

	UnknownWithdrawalTypeText string = "UNKNOWN"
)

// ModifyOrder is a an order modifyer
// ModifyOrder is a an order modifyer
type ModifyOrder struct {
	OrderID string
	OrderType
	OrderSide
	Price           float64
	Amount          float64
	LimitPriceUpper float64
	LimitPriceLower float64
	Currency        pair.CurrencyPair

	ImmediateOrCancel bool
	HiddenOrder       bool
	FillOrKill        bool
	PostOnly          bool
}

// ModifyOrderResponse is an order modifying return type
type ModifyOrderResponse struct {
	OrderID string
}

// CancelAllOrdersResponse returns the status from attempting to cancel all orders on an exchagne
type CancelAllOrdersResponse struct {
	OrderStatus map[string]string
}

// OrderType enforces a standard for Ordertypes across the code base
type OrderType string

// OrderType ...types
const (
	Limit  OrderType = "Limit"
	Market OrderType = "Market"
)

// ToString changes the ordertype to the exchange standard and returns a string
func (o OrderType) ToString() string {
	return fmt.Sprintf("%v", o)
}

// OrderSide enforces a standard for OrderSides across the code base
type OrderSide string

// OrderSide types
const (
	Buy  OrderSide = "Buy"
	Sell OrderSide = "Sell"
)

// ToString changes the ordertype to the exchange standard and returns a string
func (o OrderSide) ToString() string {
	return fmt.Sprintf("%v", o)
}

// AccountInfo is a Generic type to hold each exchange's holdings in
// all enabled currencies
type AccountInfo struct {
	ExchangeName string
	Currencies   []AccountCurrencyInfo
}

// AccountCurrencyInfo is a sub type to store currency name and value
type AccountCurrencyInfo struct {
	CurrencyName string
	TotalValue   float64
	Hold         float64
}

// TradeHistory holds exchange history data
type TradeHistory struct {
	Timestamp int64
	TID       int64
	Price     float64
	Amount    float64
	Exchange  string
	Type      string
}

// OrderDetail holds order detail data
type OrderDetail struct {
	Exchange      string
	ID            string
	BaseCurrency  string
	QuoteCurrency string
	OrderSide     string
	OrderType     string
	CreationTime  int64
	Status        string
	Price         float64
	Amount        float64
	OpenVolume    float64
}

// OrderCancellation type requred when requesting to cancel an order
type OrderCancellation struct {
	AccountID     string
	OrderID       string
	CurrencyPair  pair.CurrencyPair
	AssetType     assets.AssetType
	WalletAddress string
	Side          OrderSide
}

// FundHistory holds exchange funding history data
type FundHistory struct {
	ExchangeName      string
	Status            string
	TransferID        int64
	Description       string
	Timestamp         int64
	Currency          string
	Amount            float64
	Fee               float64
	TransferType      string
	CryptoToAddress   string
	CryptoFromAddress string
	CryptoTxID        string
	BankTo            string
	BankFrom          string
}

// WithdrawRequest used for wrapper crypto and FIAT withdraw methods
type WithdrawRequest struct {
	// General withdraw information
	Description     string
	OneTimePassword int64
	AccountID       string
	PIN             int64
	TradePassword   string
	Amount          float64
	Currency        pair.CurrencyItem
	// Crypto related information
	Address    string
	AddressTag string
	FeeAmount  float64
	// FIAT related information
	BankAccountName   string
	BankAccountNumber float64
	BankName          string
	BankAddress       string
	BankCity          string
	BankCountry       string
	BankPostalCode    string
	SwiftCode         string
	IBAN              string
	BankCode          float64
	IsExpressWire     bool
	// Intermediary bank information
	RequiresIntermediaryBank      bool
	IntermediaryBankAccountNumber float64
	IntermediaryBankName          string
	IntermediaryBankAddress       string
	IntermediaryBankCity          string
	IntermediaryBankCountry       string
	IntermediaryBankPostalCode    string
	IntermediarySwiftCode         string
	IntermediaryBankCode          float64
	IntermediaryIBAN              string
	WireCurrency                  string
}

// Features stores the supported and enabled features
// for the exchange
type Features struct {
	Supports FeaturesSupported
	Enabled  FeaturesEnabled
}

// FeaturesEnabled stores the exchange enabled features
type FeaturesEnabled struct {
	AutoPairUpdates bool
}

// ProtocolFeatures holds all variables for the exchanges supported features
// for a protocol (e.g REST or Websocket)
type ProtocolFeatures struct {
	TickerBatching      bool
	TickerFetching      bool
	OrderbookFetching   bool
	AutoPairUpdates     bool
	AccountInfo         bool
	CryptoDeposit       bool
	CryptoWithdrawal    uint32
	FiatWithdraw        bool
	GetOrder            bool
	GetOrders           bool
	CancelOrders        bool
	CancelOrder         bool
	SubmitOrder         bool
	SubmitOrders        bool
	ModifyOrder         bool
	DepositHistory      bool
	WithdrawalHistory   bool
	TradeHistory        bool
	UserTradeHistory    bool
	TradeFee            bool
	FiatDepositFee      bool
	FiatWithdrawalFee   bool
	CryptoDepositFee    bool
	CryptoWithdrawalFee bool
}

// TradingSupported stores the exchanges supported
// trading assets
type TradingSupported struct {
	Spot           bool
	Futures        bool
	Margin         bool
	PerpetualSwaps bool
	Index          bool
}

// FeaturesSupported stores the exchanges supported features
type FeaturesSupported struct {
	REST                  bool
	RESTCapabilities      ProtocolFeatures
	Websocket             bool
	WebsocketCapabilities ProtocolFeatures
	Trading               TradingSupported
}

// API stores the exchange API settings
type API struct {
	AuthenticatedSupport bool
	PEMKeySupport        bool

	Endpoints struct {
		URL                 string
		URLDefault          string
		URLSecondary        string
		URLSecondaryDefault string
		WebsocketURL        string
	}

	Credentials struct {
		Key      string
		Secret   string
		ClientID string
		PEMKey   string
	}

	CredentialsValidator struct {
		// For Huobi (optional)
		RequiresPEM bool

		RequiresKey                bool
		RequiresSecret             bool
		RequiresClientID           bool
		RequiresBase64DecodeSecret bool
	}
}

// CurrencyPair stores a list of enable/available
// currency pairs and their storage/request format
type CurrencyPair struct {
	Enabled       []string
	Available     []string
	RequestFormat config.CurrencyPairFormatConfig
	ConfigFormat  config.CurrencyPairFormatConfig
}

// CurrencyPairs stores a list of tradable currency pair settings
type CurrencyPairs struct {
	RequestFormat       config.CurrencyPairFormatConfig
	ConfigFormat        config.CurrencyPairFormatConfig
	UseGlobalPairFormat bool
	LastUpdated         int64
	AssetTypes          assets.AssetTypes
	Spot                CurrencyPair
	Futures             CurrencyPair
}

// Base stores the individual exchange information
type Base struct {
	Name    string
	Enabled bool
	Verbose bool

	APIWithdrawPermissions uint32
	API                    API
	Nonce                  nonce.Nonce
	BaseCurrencies         []string
	CurrencyPairs          CurrencyPairs

	Features      Features
	HTTPTimeout   time.Duration
	HTTPUserAgent string
	Websocket     *Websocket
	*request.Requester

	LoadedByConfig bool
	Config         *config.ExchangeConfig
}
