package port

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"git.infra.egt.com/gateway/module/internal/dto"
)

type DeviceEventProcessorInterface interface {
	Add(*dto.DeviceEvent)
	StopProcessingGracefully()
}

// DeviceEventHandler handle data request.
type DeviceEventHandler struct {
	log                  *zap.Logger
	deviceProcessorEvent DeviceEventProcessorInterface
}

// NewDeviceEventHandler builds a new DeviceEventHandler.
func NewDeviceEventHandler(log *zap.Logger, deviceProcessorEvent DeviceEventProcessorInterface) *DeviceEventHandler {
	return &DeviceEventHandler{
		log:                  log,
		deviceProcessorEvent: deviceProcessorEvent,
	}
}

func (*DeviceEventHandler) Pattern() string {
	return "/api/v1/device/event"
}

func (h *DeviceEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	deveiceEventDTO := &dto.DeviceEvent{}

	err := validateRequest(r, deveiceEventDTO)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	h.deviceProcessorEvent.Add(deveiceEventDTO)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}
