package domain

import "errors"

// DiffSide is used to refer to the side of the comparison
type DiffSide string

func (ds DiffSide) String() string {
	return string(ds)
}

// ParseDiffSide returns a DiffSide if the value is a valid side
func ParseDiffSide(value string) (DiffSide, error) {
	if value == LeftSide.String() || value == RightSide.String() {
		return DiffSide(value), nil
	}
	return DiffSide(""), errors.New("invalid side value")
}

// DiffPayload contains data to upload a side of the comparison
type DiffPayload struct {
	ID    string
	Side  DiffSide
	Value string
}

// LeftSide is the left side constant
const LeftSide = DiffSide("left")

// RightSide is the right side constant
const RightSide = DiffSide("right")

// IllegalDiffPayloadError is returned when the payload contains illegal attributes
type IllegalDiffPayloadError string

func (err IllegalDiffPayloadError) Error() string {
	return string(err)
}
