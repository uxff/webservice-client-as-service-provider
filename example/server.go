package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	//"sync"
	"time"

	websocket2 "github.com/gorilla/websocket"
)

type ServiceNode struct {
	ClientIp     string
	Cid          string
	Created      time.Time
	RequestTimes int
	Conn         *websocket2.Conn //net.Conn
}

var nodes map[string]*ServiceNode

var port int
var upgrader = websocket2.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// 初始化参数
func init() {
	//dir = path.Dir(os.Args[0])
	flag.IntVar(&port, "p", 8081, "ws服务器端口")
	flag.Parse()

	nodes = make(map[string]*ServiceNode, 0)
}

func main() {
	log.SetFlags(log.LstdFlags)

	// 注册后，复用该连接
	http.HandleFunc("/ws", ServeRegister2)
	http.HandleFunc("/", ServerForHome)
	http.HandleFunc("/s", ForwardingToClient)

	log.Printf("websocket server will start at :%v", port)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func OnMessageHandler(msg []byte, msgtype int) {

}

func ServeRegister2(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("serve register2 error:%v", err)
	}

	q := req.URL.Query()
	cid := q.Get("cid")

	node := &ServiceNode{
		Cid:          cid,
		ClientIp:     req.RemoteAddr,
		Created:      time.Now(),
		RequestTimes: 0,
		Conn:         conn,
	}

	log.Printf("a client registered 2: %v %v", req.RemoteAddr, cid)
	/*
		conn.SetPingHandler(func(msg string) error {
			conn.SetReadDeadline(time.Now().Add(1000 * time.Hour))
			log.Printf("this ping handler.\n")
			return nil
		})
	*/

	//conn.SetCloseHandler()

	go func(node *ServiceNode) {

		for {

			// mtype 只可能等于 Text Binary 不可能等于 PingMessage PongMessage
			mtype, msg, err := node.Conn.ReadMessage()
			if err != nil {
				log.Printf("READ ERROR from %v(%s):%v", node.Cid, node.ClientIp, err)
				node.Conn.Close()
				delete(nodes, node.ClientIp)
				return
			}

			log.Printf("get a message from %v:type=%d msg=%v ", node.ClientIp, mtype, string(msg))
		}

	}(node)

	nodes[req.RemoteAddr] = node
}

// 主页请求
func ServerForHome(w http.ResponseWriter, req *http.Request) {

	html := "<!Doctype html><html><ul>\n"

	if len(nodes) <= 0 {
		html += `<li class="empty">no services from client</li><br>` + "\n"
	}
	for _, n := range nodes {
		html += fmt.Sprintf(`<li> ip=%s | Cid=%s | created=%s | req times=%v</li><br/>`+"\n", n.ClientIp, n.Cid, n.Created.Format("2017-01-02 13:34:05"), n.RequestTimes)
	}

	html += "</ul></html>\n"

	io.WriteString(w, html)
}

// 转发
func ForwardingToClient(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	cid := q.Get("cid")
	path := q.Get("path")

	if len(path) == 0 {
		path = "/"
	}

	node, ok := nodes[cid]
	if !ok {
		log.Printf("client name not exist:%v", cid)
		return
	}

	// 必须使用WriteMessage()回复，不能使用io.Copy()
	err := node.Conn.WriteMessage(websocket2.TextMessage, []byte(path))
	if err != nil {
		log.Printf("write to client %v(%s) ERROR:%v", node.Cid, node.ClientIp, err)
		delete(nodes, node.ClientIp)
		node.Conn.Close()
		return
	}

	/*
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		if err != nil {
			log.Printf("read from client after send req to client error:%v", err)
			w.Write([]byte(err.Error()))
		} else {
			w.Write(msg)
		}
	*/

	log.Printf("finish forwarding: cid=%v path=%v", cid, path)
}
