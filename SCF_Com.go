package main

import (
    "bufio"
    "bytes"
    "crypto/tls"
    "encoding/base64"
    "encoding/json"
    "events"
    "fmt"
    "github.com/tencentyun/scf-go-lib/cloudfunction"
    "io/ioutil"
    // "io"
    "net/http"
    "net/http/httputil"
    "net/url"
    "regexp"
    "strings"
    "time"
)

type DefineEvent struct {
    // test event define
    URL     string                       `json:"url"`     // 目标的 URL, eg: http://cip.cc/
    Content map[string]map[string]string `json:"content"` // 最原始的 HTTP 报文, base64,多个报文，减少函数调用次数 `pwd-[url-报文]`
}

type RespEvent struct {
    Bodys map[string]string `json:"bodys"` //密码-响应结果
}

//处理报文，发送请求
func ForwardReq(req *http.Request, client http.Client) ([]byte, error) {
    resp, err := client.Do(req) //Do 方法发送请求，返回 HTTP 回复
    if err != nil {
        fmt.Println("代理服务端：请求发送发生异常")
        return nil, err
    }
    var dump []byte
    if dump, err = httputil.DumpResponse(resp, true); err != nil {
        fmt.Println("代理服务端：响应解析异常")
        return nil, err
    }
    return dump, nil
}

func ScfProxy(event events.APIGatewayRequest) (*events.APIGatewayResponse, error) {
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
    /*
       异常机制：
           dump={
               "1":"服务端发生某种异常，该密码需要记录"
           }
    */
    var url_content DefineEvent
    var dump string = "1"
    var dumpResponse []byte
    var rawreq *http.Request
    var tempurl *url.URL
    var err error
    // var rawreqContent []byte
    var bodyParamsReader *strings.Reader
    var r *regexp.Regexp
    var bodyParamEncode string

    decoder := json.NewDecoder(strings.NewReader(event.Body))
    if err := decoder.Decode(&url_content); err != nil {
        fmt.Println("解码失败", err.Error())
    }

    fmt.Println("---URL---:", url_content.URL)
    fmt.Println("----------------开始处理转发请求----------------")

    respevent := &RespEvent{
        Bodys: make(map[string]string, 6), //每次调用装6个响应值
    }

    for pwd, url_req := range url_content.Content {
        //解析请求url以及请求本体
        var content []byte
        // var duration time.Duration
        for req_url, contentbase64 := range url_req {
            content, err = base64.StdEncoding.DecodeString(string(contentbase64))
            if err != nil {
                fmt.Println("解析请求发生异常,发生异常URL：", req_url)
                dump = base64.StdEncoding.EncodeToString([]byte("1"))
                goto pack
            }
            tempurl, _ = url.Parse(req_url) //更新请求url
        }

        fmt.Println("---Password---:", pwd)
        fmt.Println("---Content---:", string(content))

        //将字符串重新转换成请求
        rawreq, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(content)))

        //=====此处需要修改，客户端需要把完整的url传过来
        //这几步是必须的，不然会无法发送请求。。。
        rawreq.URL = tempurl
        rawreq.RequestURI = ""

        //把请求体内容正则出来
        r = regexp.MustCompile(`\s\W((([0-9a-zA-Z_=]){1,50})(&|=))+`)
        bodyParamEncode = strings.Trim(r.FindString(string(content)), "\n")
        bodyParamEncode = strings.TrimSpace(bodyParamEncode)
        fmt.Println("body参数匹配结果:", bodyParamEncode, "\n")

        if bodyParamEncode != "" {
            bodyParamsReader = strings.NewReader(bodyParamEncode)
            rawreq.Body = ioutil.NopCloser(bodyParamsReader) //更新请求参数内容
        }

        // rawreqContent, _ = httputil.DumpRequestOut(rawreq, true)
        // fmt.Printf("--------rawreq------:%s\n", string(rawreqContent))

        if err != nil {
            fmt.Println("生成请求发生异常")
            dump = base64.StdEncoding.EncodeToString([]byte("1"))
            goto pack
        }

        //发送请求
        // duration, _ = time.ParseDuration("0.1s") //延长时间,尽可能让每一组乱序不同
        // time.Sleep(duration)
        dumpResponse, _ = ForwardReq(rawreq, client)

        //封装响应结果和对应的密码
        if dumpResponse == nil {
            dump = base64.StdEncoding.EncodeToString([]byte("1"))
        } else {
            dump = base64.StdEncoding.EncodeToString(dumpResponse)
        }

    pack:
        respevent.Bodys[pwd] = dump
    }

    ret, err := json.Marshal(respevent)
    if err != nil {
        fmt.Println("返回数据json编码异常")
    }
    fmt.Println("------------------返回数据------------------")
    return &events.APIGatewayResponse{
        IsBase64Encoded: false,
        StatusCode:      200,
        Headers:         map[string]string{"Content-Type": "application/json"},
        Body:            string(ret),
    }, nil
}

func main() {
    // Make the handler available for Remote Procedure Call by Cloud Function
    cloudfunction.Start(ScfProxy)
}
