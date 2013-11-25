package sns

import (
	"bytes"
	"log"
	"net/http"
	"testing"
	"github.com/smartystreets/goconvey/convey"
	. "launchpad.net/gocheck"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&S{})

type S struct {
	request              *http.Request
	expectedNotification *Notification
}

func (self *S) SetUpTest(c *C) {
	body := `{
  "Type" : "Notification",
  "MessageId" : "da41e39f-ea4d-435a-b922-c6aae3915ebe",
  "TopicArn" : "arn:aws:sns:us-east-1:123456789012:MyTopic",
  "Subject" : "test",
  "Message" : "test message",
  "Timestamp" : "2012-04-25T21:49:25.719Z",
  "SignatureVersion" : "1",
  "Signature" : "EXAMPLElDMXvB8r9R83tGoNn0ecwd5UjllzsvSvbItzfaMpN2nk5HVSw7XnOn/49IkxDKz8YrlH2qJXj2iZB0Zo2O71c4qQk1fMUDi3LGpij7RCW7AW9vYYsSqIKRnFS94ilu7NFhUzLiieYr4BKHpdTmdD6c0esKEYBpabxDSc=",
  "SigningCertURL" : "https://sns.us-east-1.amazonaws.com/SimpleNotificationService-f3ecfb7224c7233fe7bb5f59f96de52f.pem",
  "UnsubscribeURL" : "https://sns.us-east-1.amazonaws.com/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-1:123456789012:MyTopic:2bcfbf39-05c3-41de-beaa-fcfcc21c8f55"
}`
	buf := bytes.NewReader([]byte(body))
	req, err := http.NewRequest("POST", "http://example.com", buf)

	if err != nil {
		log.Fatal("Failed to build request:", err)
	}

	self.request = req

	self.expectedNotification = &Notification{
		Type:             "Notification",
		MessageId:        "da41e39f-ea4d-435a-b922-c6aae3915ebe",
		TopicArn:         "arn:aws:sns:us-east-1:123456789012:MyTopic",
		Subject:          "test",
		Message:          "test message",
		Timestamp:        "2012-04-25T21:49:25.719Z",
		SignatureVersion: "1",
		Signature:        "EXAMPLElDMXvB8r9R83tGoNn0ecwd5UjllzsvSvbItzfaMpN2nk5HVSw7XnOn/49IkxDKz8YrlH2qJXj2iZB0Zo2O71c4qQk1fMUDi3LGpij7RCW7AW9vYYsSqIKRnFS94ilu7NFhUzLiieYr4BKHpdTmdD6c0esKEYBpabxDSc=",
		SigningCertURL:   "https://sns.us-east-1.amazonaws.com/SimpleNotificationService-f3ecfb7224c7233fe7bb5f59f96de52f.pem",
		UnsubscribeURL:   "https://sns.us-east-1.amazonaws.com/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-1:123456789012:MyTopic:2bcfbf39-05c3-41de-beaa-fcfcc21c8f55",
	}
}

func (self *S) TestNewFromRequest(c *C) {
	convey.Convey("Parse SNS POST request into Notification struct", c, func() {

		req := self.request

		notification := NewFromRequest(req)
		expectedNotification := self.expectedNotification

		convey.So(notification, convey.ShouldHaveSameTypeAs, expectedNotification)
		convey.So(notification.Type, convey.ShouldEqual, expectedNotification.Type)
		convey.So(notification.MessageId, convey.ShouldEqual, expectedNotification.MessageId)
		convey.So(notification.TopicArn, convey.ShouldEqual, expectedNotification.TopicArn)
		convey.So(notification.Subject, convey.ShouldEqual, expectedNotification.Subject)
		convey.So(notification.Message, convey.ShouldEqual, expectedNotification.Message)
		convey.So(notification.Timestamp, convey.ShouldEqual, expectedNotification.Timestamp)
		convey.So(notification.SignatureVersion, convey.ShouldEqual, expectedNotification.SignatureVersion)
		convey.So(notification.Signature, convey.ShouldEqual, expectedNotification.Signature)
		convey.So(notification.SigningCertURL, convey.ShouldEqual, expectedNotification.SigningCertURL)
		convey.So(notification.UnsubscribeURL, convey.ShouldEqual, expectedNotification.UnsubscribeURL)
	})

}
