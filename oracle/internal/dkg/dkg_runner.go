package dkg

import (
	"log"
	"time"
)

type Runner struct {
	fetcher  *Fetcher
	executor *Executor
	interval time.Duration
}

func NewRunner(
	fetcher *Fetcher,
	executor *Executor,
	interval time.Duration,
) *Runner {
	return &Runner{
		fetcher:  fetcher,
		executor: executor,
		interval: interval,
	}
}

func (r *Runner) Run() {
	for {
		dkg, err := r.fetcher.Fetch()
		if err != nil {
			log.Printf("Failed to fetch DKG: %v", err)
			time.Sleep(r.interval)
			continue
		}
		r.executor.Execute(dkg)
		time.Sleep(r.interval)
	}
}
