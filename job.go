package patron

import (
	"context"
	"errors"
)

var (
	ErrPayloadNotFound = errors.New("payload not found")
)

type Job struct {
	ID      int
	Context context.Context
	Payload map[string]interface{}
}

func (j *Job) GetPayload(key string) (interface{}, error) {
	if j.Payload == nil {
		return nil, ErrPayloadNotFound
	}

	val, ok := j.Payload[key]
	if !ok {
		return nil, ErrPayloadNotFound
	}

	return val, nil
}
