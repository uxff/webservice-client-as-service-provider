package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	//"net"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	websocket2 "github.com/gorilla/websocket"
	"golang.org/x/net/websocket"
)

var dir string
var port int
var staticHandler http.Handler
var registerCenter = "ws://127.0.0.1:8081/casp?servicename=localdir"

var wsorigin = "http://127.0.0.1:8081/"
var wsurl = "ws://127.0.0.1:8081/echo"

// 初始化参数
func init() {
	flag.IntVar(&port, "p", 8082, "服务器端口")
	flag.StringVar(&dir, "dir", "./", "dir of statis server")
	flag.Parse()

	dir = path.Dir(dir)

	staticHandler = http.FileServer(http.Dir(dir))
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// 请求 register
	TestWebsocket()
	//return

	//go RegWithWebsocket()
	go RegWithWebsocket3()

	// 开启本地http服务器
	http.HandleFunc("/", StaticServer)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	// 通过websocket连接到注册服务器
	// 接受注册服务器的任务，转发到本机的服务器
	//localServiceUrl := fmt.Sprintf("http://127.0.0.1:%v", port)
	//http.Get(localServiceUrl)
}

// 静态文件处理
func StaticServer(w http.ResponseWriter, req *http.Request) {
	staticHandler.ServeHTTP(w, req)
	//io.WriteString(w, "hello, world! i am a static server\n")
}

func TestWebsocket() {
	ws, err := websocket.Dial(wsurl, "", wsorigin)
	if err != nil {
		log.Fatal(err)
	}
	message := []byte("hello, world!你好")
	_, err = ws.Write(message)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Send: %s\n", message)

	var msg = make([]byte, 512)
	m, err := ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Receive: %s\n", msg[:m])

	ws.Close() //关闭连接
}

func RegWithWebsocket() {
	// do register
	ws, err := websocket.Dial(registerCenter, "", wsorigin)
	//ts, err :=
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("register to %v SUCCESS expire:%v", registerCenter, time.Now().Add(time.Hour*1000))

	ws.SetReadDeadline(time.Now().Add(time.Hour * 1000))

	for {

		var msg = make([]byte, 4096)

		m, err := ws.Read(msg)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Receive: %s\n", msg[:m])

		message := []byte("unkown response")

		res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", port))
		if err != nil {
			message = []byte(fmt.Sprintf("HTTP/1.1 404 Not Found\r\n%v", err))
		} else {
			buf := &bytes.Buffer{} //bytes.NewBuffer(message)
			io.Copy(buf, res.Body)
			message = buf.Bytes()
		}

		_, err = ws.Write(message)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Send: %s\n", message)
	}

}

func RegWithWebsocket2() {
	c, _, err := websocket2.DefaultDialer.Dial(registerCenter, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer c.Close()
		defer close(done)
		defer wg.Done()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	wg.Wait()
}

func RegWithWebsocket3() {
	ws, _, err := websocket2.DefaultDialer.Dial(registerCenter, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(1000 * time.Hour))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(1000 * time.Hour)); return nil })
	//ws.SetPingHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(1000 * time.Hour)); return nil })

	for {
		mtype, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("read message error:%v", err)
			break
		}

		log.Printf("read msg=%v mtype=%v", string(msg), mtype)

		res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d%v", port, string(msg)))
		if err != nil {
			msg = []byte(fmt.Sprintf("HTTP/1.1 404 Not Found\r\n%v", err))
		} else {
			buf := &bytes.Buffer{} //bytes.NewBuffer(message)
			io.Copy(buf, res.Body)
			msg = buf.Bytes()
		}

		err = ws.WriteMessage(websocket2.TextMessage, msg) //ws.Write(message)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Send(len=%d): %s\n", len(msg), string(msg[:10]))

	}
}
