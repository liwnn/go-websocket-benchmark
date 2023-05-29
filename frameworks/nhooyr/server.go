package main

import (
	"context"
	"flag"
	"fmt"
	"go-websocket-benchmark/conf"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"nhooyr.io/websocket"
)

var (
	_ = flag.Int("b", 1024, `read buffer size`)
	_ = flag.Int("nb", 10000, `max blocking online num, e.g. 10000`)
)

func main() {
	flag.Parse()

	ports := strings.Split(conf.Ports[conf.Nhooyr], ":")
	minPort, err := strconv.Atoi(ports[0])
	if err != nil {
		log.Fatalf("invalid port range: %v, %v", ports, err)
	}
	maxPort, err := strconv.Atoi(ports[1])
	if err != nil {
		log.Fatalf("invalid port range: %v, %v", ports, err)
	}
	addrs := []string{}
	for i := minPort; i <= maxPort; i++ {
		addrs = append(addrs, fmt.Sprintf(":%d", i))
	}
	startServers(addrs)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}

func startServers(addrs []string) {
	for _, v := range addrs {
		go func(addr string) {
			mux := &http.ServeMux{}
			mux.HandleFunc("/ws", onWebsocket)
			server := http.Server{
				Addr:    addr,
				Handler: mux,
			}
			log.Fatalf("server exit: %v", server.ListenAndServe())
		}(v)
	}
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	for {
		mt, data, err := c.Read(context.Background())
		if err != nil {
			log.Printf("read failed: %v", err)
			break
		}
		c.Write(context.Background(), mt, data)
	}
}
