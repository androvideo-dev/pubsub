package pubsub_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/daniel-androvideo/pubsub"
	"github.com/garyburd/redigo/redis"

	"os"
	"time"
)

var _ = Describe("PubSub能夠正確傳送，以及接收資料", func() {

	redisHost := os.Getenv("REDIS_HOST")
	redisClient := &redis.Pool{
		MaxIdle:     1,
		MaxActive:   100,
		IdleTimeout: time.Duration(180) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				return nil, err
			}

			if _, err := c.Do("SELECT", "10"); err != nil {
				c.Close()
				return nil, err
			}

			return c, nil
		},
	}
	pubsub.InitPubSub(redisClient)

	Context("publisher傳送字串給subscriber", func() {
		It("單一channel", func() {
			channel := "test_channel"

			go func() {
				<-time.After(50 * time.Millisecond)
				pubsub.Broadcast(channel, "ping")
			}()

			listener, _ := pubsub.Listen(channel)
			defer listener.Close()
			var actual string
			listener.GetData(&actual, 5*time.Second)
			Expect(actual).To(Equal("ping"))
		})
	})

	Context("傳送struct", func() {
		It("單一channel", func() {
			channel := "test_channel"

			type ping struct {
				Text string
			}

			go func() {
				<-time.After(50 * time.Millisecond)
				sendData := ping{"ping"}
				pubsub.Broadcast(channel, sendData)
			}()

			listener, _ := pubsub.Listen(channel)
			defer listener.Close()
			var actual ping
			if _, err := listener.GetData(&actual, 5*time.Second); err == nil {
				Expect(actual.Text).To(Equal("ping"))
			} else {
				Fail("傳送struct receive error")
			}
		})
	})

	Context("subscriber收到訊號後，可以reply資料給publisher", func() {
		It("單一channel，一個reply", func(done Done) {
			channel := "test_channel"

			listener, _ := pubsub.Listen(channel)
			defer listener.Close()
			go func() {
				var s string
				replyer, _ := listener.GetData(&s, 3*time.Second)
				replyer.Reply("pong")
			}()

			<-time.After(50 * time.Millisecond)
			pubListener, _ := pubsub.BroadcastAndListenReply(channel, "ping")
			defer pubListener.Close()
			var actual string
			if _, err := pubListener.GetData(&actual, 2*time.Second); err != nil {
				Fail(err.Error())
			} else {
				Expect(actual).To(Equal("pong"))
				close(done)
			}
		}, 5.0)

		It("publisher沒收到reply會正確timeout", func(done Done) {
			channel := "timeout_channel"

			listener, _ := pubsub.Listen(channel)
			defer listener.Close()
			go func() {
				<-time.After(200 * time.Millisecond)
				var s string
				replyer, _ := listener.GetData(&s, 1*time.Second)
				replyer.Reply("pong")
			}()

			<-time.After(50 * time.Millisecond)
			pubListener, _ := pubsub.BroadcastAndListenReply(channel, "ping")
			defer pubListener.Close()
			var s string
			if _, err := pubListener.GetData(&s, 50*time.Millisecond); err != nil {
				Succeed()
				close(done)
			} else {
				Fail("receive data, no timeout")
				close(done)
			}
		}, 6.0)
	})
})
