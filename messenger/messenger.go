package messenger

import (
	"fmt"
	"log"
	"net/http"
)

type Messenger interface {
	Start()
	ServeSNS(resp http.ResponseWriter, req *http.Request)
	ServeSSE(resp http.ResponseWriter, req *http.Request)
	Name() string
}

type DefaultMessenger struct {
	// Create a map of clients, the keys of the map are the channels
	// over which we can push messages to attached clients. (The values
	// are just booleans and are meaningless.)
	//
	clients map[chan string]bool

	// Channel into which new clients can be pushed
	//
	newClients chan chan string

	// Channel into which disconnected clients should be pushed
	//
	defunctClients chan chan string

	// Channel into which messages are pushed to be broadcast out
	// to attahed clients.
	//
	messages chan string

	name string
}

func New() Messenger {
	return &DefaultMessenger{
		make(map[chan string]bool),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string),
		"default_messenger",
	}
}

func (this *DefaultMessenger) Start() {

	go func() {
		for {
			select {
			case s := <-this.newClients:

				// There is a new client attached and we
				// want to start sending them messages.
				this.clients[s] = true
				log.Println("Added new client")

			case s := <-this.defunctClients:

				// A client has dettached and we want to
				// stop sending them messages.
				delete(this.clients, s)
				log.Println("Removed client")

			case msg := <-this.messages:

				// There is a new message to send.  For each
				// attached client, push the new message
				// into the client's message channel.
				for s, _ := range this.clients {
					s <- msg
				}
				log.Printf("Broadcast message to %d clients", len(this.clients))
			}
		}
	}()
}

func (this *DefaultMessenger) ServeSSE(resp http.ResponseWriter, req *http.Request) {
	// Make sure that the writer supports flushing.
	//
	f, ok := resp.(http.Flusher)
	if !ok {
		http.Error(resp, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	c, ok := resp.(http.CloseNotifier)
	if !ok {
		http.Error(resp, "close notification unsupported",
			http.StatusInternalServerError)
		return
	}
	closer := c.CloseNotify()

	// Create a new channel, over which the broker can
	// send this client messages.
	messageChan := make(chan string)

	// Add this client to the map of those that should
	// receive updates
	this.newClients <- messageChan

	// Remove this client from the map of attached clients
	// when `EventHandler` exits.
	defer func() {
		this.defunctClients <- messageChan
	}()

	// Set the headers related to event streaming.
	resp.Header().Set("Content-Type", "text/event-stream")
	resp.Header().Set("Cache-Control", "no-cache")
	resp.Header().Set("Connection", "keep-alive")

	// Use the CloseNotifier interface
	// https://code.google.com/p/go/source/detail?name=3292433291b2
	//
	// NOTE: we could loop endlessly; however, then you
	// could not easily detect clients that dettach and the
	// server would continue to send them messages long after
	// they're gone due to the "keep-alive" header.  One of
	// the nifty aspects of SSE is that clients automatically
	// reconnect when they lose their connection.
	//
	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(resp, "data: Message: %s\n\n", msg)
			f.Flush()
		case <-closer:
			log.Println("Closing connection")
			return
		}
	}

	// Done.
	log.Println("Finished HTTP request at ", req.URL.Path)
}

func (this *DefaultMessenger) ServeSNS(resp http.ResponseWriter, req *http.Request) {

	test_response := `{
"Type" : "Notification",
"MessageId" : "da41e39f-ea4d-435a-b922-c6aae3915ebe",
"TopicArn" : "arn:aws:sns:us-east-1:123456789012:MyTopic",
"Subject" : "test",
"Message" : "test message",
"Timestamp" : "2012-04-25T21:49:25.719Z",
"SignatureVersion" : "1",
"Signature" : "EXAMPLElDMXvB8r9R83tGoNn0ecwd5UjllzsvSvbItzfaMpN2nk5HVSw7XnOn/49IkxDKz8YrlH2qJXj2iZB0Zo2O71c4qQk1fMUDi3LGpij7RCW7AW9vYYsSqIKRnFS94ilu7NFhUzLiieYr4BKHpdTmdD6c0esKEYBpabxDSc=",
"SigningCertURL" : "https://sns.us-east-1.amazonaws.com/SimpleNotificationService-f3ecfb7224c7233fe7bb5f59f96de52f.pem",
"UnsubscribeURL" : "https://sns.us-east-1.amazonaws.com/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-1:123456789012:MyTopic:2bcfbf39-05c3-41de-beaa-fcfcc21c8f55"
} `

	this.messages <- test_response

}

func (this *DefaultMessenger) Name() string {
	return this.name
}
