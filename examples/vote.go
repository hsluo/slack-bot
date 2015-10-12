package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/hsluo/slack-bot"

	"appengine"
	"appengine/urlfetch"
)

type Vote struct {
	UserName string
	Option   string
}

type StringSet struct {
	set map[string]struct{}
}

func newStringSet() StringSet {
	return StringSet{set: make(map[string]struct{})}
}

func (v *StringSet) add(key string) {
	v.set[key] = struct{}{}
}

func (v StringSet) contains(key string) bool {
	_, ok := v.set[key]
	return ok
}

type VoteResult map[string]StringSet

func (vr VoteResult) hasVoted(user string) bool {
	for _, v := range vr {
		if v.contains(user) {
			return true
		}
	}
	return false
}

var (
	votes      = newStringSet()
	voteResult = VoteResult{}
	m          sync.Mutex
)

func vote(rw http.ResponseWriter, req *http.Request) {
	var (
		c         = appengine.NewContext(req)
		client    = urlfetch.Client(c)
		channelId = req.PostFormValue("channel_id")
		text      = req.PostFormValue("text")
		userId    = req.PostFormValue("user_id")
	)
	m.Lock()
	if text == "start" {
		if startVote(channelId) {
			err := bot.WithClient(client).ChatPostMessage(url.Values{
				"channel": {channelId},
				"text":    {fmt.Sprintf("<@%s> just starts a vote!", userId)},
			})
			if err != nil {
				fmt.Fprintln(rw, err)
			} else {
				fmt.Fprintln(rw, votes)
			}
		} else {
			fmt.Fprintln(rw, "we're voting")
		}
	} else if text == "done" {
		fmt.Fprintln(rw, voteResult)
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
		fmt.Fprintln(rw, "vote recorded")
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

// count active users in channel with channels.info then users.getPresence
// very slow due to network
func activeUsersInChannel(c appengine.Context, channelId string) (users []string, err error) {
	bot := bot.WithClient(urlfetch.Client(c))
	members, err := bot.ChannelsInfo(channelId)
	c.Infof("check %v", members)
	active := make(chan string, len(members))
	var wg sync.WaitGroup
	for i := range members {
		wg.Add(1)
		go func(user string, active chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			c.Infof("begin " + user)
			if p, err := bot.UsersGetPresence(user); err != nil {
				c.Errorf("%s", err)
				return
			} else if p == "active" {
				active <- user
			}
			c.Infof("done " + user)
		}(members[i], active, &wg)
	}
	wg.Wait()
	c.Infof("done wait")
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
