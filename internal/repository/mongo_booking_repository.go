package repository

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/ita-av/booking-service/internal/model"
)

// BookingRepository implements repository.BookingRepository with MongoDB	
type MongoBookingRepository struct {
	collection *mongo.Collection
}

// NewBookingRepository creates a new MongoDB-backed booking repository
func NewMongoBookingRepository(db *mongo.Database) *MongoBookingRepository {
	return &MongoBookingRepository{
		collection: db.Collection("bookings"),
	}
}

// CreateBooking adds a new booking to the database
func (r *MongoBookingRepository) CreateBooking(ctx context.Context, booking *model.Booking) (*model.Booking, error) {
	// Set timestamps
	now := time.Now()
	booking.CreatedAt = now
	booking.UpdatedAt = now

	// Generate new ID if not set
	if booking.ID.IsZero() {
		booking.ID = primitive.NewObjectID()
	}

	// Insert into MongoDB
	_, err := r.collection.InsertOne(ctx, booking)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert booking")
	}

	return booking, nil
}

// GetBookingByID retrieves a booking by its ID
func (r *MongoBookingRepository) GetBookingByID(ctx context.Context, id string) (*model.Booking, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(err, "invalid booking ID format")
	}

	var booking model.Booking
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&booking)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No booking found
		}
		return nil, errors.Wrap(err, "failed to get booking")
	}

	return &booking, nil
}

// UpdateBooking updates an existing booking
func (r *MongoBookingRepository) UpdateBooking(ctx context.Context, id string, updates map[string]interface{}) (*model.Booking, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(err, "invalid booking ID format")
	}

	// Add updated timestamp
	updates["updatedAt"] = time.Now()

	update := bson.M{"$set": updates}

	// Create the options to return the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		update,
		opts,
	)

	var booking model.Booking
	if err := result.Decode(&booking); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No booking found
		}
		return nil, errors.Wrap(err, "failed to update booking")
	}

	return &booking, nil
}

// CancelBooking sets a booking's status to cancelled
func (r *MongoBookingRepository) CancelBooking(ctx context.Context, id string) (bool, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, errors.Wrap(err, "invalid booking ID format")
	}

	update := bson.M{
		"$set": bson.M{
			"status":    model.BookingStatusCancelled,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return false, errors.Wrap(err, "failed to cancel booking")
	}

	return result.ModifiedCount > 0, nil
}

// GetUserBookings retrieves all bookings for a specific user
func (r *MongoBookingRepository) GetUserBookings(ctx context.Context, userID string) ([]*model.Booking, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user bookings")
	}
	defer cursor.Close(ctx)

	var bookings []*model.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, errors.Wrap(err, "failed to decode bookings")
	}

	return bookings, nil
}

// GetBarberBookings retrieves all bookings for a specific barber
func (r *MongoBookingRepository) GetBarberBookings(ctx context.Context, barberID string, date *time.Time) ([]*model.Booking, error) {
	filter := bson.M{"barberId": barberID}

	// Add date filter if specified
	if date != nil {
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		filter["startTime"] = bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		}
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get barber bookings")
	}
	defer cursor.Close(ctx)

	var bookings []*model.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, errors.Wrap(err, "failed to decode bookings")
	}

	return bookings, nil
}

// GetBookingsInTimeRange retrieves all bookings for a barber in a time range
func (r *MongoBookingRepository) GetBookingsInTimeRange(ctx context.Context, barberID string, start, end time.Time) ([]*model.Booking, error) {
	filter := bson.M{
		"barberId": barberID,
		"status":   bson.M{"$ne": model.BookingStatusCancelled},
		"$or": []bson.M{
			{
				"startTime": bson.M{
					"$gte": start,
					"$lt":  end,
				},
			},
			{
				"endTime": bson.M{
					"$gt":  start,
					"$lte": end,
				},
			},
			{
				"startTime": bson.M{"$lte": start},
				"endTime":   bson.M{"$gte": end},
			},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bookings in time range")
	}
	defer cursor.Close(ctx)

	var bookings []*model.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, errors.Wrap(err, "failed to decode bookings")
	}

	return bookings, nil
}
