package event

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent struct {
	Event interface{}
}

type DomainDeviceEvent struct {
	ID         uuid.UUID
	DeviceId   string
	DeviceName string
	SensorName string
	Message    string
	CreatedAt  time.Time
}
