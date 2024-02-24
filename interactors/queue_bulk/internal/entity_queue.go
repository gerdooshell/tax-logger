package internal

type EntityQ[T comparable] interface {
	GetElements() []T
}

func NewEntityQ[T comparable](elements []T) EntityQ[T] {
	return &entityQ[T]{elements: elements}
}

type entityQ[T comparable] struct {
	elements []T
}

func (e *entityQ[T]) GetElements() []T {
	return e.elements
}
