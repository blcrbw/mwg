package model

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
