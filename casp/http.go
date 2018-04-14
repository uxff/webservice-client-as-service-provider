package casp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type SimpleRequest struct {
	//RequestId int      `json:"RequestId"`
	Uri    string   `json:"Uri"`
	Method string   `json:"Method"`
	Header []string `json:"Header"`
	Body   string   `json:"Body"`
	Err    error    `json:"Body"`
}

type SimpleResponse struct {
	//RequestId int      `json:"RequestId"`
	Header []string `json:"Header"`
	Body   string   `json:"Body"`
	Err    error    `json:"Err"`
}
type HttpMsg struct {
	MsgId   string        `json:"MsgId"`
	MsgType int           `json:"MsgType"`
	MsgBody SimpleRequest `json:"MsgBody"`
}

const (
	MSG_TYPE_HTTP_REQ = 1
	MSG_TYPE_HTTP_RES = 2
)

var lastErrors []error

func (h *HttpMsg) ToBytes() []byte {
	theBytes, err := json.Marshal(h)
	if err != nil {
		lastErrors = append(lastErrors, err)
		return theBytes
	}

	return theBytes
}

func ConvertBytesToHttpMsg(str []byte) (*HttpMsg, error) {
	msg := &HttpMsg{MsgBody: SimpleRequest{}}
	err := json.Unmarshal(str, msg)
	return msg, err
}

func ConvertHttpResToSimpleResponse(httpRes *http.Response) (*SimpleRequest, error) {
	res := &SimpleRequest{}
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		res.Err = err
		return res, err
	}

	res.Body = string(body)
	res.Header = make([]string, 0)

	for hk, hv := range httpRes.Header {
		res.Header = append(res.Header, fmt.Sprintf("%s: %s"), hk, strings.Join(hv, ","))
	}

	return res, err
}

func (req *SimpleRequest) Do(timeOut time.Duration) *SimpleRequest {
	res := &SimpleRequest{}

	httpReq, err := http.NewRequest(req.Method, req.Uri, strings.NewReader(req.Body))
	if err != nil {
		res.Err = err
		return res
	}

	for _, v := range req.Header {
		//headerValPos := strings.Index(v, ":")
		headerKeyVals := strings.SplitN(v, ":", 1)
		if len(headerKeyVals) > 1 {
			httpReq.Header.Add(headerKeyVals[0], headerKeyVals[1])
		}
	}

	cli := &http.Client{}
	httpRes, err := cli.Do(httpReq)
	if err != nil {
		res.Err = err
		return res
	}

	res, err = ConvertHttpResToSimpleResponse(httpRes)

	return res
}
