package casp

import (
	"log"
	"net/http"
	"sync"
	"time"

	websocket "github.com/gorilla/websocket"
)

type WsClientNode struct {
	ClientIp     string
	Created      time.Time
	RequestTimes int
	Conn         *websocket.Conn //net.Conn
	httpReq      *http.Request
}

type CaspServer struct {
	nodes map[string]*WsClientNode
	//Ws *websocket.Conn
	PingInterval time.Duration //sec
	OnOpen       func(Ws *websocket.Conn, req *http.Request)
	OnMessage    func(Ws *websocket.Conn, msg []byte, mtype int)
	OnClose      func(Ws *websocket.Conn)
	//ServeServer func(w http.ResponseWriter, req *http.Request)
}

func (cs *CaspServer) ServeWebsocket(w http.ResponseWriter, req *http.Request) {
	if cs.nodes == nil {
		cs.nodes = make(map[string]*WsClientNode, 0)
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeWebsocket error:%v", err)
	}

	//q := req.URL.Query()

	node := &WsClientNode{
		ClientIp:     req.RemoteAddr,
		Created:      time.Now(),
		RequestTimes: 0,
		Conn:         conn,
		httpReq:      req,
	}

	log.Printf("a client registered: %v ", req.RemoteAddr)

	if cs.OnOpen != nil {
		cs.OnOpen(conn, req)
	}

	if cs.PingInterval > 0 {
		go func(conn *websocket.Conn) {
			for {
				time.Sleep(cs.PingInterval)
				log.Printf("client write a ping.")
				conn.WriteMessage(websocket.PingMessage, nil)
			}
		}(conn)
	}

	/*
		conn.SetPingHandler(func(msg string) error {
			conn.SetReadDeadline(time.Now().Add(1000 * time.Hour))
			log.Printf("this ping handler.\n")
			return nil
		})
	*/

	//conn.SetCloseHandler()

	go func(node *WsClientNode) {

		for {

			// mtype 只可能等于 Text Binary 不可能等于 PingMessage PongMessage
			mtype, msg, err := node.Conn.ReadMessage()
			if err != nil {
				log.Printf("READ from %s ERROR:%v", node.ClientIp, err)
				cs.closeClient(node.ClientIp)
				return
			}

			log.Printf("get a message from %v:type=%d msg=%v ", node.ClientIp, mtype, string(msg))
			if cs.OnMessage != nil {
				cs.OnMessage(node.Conn, msg, mtype)
			}
		}

	}(node)

	cs.nodes[req.RemoteAddr] = node

}

func (cs *CaspServer) closeClient(clientIp string) {
	mutex := &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := cs.nodes[clientIp]; ok {
		if cs.OnClose != nil {
			cs.OnClose(cs.nodes[clientIp].Conn)
		}
		cs.nodes[clientIp].Conn.Close()
		delete(cs.nodes, clientIp)
	}
}

type CaspClient struct {
	Conn         *websocket.Conn
	Url          string
	PingInterval time.Duration //sec
	OnOpen       func(Ws *websocket.Conn)
	OnMessage    func(Ws *websocket.Conn, msg []byte, mtype int)
	OnClose      func(Ws *websocket.Conn)
}

func (cc *CaspClient) Open() error {
	conn, _, err := websocket.DefaultDialer.Dial(cc.Url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	cc.Conn = conn

	//defer ws.Close()
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(1000 * time.Hour))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(1000 * time.Hour))
		log.Printf("this pong handler.\n")
		return nil
	})
	//conn.SetPingHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(1000 * time.Hour)); return nil })

	if cc.OnOpen != nil {
		cc.OnOpen(conn)
	}

	if cc.PingInterval > 0 {
		go func(conn *websocket.Conn) {
			for {
				time.Sleep(cc.PingInterval)
				log.Printf("client write a ping.")
				conn.WriteMessage(websocket.PingMessage, nil)
			}
		}(conn)
	}

	go func(conn *websocket.Conn) {

		for {
			mtype, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("read message error:%v", err)
				cc.Close()
				break
			}

			log.Printf("read msg=%v mtype=%v", string(msg), mtype)

			if cc.OnMessage != nil {
				cc.OnMessage(conn, msg, mtype)
			}

		}

	}(conn)

	return nil
}

func (cc *CaspClient) Close() {
	if cc.OnClose != nil {
		cc.OnClose(cc.Conn)
	}
	cc.Conn.Close()
}

func init() {

}
