package config

// Value is a generic configuration value that updates itself when the configuration is reloaded.
type Value[T any] interface {
	Get() T
}

type StaticValue[T any] struct {
	Value T
}

func (i StaticValue[T]) Get() T {
	return i.Value
}

// valueFunc is a helper type that makes it easier to implement the Value interface.
type valueFunc[T any] func() T

func (f valueFunc[T]) Get() T {
	return f()
}
