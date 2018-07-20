package pubsub

import (
	"gopkg.in/mgo.v2/bson"
)

// Publisher is PubSub Publisher, use to broadcast message to every subscriber
type Publisher struct {
	pool *Pool
}

func NewPublisher(pool *Pool) *Publisher {
	return &Publisher{
		pool: pool,
	}
}

// Broadcast will send data to every subscriber
func (p *Publisher) Broadcast(channel string, data interface{}) error {
	conn := p.pool.GetConn()
	defer conn.Close()
	var (
		bsonData []byte
		err      error
	)
	switch data.(type) {
	case Package:
		bsonData, err = bson.Marshal(data)
	default:
		bsonData, err = bson.Marshal(NewPackage(data))
	}
	if err != nil {
		return err
	}

	_, err = conn.Do("PUBLISH", channel, bsonData)
	if err != nil {
		return err
	}
	return nil
}

// BroadcastAndListenReply use to broadcast message
// and listen subscriber reply until timeout
func (p *Publisher) BroadcastAndListenReply(channel string, data interface{}) (*Listener, error) {
	pkg := NewPackage(data)
	listener, _ := Listen(ReplyChannelPrefix + pkg.ID)
	err := p.Broadcast(channel, pkg)
	return listener, err
}
