package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	websocket "github.com/gorilla/websocket"
	"github.com/uxff/webservice-client-as-service-provider/casp"
)

var port int
var pingInterval int

// 初始化参数
func init() {
	//dir = path.Dir(os.Args[0])
	flag.IntVar(&pingInterval, "i", pingInterval, "ping interval, num of seconds")
	flag.IntVar(&port, "p", 8081, "ws服务器端口")
	flag.Parse()

	//nodes = make(map[string]*ServiceNode, 0)
}

func main() {

	log.SetFlags(log.LstdFlags)

	cs := &casp.CaspServer{
		OnMessage: func(Ws *websocket.Conn, msg []byte, mtype int) {
			log.Printf("this is main message:%v", string(msg))
		},
		OnClose: func(Ws *websocket.Conn) {
			log.Printf("this is main close")
		},
		PingInterval: time.Second * time.Duration(pingInterval),
		TimeOut:      time.Second * 10,
	}

	cs.OnOpen = func(Ws *websocket.Conn, req *http.Request) {
		log.Printf("this is main open")
		Ws.SetPingHandler(func(str string) error {
			log.Printf("ServerPingHandler: from client %s", str)
			return nil
		})
		Ws.SetPongHandler(func(str string) error {
			log.Printf("ServerPongHandler: server -> client %s", str)
			return nil
		})
	}

	cs.InitOnce()

	// 注册后，复用该连接
	http.HandleFunc("/ws", cs.ServeWebsocket)
	//http.HandleFunc("/", ServerForHome)
	//http.HandleFunc("/s", ForwardingToClient)

	log.Printf("websocket server will start at :%v", port)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
