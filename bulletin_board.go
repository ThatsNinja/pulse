package bulletin_board

import (
	"github.com/lazywei/sns-exp/bulletin_board/messenger"
	"net/http"
)

type BulletinBoard struct {
	msgers map[messenger.Messenger]bool
	port   string
}

func New(port string) *BulletinBoard {
	return &BulletinBoard{
		msgers: make(map[messenger.Messenger]bool),
		port:   port,
	}
}

func (this *BulletinBoard) RegisterBroker(msger messenger.Messenger) {
	this.msgers[msger] = true
}

func (this *BulletinBoard) Run() {

	for msger, _ := range this.msgers {
		http.HandleFunc("/brokers/"+msger.Name(), msger.ServeSSE)
		http.HandleFunc("/sns/"+msger.Name(), msger.ServeSNS)
		msger.Start()
	}

	http.ListenAndServe(this.port, nil)
}
