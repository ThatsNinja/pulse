package pulse

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/polydice/pulse/messenger"
	"github.com/polydice/pulse/sns"
)

type Pump struct {
	msgers map[string]*messenger.Messenger
	port   string
}

func New(port string) *Pump {
	return &Pump{
		msgers: make(map[string]*messenger.Messenger),
		port:   port,
	}
}

func (this *Pump) RegisterMessenger(name string, msger *messenger.Messenger) {
	this.msgers[name] = msger
}

func (this *Pump) Start(allowCrossDomain bool) {

	r := mux.NewRouter()
	r.HandleFunc("/subscribe/{event}", func(resp http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		name := vars["event"]

		var msger *messenger.Messenger
		if this.msgers[name] == nil {
			http.Error(resp, "You need to subscribe this endpoint to AWS SNS.", http.StatusNotFound)
			return
		} else {
			msger = this.msgers[name]
		}

		// Make sure that the writer supports flushing.
		//
		f, ok := resp.(http.Flusher)
		if !ok {
			http.Error(resp, "Streaming unsupported", http.StatusInternalServerError)
			return
		}
		c, ok := resp.(http.CloseNotifier)
		if !ok {
			http.Error(resp, "Close notification unsupported", http.StatusInternalServerError)
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

	r.HandleFunc("/publish/{event}", func(resp http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		name := vars["event"]

		var msger *messenger.Messenger
		if this.msgers[name] == nil {
			msger = messenger.New(name)
			this.RegisterMessenger(name, msger)
			log.Println("SNS endpoint: listening /publish/" + msger.Name())
			log.Println("SSE endpoint: listening /subscribe/" + msger.Name())
		} else {
			msger = this.msgers[name]
		}

		n := sns.NewFromRequest(req)
		msger.SendMessage(n.Message)
	})

	log.Println("HTTP server port " + this.port)
	http.ListenAndServe(this.port, r)
}
