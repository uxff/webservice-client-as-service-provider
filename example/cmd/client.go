package main

import (
	"flag"
	"log"
	"net/http"
	"os/exec"
	"path"
	"strings"
	"time"

	websocket "github.com/gorilla/websocket"
	"github.com/uxff/webservice-client-as-service-provider/casp"
)

var dir string
var port int
var staticHandler http.Handler
var registerCenter = "ws://127.0.0.1:8081/ws"
var sid string = "19283911"
var pingInterval int = 3
var delay = 0

// 初始化参数
func init() {
	flag.IntVar(&port, "p", 8082, "本地服务器端口")
	flag.StringVar(&dir, "dir", "./", "dir of statis server")
	flag.StringVar(&sid, "sid", sid, "param serial id")
	flag.StringVar(&registerCenter, "reg", registerCenter, "register center")
	flag.IntVar(&pingInterval, "i", pingInterval, "ping interval, num of seconds")
	flag.IntVar(&delay, "delay", delay, "delay request")
	flag.Parse()

	dir = path.Dir(dir)

	staticHandler = http.FileServer(http.Dir(dir))
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	maxFailedTimes := 100
	failedTimes := 0

	cc := &casp.CaspClient{
		Url: registerCenter,
		OnOpen: func(ws *websocket.Conn) {
			log.Printf("in main open a client->server")
		},
		OnMessage: func(ws *websocket.Conn, msg []byte, mtype int) {
			log.Printf("in main get msg:%v", string(msg))
			msgReq, err := casp.ConvertBytesToHttpMsg(msg)
			if err == nil {
				log.Printf("convertToHttpMsg error:%v msg=%v", err, string(msg))
			}

			if msgReq.MsgType == casp.MSG_TYPE_HTTP_REQ {

				log.Printf("a got a task of http req")
				req := msgReq.MsgBody
				log.Printf("task is:%s %s", req.Method, req.Uri)

				if delay > 0 {

					time.Sleep(time.Second * time.Duration(delay))
				}

				args := strings.Split(msgReq.MsgBody.Body, " ")

				res, err := exec.Command(args[0], args[1:]...).Output()
				//res := req.Do(time.Second * 10)
				if err != nil {
				}
				log.Printf("we got task res:%v", res)

				msgRes := &casp.HttpMsg{
					MsgId:   msgReq.MsgId,
					MsgType: casp.MSG_TYPE_HTTP_RES,
					MsgBody: casp.SimpleRequest{Body: string(res)},
				}

				err = ws.WriteMessage(websocket.TextMessage, msgRes.ToBytes())
				if err != nil {
					log.Printf("return ret to casp server error:%v", err)
				}

			}

		},
		OnClose: func(ws *websocket.Conn) {
			log.Printf("in main, ws closed")
		},
		PingInterval: time.Second * time.Duration(pingInterval),
	}

	cc.OnClose = func(ws *websocket.Conn) {
		log.Printf("in main, ws closed, try reconnect")
		failedTimes++
		if failedTimes > maxFailedTimes {
			return
		}

		time.Sleep(3 * time.Second)
		if err := cc.Open(); err != nil {
			log.Printf("connect %s error:%v", registerCenter, err)
			//return
		}
	}

	for {

		if err := cc.Open(); err != nil {
			log.Printf("connect %s error:%v", registerCenter, err)
			goto RETRY
		}
		if err := cc.Serve(); err != nil {
			log.Printf("Serve error:%v", err)
			goto RETRY
		}

	RETRY:
		time.Sleep(time.Second * 5)

	}

	select {}

}
