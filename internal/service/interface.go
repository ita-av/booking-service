// In internal/service/interface.go (create this file)
package service

import (
	"context"
	"time"

	"github.com/ita-av/booking-service/internal/model"
)

// BookingServiceInterface defines the interface for booking operations
type BookingServiceInterface interface {
	CreateBooking(ctx context.Context, userID, barberID string, startTime time.Time, serviceType model.ServiceType, notes string) (*model.Booking, error)
	GetBooking(ctx context.Context, id string) (*model.Booking, error)
	UpdateBooking(ctx context.Context, id string, startTime *time.Time, serviceType *model.ServiceType, notes *string) (*model.Booking, error)
	CancelBooking(ctx context.Context, id string) (bool, error)
	GetUserBookings(ctx context.Context, userID string) ([]*model.Booking, error)
	GetBarberBookings(ctx context.Context, barberID string, date *time.Time) ([]*model.Booking, error)
	GetAvailableTimeSlots(ctx context.Context, barberID string, date time.Time) ([]*model.TimeSlot, error)
}
