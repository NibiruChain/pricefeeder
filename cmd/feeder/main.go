package main

import (
	"github.com/NibiruChain/nibiru/app"
	"github.com/NibiruChain/price-feeder/config"
)

func main() {
	app.SetPrefixes(app.AccountAddressPrefix)
	conf := config.MustGet()
	f := conf.Feeder()

	defer f.Close()
	select {}
}
