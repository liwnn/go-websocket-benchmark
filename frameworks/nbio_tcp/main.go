package main

import (
	"flag"
	"fmt"

	"go-websocket-benchmark/config"
	"go-websocket-benchmark/frameworks"
	"go-websocket-benchmark/logging"

	"github.com/lesismal/nbio"
)

func main() {
	flag.Parse()

	addrs, err := config.GetFrameworkServerAddrs(config.NbioTcp)
	if err != nil {
		logging.Fatalf("GetFrameworkBenchmarkAddrs(%v) failed: %v", config.NbioStd, err)
	}

	engine := nbio.NewEngine(nbio.Config{
		Network:            "tcp",
		Addrs:              addrs,
		MaxWriteBufferSize: 6 * 1024 * 1024,
		Listen:             frameworks.Listen,
	})

	engine.OnOpen(func(c *nbio.Conn) {
	})

	engine.OnData(func(c *nbio.Conn, data []byte) {
		c.Write(append([]byte{}, data...))
	})

	err = engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer engine.Stop()

	<-make(chan int)
}
