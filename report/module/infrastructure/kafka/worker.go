package main

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type ConsumerInterface interface {
	Consume(ctx context.Context, worker *Wroker)
}

type Wroker struct {
	wg        sync.WaitGroup
	cancel    context.CancelFunc
	logger    *zap.Logger
	consumers []ConsumerInterface
}

func NewWorker(logger *zap.Logger, consumers []ConsumerInterface) *Wroker {
	return &Wroker{
		logger:    logger,
		consumers: consumers,
	}
}

func (worker *Wroker) Run(ctx context.Context) {
	baseContext := context.Background()
	c, cancel := context.WithCancel(baseContext)
	worker.cancel = cancel

	for _, consumer := range worker.consumers {
		worker.wg.Add(1)

		go consumer.Consume(c, worker)
	}
}

func (worker *Wroker) Done() {
	worker.wg.Done()
}

func (worker *Wroker) StopProcessingGracefully() {
	worker.cancel()
	worker.wg.Wait()
}
