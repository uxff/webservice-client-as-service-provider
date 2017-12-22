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
	"golang.org/x/net/websocket"
)

type ServiceNode struct {
	ClientIp     string
	ServiceName  string
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
	flag.IntVar(&port, "p", 8081, "服务器端口")
	flag.Parse()

	nodes = make(map[string]*ServiceNode, 0)
}

func main() {
	log.SetFlags(log.LstdFlags)

	// 注册后，复用该连接
	http.HandleFunc("/casp", ServeRegister2)
	http.HandleFunc("/", ServerForHome)
	http.HandleFunc("/s", ForwardingToClient)
	http.Handle("/echo", websocket.Handler(echoHandler))

	log.Printf("server will start at :%v", port)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func echoHandler(ws *websocket.Conn) {
	msg := make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Receive: %s\n", msg[:n])

	for i := 0; i < 2; i++ {

		send_msg := "[" + string(msg[:n]) + "]"
		m, err := ws.Write([]byte(send_msg))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Send: i=%v %s\n", i, msg[:m])
	}
}

// 注册
func ServerForRegister(ws *websocket.Conn) {

	req := ws.Request()
	q := req.URL.Query()
	serviceName := q.Get("servicename")

	node := &ServiceNode{
		ServiceName:  serviceName,
		ClientIp:     req.RemoteAddr,
		Created:      time.Now(),
		RequestTimes: 0,
		//Conn:         ws,
	}

	log.Printf("a service registered: %v %v", req.RemoteAddr, serviceName)

	nodes[serviceName] = node

}

func ServeRegister2(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("serve register2 error:%v", err)
	}

	q := req.URL.Query()
	serviceName := q.Get("servicename")

	node := &ServiceNode{
		ServiceName:  serviceName,
		ClientIp:     req.RemoteAddr,
		Created:      time.Now(),
		RequestTimes: 0,
		Conn:         conn,
	}

	log.Printf("a service registered 2: %v %v", req.RemoteAddr, serviceName)

	nodes[serviceName] = node
}

// 主页请求
func ServerForHome(w http.ResponseWriter, req *http.Request) {

	html := "<!Doctype html><html><ul>"

	if len(nodes) <= 0 {
		html += `<li class="empty">no services from client</li>`
	}
	for _, n := range nodes {
		nodeUrl := fmt.Sprintf("http://127.0.0.1:%v/s?servicename=%v&path=/", port, n.ServiceName)
		html += fmt.Sprintf(`<li><a href="%v">%s</a> | %s | %s | %v</li>`, nodeUrl, n.ClientIp, n.ServiceName, n.Created.Format("2017-01-02 13:34:05"), n.RequestTimes)
	}

	html += "</ul></html>"

	io.WriteString(w, html)
}

// 转发
func ForwardingToClient(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	serviceName := q.Get("servicename")
	path := q.Get("path")

	if len(path) == 0 {
		path = "/"
	}

	node, ok := nodes[serviceName]
	if !ok {
		log.Printf("service name not exist:%v", serviceName)
	}

	// 必须使用WriteMessage()回复，不能使用io.Copy()
	node.Conn.WriteMessage(websocket2.TextMessage, []byte(path))

	mtype, msg, err := node.Conn.ReadMessage()

	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	if err != nil {
		log.Printf("read from client after send req to client error:%v", err)
		w.Write([]byte(err.Error()))
	} else {
		w.Write(msg)
	}

	log.Printf("finish forwarding: servicename=%v path=%v mtype=%d", serviceName, path, mtype)
}
