package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/leezer3379/voice-sender/dataobj"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/runner"

	"github.com/leezer3379/voice-sender/config"
	"github.com/leezer3379/voice-sender/cron"
	"github.com/leezer3379/voice-sender/redisc"
)

var (
	vers *bool
	help *bool
	conf *string
	test *string
)

func init() {
	vers = flag.Bool("v", false, "display the version.")
	help = flag.Bool("h", false, "print this help.")
	conf = flag.String("f", "", "specify configuration file.")
	test = flag.String("t", "", "test configuration.")
	flag.Parse()

	if *vers {
		fmt.Println("version:", config.Version)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	runner.Init()
	fmt.Println("runner.cwd:", runner.Cwd)
	fmt.Println("runner.hostname:", runner.Hostname)
}

func main() {
	aconf()
	pconf()

	if *test != "" {
		config.Test(strings.Split(*test, ","))
		os.Exit(0)
	}

	config.InitLogger()
	redisc.InitRedis()

	go cron.SendVoice()

	//ending()
	startHttp()
}
func startHttp() {
	http.HandleFunc("/voice", sendVoice) //设置访问的路由
	err := http.ListenAndServe("127.0.0.1:2008", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func sendVoice(w http.ResponseWriter, r *http.Request) {

	fmt.Println("method:", r.Method) //获取请求的方法
	if r.Method == "GET" {

		fmt.Println("OK")
	} else {

		s, _ := ioutil.ReadAll(r.Body) //把  body 内容读入字符串 s
		fmt.Println("body: ", string(s))
		var v3message dataobj.V3Message
		err := json.Unmarshal(s, &v3message)
		if err != nil {
			logger.Errorf("unmarshal message failed, err: %v, redis reply: %v", err)

		}
		fmt.Println("Tos: ", v3message.Tos)
		fmt.Println("Subject: ", v3message.Subject)
		fmt.Println("Content: ", v3message.Content)
		if count := len(v3message.Tos); count > 0 {
			for _, mobile := range v3message.Tos {
				go cron.V3SendVoice(mobile, v3message.Content)

			}
		}

	}
}

func ending() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-c:
		fmt.Printf("stop signal caught, stopping... pid=%d\n", os.Getpid())
	}

	logger.Close()
	redisc.CloseRedis()
	fmt.Println("sender stopped successfully")
}

// auto detect configuration file
func aconf() {
	if *conf != "" && file.IsExist(*conf) {
		return
	}

	*conf = path.Join(runner.Cwd, "etc", "voice-sender.local.yml")
	if file.IsExist(*conf) {
		return
	}

	*conf = path.Join(runner.Cwd, "etc", "voice-sender.yml")
	if file.IsExist(*conf) {
		return
	}

	fmt.Println("no configuration file for sender")
	os.Exit(1)
}

// parse configuration file
func pconf() {
	if err := config.ParseConfig(*conf); err != nil {
		fmt.Println("cannot parse configuration file:", err)
		os.Exit(1)
	} else {
		fmt.Println("parse configuration file:", *conf)
	}
}
