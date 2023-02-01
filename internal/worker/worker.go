package worker

import (
	"context"
	"sync"

	"github.com/iunary/sesify/internal/compaign"
	"github.com/iunary/sesify/internal/sender"
)

type Worker struct {
	Count     int64
	Compaigns chan compaign.Compaign
	Results   chan compaign.Result
	Done      chan struct{}
	wg        sync.WaitGroup
	Provider  sender.Provider
}

func NewWorker(count int64, provider sender.Provider) Worker {
	return Worker{
		Count:     count,
		Compaigns: make(chan compaign.Compaign, count),
		Results:   make(chan compaign.Result, count),
		Done:      make(chan struct{}),
		Provider:  provider,
	}
}

func (s *Worker) Run(ctx context.Context) {

	for i := 0; i < int(s.Count); i++ {
		s.wg.Add(1)
		go s.run(ctx, &s.wg, s.Compaigns, s.Results)
	}

	s.wg.Wait()
	close(s.Done)
	close(s.Results)
}

func (s *Worker) run(ctx context.Context, wg *sync.WaitGroup, comp <-chan compaign.Compaign, result chan<- compaign.Result) {
	defer wg.Done()
	for {
		select {
		case j, ok := <-comp:
			if !ok {
				return
			}
			result <- s.Provider.Send(ctx, &j)
		case <-ctx.Done():
			result <- compaign.Result{
				Error: ctx.Err(),
			}
		}
	}
}
