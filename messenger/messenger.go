package messenger

import (
	"log"
)

type Messenger struct {
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

func New(name string) *Messenger {
  m := &Messenger{
		make(map[chan string]bool),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string),
		name,
	}
	m.start()
	return m
}

func (this *Messenger) Name() string {
	return this.name
}

func (this *Messenger) AddClient(messageChan chan string) {
	this.newClients <- messageChan
}

func (this *Messenger) RemoveClient(messageChan chan string) {
	this.defunctClients <- messageChan
}

func (this *Messenger) SendMessage(msg string) {
	this.messages <- msg
}

func (this *Messenger) start() {

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
