package config

import (
	"fmt"
	"os"
	"time"
	"github.com/leezer3379/voice-sender/corp"
	"github.com/toolkits/pkg/logger"
)

// InitLogger init logger toolkits
func InitLogger() {
	c := Get().Logger

	lb, err := logger.NewFileBackend(c.Dir)
	if err != nil {
		fmt.Println("cannot init logger:", err)
		os.Exit(1)
	}

	lb.SetRotateByHour(true)
	lb.SetKeepHours(c.KeepHours)

	logger.SetLogging(c.Level, lb)
}



func Test(args []string) {
	c := Get()
	voiceClient := corp.New(c.Voice.Mobiles,c.Voice.Message, c.Voice.TtsCode,c.Voice.CalledShowNumber, c.Voice.TtsParam.Sname)

	if len(args) == 0 {
		fmt.Println("token not given")
		os.Exit(1)
	}

	for i := 0; i < len(args); i++ {
		mobile := args[i]
		voiceClient.Send(mobile, "阿里云Rds",c.Voice.TtsParam.Sname)

		time.Sleep(time.Millisecond*10)
	}
}

