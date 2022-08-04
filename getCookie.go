package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
	"github.com/tencentyun/scf-go-lib/events"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type DefineEvent struct {
	URL     string `json:"url"`
	Params  string `json:"params"`
	Content string `json:"content"`
}

type RespEvent struct {
	Cookie string `json:"cookie"`
}

func GetCookie(event events.APIGatewayRequest) (*events.APIGatewayResponse, error) {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 60 * time.Second,
	}

	respevent := &RespEvent{
		Cookie: "",
	}

	var wait_proxy_req DefineEvent
	decoder := json.NewDecoder(strings.NewReader(event.Body))
	if err := decoder.Decode(&wait_proxy_req); err != nil {
		fmt.Println("解码失败", err.Error())
	}

	rawreq, err := http.ReadRequest(bufio.NewReader(strings.NewReader(wait_proxy_req.Content)))

	if wait_proxy_req.Params != "" {
		//做处理
	}

	tempurl, _ := url.Parse(wait_proxy_req.URL)
	rawreq.URL = tempurl
	rawreq.RequestURI = ""

	resp, err := client.Do(rawreq)
	if err != nil {
		fmt.Println("请求发送发生异常")
	}
	var dump []byte
	if dump, err = httputil.DumpResponse(resp, true); err != nil {
		fmt.Println("响应解析异常")
	}
	respevent.Cookie = string(dump)

	ret, err := json.Marshal(respevent)
	return &events.APIGatewayResponse{
		IsBase64Encoded: false,
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "application/json"},
		Body:            string(ret),
	}, nil

}

func main() {
	cloudfunction.Start(GetCookie)
}
