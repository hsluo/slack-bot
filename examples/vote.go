// +build appengine

package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"sync"

	"golang.org/x/net/context"

	"github.com/hsluo/slack-bot"

	"google.golang.org/appengine"
	l "google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type VoteResult map[string]StringSet

func (vr VoteResult) hasVoted(user string) bool {
	for _, v := range vr {
		if v.contains(user) {
			return true
		}
	}
	return false
}

func (vr VoteResult) String() string {
	options := make([]string, 0, len(vr))
	for k, _ := range vr {
		options = append(options, k)
	}
	sort.Strings(options)
	var buf bytes.Buffer
	buf.WriteString("Result:\n")
	for i := range options {
		buf.WriteString(fmt.Sprintf("%s: %v\n", options[i], vr[options[i]].toSlice()))
	}
	return buf.String()
}

var (
	votes      = newStringSet()
	voteResult = VoteResult{}
	m          sync.Mutex
)

func vote(rw http.ResponseWriter, req *http.Request) {
	var (
		c         = appengine.NewContext(req)
		channelId = req.PostFormValue("channel_id")
		text      = req.PostFormValue("text")
		userId    = req.PostFormValue("user_id")
	)
	l.Infof(c, "%v", req.PostForm)
	m.Lock()
	if text == "start" {
		if startVote(channelId) {
			err := annouce(c, channelId, fmt.Sprintf("<@%s> just starts a vote!", userId))
			if err != nil {
				fmt.Fprintln(rw, err)
			} else {
				fmt.Fprintln(rw, "vote starts now")
			}
		} else {
			fmt.Fprintln(rw, "we're voting")
		}
	} else if text == "done" {
		annouce(c, channelId, voteResult.String())
		fmt.Fprintln(rw, "vote ends")
		delete(votes.set, channelId)
	} else if votes.contains(channelId) {
		userName := req.PostFormValue("user_name")
		if voters, ok := voteResult[text]; ok {
			if !voteResult.hasVoted(userName) {
				voters.add(userName)
			}
		} else {
			voters = newStringSet()
			voters.add(userName)
			voteResult[text] = voters
		}
		fmt.Fprintln(rw, voteResult)
	} else {
		fmt.Fprintln(rw, "Not voting")
	}
	m.Unlock()
}

func startVote(channelId string) bool {
	if votes.contains(channelId) {
		return false
	} else {
		votes.add(channelId)
		return true
	}
}

func annouce(c context.Context, channelId, text string) error {
	client := urlfetch.Client(c)
	err := bot.WithClient(client).ChatPostMessage(url.Values{
		"channel": {channelId},
		"text":    {text},
	})
	return err
}

// count active users in channel with channels.info then users.getPresence
// very slow due to network
func activeUsersInChannel(c context.Context, channelId string) (users []string, err error) {
	bot := bot.WithClient(urlfetch.Client(c))
	members, err := bot.ChannelsInfo(channelId)
	l.Infof(c, "check %v", members)
	active := make(chan string, len(members))
	var wg sync.WaitGroup
	for i := range members {
		wg.Add(1)
		go func(user string, active chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			l.Infof(c, "begin "+user)
			if p, err := bot.UsersGetPresence(user); err != nil {
				l.Errorf(c, "%s", err)
				return
			} else if p == "active" {
				active <- user
			}
			l.Infof(c, "done "+user)
		}(members[i], active, &wg)
	}
	wg.Wait()
	l.Infof(c, "done wait")
	close(active)
	users = make([]string, len(members))
	for user := range active {
		users = append(users, user)
	}
	return
}

func init() {
	log.Println("vote init")
	http.HandleFunc("/cmds/vote",
		slack.ValidateCommand(http.HandlerFunc(vote), credentials.Commands))
}
