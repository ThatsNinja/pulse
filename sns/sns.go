package sns

import (
	"encoding/json"
	"log"
	"net/http"
)

type Notification struct {
	Message          string
	MessageId        string
	Signature        string
	SignatureVersion string
	SigningCertURL   string
	SubscribeURL     string
	Subject          string
	Timestamp        string
	TopicArn         string
	Type             string
	UnsubscribeURL   string
}

func NewFromRequest(req *http.Request) *Notification {
	var n Notification
	dec := json.NewDecoder(req.Body)
	err := dec.Decode(&n)

	if err != nil {
		log.Println("Get error when decode sns request json:", err)
	}

	if s := n.SubscribeURL; len(s) != 0 {
		log.Println("SubscribeURL detected: ", s)

		if _, err := http.Get(s); err != nil {
			log.Println("Get error when subscribe:", err)
		}
	}

	return &n
}
