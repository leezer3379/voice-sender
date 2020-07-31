package corp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"os"
)



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


// Send 发送信息
func (c *Client) Send(mobile string, sname string) error {
	//var sendttsparam map[string]string
	//sendttsparam["ops"] = c
	if mobile == "" {
		fmt.Printf("%s 未填写手机号码。",sname)
	}
	sendttsparam := make(map[string]interface{})
	sendttsparam["Sname"] = sname

	dataType , _ := json.Marshal(sendttsparam)
	endttsparam := string(dataType)
	client, err := sdk.NewClientWithAccessKey(os.Getenv("REGION_ID"), os.Getenv("ACCESS_KEY_ID"), os.Getenv("ACCESS_KEY_SECRET"))
	if err != nil {
		panic(err)
	}


	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "dyvmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SingleCallByTts"
	request.QueryParams["RegionId"] = os.Getenv("REGION_ID")
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
