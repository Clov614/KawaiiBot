package tcd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// AT错误响应
type RespErrAT struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// AT响应
type RespAT struct {
	AccessToken string `json:"access_token"`
}

// 请求参数
type ReqAT struct {
	GrantType    string `json:"grant_type"` //固定为client_credentials
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AccessTOKEN string

func (r ReqAT) GetAT() AccessTOKEN {
	url := "https://aip.baidubce.com/oauth/2.0/token?" + "grant_type=" + r.GrantType
	url = url + "&client_id=" + r.ClientId + "&client_secret=" + r.ClientSecret
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Println(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	respB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	respAT := RespAT{}
	json.Unmarshal(respB, &respAT)
	return AccessTOKEN(respAT.AccessToken)
}
