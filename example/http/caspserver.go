package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	websocket "github.com/gorilla/websocket"
	"github.com/uxff/webservice-client-as-service-provider/casp"
)

var port int
var pingInterval int
var cs *casp.CaspServer
var requestChan chan *casp.HttpMsg
var responseChan chan *casp.HttpMsg

// 初始化参数
func init() {
	//dir = path.Dir(os.Args[0])
	flag.IntVar(&pingInterval, "i", pingInterval, "ping interval, num of seconds")
	flag.IntVar(&port, "p", 8081, "ws服务器端口")

	flag.Parse()

	requestChan = make(chan *casp.HttpMsg, 10)
	responseChan = make(chan *casp.HttpMsg, 10)

	//nodes = make(map[string]*ServiceNode, 0)
}

func main() {

	log.SetFlags(log.LstdFlags)

	cs = &casp.CaspServer{
		OnMessage: func(Ws *websocket.Conn, msg []byte, mtype int) {
			log.Printf("this is main message:%v", string(msg))
			res, err := casp.ConvertBytesToHttpMsg(msg)
			if err != nil {
				go func() {
					requestChan <- res
				}()
			}
			// convert response
		},
		OnClose: func(Ws *websocket.Conn) {
			log.Printf("this is main close")
		},
		PingInterval: time.Second * time.Duration(pingInterval),
		TimeOut:      time.Second * 10,
	}

	cs.OnOpen = func(Ws *websocket.Conn, r *http.Request) {
		log.Printf("this is main open")
		Ws.SetPingHandler(func(str string) error {
			log.Printf("ServerPingHandler: from client %s", str)
			return nil
		})
		Ws.SetPongHandler(func(str string) error {
			log.Printf("ServerPongHandler: server -> client %s", str)
			return nil
		})

		go func(Conn *websocket.Conn) {
			for {
				select {
				case req := <-requestChan:
					Conn.WriteMessage(websocket.TextMessage, req.ToBytes())
				}
			}
		}(Ws)

		return

		// send a task of httpRequest to casp client
		req := casp.HttpMsg{
			MsgBody: casp.SimpleRequest{
				Method: "GET",
				Uri:    "http://www.baidu.com",
			},
			MsgId:   fmt.Sprintf("%d", time.Now().Unix()),
			MsgType: casp.MSG_TYPE_HTTP_REQ,
		}

		err := Ws.WriteMessage(websocket.TextMessage, req.ToBytes())
		if err != nil {
			log.Printf("write message error:%v", err)
		}
		log.Printf("already send httpMsg to client:%v", r.RemoteAddr)
	}

	cs.InitOnce()

	// 注册后，复用该连接
	http.HandleFunc("/ws", cs.ServeWebsocket)
	http.HandleFunc("/", ServerForHome)
	http.HandleFunc("/a", ActionToClient)

	log.Printf("websocket server will start at :%v", port)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func ServerForHome(w http.ResponseWriter, req *http.Request) {
	//casp.GetList()
}

func ActionToClient(res http.ResponseWriter, req *http.Request) {
	// q= `{"Method":"GET","Uri":"http://www.baidu.com/hello"}`
	q := req.URL.Query().Get("q")
	sreq := &casp.HttpMsg{}
	err := json.Unmarshal([]byte(q), &sreq.MsgBody)
	if err != nil {
		log.Printf("unmarshal %s error:%v", q, err)
		res.Write([]byte(fmt.Sprintf("unmarshal %s error:%v", q, err)))
		return
	}

	//n := req.URL.Query().Get("n")

	requestChan <- sreq

	sres := <-responseChan

	//err := cs.RequestFromClient(n, sreq)
	log.Printf("the res from action:%v", sres)
	res.Write(sres.ToBytes())
}
