package audit

import (
	"context"
	"fmt"
	"time"
)

type AuditEvent struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type Subscriber interface {
	Send(ctx context.Context, event AuditEvent) error
}

type Observer struct {
	subscribers []Subscriber
}

func NewObserver(subscribers ...Subscriber) *Observer {
	return &Observer{subscribers: subscribers}
}

func (o *Observer) AddSub(sub Subscriber) {
	o.subscribers = append(o.subscribers, sub)
}

func (o *Observer) Notify(ctx context.Context, metrics []string, ipAddr string) {
	if len(o.subscribers) == 0 {
		return
	}

	event := AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: ipAddr,
	}

	for _, sub := range o.subscribers {
		go func(s Subscriber) {
			err := s.Send(ctx, event)
			if err != nil {
				fmt.Println(err)
			}
		}(sub)
	}
}
