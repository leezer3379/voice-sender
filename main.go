package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/leezer3379/voice-sender/dataobj"
	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/runner"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/leezer3379/voice-sender/config"
	"github.com/leezer3379/voice-sender/cron"
	//"github.com/leezer3379/voice-sender/redisc"
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
	//redisc.InitRedis()

	//go cron.SendVoice()

	//ending()
	startHttp()
}
func startHttp() {
	http.HandleFunc("/voice", sendVoice) //设置访问的路由
	http.HandleFunc("/addwl", addWL) //设置访问的路由
	err := http.ListenAndServe("0.0.0.0:2008", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func IsAddWL(instanceid string) bool {
	// 是否添加白名单
	curtime := time.Now().Unix()
	jsonconf := config.LoadJsonConfig()
	for i := 0;i < len(jsonconf.WLs); i++ {
		if curtime > jsonconf.WLs[i].ExTime {
			//删除第i个元素
			jsonconf.WLs = append(jsonconf.WLs[:i], jsonconf.WLs[i+1:]...)
			i--
			config.SaveJsonConfig(jsonconf)
		}

	}
	for i := 0;i < len(jsonconf.WLs); i++ {
		if jsonconf.WLs[i].InstanceId == instanceid {
			return true
		}
	}
	return false
}

func AddWL(instanceid, t string) {
	jsonconf := config.LoadJsonConfig()
	//白名单过时24小时
	dd, _ := time.ParseDuration("24h")
	if t != "" {
		dd, _ = time.ParseDuration(t)
	}

	tm := time.Now()
	tm = tm.Add(dd)
	var wl config.WL

	wl.ExTime = tm.Unix()
	wl.InstanceId = instanceid
	jsonconf.WLs = append(jsonconf.WLs, wl)
	config.SaveJsonConfig(jsonconf)
}

func Isupdate(instanceid string)  bool {
	// 是否升级，判断次数，大于2次的, 屏蔽的时候删除升级规则
	jsonconf := config.LoadJsonConfig()
	for i := 0; i < len(jsonconf.Ups); i++ {
		fmt.Println("debug...................")
		fmt.Println(jsonconf.Ups[i])
		fmt.Println(instanceid)
		fmt.Println("debug...................")
		if jsonconf.Ups[i].InstanceId == instanceid {
			if jsonconf.Ups[i].Count >= 2 {
				jsonconf.Ups[i].IsUp = true
				config.SaveJsonConfig(jsonconf)
				return true
			} else {
				fmt.Println("debug+111111111...................")
				jsonconf.Ups[i].Count += 1
				fmt.Println(jsonconf.Ups[i].Count)
				config.SaveJsonConfig(jsonconf)
				return false
			}
		}
	}

	// 在新增之前清理24小时
	curtime := time.Now().Unix()
	for i := 0; i < len(jsonconf.Ups); i++ {
		if curtime > jsonconf.Ups[i].ExTime {
			//删除第i个元素
			jsonconf.Ups = append(jsonconf.Ups[:i], jsonconf.Ups[i+1:]...)
			i--
		}
	}
	//  如果未找到
	var up config.Up
	up.Count = 1
	up.IsUp = false
	up.InstanceId = instanceid
	// 升级过时24小时
	dd, _ := time.ParseDuration("24h")
	tm := time.Now()
	tm = tm.Add(dd)
	up.ExTime = tm.Unix()
	jsonconf.Ups = append(jsonconf.Ups,up)
	config.SaveJsonConfig(jsonconf)
	return false
}

func addWL(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //获取请求的方法
	if r.Method == "GET" {
		// 第二种方式
		query := r.URL.Query()
		instanceid := query.Get("instanceid")
		t := query.Get("time")
		ok := IsAddWL(instanceid)
		if ok {
			fmt.Fprintf(w,`{"code":200, "msg": "add ok"}`)
		} else {
			AddWL(instanceid, t)
			fmt.Fprintf(w,`{"code":200, "msg": "add ok"}`)
		}
	}
}


func sendVoice(w http.ResponseWriter, r *http.Request) {

	fmt.Println("method:", r.Method) //获取请求的方法
	if r.Method == "GET" {
		fmt.Fprintf(w,`{"code":200}`)

	} else {

		//无状态每次需要读取一次文件
		jsonconf := config.LoadJsonConfig()
		fmt.Println("jsonconf:")
		fmt.Println(jsonconf)

		s, _ := ioutil.ReadAll(r.Body) //把  body 内容读入字符串 s
		fmt.Println("body: ", string(s))
		var v3message dataobj.V3Message
		err := json.Unmarshal(s, &v3message)
		if err != nil {
			logger.Errorf("unmarshal message failed, err: %v, redis reply: %v", err)

		}
		//过滤不需要报警的实例
		if IsAddWL(v3message.InstanceId) {
			fmt.Fprintf(w,`{"code":200, "msg": "no alert"}`)

		}
		// 是否升级
		if Isupdate(v3message.InstanceId) {
			v3message.Tos = append(v3message.Tos, jsonconf.LeaderPhone...)
			v3message.Content += "\n已升级上报@刘翔"
		}
		fmt.Println("Tos: ", v3message.Tos)
		fmt.Println("Subject: ", v3message.Subject)
		fmt.Println("InstaceId: ", v3message.InstanceId)
		fmt.Println("Content: ", v3message.Content)

		go cron.V3SendDingTalk(jsonconf.Tk, v3message.Content, v3message.Tos)

		if count := len(v3message.Tos); count > 0 {
			for _, mobile := range v3message.Tos {
				go cron.V3SendVoice(mobile, v3message.Subject, v3message.Content)

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
	//redisc.CloseRedis()
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
