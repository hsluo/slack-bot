[![Build Status](https://travis-ci.org/hsluo/slack-bot.svg?branch=master)](https://travis-ci.org/hsluo/slack-bot)
==

A simple Slack bot using web API and RTM API. Compatible with [Google App Engine](https://cloud.google.com/appengine/docs).

### APIs used
- [chat.postMessage](https://api.slack.com/methods/chat.postMessage)
- [rtm.start](https://api.slack.com/methods/rtm.start)

### Implemented use cases
- Daily stand up meeting alert @ 10:00AM
- Loggly HTTP alert forwarding (Currently, They support [only static alert messages](https://www.loggly.com/docs/slack-alerts/) to Slack Chat)
- Outgoing webhooks
- Mentioning bot in RTM
