package domain

import (
	"time"

	"github.com/google/uuid"
)

type DeviceEvent struct {
	ID         uuid.UUID
	DeviceId   string
	DeviceName string
	SensorName string
	Message    string
	CreatedAt  *time.Time
}

func NewDeviceEvent(
	id string,
	deviceId string,
	deviceName string,
	sensorName string,
	message string,
	createAt *time.Time,
) (*DeviceEvent, error) {
	// TODO: Add domain logic here if needed
	// NOTE: this is simple validation and could be inside port, but I set it here as a domain rule - depends on the business requirements
	if err := uuid.Validate(id); err != nil {
		return nil, err
	}

	deviceEvent := &DeviceEvent{
		ID:         uuid.MustParse(id),
		DeviceId:   deviceId,
		DeviceName: deviceName,
		SensorName: sensorName,
		Message:    message,
		CreatedAt:  createAt,
	}

	return deviceEvent, nil
}
