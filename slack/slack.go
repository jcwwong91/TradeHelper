package slack

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
)

const(
	apiEndpoint = "https://slack.com/api/"
	messageURL = apiEndpoint+"chat.postMessage"
)

// Slack us a wrapper object to read/slack messages
type Slack struct {
	Channel string
	Token string
}

func LoadSlackConfig(filename string) (*Slack, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	slack := Slack{}
	err = decoder.Decode(&slack)
	if err != nil {
		return nil, err
	}
	return &slack, nil
}

// SendMessage will send a slack message to a specified channel
func (s *Slack) SendMessage(channel string, message string) {
	v := url.Values{}
	v.Add("token", s.Token)
	v.Add("channel", channel)
	v.Add("text", message)
	_, err := http.Get(messageURL+"?"+v.Encode())
	if err != nil {
		log.Println(err)
	}
}
