package dingtalk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Subscriber struct {
	AppKey    string
	AppSecret string

	accessToken string
}

func (s Subscriber) Push() {

}

type getAccessTokenRes struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
}

func (s *Subscriber) loadAccessToken() (err error) {
	resp, err := http.Get(fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s", s.AppKey, s.AppSecret))

	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	res := &getAccessTokenRes{}

	err = json.Unmarshal(body, res)

	if err != nil {
		return
	}

	s.accessToken = res.AccessToken

	return nil
}

func (s *Subscriber) GetAccessToken() (accessToken string, err error) {
	if s.accessToken == "" {
		err := s.loadAccessToken()
		if err != nil {
			return "", err
		}
	}

	return s.accessToken, nil
}

func (s *Subscriber) doRequest() {

}
