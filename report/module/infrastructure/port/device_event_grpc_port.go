package port

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	reportapiv1 "git.infra.egt.com/clients/go/report/api/v1"
	"git.infra.egt.com/report/module/internal/application"
)

type DeviceEventGPRCHandler struct {
	deviceEventFetcher *application.DeviceEventFetcher

	reportapiv1.UnimplementedDeviceEventServiceServer
}

func NewDeviceEventGPRCHandler(deviceEventFetcher *application.DeviceEventFetcher) *DeviceEventGPRCHandler {
	return &DeviceEventGPRCHandler{
		deviceEventFetcher: deviceEventFetcher,
	}
}

func RegisterDeviceEventGRPCHandler(
	grpcServer *grpc.Server,
	deviceEventGPRCHandler *DeviceEventGPRCHandler,
) {
	reportapiv1.RegisterDeviceEventServiceServer(grpcServer, deviceEventGPRCHandler)
}

func (deviceEventGPRCHandler *DeviceEventGPRCHandler) GetEvent(
	ctx context.Context,
	getRequest *reportapiv1.GetEventRequest,
) (*reportapiv1.GetEventResponse, error) {
	id := getRequest.Id
	if err := uuid.Validate(id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID: %v", err)
	}

	uuid := uuid.MustParse(id)

	deviceEvent, err := deviceEventGPRCHandler.deviceEventFetcher.GetById(ctx, &uuid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	if deviceEvent == nil {
		return &reportapiv1.GetEventResponse{}, nil
	}

	return &reportapiv1.GetEventResponse{
		Event: &reportapiv1.Event{
			Id:          deviceEvent.ID.String(),
			DeviceId:    deviceEvent.DeviceId,
			DeviceName:  deviceEvent.DeviceName,
			SensorName:  deviceEvent.SensorName,
			Message:     deviceEvent.Message,
			CreatedTime: timestamppb.New(*deviceEvent.CreatedAt),
		},
	}, nil
}

func (deviceEventGPRCHandler *DeviceEventGPRCHandler) ListEvent(
	ctx context.Context,
	listRequest *reportapiv1.ListEventRequest,
) (*reportapiv1.ListEventResponse, error) {
	deviceEvents, err := deviceEventGPRCHandler.deviceEventFetcher.GetAll(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	var events []*reportapiv1.Event
	for _, deviceEvent := range deviceEvents {
		events = append(events, &reportapiv1.Event{
			Id:          deviceEvent.ID.String(),
			DeviceId:    deviceEvent.DeviceId,
			DeviceName:  deviceEvent.DeviceName,
			SensorName:  deviceEvent.SensorName,
			Message:     deviceEvent.Message,
			CreatedTime: timestamppb.New(*deviceEvent.CreatedAt),
		})
	}

	return &reportapiv1.ListEventResponse{
		Events: events,
	}, nil
}
