package mob

// Option configures a handler during the registration process.
type Option interface {
	apply(*handler)
}

type optionFunc func(*handler)

func (f optionFunc) apply(h *handler) {
	f(h)
}

// WithName returns an Option that associates a given name with a handler.
func WithName(name string) Option {
	var opt optionFunc = func(h *handler) {
		h.name = name
	}
	return opt
}
