package pubsub

import (
	"github.com/garyburd/redigo/redis"
)

// Pool use to management Redis connection pool
type Pool struct {
	pool *redis.Pool
}

// GetConn use to get connection from Redis connection pool
func (p *Pool) GetConn() redis.Conn {
	return p.pool.Get()
}

// GetPubSubConn use to get pubsub connection
func (p *Pool) GetPubSubConn() redis.PubSubConn {
	conn := p.GetConn()
	return redis.PubSubConn{conn}
}
