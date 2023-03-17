package main

import (
	"fmt"
	"github.com/OpenWechat/Setting"
	"github.com/OpenWechat/chatgpt"
	"github.com/eatmoreapple/openwechat"
	"github.com/skip2/go-qrcode"
	"log"
	"os"
	"strings"
	"time"
)

var (
	conf   = new(Setting.Conf)
	params = chatgpt.InitParams{
		ApiURL: "https://api.openai.com/v1/chat/completions",
		//ApiKey:   "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",
		//ProxyURL: "http://127.0.0.1:7890",
		Model: "gpt-3.5-turbo",
		//TimeOut:  300,
	}
	manager = &chatgpt.Manager{}
)

func init() {
	if !conf.IsExist() { // 配置文件不存在
		conf.InitConfDefault() // 初始化默认conf
		log.Println("请配置conf.yml文件后重启程序")
		time.Sleep(time.Second * 5) // 延时5s退出
		os.Exit(1)
	}
	err := conf.ReadConf()
	if err != nil {
		log.Fatalln(err)
	}
	// 判断是否配置成功
	if conf.ProxyURL == "" || conf.TimeOut == 0 || conf.ApiKey == "" {
		log.Println("配置文件错误")
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

	//// 扫码回调
	//bot.ScanCallBack = func(body openwechat.CheckLoginResponse) {
	//	Avater, _ := body.Avatar()
	//	// 去掉Base64字符串中的非数据部分
	//	// 写正则提取出base64
	//	pattern, err := regexp.Compile("^data:img/jpg;base64,([\\s\\S]*)$")
	//	target := pattern.FindStringSubmatch(Avater)
	//	imgBase64 := ""
	//	if len(target) != 0 {
	//		imgBase64 = target[1]
	//	} else {
	//		log.Fatalln("len(target) == 0")
	//	}
	//	// 解码Base64字符串为字节数组
	//	imgData, err := base64.StdEncoding.DecodeString(imgBase64)
	//	if err != nil {
	//		fmt.Println("Error decoding image data:", err)
	//		return
	//	}
	//	f, _ := os.OpenFile("./Avater.jpg", os.O_CREATE|os.O_RDWR, 0777)
	//	_, err2 := f.Write(imgData) //buffer输出到jpg文件中（不做处理，直接写到文件）
	//	if err2 != nil {
	//		log.Fatalln("图像保存失败", err2)
	//	}
	//}

	// 创建热登录对象
	reloadStorage := openwechat.NewFileHotReloadStorage("./storage.json")

	defer reloadStorage.Close()

	// 执行热登录
	if err := bot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		log.Fatalln(err)
	}

	// 登陆
	//if err := bot.Login(); err != nil {
	//	fmt.Println(err)
	//	return
	//}

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

	//// 注册消息处理函数
	//bot.MessageHandler = func(msg *openwechat.Message) {
	//	if msg.IsText() && msg.Content == "ping" {
	//		msg.ReplyText("pong")
	//	}
	//}

	//// 发送chatgpt消息
	//bot.MessageHandler = func(msg *openwechat.Message) {
	//	if _, ok := (*manager.Conns)[msg.ToUserName]; msg.IsText() && ok {
	//		reply := manager.SendMsg(msg.ToUserName, msg.Content)
	//		msg.ReplyText("[chatGPT] " + reply)
	//	}
	//}

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
			if len(Ls) >= 2 && Ls[0] == "/chatgpt" || Ls[0] == "/Chatgpt" || Ls[0] == "/chatGPT" {
				manager.LifeCycleCtl(user.ID(), params, Ls[1])
				msg.ReplyText("新建chatGPT对话成功，赶紧开始聊天吧！\n" + "使用/q + 空格 + 内容(/q 内容)来提问吧")
				return
			} else if msg.Content == "/chatgpt" || msg.Content == "/chatGPT" {
				manager.LifeCycleCtl(user.ID(), params, "你是一个全知全能的AI助手")
				msg.ReplyText("新建chatGPT对话成功，赶紧开始聊天吧！\n" + "使用/q + 空格 + 内容(/q 内容)来提问吧")
				return
			}
			if len(Ls) >= 2 && Ls[0] == "/Q" || Ls[0] == "/q" || Ls[0] == "[Q]" || Ls[0] == "[q]" {
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

	// 消息处理
	// 构造dispatcher
	//dispatcher := openwechat.NewMessageMatchDispatcher()
	// 注册消息处理函数
	//dispatcher.RegisterHandler(matchFunc, handleInfo)
	//dispatcher.RegisterHandler(matchSendChatGPT, handleSendChatGPT)
	//dispatcher.RegisterHandler(matchFunc, handleInfo)
	// 注册消息回调函数
	//bot.MessageHandler = dispatcher.AsMessageHandler()

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}

// 控制台输出二维码
func ConsoleQrCode(uuid string) {
	q, _ := qrcode.New("https://login.weixin.qq.com/l/"+uuid, qrcode.Low)
	fmt.Println(q.ToString(true))
}

// 消息匹配函数
func matchFunc(msg *openwechat.Message) bool {
	if msg.IsText() && msg.Content == "/info" {
		return true
	}
	fmt.Println(msg.MsgType)
	fmt.Println(msg.AppMsgType)
	fmt.Println(msg.Url)
	return false
}

func matchNewChatGPT(msg *openwechat.Message) bool {
	if msg.IsText() {
		//msg.ReplyText(msg.Content)
		Ls := strings.Split(msg.Content, " ")
		if len(Ls) >= 2 && Ls[0] == "/chatgpt" || Ls[0] == "/Chatgpt" || Ls[0] == "/chatGPT" {
			return true
		} else if msg.Content == "/chatgpt" || msg.Content == "/chatGPT" {
			manager.LifeCycleCtl(msg.ToUserName, params, "你是一个全知全能的AI助手")
			msg.ReplyText("新建chatGPT对话成功，赶紧开始聊天吧！")
			return true
		}
	}
	return false
}

func handleNewChatGPT(ctx *openwechat.MessageContext) {
	Ls := strings.Split(ctx.Message.Content, " ")
	manager.LifeCycleCtl(ctx.Message.ToUserName, params, Ls[1])
	_, err := ctx.Message.ReplyText("新建chatGPT对话成功，赶紧开始聊天吧！")
	if err != nil {
		return
	}
}

func matchSendChatGPT(msg *openwechat.Message) bool {
	if _, ok := (*manager.Conns)[msg.ToUserName]; msg.IsText() && ok {
		return true
	}
	return false
}

func handleSendChatGPT(ctx *openwechat.MessageContext) {
	reply := manager.SendMsg(ctx.Message.ToUserName, ctx.Message.Content)
	_, err := ctx.Message.ReplyText("[chatGPT] " + reply)
	if err != nil {
		return
	}
}

// 文件相关测试
//func matchFunc(msg *openwechat.Message) bool {
//	fmt.Println("[msgType]", int(msg.MsgType))
//	fmt.Println("[AppMsgType]", int(msg.AppMsgType))
//	if int(msg.MsgType) == 49 {
//		fmt.Println("allin")
//		resp, err := msg.GetFile()
//		if err != nil {
//			log.Fatalln(err)
//		}
//		defer resp.Body.Close()
//		respByte, _ := ioutil.ReadAll(resp.Body)
//		file, _ := os.OpenFile("./resp.word", os.O_CREATE|os.O_RDWR, 0777)
//		file.Write(respByte)
//		defer file.Close()
//	}
//	return false
//}

//func matchFunc(msg *openwechat.Message) bool {
//	if msg.IsText() && msg.Content == "/chatgpt" {
//		manager.LifeCycleCtl(msg.ToUserName, params, "你是一个全知全能的人工智能")
//		return true
//	}
//	fmt.Println(msg.MsgType)
//	fmt.Println(msg.AppMsgType)
//	fmt.Println(msg.Url)
//	return false
//}

// 消息处理函数
//func handleInfo(ctx *openwechat.MessageContext) {
//	_, err := ctx.Message.ReplyText("这里是一个由openwechat框架搭建的一个wechat_bot,主要的功能是AI对话")
//	if err != nil {
//		return
//	}
//}
