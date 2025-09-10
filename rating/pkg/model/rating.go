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
	RecordId   string      `json:"recordId,"`
	RecordType string      `json:"recordType"`
	UserId     UserId      `json:"userId"`
	Value      RatingValue `json:"value"`
}

func (r *Rating) String() string {
	return fmt.Sprintf("Rating{recordId=%s, recordType=%s, UserId=%s, Value=%d}", r.RecordId, r.RecordType, r.UserId, r.Value)
}

// RatingEvent defines an event containing rating information.
type RatingEvent struct {
	Rating
	ProviderId string          `json:"providerId"`
	EventType  RatingEventType `json:"eventType"`
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
