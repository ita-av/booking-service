package repository

import (
	"context"
	"time"

	"github.com/ita-av/booking-service/internal/model"
)

// BookingRepository defines the interface for booking data operations
type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *model.Booking) (*model.Booking, error)
	GetBookingByID(ctx context.Context, id string) (*model.Booking, error)
	UpdateBooking(ctx context.Context, id string, updates map[string]interface{}) (*model.Booking, error)
	CancelBooking(ctx context.Context, id string) (bool, error)
	GetUserBookings(ctx context.Context, userID string) ([]*model.Booking, error)
	GetBarberBookings(ctx context.Context, barberID string, date *time.Time) ([]*model.Booking, error)
	GetBookingsInTimeRange(ctx context.Context, barberID string, start, end time.Time) ([]*model.Booking, error)
}
