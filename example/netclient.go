package main

import (
	//"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	//"net/url"
	"path"
	"strconv"
	//"strings"
	"sync"
	"time"
)

var dir string
var reg string
var port int
var staticHandler http.Handler

// 初始化参数
func init() {
	flag.IntVar(&port, "p", 8082, "静态文件服务器端口")
	flag.StringVar(&dir, "dir", "./", "dir of statis server")
	flag.StringVar(&reg, "reg", "127.0.0.1:20009", "tcp register server")
	flag.Parse()

	dir = path.Dir(dir)

	staticHandler = http.FileServer(http.Dir(dir))
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// 请求 register

	//go RegWithWebsocket()
	go Reg(reg)

	// 开启本地http服务器
	http.HandleFunc("/", StaticServer)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func Reg(registerCenter string) error {
	conn, err := net.Dial("tcp", registerCenter)

	if err != nil {
		log.Printf("dail reg(%v) error:%v", registerCenter, err)
		return err
	}

	registerMsg := "servicename=netclient\n"
	conn.Write([]byte(registerMsg))

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func(conn net.Conn) {

		conn.SetReadDeadline(time.Now().Add(time.Hour * 1000))

		handleClientRequest(conn)

		wg.Done()
	}(conn)

	wg.Wait()

	return nil
}

func handleClientRequest(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()
	log.Printf("waiting for request:%v", client.RemoteAddr().String())

	/*
		var b [1024]byte
		n, err := client.Read(b[:])
		if err != nil {
			log.Println(err)
			return
		}

		var method, host, address string
		fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
		hostPortURL, err := url.Parse(host)
		if err != nil {
			log.Println(err)
			return
		}
	*/
	address := fmt.Sprintf("127.0.0.1:%d", port)

	// 向本地http服务拨号
	server, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}

	/*
		if method == "CONNECT" {
			fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n")
		} else {
			server.Write(b[:n])
		}
	*/

	wg := &sync.WaitGroup{}
	//进行转发
	wg.Add(1)
	go func() {
		//io.Copy(server, client)
		buf := make([]byte, 4096)
		io.CopyBuffer(server, client, buf)
		log.Printf("recv(%d) from casp server and transfer to local server: %v", len(buf), string(buf[:10]))
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		//io.Copy(client, server)
		buf := make([]byte, 4096)
		io.CopyBuffer(client, server, buf)
		log.Printf("send(%d) local server res to casp server: %v", len(buf), string(buf[:10]))
		wg.Done()
	}()

	wg.Wait()
}

// 静态文件处理
func StaticServer(w http.ResponseWriter, req *http.Request) {
	staticHandler.ServeHTTP(w, req)
	//io.WriteString(w, "hello, world! i am a static server\n")
}
