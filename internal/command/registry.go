package command

import "strings"

type Handler func(args []string) (string, error)

type Registry struct {
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

func (r *Registry) Register(name string, handler Handler) {
	r.handlers[strings.ToUpper(name)] = handler
}

func (r *Registry) Lookup(name string) (Handler, bool) {
	h, ok := r.handlers[strings.ToUpper(name)]
	return h, ok
}
