package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type stockUptReq struct {
	Stocks []string `json:"stocks"`
}

var stocks = map[string]float32{
	"AAPL": 95.0,
	"MSFT": 50.0,
	"AMZN": 300.0,
	"GOOG": 550.0,
	"YHOO": 35.0,
}

var addr = flag.String("addr", "localhost:8181", "http service address")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func sendStockUpdates(wc *websocket.Conn) {
	rand.Seed(42)
	for k := range stocks {
		if rand.Float32() > 0.5 {
			stocks[k] = stocks[k] + rand.Float32()
		}

		stocks[k] = stocks[k] - rand.Float32()
	}

	if err := wc.WriteJSON(stocks); err != nil {
		log.Println(err)
	}

}

func updateStock(w http.ResponseWriter, r *http.Request) {
	wc, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer wc.Close()

	go func() {
		for range time.Tick(time.Second) {
			sendStockUpdates(wc)
		}
	}()

	for {
		var req *stockUptReq
		if err := wc.ReadJSON(req); err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", req)
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", updateStock)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
