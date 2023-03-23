package tcd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type ReqTC struct {
	Text        string `json:"text"`
	AccessToken string `json:"access_token"`
}

// 响应
type RespTC struct {
	Conclusion     string `json:"conclusion"`
	LogID          int64  `json:"log_id"`
	Data           []Data `json:"data"`
	IsHitMd5       bool   `json:"isHitMd5"`
	ConclusionType int    `json:"conclusionType"`
}
type Hits struct {
	WordHitPositions  []interface{} `json:"wordHitPositions"`
	Probability       float64       `json:"probability"`
	DatasetName       string        `json:"datasetName"`
	Words             []interface{} `json:"words"`
	ModelHitPositions [][]float64   `json:"modelHitPositions"`
}
type Data struct {
	Msg            string `json:"msg"`
	Conclusion     string `json:"conclusion"`
	Hits           []Hits `json:"hits"`
	SubType        int    `json:"subType"`
	ConclusionType int    `json:"conclusionType"`
	Type           int    `json:"type"`
}

func (r ReqTC) Detect(respTc *RespTC) bool {
	url := "https://aip.baidubce.com/rest/2.0/solution/v1/text_censor/v2/user_defined?access_token=" + r.AccessToken
	r.Text = strings.ReplaceAll(r.Text, "\n", "")
	url = url + "&text=" + r.Text
	client := &http.Client{}
	//data := fmt.Sprintf(`{"text":"%s"}`, r.Text)
	req, err := http.NewRequest("POST", url, nil) //bytes.NewBufferString(data)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Println(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	respB, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(respB, respTc)
	if respTc.Conclusion == "不合规" {
		return false
	} else {
		return true
	}
}
