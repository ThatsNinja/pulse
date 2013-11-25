package pulse

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/polydice/pulse/testutil"
	. "launchpad.net/gocheck"
)

var port string = ":4000"

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&S{})

type S struct {
	pump *Pump
}

func (self *S) SetUpTest(c *C) {
	self.pump = New(port)
	go self.pump.Start(true)
}

func (self *S) TestPublish(c *C) {
	req := testutil.RequestFromSNS()
	resp, err := http.Post("http://localhost"+port+"/publish/foo", "text/plain", req.Body)
	c.Check(err, IsNil)
	c.Check(resp.StatusCode, Equals, 200)
}

func (self *S) TestSubscribe(c *C) {
	go func() {
		time.Sleep(time.Second)
		req := testutil.RequestFromSNS()
		http.Post("http://localhost"+port+"/publish/foo", "text/plain", req.Body)
	}()

	p := make([]byte, 50)
	resp, err := http.Get("http://localhost" + port + "/subscribe/foo")
	n, err := resp.Body.Read(p)
	text := string(p[:n])

	c.Check(err, IsNil)
	c.Check(strings.Trim(text, "\n"), Equals, "data: test message")
}
