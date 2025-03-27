package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ita-av/booking-service/internal/auth"
	"github.com/ita-av/booking-service/internal/model"
	"github.com/ita-av/booking-service/internal/service"
	pb "github.com/ita-av/booking-service/pkg/api/proto"
)

// MockBookingService is a mock implementation of the booking service
type MockBookingService struct {
	mock.Mock
}

var _ service.BookingServiceInterface = (*MockBookingService)(nil)

// Implement all service methods...

func (m *MockBookingService) CreateBooking(ctx context.Context, userID, barberID string, startTime time.Time, serviceType model.ServiceType, notes string) (*model.Booking, error) {
	args := m.Called(ctx, userID, barberID, startTime, serviceType, notes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Booking), args.Error(1)
}

func (m *MockBookingService) GetBooking(ctx context.Context, id string) (*model.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Booking), args.Error(1)
}

func (m *MockBookingService) UpdateBooking(ctx context.Context, id string, startTime *time.Time, serviceType *model.ServiceType, notes *string) (*model.Booking, error) {
	args := m.Called(ctx, id, startTime, serviceType, notes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Booking), args.Error(1)
}

func (m *MockBookingService) CancelBooking(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockBookingService) GetUserBookings(ctx context.Context, userID string) ([]*model.Booking, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Booking), args.Error(1)
}

func (m *MockBookingService) GetBarberBookings(ctx context.Context, barberID string, date *time.Time) ([]*model.Booking, error) {
	args := m.Called(ctx, barberID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Booking), args.Error(1)
}

func (m *MockBookingService) GetAvailableTimeSlots(ctx context.Context, barberID string, date time.Time) ([]*model.TimeSlot, error) {
	args := m.Called(ctx, barberID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.TimeSlot), args.Error(1)
}

// Mock context with user claims
func mockContextWithClaims(userID string, isBarber bool) context.Context {
	claims := &auth.Claims{
		IsBarber: isBarber,
	}
	claims.Subject = userID
	return context.WithValue(context.Background(), "user_claims", claims)
}

// Test: Regular user creates booking for themselves (should succeed)
func TestCreateBooking_RegularUserForSelf(t *testing.T) {
	mockService := new(MockBookingService)
	server := &BookingServer{service: mockService}

	// Create test data
	objectID := primitive.NewObjectID()
	startTime := time.Now().Round(time.Second)
	endTime := startTime.Add(30 * time.Minute)

	booking := &model.Booking{
		ID:          objectID,
		UserID:      "user1",
		BarberID:    "barber1",
		StartTime:   startTime,
		EndTime:     endTime,
		ServiceType: model.ServiceTypeHaircut,
		Status:      model.BookingStatusPending,
		Notes:       "Test booking",
		CreatedAt:   startTime,
		UpdatedAt:   startTime,
	}

	// Set up mock expectations
	mockService.On("CreateBooking",
		mock.Anything,
		"user1",
		"barber1",
		mock.AnythingOfType("time.Time"),
		model.ServiceTypeHaircut,
		"Test booking").Return(booking, nil)

	// Create the request
	req := &pb.CreateBookingRequest{
		UserId:      "user1",
		BarberId:    "barber1",
		StartTime:   startTime.Format(time.RFC3339),
		ServiceType: pb.ServiceType_HAIRCUT,
		Notes:       "Test booking",
	}

	// Create context with claims (regular user)
	ctx := mockContextWithClaims("user1", false)

	// Call the method
	resp, err := server.CreateBooking(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, objectID.Hex(), resp.Id)
}

// Test: Regular user tries to create booking for another user (should fail)
func TestCreateBooking_RegularUserForOther(t *testing.T) {
	mockService := new(MockBookingService)
	server := &BookingServer{service: mockService}

	// Create the request
	req := &pb.CreateBookingRequest{
		UserId:      "user2", // Different from authenticated user
		BarberId:    "barber1",
		StartTime:   time.Now().Format(time.RFC3339),
		ServiceType: pb.ServiceType_HAIRCUT,
	}

	// Create context with claims (regular user)
	ctx := mockContextWithClaims("user1", false)

	// Call the method
	resp, err := server.CreateBooking(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Verify that the error is permission denied
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())

	// Verify that the service was never called
	mockService.AssertNotCalled(t, "CreateBooking")
}

