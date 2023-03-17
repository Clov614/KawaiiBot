package Setting

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Conf struct {
	ProxyURL      string `yaml:"proxy_url"`
	ApiKey        string `yaml:"api_key"`
	TimeOut       int    `yaml:"time_out"`
	ConsoleQrCode bool   `yaml:"console_qr_code"`
}

var (
	InfoHelp = "# 配置文件\n# proxy_url 代理地址【必要】 （true or false)\n# api_key chatGPT的APIKEY \n# time_out 每个对话超时的时间\n"
	// 默认配置
	DefaultConf = Conf{
		ProxyURL:      "http://127.0.0.1:7890",
		TimeOut:       1800,  // 默认对话超时时间为30分钟
		ConsoleQrCode: false, // 微信登录二维码是否控制台输出
		ApiKey:        "",
	}
)

// 判断conf是否存在
func (c *Conf) IsExist() bool {
	if !PathExists("./conf/conf.yaml") {
		return false
	}
	return true
}

// 初始化conf
func (c *Conf) InitConfDefault() {
	log.Println("生成默认conf文件中")
	err := os.MkdirAll("./conf", 0766)
	if err != nil {
		log.Fatalln("创建conf目录失败:", err)
	}
	file, err := os.OpenFile("./conf/conf.yaml", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatalln("创建conf.yaml失败:", err)
	}
	defer file.Close()
	file.Write([]byte(InfoHelp))
	// 初始化conf.yaml
	dataStr, _ := yaml.Marshal(&DefaultConf)
	file.Write(dataStr)
}

// 加载配置
func (c *Conf) ReadConf() error {
	err := ReadYaml(&c, "./conf/conf.yaml")
	if err != nil {
		return err
	}
	return nil
}

// 判断路径是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// 读取yaml
func ReadYaml(_type interface{}, path string) (err error) {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln("读取Error path: "+path, err)
	}
	err = yaml.Unmarshal(file, _type)
	if err != nil {
		return err
	}
	return nil
}

// 处理conf
func (c *Conf) Handler() {
	exist := c.IsExist()
	if !exist {
		c.InitConfDefault()
	}
	err := c.ReadConf()
	if err != nil {
		log.Fatalln(err)
	}
}
