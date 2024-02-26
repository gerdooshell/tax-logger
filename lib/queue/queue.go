package queue

import (
	"errors"
	"sync"
	"time"
)

type Queue[T comparable] interface {
	// Insert adds an item to the queue or returns an error if queue is full or the element is nil.
	Insert(element T) error
	// Read reads one item from queue. call IsDone(success) callback if you want to dispose/keep the read value. Read is thread safe.
	Read() Output[T]
	// ReadAll keeps reading from queue forever. This method should be called only once. call IsDone(success) callback if you want to dispose/keep the read value.
	ReadAll() <-chan Output[T]
}

func NewQueue[T comparable](bufferSize int, isSticky bool) Queue[T] {
	return &queue[T]{
		channel:     make(chan T, bufferSize),
		maxLength:   bufferSize,
		readTimeout: time.Second * 3,
		isSticky:    isSticky,
	}
}

type queue[T comparable] struct {
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

// Insert adds an item to the queue or returns an error if queue is full or the element is nil.
func (q *queue[T]) Insert(element T) error {
	if q.length >= q.maxLength {
		return errors.New("queue reached its max length")
	}
	if element == *new(T) {
		return errors.New("nil item is passed to queue")
	}
	q.channel <- element
	q.length++
	return nil
}

// Read reads one item from queue. call IsDone(success) callback if you want to dispose/keep the read value. Read is thread safe.
func (q *queue[T]) Read() Output[T] {
	q.mu.Lock()
	var value T
	var err error
	if q.firstItem != nil {
		value = *q.firstItem
	} else {
		select {
		case value = <-q.channel:
			q.firstItem = &value
			q.length--
		default:
			err = errors.New("empty queue")
		}
	}
	isTimedOut := false
	isDoneCalled := make(chan struct{}, 1)
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
	go func() {
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
	return qOut
}

// ReadAll keeps reading from queue forever. This method should be called only once. call IsDone(success) callback if you want to dispose/keep the read value.
func (q *queue[T]) ReadAll() <-chan Output[T] {
	// TODO: add stop callback for graceful stop of a docker container
	out := make(chan Output[T], 1)
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
					q.length--
				case <-time.After(retryTimeKeepAlive):
					continue
				}
			}
			isTimedOut := false
			isDoneCalled := make(chan struct{}, 1)
			qOut := Output[T]{Value: value, Err: err, IsDone: func(success bool) {
				if isTimedOut {
					return
				}
				defer close(isDoneCalled)
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
