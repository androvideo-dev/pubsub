package pubsub

import (
	"github.com/satori/go.uuid"

	"fmt"
)

type Package struct {
	Data interface{}
	ID   string
}

func NewPackage(data interface{}) Package {
	id, _ := uuid.NewV4()

	return Package{
		Data: data,
		ID:   fmt.Sprintf("%s", id),
	}
}
