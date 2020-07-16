# voice-sender

Nightingale的理念，是将告警事件扔到redis里就不管了，接下来由各种sender来读取redis里的事件并发送，毕竟发送报警的方式太多了，适配起来比较费劲，希望社区同仁能够共建。

这里提供一个钉钉的sender，参考了[https://github.com/n9e/wechat-sender](https://github.com/n9e/wechat-sender) 及 [https://github.com/wulorn/dingtalk](https://github.com/wulorn/dingtalk)，具体如何获取钉钉机器人token，也可以参看钉钉官网

## compile

```bash
cd $GOPATH/src
mkdir -p github.com/n9e
cd github.com/n9e
git clone https://github.com/leezer3379/voice-sender.git
cd voice-sender
./control build
```

如上编译完就可以拿到二进制了。

## configuration

直接修改etc/voice-sender.yml即可

## 注意

voice-sender仅支持阿里云语音服务且只能传递一个变量sname策略名称, 设置环境变量：
```bash
ACCESS_KEY_ID=
ACCESS_KEY_SECRET=
REGION_ID=cn-hangzhou

```

需monapi.yaml设置里的notify添加voice告警，如下：

```yaml
notify:
  p1: ["voice"]
  p2: ["voice"]
  p3: ["voice"]
```

## pack

编译完成之后可以打个包扔到线上去跑，将二进制和配置文件打包即可：

```bash
tar zcvf voice-sender.tar.gz voice-sender etc/voice-sender.yml etc/voice.tpl
```

## voice-sender.yml

```yaml
voice:
  ttscode: "TTS_xxx"  //阿里云 语音模板
  calledshownumber: "00000000000"    //阿里云语音主叫号码
  ttsparam:
    sname: "策略名称"   // 传递策略名
```
## test

配置etc/voice-sender.yml，相关配置修改好，我们先来测试一下是否好使， `./voice-sender -p phone`，token为钉钉群机器人的token值，程序会自动读取etc目录下的配置文件，发一个测试消息给钉钉群`token`

## run

如果测试发送没问题，扔到线上跑吧，使用systemd或者supervisor之类的托管起来，systemd的配置实例：


```
$ cat voice-sender.service
[Unit]
Description=Nightingale voice sender
After=network-online.target
Wants=network-online.target

[Service]
User=root
Group=root

Type=simple
ExecStart=/home/n9e/voice-sender
WorkingDirectory=/home/n9e

Restart=always
RestartSec=1
StartLimitInterval=0

[Install]
WantedBy=multi-user.target
```