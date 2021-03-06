package pulse

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/garyburd/redigo/redis"
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

func (self *Pump) RegisterMessenger(name string, msger *messenger.Messenger) {
	self.msgers[name] = msger
}

func (self *Pump) Start(allowCrossDomain bool) {

	r := mux.NewRouter()
	r.HandleFunc("/subscribe/{event}", func(resp http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		event := vars["event"]

		self.runMsger(event, resp, allowCrossDomain)

		// Done.
		log.Println("Finished HTTP request at ", req.URL.Path)
	})

	r.HandleFunc("/subscribe/{event}/{id:[0-9]}+", func(resp http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		event := vars["event"]
		eventId := vars["id"]
		msgerName := fmt.Sprintf("%v.%v", event, eventId)

		self.runMsger(msgerName, resp, allowCrossDomain)

		// Done.
		log.Println("Finished HTTP request at ", req.URL.Path)
	})

	r.HandleFunc("/publish/{event}", func(resp http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		name := vars["event"]

		var msger *messenger.Messenger
		if self.msgers[name] == nil {
			msger = messenger.New(name)
			self.RegisterMessenger(name, msger)
			log.Println("SNS endpoint: listening /publish/" + msger.Name())
			log.Println("SSE endpoint: listening /subscribe/" + msger.Name())
		} else {
			msger = self.msgers[name]
		}

		n := sns.NewFromRequest(req)
		msger.SendMessage(n.Message)
	})

	log.Println("HTTP server port " + self.port)
	http.ListenAndServe(self.port, r)
}

func (self *Pump) runMsger(msgerName string, resp http.ResponseWriter, allowCrossDomain bool) {
	var msger *messenger.Messenger
	if self.msgers[msgerName] == nil {
		msger = messenger.New(msgerName)
		self.RegisterMessenger(msgerName, msger)
		// TODO: fix display format to subscribe/events/id
		log.Println("SSE endpoint: listening /subscribe/" + msger.Name())
	} else {
		msger = self.msgers[msgerName]
	}

	go self.subscribeRedis(msgerName, msger)

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
	// send self client messages.
	messageChan := make(chan string)

	// Add self client to the map of those that should
	// receive updates
	msger.AddClient(messageChan)

	// Remove self client from the map of attached clients
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
}

func (self *Pump) subscribeRedis(channel string, msger *messenger.Messenger) {

	c, err := redis.Dial("tcp", os.Getenv("REDIS_ADDR"))
	if err != nil {
		panic(err)
	}
	defer c.Close()

	psc := redis.PubSubConn{c}
	psc.Subscribe(channel)
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("PMessage: channel:%s data:%s\n", v.Channel, v.Data)
			msger.SendMessage(string(v.Data))
		case redis.Subscription:
			log.Printf("Subscription: kind:%s channel:%s count:%d\n", v.Kind, v.Channel, v.Count)
			if v.Count == 0 {
				return
			}
		case error:
			log.Printf("error: %v\n", v)
			return
		}
	}
}
