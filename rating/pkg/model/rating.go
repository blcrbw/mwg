package model

import "fmt"

// RecordId defines a record id. Together with RecordType
// identifies unique record across all types.
type RecordId string

// RecordType defines a record type. Together with RecordId
// identifies unique record across all types.
type RecordType string

// Existing record types.
const (
	RecordTypeMovie = RecordType("movie")
)

// UserId defines a user id.
type UserId string

// RatingValue defines a value of a rating record.
type RatingValue int

// Rating defines an individual rating created by a user for some record.
type Rating struct {
	RecordId   string      `json:"recordId" validate:"required"`
	RecordType string      `json:"recordType" validate:"required"`
	UserId     UserId      `json:"userId" validate:"required"`
	Value      RatingValue `json:"value" validate:"gte=1,lte=5"`
}

func (r *Rating) String() string {
	return fmt.Sprintf("Rating{recordId=%s, recordType=%s, UserId=%s, Value=%d}", r.RecordId, r.RecordType, r.UserId, r.Value)
}

// RatingEvent defines an event containing rating information.
type RatingEvent struct {
	Rating
	ProviderId string          `json:"providerId" validate:"required"`
	EventType  RatingEventType `json:"eventType" validate:"required"`
}

func (ev *RatingEvent) String() string {
	return fmt.Sprintf("RatingEvent{Rating=%s, ProviderId=%s, EventType=%s}", ev.Rating.String(), ev.ProviderId, ev.EventType)
}

// RatingEventType defines the type of rating event.
type RatingEventType string

// Rating event types.
const (
	RatingEventTypePut    = RatingEventType("put")
	RatingEventTypeDelete = RatingEventType("delete")
)
