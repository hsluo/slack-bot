package slack

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/websocket"
)

const (
	API_BASE           = "https://slack.com/api/"
	ChatPostMessageApi = API_BASE + "chat.postMessage"
	ChannelsInfoApi    = API_BASE + "channels.info"
)

type Credentials struct {
	HookToken   string
	BotToken    string
	SlackbotUrl string
	Commands    map[string]string
}

type Message struct {
	Type    string     `json:"type"`
	SubType string     `json:"subtype"`
	Channel string     `json:"channel"`
	User    string     `json:"user"`
	Text    string     `json:"text"`
	Ts      string     `json:"ts"`
	File    FileObject `json:"file"`
}

type FileObject struct {
	Mimetype   string `json:"mimetype"`
	Filetype   string `json:"filetype"`
	PrettyType string `json:"pretty_type"`
}

type Attachment struct {
	Fallback   string   `json:"fallback"`
	Color      string   `json:"color"`
	Pretext    string   `json:"pretext"`
	AuthorName string   `json:"author_name"`
	Title      string   `json:"title"`
	TitleLink  string   `json:"title_link"`
	Text       string   `json:"text"`
	Fields     []Field  `json:"fields"`
	MrkdwnIn   []string `json:"mrkdwn_in"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Channel struct {
	Members []string `json:"members"`
}

type Bot struct {
	Token, UserId, User string
	Client              *http.Client
}

var Creds Credentials

func NewBot(c *http.Client, token string) (b Bot, err error) {
	resp, err := c.PostForm("https://slack.com/api/auth.test", url.Values{"token": {token}})
	if err != nil {
		return
	}
	respAuthTest, err := asJson(resp)
	if err != nil {
		return
	} else if !respAuthTest["ok"].(bool) {
		err = errors.New(respAuthTest["error"].(string))
		return
	} else {
		b = Bot{
			Token:  token,
			UserId: respAuthTest["user_id"].(string),
			User:   respAuthTest["user"].(string),
		}
	}
	return
}

func (b Bot) WithClient(c *http.Client) Bot {
	b.Client = c
	return b
}

func (b Bot) PostForm(url string, data url.Values) (respJson map[string]interface{}, err error) {
	data.Add("token", b.Token)
	if _, ok := data["as_user"]; !ok {
		data.Add("as_user", "true")
	}

	if b.Client == nil {
		b.Client = http.DefaultClient
	}
	resp, err := b.Client.PostForm(url, data)
	if err != nil {
		return
	}
	respJson, err = asJson(resp)
	if err != nil {
		return
	}
	if !respJson["ok"].(bool) {
		err = errors.New(respJson["error"].(string))
	}
	return
}

func (b Bot) ChatPostMessage(data url.Values) (err error) {
	_, err = b.PostForm("https://slack.com/api/chat.postMessage", data)
	return
}

func (b Bot) ChannelsInfo(channelId string) ([]string, error) {
	resp, err := b.PostForm(ChannelsInfoApi, url.Values{"channel": {channelId}})
	if err != nil {
		return nil, err
	}
	members := resp["channel"].(map[string]interface{})["members"].([]interface{})
	ret := make([]string, len(members))
	for i := range members {
		ret[i] = members[i].(string)
	}
	return ret, nil
}

func (b Bot) UsersGetPresence(user string) (presence string, err error) {
	resp, err := b.PostForm("https://slack.com/api/users.getPresence", url.Values{"user": {user}})
	if err != nil {
		return
	} else {
		presence = resp["presence"].(string)
		return
	}
}

// for presence now
func (b Bot) UsersList(presence string) (present []string, err error) {
	resp, err := b.PostForm("https://slack.com/api/users.list", url.Values{"presence": {presence}})
	if err != nil {
		return
	}
	members := resp["members"].([]interface{})
	present = make([]string, len(members))
	for i := range members {
		m := members[i].(map[string]interface{})
		p := m["presence"].(string)
		if p == "active" {
			present = append(present, m["id"].(string))
		}
	}
	return
}

func asJson(resp *http.Response) (map[string]interface{}, error) {
	defer resp.Body.Close()
	m := make(map[string]interface{})
	dec := json.NewDecoder(resp.Body)
	for {
		if err := dec.Decode(&m); err == io.EOF {
			return m, nil
		} else if err != nil {
			return nil, err
		}
	}
}

// Calls rtm.start API, return websocket url and bot id
func RtmStart(token string) (wsurl string, id string) {
	resp, err := http.PostForm("https://slack.com/api/rtm.start", url.Values{"token": {token}})
	if err != nil {
		log.Fatal(err)
	}
	respRtmStart, err := asJson(resp)
	if err != nil {
		log.Fatal(err)
	}
	wsurl = respRtmStart["url"].(string)
	id = respRtmStart["self"].(map[string]interface{})["id"].(string)
	return
}

func RtmReceive(ws *websocket.Conn, incoming chan<- Message) {
	for {
		var m Message
		if err := websocket.JSON.Receive(ws, &m); err != nil {
			log.Println(err)
		} else {
			log.Printf("read %v", m)
			incoming <- m
		}
	}
}

func RtmSend(ws *websocket.Conn, outgoing <-chan Message) {
	for m := range outgoing {
		m.Ts = fmt.Sprintf("%f", float64(time.Now().UnixNano())/1000000000.0)
		log.Printf("send %v", m)
		if err := websocket.JSON.Send(ws, m); err != nil {
			log.Println(err)
		}
	}
}

func LoadCredentials(filename string) (err error) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = json.Unmarshal(f, &Creds)
	return
}

func ValidateCommand(req *http.Request) bool {
	return Creds.Commands[req.PostFormValue("command")] == req.PostFormValue("token")
}
