package pulse

import (
	"fmt"
	"github.com/polydice/pulse/messenger"
	"github.com/polydice/pulse/sns"
	"log"
	"net/http"
)

type Pump struct {
	msgers map[messenger.Messenger]bool
	port   string
}

func New(port string) *Pump {
	return &Pump{
		msgers: make(map[messenger.Messenger]bool),
		port:   port,
	}
}

func (this *Pump) RegisterMessenger(msger messenger.Messenger) {
	this.msgers[msger] = true
}

func (this *Pump) Start(allowCrossDomain bool) {

	for msger, _ := range this.msgers {

		log.Println("SSE server: listening /messengers/" + msger.Name())
		http.HandleFunc("/messengers/"+msger.Name(),

			func(resp http.ResponseWriter, req *http.Request) {
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
				msger.AddClient(messageChan)

				// Remove this client from the map of attached clients
				// when `EventHandler` exits.
				defer func() {
					msger.RemoveClient(messageChan)
				}()

				// Set the headers related to event streaming.
				resp.Header().Set("Content-Type", "text/event-stream")
				resp.Header().Set("Cache-Control", "no-cache")
				resp.Header().Set("Connection", "keep-alive")
				if allowCrossDomain {
					resp.Header().Set("Access-Control-Allow-Origin", "*")
				}

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
						fmt.Fprintf(resp, "data: %s\n\n", msg)
						f.Flush()
					case <-closer:
						log.Println("Closing connection")
						return
					}
				}

				// Done.
				log.Println("Finished HTTP request at ", req.URL.Path)
			})

		log.Println("SNS endpoint: listening /sns/" + msger.Name())
		http.HandleFunc("/sns/"+msger.Name(),

			func(resp http.ResponseWriter, req *http.Request) {

				n := sns.NewFromRequest(req)

				msger.SendMessage(n.Message)

			})
		msger.Start()
	}

	log.Println("HTTP server port " + this.port)
	http.ListenAndServe(this.port, nil)
}
