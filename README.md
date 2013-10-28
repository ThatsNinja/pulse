Bulletin Board
==============

A [AWS SNS](http://aws.amazon.com/sns/)(endpoint) and SSE(server side) integration library written in GO


##Install

```
go get github.com/lazywei/bulletin_board
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
	"github.com/lazywei/bulletin_board"
	"github.com/lazywei/bulletin_board/messenger"
)

func main() {

	board := bulletin_board.New(":8000")

	latestTopicMsger := messenger.New("latest_topic")
	board.RegisterMessenger(latestTopicMsger)

	latestCommentMsger := messenger.New("latest_comment")
	board.RegisterMessenger(latestCommentMsger)

	board.Run()
}
```

##Contact

Bert Chang

- Twitter https://twitter.com/jrweizhang

##TODO

- SNS signature verification
- examples

