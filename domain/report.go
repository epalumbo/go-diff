package domain

// DiffResult defines the possible outcomes of comparing data
type DiffResult string

// DiffResult constants
const (
	Equal        = DiffResult("EQUAL")
	NotEqual     = DiffResult("NOT_EQUAL")
	SizeMismatch = DiffResult("SIZE_MISMATCH")
)

func (dr DiffResult) String() string {
	return string(dr)
}

// DiffInsight contains information about a single difference in the data
type DiffInsight struct {
	Offset uint
	Length uint
}

// DiffReport contains information about differences in the data
type DiffReport struct {
	Result   DiffResult
	Insights []DiffInsight
}

// DiffNotFoundError is the error returned when no data is found for a given ID
type DiffNotFoundError struct {
	ID string
}

func (e DiffNotFoundError) Error() string {
	return "diff not found for ID: " + e.ID
}
