Pulse
==============

A [AWS SNS](http://aws.amazon.com/sns/)(endpoint) and SSE(server side) integration library written in GO


##Install

```
go get github.com/polydice/pulse
```

##Features

- Auto confirm SNS subscribing
- Auto parse SNS notification
- Built-in http server for SNS request and SSE connection
- Auto send notification to SSE client

##Usage

```go
package main

import (
	"github.com/polydice/pulse"
	"github.com/polydice/pulse/messenger"
)

func main() {

	pump := pulse.New(":8000")

	latestTopicMsger := messenger.New("latest_topic")
	pump.RegisterMessenger(latestTopicMsger)

	latestCommentMsger := messenger.New("latest_comment")
	pump.RegisterMessenger(latestCommentMsger)

	pump.Start()
}
```

##TODO

- SNS signature verification
- examples

