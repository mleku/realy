package eventstore

import "errors"

var (
	ErrDupEvent       = errors.New("duplicate: event already exists")
	ErrEventNotExists = errors.New("unknown: event not known by any source of this relay")
)
