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

	registerMsg := "servicename=netclient"
	conn.Write([]byte(registerMsg))

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func(conn net.Conn) {

		conn.SetReadDeadline(time.Now().Add(time.Hour * 1000))

		//handleClientRequest(conn)
		handleBasicQueryRequest(conn)

		wg.Done()
	}(conn)

	wg.Wait()

	return nil
}

func handleClientRequest(caspServer net.Conn) {
	if caspServer == nil {
		return
	}
	//defer caspServer.Close()
	log.Printf("waiting for request:%v", caspServer.RemoteAddr().String())

	/*
		var b [1024]byte
		n, err := caspServer.Read(b[:])
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
	localServer, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}

	/*
		if method == "CONNECT" {
			fmt.Fprint(caspServer, "HTTP/1.1 200 Connection established\r\n")
		} else {
			localServer.Write(b[:n])
		}
	*/

	wg := &sync.WaitGroup{}
	//进行转发
	wg.Add(1)
	go func() {

		errTimes := 0
		for {

			if errTimes > 10 {
				break
			}

			//io.Copy(localServer, caspServer)
			buf := make([]byte, 4096)
			n, err := caspServer.Read(buf)

			if err != nil {
				log.Printf("read from casp server error: n=%d, err=%v", n, err)
				errTimes++
				continue
			}

			log.Printf("read success from casp server: n=%v len=%v buf=%v", n, len(buf), string(buf[:10]))

			n, err = localServer.Write(buf)
			if err != nil {
				log.Printf("write to local server error: n=%d, err=%v", n, err)
				errTimes++
				continue
			}
			log.Printf("write success to local server: n=%v len=%v buf=%v", n, len(buf), string(buf[:10]))

			//io.CopyBuffer(localServer, caspServer, buf)
			//log.Printf("recv(%d) from casp server and transfer to local server: %v", len(buf), string(buf[:10]))
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		errTimes := 0
		for {

			if errTimes > 10 {
				break
			}

			//io.Copy(localServer, caspServer)
			buf := make([]byte, 4096)
			n, err := localServer.Read(buf)

			if err != nil {
				log.Printf("read from local server error: n=%d, err=%v", n, err)
				errTimes++
				continue
			}

			log.Printf("read success from local server: n=%v len=%v buf=%v", n, len(buf), string(buf[:10]))

			n, err = caspServer.Write(buf)
			if err != nil {
				log.Printf("write to casp server error: n=%d, err=%v", n, err)
				errTimes++
				continue
			}
			log.Printf("write success to casp server: n=%v len=%v buf=%v", n, len(buf), string(buf[:10]))

			//io.CopyBuffer(localServer, caspServer, buf)
			//log.Printf("recv(%d) from casp server and transfer to local server: %v", len(buf), string(buf[:10]))
		}
		wg.Done()
	}()

	wg.Wait()
	if false {
		io.Copy(caspServer, localServer)
		io.Copy(localServer, caspServer)
	}
}

func handleBasicQueryRequest(caspServer net.Conn) {
	errTimes := 0
	for {

		if errTimes > 10 {
			break
		}

		//io.Copy(localServer, caspServer)
		buf := make([]byte, 4096)
		n, err := caspServer.Read(buf)

		if err != nil {
			log.Printf("read from casp server error: n=%d, err=%v", n, err)
			errTimes++
			continue
		}

		log.Printf("read success from casp server: n=%v len=%v buf=%v", n, len(buf), string(buf[:10]))

		r, err := http.Get(string(buf))
		if err != nil {
			caspServer.Write([]byte(fmt.Sprintf("error when get [%v]:%v", string(buf), err)))
			//errTimes++
			continue
		}

		io.Copy(caspServer, r.Body)

		log.Printf("write success to local server: n=%v len=%v buf=%v", n, len(buf), string(buf[:10]))

		//io.CopyBuffer(localServer, caspServer, buf)
		//log.Printf("recv(%d) from casp server and transfer to local server: %v", len(buf), string(buf[:10]))
	}
}

// 静态文件处理
func StaticServer(w http.ResponseWriter, req *http.Request) {
	staticHandler.ServeHTTP(w, req)
	//io.WriteString(w, "hello, world! i am a static server\n")
}
