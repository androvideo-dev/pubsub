package pubsub

// Replyer use to reply data, when subscriber receive publisher message
type Replyer struct {
	Channel string
}

const (
	ReplyChannelPrefix = "$reply:"
)

// Generate a Replyer with a data id.
func NewReplyer(dataID string) Replyer {
	return Replyer{
		Channel: ReplyChannelPrefix + dataID,
	}
}

// Reply use to reply message to reply channel
func (r *Replyer) Reply(data interface{}) error {
	return Broadcast(r.Channel, data)
}
