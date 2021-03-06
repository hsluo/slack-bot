package slack

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
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
	Type        string       `json:"type"`
	SubType     string       `json:"subtype"`
	Channel     string       `json:"channel"`
	User        string       `json:"user"`
	Text        string       `json:"text"`
	Ts          string       `json:"ts"`
	File        FileObject   `json:"file"`
	Attachments []Attachment `json:"attachments"`
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

type Bot struct {
	Token  string `json:"token"`
	UserId string `json:"user_id"`
	User   string `json:"user"`
	Client *http.Client
}

func NewBot(token string) *Bot {
	return &Bot{Token: token}
}

func (b *Bot) WithClient(c *http.Client) *Bot {
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

func (b *Bot) Get(endpoint string, params url.Values) (*http.Response, error) {
	params.Set("token", b.Token)
	if b.Client == nil {
		b.Client = http.DefaultClient
	}
	return b.Client.Get(API_BASE + endpoint + "?" + params.Encode())
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
	present = make([]string, 0)
	for i := range members {
		m := members[i].(map[string]interface{})
		if p, ok := m["presence"]; ok {
			if p == "active" {
				present = append(present, m["id"].(string))
			}
		}
	}
	return
}

// Calls rtm.start API, return websocket url and bot id
func RtmStart(token string) (wsurl string, err error) {
	resp, err := http.PostForm(API_BASE+"rtm.start", url.Values{"token": {token}})
	if err != nil {
		return
	}
	respRtmStart, err := asJson(resp)
	if err != nil {
		return
	}
	wsurl = respRtmStart["url"].(string)
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
		command := r.PostFormValue("command")
		token := r.PostFormValue("token")
		if t, ok := commands[command]; ok && t == token {
			handler.ServeHTTP(w, r)
		} else {
			w.WriteHeader(400)
		}
	})
}
