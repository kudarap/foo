package fakeproducer

import (
	"fmt"
	"log"
	"time"

	"github.com/kudarap/foo/worker"
)

type Job struct {
	sleep time.Duration
}

func New(sleep time.Duration) *Job {
	return &Job{sleep}
}

func (j *Job) Listen(topics []string, q chan<- worker.Job) (stop func() error, err error) {
	log.Println("fake subscribed to", topics)

	quit := make(chan struct{}, 1)
	done := make(chan error, 1)
	stop = func() error {
		quit <- struct{}{}
		return <-done
	}

	go func() {
		var i int
		for {
			select {
			case <-quit:
				log.Println("listener stopped")
				done <- nil // place error here
				return

			default:
				i++
				q <- worker.Job{
					Topic:   "demo",
					Payload: []byte(fmt.Sprintf(`{"faker": %d}`, i)),
					Done: func() error {
						log.Println("fake done!", i)
						return nil
					},
				}
				log.Println("fake produced job", i)
				time.Sleep(j.sleep)
			}
		}
	}()
	return
}

func (j *Job) Close() error {
	log.Println("fake invoked close")
	return nil
}
