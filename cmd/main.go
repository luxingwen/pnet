package main

import (
	"log"

	"github.com/luxingwen/pnet"
	"github.com/luxingwen/pnet/config"
)

func main() {
	conf := &config.Config{}
	conf.Port = 9000
	conf.Name = "osp"

	pn, err := pnet.NewPNet("tes1", conf)
	if err != nil {
		log.Fatal(err)
	}

	pn.Start()

	pn.Join("tcp://wwh.biggerforum.org:9001")

	select {}
}
