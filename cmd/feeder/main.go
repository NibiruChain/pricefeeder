package main

import "github.com/NibiruChain/price-feeder/config"

func main() {
	conf := config.MustGet()
	f := conf.Feeder()

	defer f.Close()
	select {}
}
