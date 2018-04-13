package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	websocket "github.com/gorilla/websocket"
	"github.com/uxff/webservice-client-as-service-provider/casp"
)

var port int

// 初始化参数
func init() {
	//dir = path.Dir(os.Args[0])
	flag.IntVar(&port, "p", 8081, "ws服务器端口")
	flag.Parse()

	//nodes = make(map[string]*ServiceNode, 0)
}

func main() {

	log.SetFlags(log.LstdFlags)

	cs := casp.CaspServer{
		OnOpen: func(Ws *websocket.Conn, req *http.Request) {
			log.Printf("this is main open")
			Ws.SetPingHandler(func(str string) error {
				log.Printf("this is ping, from client ")
				return nil
			})
			Ws.SetPongHandler(func(str string) error {
				log.Printf("this is pong, server -> client")
				return nil
			})
		},
		OnMessage: func(Ws *websocket.Conn, msg []byte, mtype int) {
			log.Printf("this is main message:%v", string(msg))
		},
		OnClose: func(Ws *websocket.Conn) {
			log.Printf("this is main close")
		},
	}

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
