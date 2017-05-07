package shakespearebot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

var (
	// ErrEmptyToken user has not provided a API token
	ErrEmptyToken = errors.New("token cannot not be empty")
	// ErrNotConnected not connected to the slack messaging service.
	ErrNotConnected = errors.New("not connected to slack api")
)

const rtmURL = "https://slack.com/api/rtm.start"
const apiURL = "https://api.slack.com/"

// Bot represents and single bot instance
type Bot struct {
	Token   string
	ID      string
	wsConn  *websocket.Conn
	running bool
}

// message sent and received to and from slack
type message struct {
	ID      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// rtmStartResponse fields to extract from real time messaging API response
type rtmStartResponse struct {
	Ok    bool         `json:"ok"`
	Error string       `json:"error"`
	URL   string       `json:"url"`
	Self  selfResponse `json:"self"`
}

// selfResponse fields to extract from the real time messaging API response
type selfResponse struct {
	ID string `json:"id"`
}

// NewBot creates a new *Bot instance
func NewBot(token string) (*Bot, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}
	return &Bot{Token: token}, nil
}

// startHelper helper which connects to the slack api. Returns websocket
// connection URL and slack assigned ID.
func (b *Bot) startHelper() (string, string, error) {
	url := fmt.Sprintf("%s?token=%s", rtmURL, b.Token)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}
	req.Close = true

	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("api request failed with code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var data rtmStartResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", "", err
	}

	if !data.Ok {
		return "", "", fmt.Errorf("slack error: %s", data.Error)
	}

	return data.URL, data.Self.ID, nil
}

// Start connect
func (b *Bot) Start() error {
	wsURL, id, err := b.startHelper()
	if err != nil {
		return err
	}
	wsConn, err := websocket.Dial(wsURL, "", apiURL)
	if err != nil {
		return err
	}
	b.wsConn = wsConn
	b.ID = id
	b.running = true
	return nil
}

// Stop the bot from processing and responding to events
func (b *Bot) Stop() {
	if err := b.wsConn.Close(); err != nil {
		log.Println(err)
	}
	b.running = false
}

// ListenAndRespond listen, processes and responds to messages
func (b *Bot) ListenAndRespond() error {
	if b.wsConn == nil {
		return ErrNotConnected
	}
	botName := "<@" + b.ID + ">"
	botPrompt := "insult"
	for b.running {
		var m message
		err := websocket.JSON.Receive(b.wsConn, &m)
		if err != nil {
			log.Println(err)
			continue
		}
		// see if we are mentioned: "@shakespearebot insult"
		// {id:0 type:message channel:C09C94QD8 text:<@U53T8CCJK> BBB}
		if m.Type == "message" {
			msg := strings.Split(m.Text, " ")
			if len(msg) == 2 {
				if msg[0] == botName && msg[1] == botPrompt {
					websocket.JSON.Send(b.wsConn, message{
						ID:      m.ID,
						Type:    "message",
						Channel: m.Channel,
						Text:    Insult(),
					})
				}
			}
		}
	}
	return nil
}
