package main

import (
	"fmt"
	"log"
	"time"

	"github.com/luxingwen/pnet/config"
	"github.com/luxingwen/pnet/node"
)

func main() {
	fmt.Println("hello")

	cfg := config.DefaultConfig()
	cfg.Hostname = "127.0.0.1"
	cfg.Port = 23333
	locl, err := node.NewLocalNode("hello1", cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = locl.Start()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("node str:", locl.Node.String())

	cfg2 := config.DefaultConfig()
	cfg2.Hostname = "127.0.0.1"
	cfg2.Port = 23334

	locl2, err := node.NewLocalNode("hello2", cfg2)
	if err != nil {
		log.Fatal(err)
	}

	err = locl2.Start()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second * 3)

	fmt.Println("node2 str:", locl2.Node.String())

	remoteNode, ok, err := locl2.Connect(locl.Node.Node)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("ok:", ok)

	fmt.Println("remoteNode:", remoteNode.Node.String())

	select {}
}
