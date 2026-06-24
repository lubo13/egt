package domain

import "time"

type Message struct {
	Headers   string
	Message   string
	CreatedAt time.Time
}
