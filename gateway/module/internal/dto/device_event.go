package dto

type DeviceEvent struct {
	ID         string `json:"id" validate:"required,min=36,max=36"`
	DeviceId   string `json:"device_id" validate:"required,min=36,max=36"`
	DeviceName string `json:"device_name" validate:"required,min=5,max=60"`
	SensorName string `json:"sensor_name" validate:"required,min=5,max=60"`
	Message    string `json:"message" validate:"required,min=10,max=255"`
}
