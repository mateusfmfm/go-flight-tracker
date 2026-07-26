package flight

import (
	"context"
	"log"
	"time"
)

type Poller struct {
	client     *Client
	interval   time.Duration
	outputChan chan []*Aircraft
}

func NewPoller(client *Client, interval time.Duration) *Poller {
	return &Poller{
		client:     client,
		interval:   interval,
		outputChan: make(chan []*Aircraft, 1),
	}
}

func (p *Poller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	p.fetchAndSend(ctx)

	for {
		select {
		case <-ticker.C:
			p.fetchAndSend(ctx)
		case <-ctx.Done():
			log.Println("poller: shutting down...")
			return
		}
	}
}

func (p *Poller) Output() <-chan []*Aircraft {
	return p.outputChan
}

func (p *Poller) fetchAndSend(ctx context.Context) {
	response, err := p.client.GetActiveFlights(ctx)
	if err != nil {
		log.Printf("error fetching active flights: %v", err)
		return
	}

	p.outputChan <- response
}
