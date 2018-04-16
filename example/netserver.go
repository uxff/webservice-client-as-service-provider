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

	//line := fmt.Sprintf("%s %s %s\r\n", r.Method, r.RequestURI, r.Proto)

	q := r.URL.Query()
	targetUrl := q.Get("url")

	log.Printf("recv a request from normal client: line=[%v], casp client (%v) will serve", r.RequestURI, n.ServiceName)

	n.RequestTimes++

	var err error

	_, err = n.Conn.Write([]byte(targetUrl))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Write line to casp conn error:%v", err)))
	}

	//err = r.Header.Write(n.Conn)

	if err != nil {
		w.Write([]byte(fmt.Sprintf("Write head to casp conn error:%v", err)))
	}

	/*
		_, err = io.Copy((n.Conn).(io.ReadWriteCloser), r.Body)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("Write copy error:%v", err)))
		}
	*/

	//_, err = io.Copy(w, n.Conn)
	go func() {
		for {
			buf := make([]byte, 4096)
			wn, err := io.CopyBuffer(w, n.Conn, buf)
			log.Printf("recv(%d) from casp server and transfer to local server: %v, writen=%d err=%v", len(buf), string(buf[:10]), wn, err)
		}
		return

		errTimes := 0
		for {

			if errTimes > 10 {
				break
			}

			buf := make([]byte, 4096)
			//io.Copy(localServer, caspServer)
			n, err := n.Conn.Read(buf)

			if err != nil {
				log.Printf("read from casp client error: n=%d, err=%v", n, err)
				errTimes++
				continue
			}

			log.Printf("read success from casp client: n=%v len=%v buf=%v", n, len(buf), string(buf[:10]))

			n, err = w.Write(buf)
			if err != nil {
				log.Printf("write to local client error: n=%d, err=%v", n, err)
				errTimes++
				continue
			}
			log.Printf("write success to local client: n=%v len=%v buf=%v", n, len(buf), string(buf[:10]))

		}
	}()

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
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	go ServeRegister()

	// 注册后，复用该连接
	http.HandleFunc("/_home", ServerForHome)
	//http.DefaultTransport.RoundTrip()
	http.Handle("/", &ForwardingHandler{})
	http.Handle("/q", &ForwardingHandler{})

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

			// 首次，读取casp client请求的注册名 应该是servicename=xxxx
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
				log.Printf("get service name from client failed, b=%v", string(b[:100]))
				return
			}

			node := &ServiceNode{
				ServiceName:  serviceName,
				ClientIp:     client.RemoteAddr().String(),
				Created:      time.Now(),
				RequestTimes: 0,
				Conn:         client,
			}

			log.Printf("a casp service registered 2: %v %v curl http://127.0.0.1:%v/", client.RemoteAddr().String(), serviceName, port)

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
