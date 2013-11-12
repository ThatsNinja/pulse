package sns

import (
	"crypto/x509"
	//"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
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

	log.Println("SignatureVersion:", n.SignatureVersion)
	log.Println("Signature:", n.Signature)
	log.Println("SigningCertURL:", n.SigningCertURL)
	log.Println("Type:", n.Type)

	n.verifySignature()

	return &n
}

func (self *Notification) verifySignature() bool {

	switch self.SignatureVersion {
	case "1":
		return self.v1Verification()
	default:
		log.Println("Unexpected signature version. Unable to verify signature.")
		return false
	}
}

func (self *Notification) v1Verification() bool {

	signString := fmt.Sprintf(`Message
%v
MessageId
%v`, self.Message, self.MessageId)

	if self.Subject != "" {
		signString = signString + fmt.Sprintf(`
Subject
%v`, self.Subject)
	}

	signString = signString + fmt.Sprintf(`
Timestamp
%v
TopicArn
%v
Type
%v`, self.Timestamp, self.TopicArn, self.Type)

	//signed, err := base64.StdEncoding.DecodeString(self.Signature)
	signed := []byte(self.Signature)

	//if err != nil {
	//	log.Println("Got error when decoding signature (w/ base64):", err)
	//	return false
	//}

	resp, _ := http.Get(self.SigningCertURL)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	p, _ := pem.Decode(body)

	cert, err := x509.ParseCertificate(p.Bytes)

	if err != nil {
		fmt.Println(err)
		return false
	}

	log.Println(signString)

	if err := cert.CheckSignature(x509.SHA1WithRSA, signed, []byte(signString)); err != nil {
		log.Println(err)
		return false
	}
	return false
}
