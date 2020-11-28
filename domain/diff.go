package domain

import "errors"

// DifferImpl is the implementation of the diff logic between two binary streams
type DifferImpl struct {
}

// NewDifferImpl creates a default DifferImpl
func NewDifferImpl() *DifferImpl {
	return &DifferImpl{}
}

type counter struct {
	offset, length uint
	insights       []DiffInsight
}

func (c *counter) count(equal bool) {
	if equal {
		c.save()
		c.offset++
	} else {
		c.length++
	}
}

func (c *counter) save() {
	if c.length > 0 {
		d := DiffInsight{
			Offset: c.offset,
			Length: c.length,
		}
		c.insights = append(c.insights, d)
		c.offset += c.length
		c.length = 0
	}
}

// Diff compares two byte slices and returns a DiffReport with insights on the differences
func (d *DifferImpl) Diff(left, right []byte) (DiffReport, error) {
	var r DiffReport

	if left == nil || right == nil {
		return r, errors.New("missing input")
	}

	if len(left) != len(right) {
		r.Result = SizeMismatch
		return r, nil
	}

	r = generateReport(left, right)

	return r, nil
}

func generateReport(left, right []byte) DiffReport {
	var c counter
	for i := range left {
		c.count(left[i] == right[i])
	}
	c.save()
	return toDiffReport(c)
}

func toDiffReport(c counter) (r DiffReport) {
	if len(c.insights) == 0 {
		r.Result = Equal
	} else {
		r.Result = NotEqual
		r.Insights = c.insights
	}
	return
}
