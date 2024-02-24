package queueBulk

import (
	"time"

	"github.com/gerdooshell/tax-logger/interactors/queue_bulk/internal"
	"github.com/gerdooshell/tax-logger/lib/queue"
)

type Output[T any] struct {
	Value  []T
	Err    error
	IsDone func(success bool)
}

type QueueBulk[T comparable] interface {
	Insert(element T) error
	ReadAll() <-chan Output[T]
}

func NewQueueBulk[T comparable](batchSize int, bufferSize int, purgeTimeout time.Duration) QueueBulk[T] {
	qb := &queueBulk[T]{
		singularQueue: queue.NewQueue[T](batchSize*2, true),
		pluralQueue:   queue.NewQueue[internal.EntityQ[T]](bufferSize, true),
		batchSize:     batchSize,
		purgeTimeout:  purgeTimeout,
		listenPeriod:  time.Microsecond * 100,
		length:        0,
	}
	go qb.listen()
	return qb
}

type queueBulk[T comparable] struct {
	singularQueue queue.Queue[T]
	pluralQueue   queue.Queue[internal.EntityQ[T]]
	batchSize     int
	purgeTimeout  time.Duration
	listenPeriod  time.Duration
	length        int
}

func (q *queueBulk[T]) Insert(element T) error {
	if err := q.singularQueue.Insert(element); err != nil {
		return err
	}
	q.length++
	return nil
}

func (q *queueBulk[T]) ReadAll() <-chan Output[T] {
	pluralOutChan := q.pluralQueue.ReadAll()
	out := make(chan Output[T], 1)
	go func() {
		defer close(out)
		for pluralOut := range pluralOutChan {
			out <- Output[T]{
				Value:  pluralOut.Value.GetElements(),
				Err:    pluralOut.Err,
				IsDone: pluralOut.IsDone,
			}
		}
	}()
	return out
}

func (q *queueBulk[T]) listen() {
	t0 := time.Now()
	for {
		if q.length < q.batchSize && (time.Since(t0) < q.purgeTimeout || q.length == 0) {
			<-time.After(q.listenPeriod)
			continue
		}
		t0 = time.Now()
		count := min(q.length, q.batchSize)
		batch := make([]T, 0, count)
		for i := 0; i < count; i++ {
			element := q.singularQueue.Read()
			element.IsDone(true)
			if element.Err != nil {
				continue
			}
			batch = append(batch, element.Value)
		}
		eq := internal.NewEntityQ[T](batch)
		err := q.pluralQueue.Insert(eq)
		if err != nil {
			continue
		}
		q.length -= count
	}
}
