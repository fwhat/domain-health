package dingtalk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/qjues/domain-health/config"
	"github.com/qjues/domain-health/domain_health/subscriber"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Subscriber struct {
	Secret  string `json:"secret"`
	WebHook string `json:"web_hook"`

	messages    []subscriber.Message
	messagesMux sync.Mutex
}

func (s *Subscriber) InitFromMap(config map[string]interface{}) (err error) {
	marshal, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(marshal, s)
	if err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) AddMessage(msg subscriber.Message) {
	s.messagesMux.Lock()
	defer s.messagesMux.Unlock()

	s.messages = append(s.messages, msg)
}

func (s *Subscriber) Delivery() error {
	s.messagesMux.Lock()
	defer s.messagesMux.Unlock()
	if len(s.messages) == 0 {
		return nil
	}

	text := ""

	expiredTemplate := "域名名称: **%s**  \n  报警类型: **%s**  \n  过期时间: **%s**  \n  ________________________  \n"
	preExpiredTemplate := "域名名称: **%s**  \n  报警类型: **%s**  \n  过期时间: **%s**  \n  ________________________  \n"
	connectTimeoutTemplate := "域名名称: **%s**  \n  报警类型: **%s**  \n  连接耗时: **%ds**  \n  ________________________  \n"

	for _, message := range s.messages {
		switch message.Type {
		case subscriber.CretExpired:
			text += fmt.Sprintf(expiredTemplate, message.Domain.Address, "域名证书已过期", formatTimestamp(message.Domain.CertInfo.ExpireTime))
		case subscriber.CretPreExpired:
			text += fmt.Sprintf(preExpiredTemplate, message.Domain.Address, "域名证书即将过期", formatTimestamp(message.Domain.CertInfo.ExpireTime))
		case subscriber.ConnectTimeout:
			text += fmt.Sprintf(connectTimeoutTemplate, message.Domain.Address, "域名连接超时", config.Instance.ConnectTimeout)
		}
	}

	msg := `{
     "msgtype": "markdown",
     "markdown": {
         "title":"域名异常警告",
         "text": "# 域名异常警告  \n  ________________________  \n %s  提醒时间: **%s**"
     },
      "at": {
          "isAtAll": true
      }
 }`
	s.messages = []subscriber.Message{}

	return s.webHookSend(bytes.NewReader([]byte(fmt.Sprintf(msg, text, formatTimestamp(time.Now().Unix())))))
}

func formatTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

func (s *Subscriber) getSign() (timestamp int64, sign string) {
	timestamp = time.Now().UnixNano() / 1e6

	h := hmac.New(sha256.New, []byte(s.Secret))
	h.Write([]byte(fmt.Sprintf("%d\n%s", timestamp, s.Secret)))

	return timestamp, base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type baseRes struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (s *Subscriber) webHookSend(reqBody io.Reader) (err error) {
	apiUrl, _ := url.Parse(s.WebHook)
	if s.Secret != "" {
		timestamp, sign := s.getSign()

		query := apiUrl.Query()
		query.Add("timestamp", fmt.Sprintf("%d", timestamp))
		query.Add("sign", sign)
		apiUrl.RawQuery = query.Encode()
	}

	resp, err := http.Post(apiUrl.String(), "application/json", reqBody)
	if err != nil {
		log.Debug(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debug(err)
		return
	}

	res := &baseRes{}

	err = json.Unmarshal(body, res)

	if err != nil {
		log.Debug(err)
		return
	}

	if res.ErrCode != 0 {
		err = errors.New(fmt.Sprintf("%d %s", res.ErrCode, res.ErrMsg))
		log.Error(err)
		return
	}

	return
}
