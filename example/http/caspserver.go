package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	websocket "github.com/gorilla/websocket"
	"github.com/uxff/webservice-client-as-service-provider/casp"
)

type CaspRequest struct {
	req *casp.HttpMsg
	res *casp.HttpMsg
}

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

var requestMap = make(map[string]*casp.HttpMsg, 0)

func AddReq(req *casp.HttpMsg) {
	mutex := &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()

	requestMap[req.MsgId] = req
}

func DelReq(req *casp.HttpMsg) {
	mutex := &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()

	delete(requestMap, req.MsgId)

}

func ResponseDispatch() {
	for {
		select {
		case res := <-responseChan:
			if req, ok := requestMap[res.MsgId]; ok {
				req.ResChan <- res
				DelReq(req)
			} else {
				log.Printf("nobody own this res:%s", res.MsgId)
			}
		}
	}
}

func main() {

	log.SetFlags(log.LstdFlags)

	cs = &casp.CaspServer{
		OnMessage: func(Ws *websocket.Conn, msg []byte, mtype int) {
			log.Printf("this is main message:%v", string(msg))
			res, _ := casp.ConvertBytesToHttpMsg(msg)
			go func() {
				// unable to implement multi-request
				responseChan <- res
			}()
		},
		OnClose: func(Ws *websocket.Conn) {
			log.Printf("this is main close")
		},
		PingInterval: time.Second * time.Duration(pingInterval),
		TimeOut:      time.Second * 30,
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
	nodes := cs.GetNodes()

	nodesStr := make([]string, 0)

	for _, node := range nodes {
		nodesStr = append(nodesStr, node.ClientIp)
	}

	buf, _ := json.Marshal(nodesStr)

	w.Write(buf)
	//casp.GetList()
}

func ActionToClient(res http.ResponseWriter, req *http.Request) {
	// q= `{"Method":"GET","Uri":"http://www.baidu.com/hello"}`
	//q := req.URL.Query().Get("q")
	q, _ := ioutil.ReadAll(req.Body)
	sreq := &casp.HttpMsg{
		MsgId:   fmt.Sprintf("%d", time.Now().UnixNano()),
		MsgType: casp.MSG_TYPE_HTTP_REQ,
		MsgBody: casp.SimpleRequest{},
	}

	err := json.Unmarshal(q, &sreq.MsgBody)
	if err != nil {
		log.Printf("unmarshal %s error:%v", q, err)
		res.Write([]byte(fmt.Sprintf("unmarshal %s error:%v", q, err)))
		return
	}

	requestChan <- sreq

	sres := <-responseChan

	//err := cs.RequestFromClient(n, sreq)
	log.Printf("the res from action:%v", sres)
	res.Write(sres.ToBytes())
}
