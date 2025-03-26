# Booking Service

Booking service for a barbershop management system built with Go, gRPC, and MongoDB.

## Features

- Create, retrieve, update, and cancel bookings
- Manage user and barber booking histories
- TODO: Check available time slots

## Technologies

- Language: Go
- RPC Framework: gRPC
- Database: MongoDB
- Containerization: Docker

## Prerequisites

- Go 1.24+
- Docker
- Docker Compose

## Local Development Setup

### Clone Repository

```bash
git clone https://github.com/ita-av/booking-service.git
cd booking-service
```

### Configuration

Configuration is managed through environment variables:

- `SERVER_PORT`: gRPC server listening port
- `MONGO_URI`: MongoDB connection string
- `MONGO_DB`: Database name
- `LOG_LEVEL`: Logging verbosity (debug, info, warn, error)

### Running with Docker Compose

```bash
docker-compose up --build
```

### Local Development (without Docker)

```bash
# Install dependencies
go mod download

# Run the service
go run cmd/main.go
```

## gRPC Methods

### CreateBooking

Create a new booking

- Input: User ID, Barber ID, Start Time, Service Type
- Output: Created Booking Details

### GetBooking

Retrieve booking details by ID

### UpdateBooking

Modify an existing booking

### CancelBooking

Cancel a specific booking

### GetUserBookings

Fetch all bookings for a user

### GetBarberBookings

Retrieve bookings for a specific barber

### GetAvailableTimeSlots

Find available booking slots for a barber
