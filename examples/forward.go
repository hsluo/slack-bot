package main

import (
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
)

const KEY = "fwdtable"

func handleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	channelId := r.PostFormValue("channel_id")
	text := r.PostFormValue("text")
	splits := strings.SplitN(text, " ", 2)
	chanName, query := splits[0], splits[1]

	client := urlfetch.Client(ctx)
	channels, err := bot.WithClient(client).ChannelsList()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	for i := range channels {
		if chanName == channels[i].Name {
			fwdtable := make(map[string][]string)
			if _, err := memcache.JSON.Get(ctx, KEY, &fwdtable); err == memcache.ErrCacheMiss {
				if _, ok := fwdtable[channelId]; !ok {
					fwdtable[channelId] = []string{channels[i].Id, query}
					memcache.JSON.Set(ctx, &memcache.Item{
						Key:    KEY,
						Object: fwdtable,
					})
					fmt.Fprintln(w, "added")
				}
			} else if err != nil {
				log.Errorf(ctx, "memcache get error: %v", err)
			} else {
				fwdtable[channelId] = []string{channels[i].Id, query}
				if err := memcache.JSON.Set(ctx, &memcache.Item{
					Key:    KEY,
					Object: fwdtable,
				}); err != nil {
					log.Errorf(ctx, "memcache set error: %v", err)
				} else {
					fmt.Fprintln(w, "updated")
				}
			}
			return
		}
	}
}
