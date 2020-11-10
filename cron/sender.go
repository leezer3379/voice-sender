package cron

import (
	//"bytes"
	"encoding/json"
	"fmt"
	//"github.com/toolkits/pkg/runner"
	//"path"
	"strings"
	//"text/template"
	"time"

	"github.com/leezer3379/voice-sender/config"
	"github.com/leezer3379/voice-sender/corp"
	"github.com/leezer3379/voice-sender/dataobj"
	"github.com/leezer3379/voice-sender/redisc"
	"github.com/toolkits/pkg/logger"
)

var (
	semaphore  chan int
	voiceClient *corp.Client
)

func SendVoice() {
	c := config.Get()

	semaphore = make(chan int, c.Consumer.Worker)

	voiceClient = corp.New(c.Voice.Mobiles,c.Voice.Message,c.Voice.TtsCode,c.Voice.CalledShowNumber, c.Voice.TtsParam.Sname)

	for {
		messages := redisc.Pop(1, c.Consumer.Queue)
		if len(messages) == 0 {
			time.Sleep(time.Duration(300) * time.Millisecond)
			continue
		}

		SendAllVoice(messages)
	}
}

func V3SendVoice(mobile,context string) {
	err := voiceClient.Send(mobile,context)
	if err != nil {
		logger.Errorf("api send to %s fail:  %v", mobile, err)
	}
}

func SendAllVoice(messages []*dataobj.Message) {
	for _, message := range messages {
		semaphore <- 1
		go sendVoice(message)
	}
}

func sendVoice(message *dataobj.Message) {
	defer func() {
		<-semaphore
	}()

	content := genContent(message)
	mobiles := pasteMobile(message)

	logger.Info("<-- hashid: %v -->", message.Event.HashId)
	logger.Infof("hashid: %d: endpoint: %s, metric: %s, tags: %s", message.Event.HashId, message.ReadableEndpoint, strings.Join(message.Metrics, ","), message.ReadableTags)
	// 重置发送数据
	if count := len(mobiles); count > 0 {
		for _, mobile := range mobiles {
			err := voiceClient.Send(mobile, content)
			if err != nil {
				logger.Errorf("send to %s fail:  %v", message.Tos, err)
			}
		}
	}
	logger.Info("<-- /hashid: %v -->", message.Event.HashId)
}

var ET = map[string]string{
	"alert":    "告警",
	"recovery": "恢复",
}

func parseEtime(etime int64) string {
	t := time.Unix(etime, 0)
	return t.Format("2006-01-02 15:04:05")
}

func pasteMobile(message *dataobj.Message) []string {
	var MobilesStd []string
	for _, v := range message.Event.RecvUser {
		fmt.Printf("%s", v)
		MobilesStd = append(MobilesStd, string(v.Phone))
	}
	return MobilesStd
}

func genContent(message *dataobj.Message) string {

	sendttsparam := make(map[string]interface{})
	sendttsparam["Sname"] = message.Event.Sname

	dataType , _ := json.Marshal(sendttsparam)
	endttsparam := string(dataType)
	return endttsparam
	//fp := path.Join(runner.Cwd, "etc", "sms.tpl")
	//t, err := template.ParseFiles(fp)
	//if err != nil {
	//	payload := fmt.Sprintf("InternalServerError: cannot parse %s %v", fp, err)
	//	logger.Errorf(payload)
	//	return fmt.Sprintf(payload)
	//}
	//
	//var body bytes.Buffer
	//err = t.Execute(&body, map[string]interface{}{
	//	"IsAlert":   message.Event.EventType == "alert",
	//	"Status":    ET[message.Event.EventType],
	//	"Sname":     message.Event.Sname,
	//	"Endpoint":  message.ReadableEndpoint,
	//	"Metric":    strings.Join(message.Metrics, ","),
	//	"Tags":      message.ReadableTags,
	//	"Value":     message.Event.Value,
	//	"Info":      message.Event.Info,
	//	"Etime":     parseEtime(message.Event.Etime),
	//	"Elink":     message.EventLink,
	//	"Slink":     message.StraLink,
	//	"Clink":     message.ClaimLink,
	//	"IsUpgrade": message.IsUpgrade,
	//	"Bindings":  message.Bindings,
	//	"Priority":  message.Event.Priority,
	//})
	//
	//if err != nil {
	//	logger.Errorf("InternalServerError: %v", err)
	//	return fmt.Sprintf("InternalServerError: %v", err)
	//}
	//
	//return body.String()
}
