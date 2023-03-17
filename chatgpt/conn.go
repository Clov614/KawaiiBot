package chatgpt

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type ApiURL string
type ApiKey string
type ProxyURL string

type Conn struct {
	Request  *Request  // 请求指针
	Response *Response // 响应指针
	//MsgUser      []Message // 用户消息
	//MsgAssistant []Message // chatGPT 消息
	//MsgSys       Message   // 系统消息
	RespErr *RespErr
	Data    *Data
}

type RespErr struct {
	Error Error `json:"error"`
}
type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

type Request struct {
	Url      string
	ApiKey   string
	ProxyURL string
}

type Data struct {
	Model    string    `json:"model"`
	Messages *Messages `json:"messages"`
}

type Messages []Message

type Message map[string]interface{}

type Response struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Conn 构造函数
func NewConn() Conn {
	conn := Conn{
		Request:  new(Request),
		Response: new(Response),
		Data:     new(Data),
		RespErr:  new(RespErr),
	}
	return conn
}

// 返回一个request对象
func NewReq(url string, apiKey string, data *Data, ProxyURL string) *Request {
	req := new(Request)
	req.Url = url
	req.ApiKey = apiKey
	req.ProxyURL = ProxyURL
	return req
}

// 初始化request
func (c *Conn) InitReq(url string, apiKey string, data *Data, ProxyURL string) {
	//c.Request = new(Request)
	c.Request.Url = url
	c.Request.ApiKey = apiKey
	c.Data = data
	c.Request.ProxyURL = ProxyURL
}

// 发送消息
func (c *Conn) PostMsg() (Response, Error) {
	r := c.Request

	// 配置好data
	Msgs := *c.Data.Messages

	data := map[string]interface{}{
		"model":    c.Data.Model,
		"messages": Msgs,
	}

	// 将data转为json格式
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	// 配置代理
	proxy, err := url.Parse(r.ProxyURL)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}

	req, err := http.NewRequest("POST", r.Url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("发起请求错误(未配置正确代理)")
		time.Sleep(time.Second * 5)
		os.Exit(-1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.ApiKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()
	// 获取响应的主体内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
	}

	if resp.StatusCode != 200 {
		log.Error("请求状态码: ", resp.StatusCode)
		//time.Sleep(time.Second * 5)
		//os.Exit(resp.StatusCode)
		//fmt.Println(string(body))
		// 错误信息
		if err := json.Unmarshal(body, c.RespErr); err != nil {
			panic(err)
		}
	} else {
		// 重置错误代码
		c.RespErr = new(RespErr)
	}

	// 将resp添加入Conn.Resp中
	//c.Response = new(Response) // 注意初始化指针
	if err := json.Unmarshal(body, c.Response); err != nil {
		panic(err)
	}

	// 消息列表加入AI的回答
	c.AddAiMsg(c.Response.Choices[0].Message.Role, c.Response.Choices[0].Message.Content)

	return *c.Response, c.RespErr.Error
}

// 配置消息 DiaLogTe: 系统消息system model
func (c *Conn) InitMsgs(model string, diaLogType string) {
	var sysMsg = Message{
		"role":    "system",
		"content": diaLogType,
	}

	msgs := &Messages{
		sysMsg,
	}
	c.Data.Model = model
	c.Data.Messages = msgs
}

// 添加用户消息
func (c *Conn) AddUserMsg(content string) {
	msg := Message{
		"role":    "user",
		"content": content,
	}
	pMsgs := c.Data.Messages
	msgs := *pMsgs
	msgs = append(msgs, msg)
	c.Data.Messages = &msgs
}

// 添加AI回复消息
func (c *Conn) AddAiMsg(role string, content string) {
	msg := Message{
		"role":    role,
		"content": content,
	}
	pMsgs := c.Data.Messages
	msgs := *pMsgs
	msgs = append(msgs, msg)
	c.Data.Messages = &msgs
}

// 上下文超限 删除最早对话
func (c *Conn) DelEarlyMsg() (errStr string) {
	pMsgs := c.Data.Messages
	msgs := *pMsgs
	if len(msgs) <= 3 {
		errStr = "By Func DelEarlyMsg: 上下文超限"
		return errStr
	}
	sysMsgs := msgs[0:1]
	msgs = msgs[3:]
	for _, v := range msgs {
		sysMsgs = append(sysMsgs, v)
	}
	c.Data.Messages = &sysMsgs
	return ""
}
