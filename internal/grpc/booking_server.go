package grpc

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ita-av/booking-service/internal/model"
	"github.com/ita-av/booking-service/internal/service"
	pb "github.com/ita-av/booking-service/pkg/api/proto"
)

// BookingServer implements the gRPC BookingService
type BookingServer struct {
	pb.UnimplementedBookingServiceServer
	service *service.BookingService
}

// NewBookingServer creates a new booking gRPC server
func NewBookingServer(service *service.BookingService) *BookingServer {
	return &BookingServer{
		service: service,
	}
}

// CreateBooking creates a new booking
func (s *BookingServer) CreateBooking(ctx context.Context, req *pb.CreateBookingRequest) (*pb.Booking, error) {
	// Parse start time
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid start time format: %v", err)
	}

	// Convert service type
	serviceType := model.ServiceType(req.ServiceType)

	// Create booking
	booking, err := s.service.CreateBooking(
		ctx,
		req.UserId,
		req.BarberId,
		startTime,
		serviceType,
		req.Notes,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create booking")
		return nil, status.Errorf(codes.Internal, "failed to create booking: %v", err)
	}

	// Convert to protobuf message
	return convertBookingToProto(booking), nil
}

// GetBooking retrieves a booking by ID
func (s *BookingServer) GetBooking(ctx context.Context, req *pb.GetBookingRequest) (*pb.Booking, error) {
	booking, err := s.service.GetBooking(ctx, req.Id)
	if err != nil {
		if errors.Is(err, errors.New("booking not found")) {
			return nil, status.Errorf(codes.NotFound, "booking not found")
		}

		log.Error().Err(err).Msg("Failed to get booking")
		return nil, status.Errorf(codes.Internal, "failed to get booking: %v", err)
	}

	return convertBookingToProto(booking), nil
}

// UpdateBooking updates an existing booking
func (s *BookingServer) UpdateBooking(ctx context.Context, req *pb.UpdateBookingRequest) (*pb.Booking, error) {
	var startTime *time.Time
	var serviceType *model.ServiceType

	// Parse start time if provided
	if req.StartTime != "" {
		t, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid start time format: %v", err)
		}
		startTime = &t
	}

	// Convert service type if provided
	if req.ServiceType != pb.ServiceType_HAIRCUT {
		st := model.ServiceType(req.ServiceType)
		serviceType = &st
	}

	// Update booking
	booking, err := s.service.UpdateBooking(ctx, req.Id, startTime, serviceType, &req.Notes)
	if err != nil {
		if errors.Is(err, errors.New("booking not found")) {
			return nil, status.Errorf(codes.NotFound, "booking not found")
		}

		log.Error().Err(err).Msg("Failed to update booking")
		return nil, status.Errorf(codes.Internal, "failed to update booking: %v", err)
	}

	return convertBookingToProto(booking), nil
}

// Cancel
// CancelBooking cancels an existing booking
func (s *BookingServer) CancelBooking(ctx context.Context, req *pb.CancelBookingRequest) (*pb.CancelBookingResponse, error) {
	success, err := s.service.CancelBooking(ctx, req.Id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to cancel booking")
		return nil, status.Errorf(codes.Internal, "failed to cancel booking: %v", err)
	}

	var message string
	if success {
		message = "Booking cancelled successfully"
	} else {
		message = "Booking not found or already cancelled"
	}

	return &pb.CancelBookingResponse{
		Success: success,
		Message: message,
	}, nil
}

// GetUserBookings retrieves all bookings for a user
func (s *BookingServer) GetUserBookings(ctx context.Context, req *pb.GetUserBookingsRequest) (*pb.BookingList, error) {
	bookings, err := s.service.GetUserBookings(ctx, req.UserId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user bookings")
		return nil, status.Errorf(codes.Internal, "failed to get user bookings: %v", err)
	}

	// Convert to proto message
	pbBookings := make([]*pb.Booking, len(bookings))
	for i, booking := range bookings {
		pbBookings[i] = convertBookingToProto(booking)
	}

	return &pb.BookingList{
		Bookings: pbBookings,
	}, nil
}

// GetBarberBookings retrieves all bookings for a barber
func (s *BookingServer) GetBarberBookings(ctx context.Context, req *pb.GetBarberBookingsRequest) (*pb.BookingList, error) {
	var date *time.Time

	// Parse date if provided
	if req.Date != "" {
		t, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid date format: %v", err)
		}
		date = &t
	}

	bookings, err := s.service.GetBarberBookings(ctx, req.BarberId, date)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get barber bookings")
		return nil, status.Errorf(codes.Internal, "failed to get barber bookings: %v", err)
	}

	// Convert to proto message
	pbBookings := make([]*pb.Booking, len(bookings))
	for i, booking := range bookings {
		pbBookings[i] = convertBookingToProto(booking)
	}

	return &pb.BookingList{
		Bookings: pbBookings,
	}, nil
}

// GetAvailableTimeSlots retrieves available time slots for a barber on a specific date
func (s *BookingServer) GetAvailableTimeSlots(ctx context.Context, req *pb.GetAvailableTimeSlotsRequest) (*pb.TimeSlotList, error) {
	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid date format: %v", err)
	}

	availableSlots, err := s.service.GetAvailableTimeSlots(ctx, req.BarberId, date)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get available time slots")
		return nil, status.Errorf(codes.Internal, "failed to get available time slots: %v", err)
	}

	// Convert to proto message
	pbTimeSlots := make([]*pb.TimeSlot, len(availableSlots))
	for i, slot := range availableSlots {
		pbTimeSlots[i] = &pb.TimeSlot{
			StartTime: slot.StartTime.Format(time.RFC3339),
			EndTime:   slot.EndTime.Format(time.RFC3339),
		}
	}

	return &pb.TimeSlotList{
		TimeSlots: pbTimeSlots,
	}, nil
}

// Helper function to convert a model.Booking to a proto Booking
func convertBookingToProto(booking *model.Booking) *pb.Booking {
	return &pb.Booking{
		Id:          booking.ID.Hex(),
		UserId:      booking.UserID,
		BarberId:    booking.BarberID,
		StartTime:   booking.StartTime.Format(time.RFC3339),
		EndTime:     booking.EndTime.Format(time.RFC3339),
		ServiceType: pb.ServiceType(booking.ServiceType),
		Status:      pb.BookingStatus(booking.Status),
		Notes:       booking.Notes,
		CreatedAt:   booking.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   booking.UpdatedAt.Format(time.RFC3339),
	}
}
