package engine

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/thrasher-/gocryptotrader/common"
	log "github.com/thrasher-/gocryptotrader/logger"
)

// RESTLogger logs the requests internally
func RESTLogger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)

		log.Debugf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

// StartRESTServer starts a REST server
func StartRESTServer() {
	listenAddr := Bot.Config.RemoteControl.DeprecatedRPC.ListenAddress
	log.Debugf("Deprecated RPC server support enabled. Listen URL: http://%s:%d\n", common.ExtractHost(listenAddr), common.ExtractPort(listenAddr))
	err := http.ListenAndServe(listenAddr, NewRouter(true))
	if err != nil {
		log.Fatal(err)
	}
}

// StartWebsocketServer starts a Websocket server
func StartWebsocketServer() {
	listenAddr := Bot.Config.RemoteControl.WebsocketRPC.ListenAddress
	log.Debugf("Websocket RPC support enabled. Listen URL: ws://%s:%d/ws\n", common.ExtractHost(listenAddr), common.ExtractPort(listenAddr))
	err := http.ListenAndServe(listenAddr, NewRouter(false))
	if err != nil {
		log.Fatal(err)
	}
}

// NewRouter takes in the exchange interfaces and returns a new multiplexor
// router
func NewRouter(isREST bool) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	var routes []Route
	listenAddr := Bot.Config.RemoteControl.DeprecatedRPC.ListenAddress

	if isREST {
		routes = []Route{
			Route{"", "GET", "/", getIndex},
			Route{"GetAllSettings", "GET", "/config/all", RESTGetAllSettings},
			Route{"SaveAllSettings", "POST", "/config/all/save", RESTSaveAllSettings},
			Route{"AllEnabledAccountInfo", "GET", "/exchanges/enabled/accounts/all", RESTGetAllEnabledAccountInfo},
			Route{"AllActiveExchangesAndCurrencies", "GET", "/exchanges/enabled/latest/all", RESTGetAllActiveTickers},
			Route{"GetPortfolio", "GET", "/portfolio/all", RESTGetPortfolio},
			Route{"AllActiveExchangesAndOrderbooks", "GET", "/exchanges/orderbook/latest/all", RESTGetAllActiveOrderbooks},
		}
	} else {
		listenAddr = Bot.Config.RemoteControl.WebsocketRPC.ListenAddress
		routes = []Route{
			Route{"ws", "GET", "/ws", WebsocketClientHandler},
		}
	}

	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = RESTLogger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler).
			Host(common.ExtractHost(listenAddr))
	}
	return router
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html>GoCryptoTrader RESTful interface. For the web GUI, please visit the <a href=https://github.com/thrasher-/gocryptotrader/blob/master/web/README.md>web GUI readme.</a></html>")
	w.WriteHeader(http.StatusOK)
}
