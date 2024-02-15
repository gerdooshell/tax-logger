package queue

import (
	"errors"
	"sync"
	"time"
)

type Queue[T any] interface {
	Add(element T) error
	Read() <-chan Output[T]
	ReadAll() <-chan Output[T]
}

func NewQueue[T any](bufferSize int, isSticky bool) Queue[T] {
	return &queue[T]{
		channel:     make(chan T, bufferSize),
		maxLength:   bufferSize,
		readTimeout: time.Second * 3,
		isSticky:    isSticky,
	}
}

type queue[T any] struct {
	channel     chan T
	mu          sync.Mutex
	firstItem   *T
	maxLength   int
	length      int
	readTimeout time.Duration
	isSticky    bool
	readAllChan chan Output[T]
}

type Output[T any] struct {
	Value  T
	Err    error
	IsDone func(success bool)
}

func (q *queue[T]) Add(element T) error {
	if q.length >= q.maxLength {
		return errors.New("queue reached its max length")
	}
	q.channel <- element
	q.length++
	return nil
}

func (q *queue[T]) Read() <-chan Output[T] {
	out := make(chan Output[T])
	go func() {
		defer close(out)
		q.mu.Lock()
		var value T
		var err error
		if q.firstItem != nil {
			value = *q.firstItem
		} else {
			select {
			case value = <-q.channel:
				q.firstItem = &value
			default:
				err = errors.New("empty queue")
			}
		}
		isTimedOut := false
		isDoneCalled := make(chan struct{})
		qOut := Output[T]{Value: value, Err: err, IsDone: func(success bool) {
			defer close(isDoneCalled)
			if isTimedOut {
				return
			}
			if success {
				q.firstItem = nil
			}
			isDoneCalled <- struct{}{}
			q.mu.Unlock()
		}}
		out <- qOut
		select {
		case <-isDoneCalled:
		case <-time.After(q.readTimeout):
			isTimedOut = true
			close(isDoneCalled)
			if !q.isSticky {
				q.firstItem = nil
			}
			q.mu.Unlock()
		}
	}()
	return out
}

func (q *queue[T]) ReadAll() <-chan Output[T] {
	out := make(chan Output[T])
	retryTimeKeepAlive := time.Minute
	go func() {
		defer close(out)
		for {
			var value T
			var err error
			if q.firstItem != nil {
				value = *q.firstItem
			} else {
				select {
				case value = <-q.channel:
					q.firstItem = &value
				case <-time.After(retryTimeKeepAlive):
					continue
				}
			}
			isTimedOut := false
			isDoneCalled := make(chan struct{})
			qOut := Output[T]{Value: value, Err: err, IsDone: func(success bool) {
				defer close(isDoneCalled)
				if isTimedOut {
					return
				}
				if success {
					q.firstItem = nil
				}
				isDoneCalled <- struct{}{}
			}}
			out <- qOut
			select {
			case <-isDoneCalled:
			case <-time.After(q.readTimeout):
				isTimedOut = true
				close(isDoneCalled)
				if !q.isSticky {
					q.firstItem = nil
				}
			}
		}
	}()
	return out
}
