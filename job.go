package patron

import (
	"context"
	"errors"
)

var (
	ErrJobPayloadNotFound = errors.New("payload not found")
)

type Job struct {
	ID      int
	Context context.Context
	Payload map[string]interface{}
}

func (j *Job) Get(key string) (interface{}, error) {
	if j.Payload == nil {
		return nil, ErrJobPayloadNotFound
	}
	val, ok := j.Payload[key]
	if !ok {
		return nil, ErrJobPayloadNotFound
	}
	return val, nil
}