// Test: Barber creates booking for another user (should succeed)
func TestCreateBooking_BarberForOther(t *testing.T) {
	mockService := new(MockBookingService)
	server := &BookingServer{service: mockService}

	// Create test data
	objectID := primitive.NewObjectID()
	startTime := time.Now().Round(time.Second)
	endTime := startTime.Add(30 * time.Minute)

	booking := &model.Booking{
		ID:          objectID,
		UserID:      "user1",
		BarberID:    "barber1",
		StartTime:   startTime,
		EndTime:     endTime,
		ServiceType: model.ServiceTypeHaircut,
		Status:      model.BookingStatusPending,
		CreatedAt:   startTime,
		UpdatedAt:   startTime,
	}

	// Set up mock expectations
	mockService.On("CreateBooking",
		mock.Anything,
		"user1",
		"barber1",
		mock.AnythingOfType("time.Time"),
		model.ServiceTypeHaircut,
		"").Return(booking, nil)

	// Create the request
	req := &pb.CreateBookingRequest{
		UserId:      "user1", // Different from authenticated user (barber2)
		BarberId:    "barber1",
		StartTime:   startTime.Format(time.RFC3339),
		ServiceType: pb.ServiceType_HAIRCUT,
	}

	// Create context with claims (barber)
	ctx := mockContextWithClaims("barber2", true)

	// Call the method
	resp, err := server.CreateBooking(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, objectID.Hex(), resp.Id)
}

// Test: Regular user tries to get barber bookings (should fail)
func TestGetBarberBookings_RegularUser(t *testing.T) {
	mockService := new(MockBookingService)
	server := &BookingServer{service: mockService}

	// Create the request
	req := &pb.GetBarberBookingsRequest{
		BarberId: "barber1",
	}

	// Create context with claims (regular user)
	ctx := mockContextWithClaims("user1", false)

	// Call the method
	resp, err := server.GetBarberBookings(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Verify that the error is permission denied
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())

	// Verify that the service was never called
	mockService.AssertNotCalled(t, "GetBarberBookings")
}

// Test: Barber gets barber bookings (should succeed)
func TestGetBarberBookings_Barber(t *testing.T) {
	mockService := new(MockBookingService)
	server := &BookingServer{service: mockService}

	// Create test data
	startTime := time.Now().Round(time.Second)
	bookings := []*model.Booking{
		{
			ID:        primitive.NewObjectID(),
			UserID:    "user1",
			BarberID:  "barber1",
			StartTime: startTime,
			EndTime:   startTime.Add(30 * time.Minute),
		},
	}

	// Set up mock expectations
	mockService.On("GetBarberBookings",
		mock.Anything,
		"barber1",
		mock.Anything).Return(bookings, nil)

	// Create the request
	req := &pb.GetBarberBookingsRequest{
		BarberId: "barber1",
	}

	// Create context with claims (barber)
	ctx := mockContextWithClaims("barber2", true)

	// Call the method
	resp, err := server.GetBarberBookings(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Bookings, 1)
}

// Test: Regular user tries to view another user's bookings (should fail)
func TestGetUserBookings_RegularUserForOther(t *testing.T) {
	mockService := new(MockBookingService)
	server := &BookingServer{service: mockService}

	// Create the request
	req := &pb.GetUserBookingsRequest{
		UserId: "user2", // Different from authenticated user
	}

	// Create context with claims (regular user)
	ctx := mockContextWithClaims("user1", false)

	// Call the method
	resp, err := server.GetUserBookings(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Verify that the error is permission denied
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())

	// Verify that the service was never called
	mockService.AssertNotCalled(t, "GetUserBookings")
}

// Test: Regular user views their own bookings (should succeed)
func TestGetUserBookings_RegularUserForSelf(t *testing.T) {
	mockService := new(MockBookingService)
	server := &BookingServer{service: mockService}

	// Create test data
	startTime := time.Now().Round(time.Second)
	bookings := []*model.Booking{
		{
			ID:        primitive.NewObjectID(),
			UserID:    "user1",
			BarberID:  "barber1",
			StartTime: startTime,
			EndTime:   startTime.Add(30 * time.Minute),
		},
	}

	// Set up mock expectations
	mockService.On("GetUserBookings",
		mock.Anything,
		"user1").Return(bookings, nil)

	// Create the request
	req := &pb.GetUserBookingsRequest{
		UserId: "user1",
	}

	// Create context with claims (regular user)
	ctx := mockContextWithClaims("user1", false)

	// Call the method
	resp, err := server.GetUserBookings(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Bookings, 1)
}

// Test: Barber views another user's bookings (should succeed)
func TestGetUserBookings_BarberForOther(t *testing.T) {
	mockService := new(MockBookingService)
	server := &BookingServer{service: mockService}

	// Create test data
	startTime := time.Now().Round(time.Second)
	bookings := []*model.Booking{
		{
			ID:        primitive.NewObjectID(),
			UserID:    "user1",
			BarberID:  "barber1",
			StartTime: startTime,
			EndTime:   startTime.Add(30 * time.Minute),
		},
	}

	// Set up mock expectations
	mockService.On("GetUserBookings",
		mock.Anything,
		"user1").Return(bookings, nil)

	// Create the request
	req := &pb.GetUserBookingsRequest{
		UserId: "user1",
	}

	// Create context with claims (barber)
	ctx := mockContextWithClaims("barber2", true)

	// Call the method
	resp, err := server.GetUserBookings(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Bookings, 1)
}
