package main

import (
	//"bytes"
	"flag"
	"fmt"
	//"io"
	"log"
	//"net"
	"net/http"
	"os"
	"path"
	"time"

	websocket2 "github.com/gorilla/websocket"
)

type Msg struct {
	MsgId   string      `json:"MsgId"`
	MsgType int         `json:"MsgType"`
	MsgBody interface{} `json:"MsgBody"`
}

type SimpleRequest struct {
	//RequestId int      `json:"RequestId"`
	Uri    string   `json:"Uri"`
	Method string   `json:"Method"`
	Header []string `json:"Header"`
	Body   string   `json:"Body"`
}

type SimpleResponse struct {
	//RequestId int      `json:"RequestId"`
	Header []string `json:"Header"`
	Body   string   `json:"Body"`
}

var dir string
var port int
var staticHandler http.Handler
var registerCenter = "ws://127.0.0.1:8081/ws"
var sid string = "19283911"

var wsorigin = "http://127.0.0.1:8081/"
var wsurl = "ws://127.0.0.1:8081/echo"

// 初始化参数
func init() {
	flag.IntVar(&port, "p", 8082, "本地服务器端口")
	flag.StringVar(&dir, "dir", "./", "dir of statis server")
	flag.StringVar(&sid, "sid", sid, "param sid")
	flag.StringVar(&registerCenter, "reg", registerCenter, "register center")
	flag.Parse()

	dir = path.Dir(dir)

	staticHandler = http.FileServer(http.Dir(dir))
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 请求 register
	wsConn, err := RegWithWebsocket3(fmt.Sprintf("%s?cid=%s", registerCenter, sid), func(msg []byte, mtype int, wsconn *websocket2.Conn) {
		log.Printf("on message:%v", string(msg))

		//parse body
		//res = get hhtpp
		//wsconn.Write(ret)

	})
	if err != nil {
		log.Printf("register error:%v", err)
		return
	}

	defer wsConn.Close()

	go func() {
		for {
			// ping message
			log.Printf("will send ping:%s", wsConn.LocalAddr().String())
			//wsConn.WriteMessage(websocket2.TextMessage, []byte("i am online:"+wsConn.LocalAddr().String()))
			err := wsConn.WriteMessage(websocket2.PingMessage, nil)
			if err != nil {
				log.Printf("err when send ping:%v", err)
				os.Exit(0)
			}
			time.Sleep(3 * time.Second)
		}
	}()

	select {}
}

func RegWithWebsocket3(registerCenter string, msgproc func([]byte, int, *websocket2.Conn)) (*websocket2.Conn, error) {
	ws, _, err := websocket2.DefaultDialer.Dial(registerCenter, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	//defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(1000 * time.Hour))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(1000 * time.Hour))
		log.Printf("this pong handler.\n")
		return nil
	})
	//ws.SetPingHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(1000 * time.Hour)); return nil })

	go func() {

		for {
			mtype, msg, err := ws.ReadMessage()
			if err != nil {
				log.Printf("read message error:%v", err)
				ws.Close()
				os.Exit(0)
				break
			}

			log.Printf("read msg=%v mtype=%v", string(msg), mtype)

			if mtype == websocket2.TextMessage {
				if msgproc != nil {
					msgproc(msg, mtype, ws)
				}
			}

		}

	}()

	return ws, err
}
