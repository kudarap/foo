package worker

import (
	"context"
	"log/slog"
)

const defaultJobQueueSize = 10

// Job represents a task details for a worker.
type Job struct {
	Topic   string
	Payload []byte
	Done    func() error
}

// JobHandler represents worker handler functions
type JobHandler func(context.Context, Job) error

// MiddlewareFunc represents worker job middleware
type MiddlewareFunc func(JobHandler) JobHandler

// Worker represents a worker that waits for a job and process base on handler.
type Worker struct {
	queue       chan Job
	quit        chan struct{}
	done        chan error
	router      map[string]JobHandler
	middlewares []MiddlewareFunc
	listener    jobListener
	logger      *slog.Logger
}

// jobListener provides access to job producers.
type jobListener interface {
	// Listen starts subscription to topics and listen for the jobs.
	Listen(topics []string, q chan<- Job, stop chan struct{}) (err error)
}

// New create new instance of worker.
func New(listener jobListener, queueSize int, logger *slog.Logger) *Worker {
	if queueSize == 0 {
		queueSize = defaultJobQueueSize
	}

	logger = logger.With("pkg", "worker")
	logger.Info("init", "queue-size", queueSize)
	return &Worker{
		queue:    make(chan Job, queueSize),
		quit:     make(chan struct{}, 1),
		done:     make(chan error, 1),
		router:   map[string]JobHandler{},
		listener: listener,
		logger:   logger,
	}
}

// Run starts listening to jobs base on topics and process according to their handlers.
func (w *Worker) Run() error {
	var tt []string
	for t := range w.router {
		w.logger.Info("register topic", "topic", t)
		tt = append(tt, t)
	}

	stopListen := make(chan struct{}, 1)
	if err := w.listener.Listen(tt, w.queue, stopListen); err != nil {
		return err
	}

	// Process job received from the job listener.
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		for {
			select {
			// Quitting will not finish un-started job on the queue but guartees to finish
			// processing job.
			case <-w.quit:
				w.logger.Info("worker quiting...")
				stopListen <- struct{}{}
				w.done <- nil
				return

			case job := <-w.queue:
				handle, ok := w.router[job.Topic]
				if !ok {
					w.logger.Debug("topic not handled", "topic", job.Topic)
					continue
				}

				// Apply registered middlewares.
				for _, m := range w.middlewares {
					handle = m(handle)
				}

				if err := handle(ctx, job); err != nil {
					continue
				}
				if err := job.Done(); err != nil {
					w.logger.Error("job done", "err", err, "topic", job.Topic)
					continue
				}
			}
		}
	}()

	w.logger.Info("worker running", "queue_size", cap(w.queue))
	return nil
}

// Stop gracefully stops listener and closes job queue.
func (w *Worker) Stop() error {
	w.logger.Info("stopping worker...")
	w.quit <- struct{}{}
	return <-w.done
}

// HandleFunc registers handler by topic that routes job to a handler.
func (w *Worker) HandleFunc(topic string, f JobHandler) {
	w.router[topic] = f
}

// Use registers middlewares for job handlers.
func (w *Worker) Use(mm ...MiddlewareFunc) {
	for _, m := range mm {
		w.middlewares = append(w.middlewares, m)
	}
}
