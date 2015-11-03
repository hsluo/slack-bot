[![Build Status](https://travis-ci.org/hsluo/slack-bot.svg?branch=master)](https://travis-ci.org/hsluo/slack-bot)
==

A simple Slack bot using web API and RTM API. Compatible with [Google App Engine](https://cloud.google.com/appengine/docs).

### APIs used
- [chat.postMessage](https://api.slack.com/methods/chat.postMessage)
- [rtm.start](https://api.slack.com/methods/rtm.start)

### Use cases implemented
- Daily stand up meeting alert @ 10:00AM
- Loggly HTTP alert to Slack (Currently, They support [only static alert messages](https://www.loggly.com/docs/slack-alerts/) to Slack Chat). 
  - Example of the alert:

    ![alert](http://i.imgur.com/G45W1M6.png)
- Loggly search API
- Outgoing webhooks
- Mentioning bot in RTM
- Slash command
  - Get a commit message from [whatthecommit.com](http://whatthecommit.com/)
  - Vote
