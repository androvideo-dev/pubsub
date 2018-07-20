package pubsub

import (
	"github.com/garyburd/redigo/redis"
)

var (
	pool *Pool
	pub  *Publisher
	sub  *Subscriber
)

func InitPubSub(client *redis.Pool) {
	pool = &Pool{client}
	pub = NewPublisher(pool)
	sub = NewSubscriber(pool)
}

func Broadcast(channel string, data interface{}) error {
	return pub.Broadcast(channel, data)
}

func BroadcastAndListenReply(channel string, data interface{}) (*Listener, error) {
	return pub.BroadcastAndListenReply(channel, data)
}

func Listen(channel string) (*Listener, error) {
	return sub.Listen(channel)
}
