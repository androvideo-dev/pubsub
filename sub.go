package pubsub

import (
	"github.com/garyburd/redigo/redis"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/mgo.v2/bson"

	"errors"
	"time"
)

// Subscriber is PubSub Subscriber, use to listen channel message
type Subscriber struct {
	connection redis.PubSubConn
	channels   map[string]chan interface{}
	errChan    chan error
}

type Listener struct {
	subscriber *Subscriber
	channel    string
	notify     chan interface{}
}

// Generate a Subscriber with redis pool.
func NewSubscriber(pool *Pool) *Subscriber {
	channels := make(map[string]chan interface{})
	errChan := make(chan error)
	psConn := pool.GetPubSubConn()
	go func() {
		for {
			receive := psConn.Receive()
			switch val := receive.(type) {
			case redis.Message:
				if channel, ok := channels[val.Channel]; ok {
					channel <- val.Data
				} else {
					errChan <- errors.New("miss channel")
				}
			case error:
				errChan <- val
			}
		}
	}()

	subscriber := Subscriber{
		connection: psConn,
		channels:   channels,
		errChan:    errChan,
	}
	return &subscriber
}

// Generate listener with channel id from a Subscriber.
func (subscriber *Subscriber) Listen(channel string) (*Listener, error) {
	if _, ok := subscriber.channels[channel]; ok {
		return nil, errors.New("this channel is exist")
	}
	notify := make(chan interface{}, 4) // some buffer size to avoid blocking
	subscriber.channels[channel] = notify
	subscriber.connection.Subscribe(channel)
	listener := &Listener{
		subscriber: subscriber,
		channel:    channel,
		notify:     notify,
	}
	return listener, nil
}

// Wait data from a channel.
func (listener *Listener) GetData(out interface{}, timeout ...time.Duration) (*Replyer, error) {
	var timeoutChannel <-chan time.Time
	if len(timeout) > 0 {
		timeoutChannel = time.After(timeout[0])
	}

	select {
	case reply := <-listener.notify:
		if reply == nil {
			// the channel was closed
			return nil, errors.New("close")
		}
		if errMsg, err := reply.(error); err {
			return nil, errMsg
		}
		result := Package{}
		err := bson.Unmarshal(reply.([]byte), &result)
		if err != nil {
			return nil, err
		}
		mapstructure.Decode(result.Data, &out)

		replayer := NewReplyer(result.ID)
		return &replayer, nil
	case <-timeoutChannel:
		return nil, errors.New("timeout")
	}
}

// Close redis subscribe.
func (listener *Listener) Close() {
	listener.subscriber.connection.Unsubscribe(listener.channel)
	delete(listener.subscriber.channels, listener.channel)
	close(listener.notify)
}
