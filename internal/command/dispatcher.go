package command

import "errors"

var ErrEmptyCommand = errors.New("empty command")
var ErrUnknownCommand = errors.New("unknown command")

type Dispatcher struct {
	registry *Registry
}

func NewDispatcher(registry *Registry) *Dispatcher {
	return &Dispatcher{registry: registry}
}

func (d *Dispatcher) Dispatch(argv []string) (string, error) {
	if len(argv) == 0 {
		return "", ErrEmptyCommand
	}
	handler, ok := d.registry.Lookup(argv[0])
	if !ok {
		return "", ErrUnknownCommand
	}
	return handler(argv[1:])
}
