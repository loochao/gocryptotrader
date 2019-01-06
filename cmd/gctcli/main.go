package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/core"
	"github.com/thrasher-/gocryptotrader/gctrpc/auth"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultHost     = "localhost:9052"
	defualtUsername = "admin"
	defaultPassword = "Password"
)

func jsonOutput(in interface{}) {
	j, err := json.MarshalIndent(in, "", " ")
	if err != nil {
		return
	}
	fmt.Print(string(j))
}

func setupClient() (*grpc.ClientConn, error) {
	targetPath := filepath.Join(common.GetDefaultDataDir(runtime.GOOS), "tls", "cert.pem")
	creds, err := credentials.NewClientTLSFromFile(targetPath, "")
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(auth.BasicAuth{
			Username: defualtUsername,
			Password: defaultPassword,
		}),
	}
	conn, err := grpc.Dial(defaultHost, opts...)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func main() {
	app := cli.NewApp()
	app.Name = "gctcli"
	app.Version = core.Version(true)
	app.Usage = "command line interface for managing the gocryptotrader daemon"
	app.Commands = []cli.Command{
		getExchangesCommand,
		enableExchangeCommand,
		disableExchangeCommand,
		getTickerCommand,
		getTickersCommand,
		getOrderbookCommand,
		getOrderbooksCommand,
		getConfigCommand,
		getPortfolioCommand,
		addPortfolioAddressCommand,
		removePortfolioAddressCommand,
		getForexRatesCommand,
		getOrdersCommand,
		getOrderCommand,
		submitOrderCommand,
		cancelOrderCommand,
		cancelAllOrdersCommand,
		getEventsCommand,
		addEventCommand,
		removeEventCommand,
		getCryptocurrencyDepositAddressesCommand,
		getCryptocurrencyDepositAddressCommand,
		withdrawCryptocurrencyFundsCommand,
		withdrawFiatFundsCommand,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
