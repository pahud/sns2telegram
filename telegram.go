package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

// Telegram struct type
type Telegram struct {
	token    string
	endpoint string
	url      string
}

// NewTelegram return an Telegram struct
func NewTelegram() *Telegram {
	return &Telegram{}
}

func (t Telegram) sendMessage(m string, chatID string) {
	t.url = "https://api.telegram.org/bot" + t.token + "/" + "sendMessage"
	resp, err := http.PostForm(t.url,
		url.Values{
			"chat_id": {chatID},
			"text":    {m},
		})
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		// handle error
	}
	// log.Println(json.MarshalIndent(string(body), "", "  "))
	var obj map[string]interface{}
	json.Unmarshal(body, &obj)
	s, _ := json.MarshalIndent(obj, "", "  ")
	log.Println(string(s))
}
