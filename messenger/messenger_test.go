package messenger

import (
	"testing"
	"time"
	. "launchpad.net/gocheck"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&S{})

type S struct {
	msger            *DefaultMessenger
	msgerWithClients *DefaultMessenger
	client           chan string
}

func (self *S) SetUpTest(c *C) {
	self.msger = New("foo").(*DefaultMessenger)
	self.msger.Start()

	self.client = make(chan string)
	self.msgerWithClients = New("bar").(*DefaultMessenger)
	self.msgerWithClients.Start()
	self.msgerWithClients.AddClient(self.client)
}

func (self *S) TestAddClient(c *C) {
	self.msger.AddClient(self.client)
	time.Sleep(time.Millisecond)
	c.Check(self.msger.clients, HasLen, 1, Commentf("Should add client"))

	self.msger.AddClient(self.client)
	time.Sleep(time.Millisecond)
	c.Check(self.msger.clients, HasLen, 1, Commentf("Should ignore duplicated client"))
}

func (self *S) TestRemoveClient(c *C) {
	self.msgerWithClients.RemoveClient(self.client)
	time.Sleep(time.Millisecond)
	c.Check(self.msgerWithClients.clients, HasLen, 0, Commentf("Should remove client"))
}

func (self *S) TestSendMessage(c *C) {
	self.msgerWithClients.SendMessage("hello")
	c.Check(<-self.client, Equals, "hello")
}
