package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ServiceType represents the type of barbershop service
type ServiceType int

// BookingStatus represents the status of a booking
type BookingStatus int

// Constants for ServiceType
const (
	ServiceTypeHaircut ServiceType = iota
	ServiceTypeBeardTrim
	ServiceTypeHairWash
	ServiceTypeFullService
)

// Constants for BookingStatus
const (
	BookingStatusPending BookingStatus = iota
	BookingStatusConfirmed
	BookingStatusCancelled
	BookingStatusCompleted
)

// Booking represents a barbershop appointment
type Booking struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"userId" json:"userId"`
	BarberID    string             `bson:"barberId" json:"barberId"`
	StartTime   time.Time          `bson:"startTime" json:"startTime"`
	EndTime     time.Time          `bson:"endTime" json:"endTime"`
	ServiceType ServiceType        `bson:"serviceType" json:"serviceType"`
	Status      BookingStatus      `bson:"status" json:"status"`
	Notes       string             `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// TimeSlot represents an available time slot for booking
type TimeSlot struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// GetDuration returns the duration for a service type in minutes
func (s ServiceType) GetDuration() int {
	switch s {
	case ServiceTypeHaircut:
		return 30
	case ServiceTypeBeardTrim:
		return 15
	case ServiceTypeHairWash:
		return 20
	case ServiceTypeFullService:
		return 60
	default:
		return 30
	}
}

// CalculateEndTime calculates the end time based on the start time and service type
func CalculateEndTime(startTime time.Time, serviceType ServiceType) time.Time {
	duration := serviceType.GetDuration()
	return startTime.Add(time.Minute * time.Duration(duration))
}
