package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	//"sync"
	"net/url"
	"time"
)

type ServiceNode struct {
	ClientIp     string
	ServiceName  string
	Created      time.Time
	RequestTimes int
	Conn         net.Conn //net.Conn
}

var nodes map[string]*ServiceNode

var port int
var regPort int

//var forwardingHandler http.Handler

type ForwardingHandler struct{}

func (this *ForwardingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// w is response to general
	// r : from general client, to general
	// r will request to casp client
	// casp client response will copy to w
	if len(nodes) == 0 {
		w.Write([]byte("NO client can serve you\n"))
		return
	}

	//n := &nodes[0]
	var n *ServiceNode
	for i, _ := range nodes {
		n = nodes[i]
		break
	}

	line := fmt.Sprintf("%s %s %s\r\n", r.Method, r.RequestURI, r.Proto)
	log.Printf("i have recv a request:%v", line)

	var err error

	_, err = n.Conn.Write([]byte(line))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Write line to casp conn error:%v", err)))
	}

	err = r.Header.Write((n.Conn).(io.WriteCloser))

	if err != nil {
		w.Write([]byte(fmt.Sprintf("Write head to casp conn error:%v", err)))
	}

	/*
		_, err = io.Copy((n.Conn).(io.ReadWriteCloser), r.Body)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("Write copy error:%v", err)))
		}
	*/

	_, err = io.Copy(w, n.Conn)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Write copy error:%v", err)))
	}

}

// 初始化参数
func init() {
	//dir = path.Dir(os.Args[0])
	flag.IntVar(&port, "p", 8081, "服务器端口")
	flag.IntVar(&regPort, "r", 20009, "register port")
	flag.Parse()

	nodes = make(map[string]*ServiceNode, 0)
}

func main() {
	log.SetFlags(log.LstdFlags)

	go ServeRegister()

	// 注册后，复用该连接
	http.HandleFunc("/_home", ServerForHome)
	//http.DefaultTransport.RoundTrip()
	http.Handle("/", &ForwardingHandler{})

	log.Printf("server will start at :%v", port)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func ServeRegister() {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(regPort))
	if err != nil {
		log.Panic(err)
	}

	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}

		go func(client net.Conn) {
			//defer client.Close()

			var b = make([]byte, 4096)
			n, err := client.Read(b)
			if err != nil {
				log.Println(err)
				return
			}

			q, err := url.ParseQuery(string(b[:n]))
			if err != nil {
				log.Printf("url(%v) parse error:%v", string(b[:n]), err)
			}

			serviceName := q.Get("servicename")

			if len(serviceName) <= 0 {
				log.Printf("get service name from client failed")
				return
			}

			node := &ServiceNode{
				ServiceName:  serviceName,
				ClientIp:     client.RemoteAddr().String(),
				Created:      time.Now(),
				RequestTimes: 0,
				Conn:         client,
			}

			log.Printf("a service registered 2: %v %v", client.RemoteAddr().String(), serviceName)

			nodes[serviceName] = node

		}(client)
	}
}

// 主页请求
func ServerForHome(w http.ResponseWriter, req *http.Request) {

	html := "<!Doctype html><html><body><ul>"

	if len(nodes) <= 0 {
		html += `<li class="empty">no services from client</li>`
	}
	for _, n := range nodes {
		nodeUrl := fmt.Sprintf("http://127.0.0.1:%v/", port)
		html += fmt.Sprintf(`<li><a href="%v">%s</a> | %s | %s | %v</li>`, nodeUrl, n.ClientIp, n.ServiceName, n.Created.Format("2017-01-02 13:34:05"), n.RequestTimes)
	}

	html += "</ul></body></html>"

	io.WriteString(w, html)
}
