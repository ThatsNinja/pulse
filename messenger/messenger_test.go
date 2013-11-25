package messenger

import (
	"testing"
	. "launchpad.net/gocheck"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&S{})

type S struct {
	msger            *Messenger
	msgerWithClients *Messenger
	client           chan string
}

func (self *S) SetUpTest(c *C) {
	self.client = make(chan string)
	self.msgerWithClients = New("foo")
	self.msgerWithClients.AddClient(self.client)
}

func (self *S) TestSendMessage(c *C) {
	self.msgerWithClients.SendMessage("hello")
	c.Check(<-self.client, Equals, "hello")
}
