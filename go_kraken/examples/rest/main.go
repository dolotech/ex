package main

import (
	"log"
	"os"

	"github.com/aopoltorzhicky/go_kraken/rest"
)

//https://futures.kraken.com/derivatives/api/v4/charts/trade/PF_XBTUSD/15m?from=1721110189&to=1721214529

func main() {
	api := rest.New(os.Getenv("KRAKEN_API_KEY"), os.Getenv("KRAKEN_SECRET"))
	data, err := api.Ticker("XXBTZUSD")
	if err != nil {
		log.Panicln(err)
		return
	}
	for _, ticker := range data {
		log.Printf("ask %s", ticker.Ask.Price)
	}
}
