package corp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/toolkits/pkg/logger"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)


const dingTimeOut = time.Second * 1

// Client
type Client struct {
	Mobiles          []string
	Message          string
	TtsParam         map[string]interface{}
	TtsCode          string
	CalledShowNumber string
	AccessKeyId      string
	AccessKeySecret  string
}
// DingClient
type DingClient struct {
	Mobiles []string
	token   string
	openUrl string
	IsAtAll bool
}

// DingNew
func DingNew(token string, mobiles []string, isAtAll bool) *DingClient {
	c := new(DingClient)
	c.openUrl = "https://oapi.dingtalk.com/robot/send?access_token="
	c.token = token
	c.Mobiles = mobiles
	c.IsAtAll = isAtAll
	return c
}

// New
func New(mobiles []string, message, ttscode, calledshownumber, sname string) *Client {
	c := new(Client)
	sendttsparam := make(map[string]interface{})
	sendttsparam["Sname"] = sname
	c.Mobiles = mobiles
	c.Message = message
	c.TtsParam = sendttsparam
	c.TtsCode = ttscode
	c.CalledShowNumber = calledshownumber

	return c
}

func (c *Client) GetMobiles() []string {
	return c.Mobiles
}

func (c *DingClient) generateData(mobile []string, msg string) interface{} {
	postData := make(map[string]interface{})
	postData["msgtype"] = "text"
	sendContext := make(map[string]interface{})
	sendContext["content"] = msg
	postData["text"] = sendContext

	at := make(map[string]interface{})
	if !c.IsAtAll && len(c.Mobiles) > 0 && c.token != "" {
		at["atMobiles"] = c.Mobiles // 根据手机号@指定人
	} else if len(mobile) > 0{
		at["atMobiles"] = mobile // 根据手机号@指定人
	} else {
		c.IsAtAll = true
	}

	at["isAtAll"] = c.IsAtAll // s是否@所有人
	postData["at"] = at

	return postData
}

func (c DingClient) GetToken() string {
	return c.token
}
// Err
type Err struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
// Result 发送消息返回结果
type Result struct {
	Err
}
// DingSend 发送信息
func (c *DingClient) DingSend(token string, mobile []string, msg string) error {

	postData := c.generateData(mobile, msg)
	if c.GetToken() != "" {
		// 配置了token 说明采用配置文件的token
		token = c.GetToken()
	}
	url := c.openUrl + token
	fmt.Println(url,postData)
	resultByte, err := DingjsonPost(url, postData)
	if err != nil {
		return fmt.Errorf("invoke send api fail: %v", err)
	}

	result := Result{}
	err = json.Unmarshal(resultByte, &result)
	if err != nil {
		return fmt.Errorf("parse send api response fail: %v", err)
	}

	if result.ErrCode != 0 || result.ErrMsg != "ok" {
		err = fmt.Errorf("invoke send api return ErrCode = %d, ErrMsg = %s ", result.ErrCode, result.ErrMsg)
	}

	return err
}



// Send 发送信息
func (c *Client) Send(mobile string, sname string) error {
	//var sendttsparam map[string]string
	//sendttsparam["ops"] = c
	if mobile == "" {
		fmt.Printf("%s 未填写手机号码。",sname)
		return nil
	}
	sendttsparam := make(map[string]interface{})
	sendttsparam["Sname"] = sname

	dataType , _ := json.Marshal(sendttsparam)
	endttsparam := string(dataType)
	client, err := sdk.NewClientWithAccessKey(os.Getenv("VOICE_REGION_ID"), os.Getenv("VOICE_ACCESS_KEY_ID"), os.Getenv("VOICE_ACCESS_KEY_SECRET"))

	if err != nil {
		panic(err)
	}


	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "dyvmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SingleCallByTts"
	request.QueryParams["RegionId"] = "cn-beijing"
	request.QueryParams["CalledNumber"] = mobile
	request.QueryParams["TtsCode"] = c.TtsCode
	request.QueryParams["CalledShowNumber"] = c.CalledShowNumber
	request.QueryParams["TtsParam"] = endttsparam

	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		return err
	}
	fmt.Print(response.GetHttpContentString())
	return err
}

func DingjsonPost(url string, data interface{}) ([]byte, error) {
	jsonBody, err := encodeJSON(data)
	if err != nil {
		return nil, err
	}
	fmt.Println(strings.NewReader(string(jsonBody)))
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		logger.Info("ding talk new post request err =>", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := getClient()
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("ding talk post request err =>", err)
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func getClient() *http.Client {
	// 通过http.Client 中的 DialContext 可以设置连接超时和数据接受超时 （也可以使用Dial, 不推荐）
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
				conn, err := net.DialTimeout(network, addr, dingTimeOut) // 设置建立链接超时
				if err != nil {
					return nil, err
				}
				_ = conn.SetDeadline(time.Now().Add(dingTimeOut)) // 设置接受数据超时时间
				return conn, nil
			},
			ResponseHeaderTimeout: dingTimeOut, // 设置服务器响应超时时间
		},
	}
}

func jsonPost(url string, data url.Values) ([]byte, error) {
	//jsonBody, err := encodeJSON(data)
	//if err != nil {
	//	return nil, err
	//}
	//req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))

	req, err := http.PostForm(url, data)
	if err != nil {
		panic(err)
	}
	//if err != nil {
	//	logger.Info("ding talk new post request err =>", err)
	//	return nil, err
	//}

	//req.Header.Set("Content-Type", "application/json")

	//client := getClient()
	//resp, err := client.Do(req)
	//if err != nil {
	//	logger.Error("ding talk post request err =>", err)
	//	return nil, err
	//}

	defer req.Body.Close()
	return ioutil.ReadAll(req.Body)
}

func encodeJSON(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Client) generateData(mobile string, msg string) url.Values {
	params := url.Values{
		"mobile":  {mobile},
		"message": {msg},
	}
	return params
}

//func getClient() *http.Client {
//	// 通过http.Client 中的 DialContext 可以设置连接超时和数据接受超时 （也可以使用Dial, 不推荐）
//	return &http.Client{
//		Transport: &http.Transport{
//			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
//				conn, err := net.DialTimeout(network, addr, dingTimeOut) // 设置建立链接超时
//				if err != nil {
//					return nil, err
//				}
//				_ = conn.SetDeadline(time.Now().Add(dingTimeOut)) // 设置接受数据超时时间
//				return conn, nil
//			},
//			ResponseHeaderTimeout: dingTimeOut, // 设置服务器响应超时时间
//		},
//	}
//}
