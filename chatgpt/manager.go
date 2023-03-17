package chatgpt

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type Manager struct {
	Conns *Conns
	Done  chan struct{} // Conns为空时de退出信号
}

type Conns map[string]*Ctx // User:[Conn,Chan,Timer]

// 每个连接的列表
type Ctx struct {
	Conn       Conn
	DeadSignal chan struct{} // 存在或销毁的状态
	Timer
}

// 超时控制
type Timer struct {
	Start   *time.Time
	TimeOut int
}

type InitParams struct {
	ApiURL
	ApiKey
	ProxyURL
	Model   string `json:"model"`
	TimeOut int
}

// 初始化连接
func (m *Manager) initConn(user string, p InitParams, diaLogType string) {
	conn := NewConn()
	conn.InitMsgs(p.Model, diaLogType) // 初始化消息，并在消息列表头部添加 system
	conn.InitReq(string(p.ApiURL), string(p.ApiKey), conn.Data, string(p.ProxyURL))
	conns := make(Conns, 0)
	m.Conns = &conns
	//m.Conns = conns
	startTime := time.Now()
	ctx := &Ctx{
		Conn:       conn,
		DeadSignal: make(chan struct{}),
		Timer: Timer{
			Start:   &startTime,
			TimeOut: p.TimeOut,
		},
	}
	//conns = *m.Conns
	(*m.Conns)[user] = ctx
}

// 发送消息
func (m *Manager) SendMsg(user string, msg string) (replyMsg string) {
	conns := *m.Conns

	// 用户名检测(检测map中的key是否存在)
	if _, ok := conns[user]; !ok {
		return "[WARN] 该用户连接不存在或已经超时关闭"
	}
	ctx := conns[user]
	conn := ctx.Conn
	// 添加消息
	conn.AddUserMsg(msg)
	replyMsg = m.doSend(ctx, conn)
	return replyMsg
}

func (m *Manager) doSend(ctx *Ctx, conn Conn) (replyMsg string) {
	// 刷新时间
	nowTime := time.Now()
	ctx.Start = &nowTime
	resp, respErr := conn.PostMsg()
	if respErr.Code != "" {
		log.Info("[ERR]" + respErr.Code)
		// 错误类型为上下文超限
		if respErr.Code == "context_length_exceeded" {
			// 删除最开始对话
			log.Info(conn.DelEarlyMsg())
			// 递归调用直至上下文不超限
			return m.doSend(ctx, conn)
		} else {
			return resp.Choices[0].Message.Content
		}
	}
	return resp.Choices[0].Message.Content
}

// 当作函数回调传入 是否超时判断
func (c *Ctx) checkTimeOut() {
	// 死循环判断是否超时
	for {
		t := c.Timer
		nowTime := time.Now()
		time.Sleep(time.Second * 1) // 每隔一秒检测一次
		// 初始化连接直至未操作后的时间差是否大于定义的时间差(超时判断)
		if timeDiff := int(nowTime.Sub(*t.Start).Seconds()); timeDiff >= t.TimeOut {
			// 写入终止信号
			close(c.DeadSignal)
			return
		}
	}
}

// TODO: checkTime 可以用一个回调函数 在高阶函数中执行死循环判断是否超时
// TODO: 还需写一个销毁Conn（连接）的函数，通过监听DeadSignal
func (m *Manager) destoryConn(user string) {
	conns := *m.Conns
	select {
	case <-conns[user].DeadSignal:
		delete(*m.Conns, user) // 删除连接
		log.Info(user + ": 连接已关闭")
	}
}

// TODO: 控制生命周期的函数 初始化-->超时检测-->销毁-->结束 name: lifeCycleCtl
func (m *Manager) LifeCycleCtl(user string, p InitParams, diaLogType string) {
	m.initConn(user, p, diaLogType)
	ctx := (*m.Conns)[user]
	go ctx.checkTimeOut()  // 检测是否超时
	go m.destoryConn(user) // 超时删除conn

}

// 检测是否Conns为空
func (m *Manager) ConnsIsEmpty() {
	m.Done = make(chan struct{})
	go func() {
		for {
			if len(*m.Conns) == 0 {
				m.Done <- struct{}{}
			}
		}
	}()
}

// 销毁Manager 关闭通道
func (m *Manager) Destory() {
	close(m.Done)
}
