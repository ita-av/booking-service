package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/ita-av/booking-service/internal/model"
	"github.com/ita-av/booking-service/internal/repository"
)

// BookingService handles business logic for bookings
type BookingService struct {
	repo repository.BookingRepository
}

var _ BookingServiceInterface = (*BookingService)(nil)

// NewBookingService creates a new booking service
func NewBookingService(repo repository.BookingRepository) *BookingService {
	return &BookingService{
		repo: repo,
	}
}

// CreateBooking creates a new booking
func (s *BookingService) CreateBooking(ctx context.Context, userID, barberID string, startTime time.Time, serviceType model.ServiceType, notes string) (*model.Booking, error) {
	// Check if the barber is available at the requested time
	endTime := model.CalculateEndTime(startTime, serviceType)

	existingBookings, err := s.repo.GetBookingsInTimeRange(ctx, barberID, startTime, endTime)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check barber availability")
	}

	if len(existingBookings) > 0 {
		return nil, errors.New("barber is not available at the requested time")
	}

	// Create the booking
	booking := &model.Booking{
		UserID:      userID,
		BarberID:    barberID,
		StartTime:   startTime,
		EndTime:     endTime,
		ServiceType: serviceType,
		Status:      model.BookingStatusPending,
		Notes:       notes,
	}

	createdBooking, err := s.repo.CreateBooking(ctx, booking)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create booking")
	}

	log.Info().
		Str("bookingID", createdBooking.ID.Hex()).
		Str("userID", userID).
		Str("barberID", barberID).
		Time("startTime", startTime).
		Msg("Booking created successfully")

	return createdBooking, nil
}

// GetBooking retrieves a booking by ID
func (s *BookingService) GetBooking(ctx context.Context, id string) (*model.Booking, error) {
	booking, err := s.repo.GetBookingByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get booking")
	}

	if booking == nil {
		return nil, errors.New("booking not found")
	}

	return booking, nil
}

// UpdateBooking updates an existing booking
func (s *BookingService) UpdateBooking(ctx context.Context, id string, startTime *time.Time, serviceType *model.ServiceType, notes *string) (*model.Booking, error) {
	// Get the existing booking
	existingBooking, err := s.repo.GetBookingByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get booking for update")
	}

	if existingBooking == nil {
		return nil, errors.New("booking not found")
	}

	// Prepare updates
	updates := map[string]interface{}{}

	if startTime != nil {
		updates["startTime"] = *startTime

		// Recalculate end time if start time or service type changes
		newServiceType := existingBooking.ServiceType
		if serviceType != nil {
			newServiceType = *serviceType
		}

		endTime := model.CalculateEndTime(*startTime, newServiceType)
		updates["endTime"] = endTime

		// Check availability
		bookings, err := s.repo.GetBookingsInTimeRange(ctx, existingBooking.BarberID, *startTime, endTime)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check barber availability")
		}

		// Filter out the current booking from results
		for i, b := range bookings {
			if b.ID == existingBooking.ID {
				bookings = append(bookings[:i], bookings[i+1:]...)
				break
			}
		}

		if len(bookings) > 0 {
			return nil, errors.New("barber is not available at the requested time")
		}
	}

	if serviceType != nil {
		updates["serviceType"] = *serviceType

		// Recalculate end time if service type changes but start time doesn't
		if startTime == nil {
			endTime := model.CalculateEndTime(existingBooking.StartTime, *serviceType)
			updates["endTime"] = endTime

			// Check availability with the new end time
			bookings, err := s.repo.GetBookingsInTimeRange(ctx, existingBooking.BarberID, existingBooking.StartTime, endTime)
			if err != nil {
				return nil, errors.Wrap(err, "failed to check barber availability")
			}

			// Filter out the current booking from results
			for i, b := range bookings {
				if b.ID == existingBooking.ID {
					bookings = append(bookings[:i], bookings[i+1:]...)
					break
				}
			}

			if len(bookings) > 0 {
				return nil, errors.New("barber is not available for the requested service duration")
			}
		}
	}

	if notes != nil {
		updates["notes"] = *notes
	}

	// Update the booking
	updatedBooking, err := s.repo.UpdateBooking(ctx, id, updates)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update booking")
	}

	log.Info().
		Str("bookingID", id).
		Msg("Booking updated successfully")

	return updatedBooking, nil
}

// CancelBooking cancels a booking
func (s *BookingService) CancelBooking(ctx context.Context, id string) (bool, error) {
	success, err := s.repo.CancelBooking(ctx, id)
	if err != nil {
		return false, errors.Wrap(err, "failed to cancel booking")
	}

	if success {
		log.Info().
			Str("bookingID", id).
			Msg("Booking cancelled successfully")
	} else {
		log.Info().
			Str("bookingID", id).
			Msg("Booking not found or already cancelled")
	}

	return success, nil
}

// GetUserBookings retrieves all bookings for a user
func (s *BookingService) GetUserBookings(ctx context.Context, userID string) ([]*model.Booking, error) {
	bookings, err := s.repo.GetUserBookings(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user bookings")
	}

	return bookings, nil
}

// GetBarberBookings retrieves all bookings for a barber
func (s *BookingService) GetBarberBookings(ctx context.Context, barberID string, date *time.Time) ([]*model.Booking, error) {
	bookings, err := s.repo.GetBarberBookings(ctx, barberID, date)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get barber bookings")
	}

	return bookings, nil
}

// GetAvailableTimeSlots gets available time slots for a barber on a specific day
func (s *BookingService) GetAvailableTimeSlots(ctx context.Context, barberID string, date time.Time) ([]*model.TimeSlot, error) {
	// Define working hours (9 AM to 5 PM)
	workStartHour := 9
	workEndHour := 17

	// Create start and end of the day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	workStart := time.Date(date.Year(), date.Month(), date.Day(), workStartHour, 0, 0, 0, date.Location())
	workEnd := time.Date(date.Year(), date.Month(), date.Day(), workEndHour, 0, 0, 0, date.Location())

	// Get all bookings for the barber on that day
	bookings, err := s.repo.GetBarberBookings(ctx, barberID, &startOfDay)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get barber bookings")
	}

	// Create 30-minute time slots
	slotDuration := 30 * time.Minute
	var availableSlots []*model.TimeSlot

	for slotStart := workStart; slotStart.Before(workEnd); slotStart = slotStart.Add(slotDuration) {
		slotEnd := slotStart.Add(slotDuration)

		// Check if this slot overlaps with any booking
		isAvailable := true
		for _, booking := range bookings {
			if booking.Status == model.BookingStatusCancelled {
				continue
			}

			if (slotStart.Equal(booking.StartTime) || slotStart.After(booking.StartTime)) && slotStart.Before(booking.EndTime) ||
				(slotEnd.After(booking.StartTime) && (slotEnd.Equal(booking.EndTime) || slotEnd.Before(booking.EndTime))) ||
				(slotStart.Before(booking.StartTime) && slotEnd.After(booking.EndTime)) {
				isAvailable = false
				break
			}
		}

		if isAvailable {
			availableSlots = append(availableSlots, &model.TimeSlot{
				StartTime: slotStart,
				EndTime:   slotEnd,
			})
		}
	}

	return availableSlots, nil
}
