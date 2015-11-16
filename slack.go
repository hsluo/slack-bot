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

	"golang.org/x/net/websocket"
)

const (
	API_BASE = "https://slack.com/api/"
)

type Credentials struct {
	HookToken   string
	Bot         Bot
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
	Token  string `json:"token"`
	UserId string `json:"user_id"`
	User   string `json:"user"`
	Client *http.Client
}

func (b Bot) WithClient(c *http.Client) Bot {
	b.Client = c
	return b
}

func (b Bot) PostForm(endpoint string, data url.Values) (respJson map[string]interface{}, err error) {
	data.Add("token", b.Token)
	if _, ok := data["as_user"]; !ok {
		data.Add("as_user", "true")
	}

	if b.Client == nil {
		b.Client = http.DefaultClient
	}
	resp, err := b.Client.PostForm(API_BASE+endpoint, data)
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

func (b Bot) ChatPostMessage(data url.Values) (err error) {
	_, err = b.PostForm("chat.postMessage", data)
	return
}

func (b Bot) ChannelsInfo(channelId string) ([]string, error) {
	resp, err := b.PostForm("channels.info", url.Values{"channel": {channelId}})
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
	resp, err := b.PostForm("users.getPresence", url.Values{"user": {user}})
	if err != nil {
		return
	} else {
		presence = resp["presence"].(string)
		return
	}
}

// for presence now
func (b Bot) UsersList(presence string) (present []string, err error) {
	resp, err := b.PostForm("users.list", url.Values{"presence": {presence}})
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

// Calls rtm.start API, return websocket url and bot id
func RtmStart(token string) (wsurl string, id string) {
	resp, err := http.PostForm("rtm.start", url.Values{"token": {token}})
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

func RtmReceive(ws *websocket.Conn) (m Message, err error) {
	err = websocket.JSON.Receive(ws, &m)
	return
}

func RtmSend(ws *websocket.Conn, m Message) error {
	return websocket.JSON.Send(ws, m)
}

func LoadCredentials(filename string) (credentials Credentials, err error) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = json.Unmarshal(f, &credentials)
	return
}

func ValidateCommand(handler http.Handler, commands map[string]string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if commands[r.PostFormValue("command")] == r.PostFormValue("token") {
			handler.ServeHTTP(w, r)
		} else {
			w.WriteHeader(404)
			fmt.Fprintln(w, "command not found")
		}
	})
}
