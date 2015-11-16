package slack

import "net/url"

type Channel struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	IsMember bool     `json:"is_member"`
	Members  []string `json:"members"`
}

type ChannelsListResp struct {
	ok       bool      `json:"ok"`
	channels []Channel `json:"channels"`
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

func (b Bot) ChannelsList() ([]Channel, error) {
	resp, err := b.PostForm("channels.list", url.Values{})
	if err != nil {
		return nil, err
	}
	array := resp["channels"].([]interface{})
	channels := make([]Channel, len(array))
	for i := range array {
		el := array[i].(map[string]interface{})
		channel := Channel{
			Id:       el["id"].(string),
			Name:     el["name"].(string),
			IsMember: el["is_member"].(bool),
		}
		channels = append(channels, channel)
	}
	return channels, nil
}
