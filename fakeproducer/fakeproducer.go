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

func (j *Job) Listen(topics []string, q chan<- worker.Job, stop chan struct{}) error {
	log.Println("fake subscribed to", topics)

	go func() {
		var i int
		for {
			select {
			case <-stop:
				log.Println("listener stopped")
				return

			default:
				i++
				log.Println("fake producing job...", i)
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

	return nil
}

func (j *Job) Close() error {
	log.Println("fake invoked close")
	return nil
}
