package main

import (
	"fmt"
	"github.com/KawaiiBot/Setting"
	"github.com/KawaiiBot/chatgpt"
	_ "github.com/KawaiiBot/logger"
	"github.com/eatmoreapple/openwechat"
	log "github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"os"
	"strings"
	"time"
)

var (
	conf   = new(Setting.Conf)
	params = chatgpt.InitParams{
		ApiURL: "https://api.openai.com/v1/chat/completions",
		Model:  "gpt-3.5-turbo",
	}
	manager = &chatgpt.Manager{}
)

func init() {
	if !conf.IsExist() { // 配置文件不存在
		conf.InitConfDefault() // 初始化默认conf
		log.Info("请配置conf.yml文件后重启程序")
		time.Sleep(time.Second * 5) // 延时5s退出
		os.Exit(1)
	}
	err := conf.ReadConf()
	if err != nil {
		log.Error(err)
	}
	// 判断是否配置成功
	if conf.ProxyURL == "" || conf.TimeOut == 0 || conf.ApiKey == "" {
		log.Error("配置文件错误")
		time.Sleep(time.Second * 5) // 延时5s退出
		os.Exit(2)
	}
	// 将读取的配置加载
	params.ApiKey = chatgpt.ApiKey(conf.ApiKey)
	params.ProxyURL = chatgpt.ProxyURL(conf.ProxyURL)
	params.TimeOut = conf.TimeOut
}

func main() {
	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式，上面登录不上的可以尝试切换这种模式

	// 注册消息处理函数
	bot.MessageHandler = func(msg *openwechat.Message) {
		if msg.IsText() && msg.Content == "ping" {
			msg.ReplyText("pong")
		}
	}
	// 注册登陆二维码回调
	if conf.ConsoleQrCode {
		bot.UUIDCallback = ConsoleQrCode
	} else {
		bot.UUIDCallback = openwechat.PrintlnQrcodeUrl
	}

	// 创建热登录对象
	reloadStorage := openwechat.NewFileHotReloadStorage("./storage.json")

	defer reloadStorage.Close()

	// 执行热登录
	if err := bot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		log.Error(err)
	}

	// 获取登陆的用户
	self, err := bot.GetCurrentUser()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 获取所有的好友
	friends, err := self.Friends()
	fmt.Println(friends, err)

	// 获取所有的群组
	groups, err := self.Groups()
	fmt.Println(groups, err)

	// 消息处理的回调函数
	bot.MessageHandler = func(msg *openwechat.Message) {
		user, _ := msg.Sender()

		if msg.IsTickledMe() {
			msg.ReplyText("别拍了，机器人是会被拍坏掉的。")
			return
		}
		// 回复文本消息
		if msg.IsText() {
			Ls := strings.SplitN(msg.Content, " ", 2)
			fmt.Println(Ls)
			if msg.Content == "/info" {
				_, err := msg.ReplyText("这里是一个由openwechat框架搭建的一个wechat_bot,主要的功能是AI对话")
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			//fmt.Println(user.ID())
			if len(Ls) >= 2 && (Ls[0] == "/chatgpt" || Ls[0] == "/Chatgpt" || Ls[0] == "/chat"+
				"GPT" || Ls[0] == "/ChatGPT") {
				manager.LifeCycleCtl(user.ID(), params, Ls[1])
				msg.ReplyText("新建chatGPT对话成功，赶紧开始聊天吧！\n" + "使用/q + 空格 + 内容(/q 内容)来提问吧")
				return
			} else if msg.Content == "/chatgpt" || msg.Content == "/chatGPT"+
				"" || msg.Content == "/ChatGPT" || msg.Content == "/Chatgpt" {
				manager.LifeCycleCtl(user.ID(), params, "你是一个全知全能的AI助手")
				msg.ReplyText("新建chatGPT对话成功，赶紧开始聊天吧！\n" + "使用/q + 空格 + 内容(/q 内容)来提问吧")
				return
			}
			if len(Ls) >= 2 && (Ls[0] == "/Q" || Ls[0] == "/q" || Ls[0] == "[Q]" || Ls[0] == "[q]") {
				fmt.Println("msg:" + Ls[1])
				if manager.Conns != nil {
					if _, ok := (*manager.Conns)[user.ID()]; ok {
						reply := manager.SendMsg(user.ID(), Ls[1])
						msg.ReplyText("[chatGPT] " + reply)
						return
					} else {
						msg.ReplyText("[Error] " + "对话连接已超时关闭\n使用/chatGPT新建连接吧")
					}
				} else {
					msg.ReplyText("[Error] " + "请初始化chatGPT\n使用/chatGPT新建连接吧")
				}

			}

		}
	}

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}

// 控制台输出二维码
func ConsoleQrCode(uuid string) {
	q, _ := qrcode.New("https://login.weixin.qq.com/l/"+uuid, qrcode.Low)
	fmt.Println(q.ToString(true))
}
