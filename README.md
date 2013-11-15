Pulse
==============

A [AWS SNS](http://aws.amazon.com/sns/)(endpoint) and SSE(server side) integration library written in GO


##Install

```
go get github.com/polydice/pulse
```

##Features

- Works out of the box, without any configuration
- Auto confirm SNS subscribing
- Auto parse SNS notification
- Built-in http server for SNS request and SSE connection
- Auto send notification to SSE client

##Usage

```go
package main

import (
	"github.com/polydice/pulse"
)

func main() {
	pump := pulse.New(":8000")
	pump.Start(true) // Allow cross domain or not.
}
```

Then subscribe `/subscribe/any_event_name_you_want` to your AWS SNS http endpoint.
Now there is a SSE server on `/publish/any_event_name_you_want`, example usage:

```javascript
var source = new EventSource('http://localhost:8000/publish/any_event_name_you_want');

// Create a callback for when a new message is received.
source.onmessage = function(e) {
  console.log(e.data)
};
```

##TODO

- SNS signature verification
- examples
