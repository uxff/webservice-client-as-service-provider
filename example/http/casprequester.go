package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/uxff/webservice-client-as-service-provider/casp"
)

var caspServerAddr = "http://127.0.0.1:12000"
var caspServerName = "all"

func main() {
	flag.StringVar(&caspServerAddr, "server", caspServerAddr, "casp server addr")
	flag.StringVar(&caspServerName, "name", caspServerName, "casp server name")

	allNodes := "" //casp.GetNodes()

	if len(allNodes) == 0 {
		fmt.Printf("the no nodes can serve you\n")
		//return
	}

	if len(allNodes) > 0 {
		//casp.Request(allNodes[0], &casp.SimpleRequest{Method: "GET", Uri: "http://www.soso.com/", Header: []string{"None: none"}})
	}

	req := casp.SimpleRequest{Method: "GET", Uri: "http://www.soso.com/", Header: []string{"None: none"}}
	reqHttp, err := http.NewRequest(req.Method, req.Uri, nil)
	if err != nil {
		//
	}
	cli := &http.Client{}
	resHttp, err := cli.Do(reqHttp)
	if err != nil {

	}

	return
}
