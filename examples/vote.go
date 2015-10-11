package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/hsluo/slack-bot"

	"appengine"
	"appengine/urlfetch"
)

type Vote struct {
	UserName string
	Option   string
}

type Voters map[string]struct{}

func (v Voters) add(user string) {
	v[user] = struct{}{}
}

func (v Voters) contains(user string) bool {
	_, ok := v[user]
	return ok
}

type VoteResult map[string]Voters

func (vr VoteResult) hasVoted(user string) bool {
	for _, v := range vr {
		if v.contains(user) {
			return true
		}
	}
	return false
}

var (
	votes      = make(map[string]bool)
	voteResult = VoteResult{}
	m          sync.Mutex
)

func vote(rw http.ResponseWriter, req *http.Request) {
	if !slack.ValidateCommand(req) {
		return
	}
	if bot.Token == "" {
		warmUp(rw, req)
	}
	c := appengine.NewContext(req)
	//client := urlfetch.Client(c)
	channel := req.PostFormValue("channel_name")
	text := req.PostFormValue("text")
	m.Lock()
	if text == "start" {
		votes[channel] = true
		members, err := activeUsersInChannel(c, req.PostFormValue("channel_id"))
		if err != nil {
			fmt.Fprintln(rw, err)
		} else {
			fmt.Fprintln(rw, members)
		}
	} else if votes[channel] {
		userName := req.PostFormValue("user_name")
		if voters, ok := voteResult[text]; ok {
			if !voteResult.hasVoted(userName) {
				voters.add(userName)
			}
		} else {
			voters = Voters{}
			voters.add(userName)
			voteResult[text] = voters
		}
		fmt.Fprintln(rw, "vote recorded")
	} else {
		fmt.Fprintln(rw, "Not voting")
	}
	m.Unlock()
}

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
