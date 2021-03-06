package main

import (
	"flag"
	"log"
	"net/http"
	"path"
	"time"

	websocket "github.com/gorilla/websocket"
	"github.com/uxff/webservice-client-as-service-provider/casp"
)

var dir string
var port int
var staticHandler http.Handler
var registerCenter = "ws://127.0.0.1:8081/ws"
var sid string = "19283911"
var pingInterval int = 3

var wsorigin = "http://127.0.0.1:8081/"
var wsurl = "ws://127.0.0.1:8081/echo"

// 初始化参数
func init() {
	flag.IntVar(&port, "p", 8082, "本地服务器端口")
	flag.StringVar(&dir, "dir", "./", "dir of statis server")
	flag.StringVar(&sid, "sid", sid, "param sid")
	flag.StringVar(&registerCenter, "reg", registerCenter, "register center")
	flag.IntVar(&pingInterval, "i", pingInterval, "ping interval, num of seconds")
	flag.Parse()

	dir = path.Dir(dir)

	staticHandler = http.FileServer(http.Dir(dir))
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	maxFailedTimes := 100
	failedTimes := 0

	cc := &casp.CaspClient{
		Url: registerCenter,
		OnOpen: func(ws *websocket.Conn) {
			log.Printf("in main open a client->server")
			ws.SetPingHandler(func(str string) error {
				log.Printf("ClientPingHandler: %v", str)
				return nil
			})
			ws.SetPongHandler(func(str string) error {
				log.Printf("ClientPongHandler: %v", str)
				return nil
			})
		},
		OnMessage: func(ws *websocket.Conn, msg []byte, mtype int) {
			log.Printf("in main get msg:%v", string(msg))

		},
		OnClose: func(ws *websocket.Conn) {
			log.Printf("in main, ws closed")
		},
		PingInterval: time.Second * time.Duration(pingInterval),
	}

	cc.OnClose = func(ws *websocket.Conn) {
		log.Printf("in main, ws closed, try reconnect")
		failedTimes++
		if failedTimes > maxFailedTimes {
			return
		}
		if err := cc.Open(); err != nil {
			log.Fatal("connect %s error:%v", registerCenter, err)
		}
	}

	if err := cc.Open(); err != nil {
		log.Fatal("connect %s error:%v", registerCenter, err)
	}

	select {}

}
