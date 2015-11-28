[![Build Status](https://travis-ci.org/hsluo/slack-bot.svg?branch=master)](https://travis-ci.org/hsluo/slack-bot)
==

A simple Slack bot using web API and RTM API. Compatible with [Google App Engine](https://cloud.google.com/appengine/docs).

### APIs used
- [chat.postMessage](https://api.slack.com/methods/chat.postMessage)
- [rtm.start](https://api.slack.com/methods/rtm.start)

### Use cases implemented
- Daily stand up meeting alert @ 10:00AM
- Loggly alert and search (See https://github.com/hsluo/slack-loggly-alert)
- Outgoing webhooks
- RTM (See https://github.com/hsluo/slack-message-forward)
- Slash command
  - Get a commit message from [whatthecommit.com](http://whatthecommit.com/)
  - Vote
