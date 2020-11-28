package domain_test

import (
	"testing"

	"github.com/ehpalumbo/go-diff/domain"
)

func TestDiffRejectsInputIfOneOfTheSidesIsMissing(t *testing.T) {
	// given
	d := domain.NewDifferImpl()

	// when
	_, err := d.Diff(nil, nil)

	// then
	if err == nil {
		t.Error("accepted nil inputs")
	}
	if err.Error() != "missing input" {
		t.Errorf("wrong error message, got: %s", err)
	}
}

func TestDiffReport(t *testing.T) {
	// given
	d := domain.NewDifferImpl()

	cases := []struct {
		name        string
		left, right string
		report      domain.DiffReport
	}{
		{
			name:   "different size",
			left:   "1",
			right:  "22",
			report: domain.DiffReport{domain.SizeMismatch, nil},
		},
		{
			name:   "equal",
			left:   "123",
			right:  "123",
			report: domain.DiffReport{domain.Equal, nil},
		},
		{
			name:  "not equal",
			left:  "123456",
			right: "120006",
			report: domain.DiffReport{
				Result: domain.NotEqual,
				Insights: []domain.DiffInsight{
					{
						Offset: 2,
						Length: 3,
					},
				},
			},
		},
		{
			name:  "not equal last byte",
			left:  "123456",
			right: "123450",
			report: domain.DiffReport{
				Result: domain.NotEqual,
				Insights: []domain.DiffInsight{
					{
						Offset: 5,
						Length: 1,
					},
				},
			},
		},
		{
			name:  "not equal with many insights",
			left:  "023456",
			right: "120406",
			report: domain.DiffReport{
				Result: domain.NotEqual,
				Insights: []domain.DiffInsight{
					{
						Offset: 0,
						Length: 1,
					},
					{
						Offset: 2,
						Length: 1,
					},
					{
						Offset: 4,
						Length: 1,
					},
				},
			},
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			// when
			r, err := d.Diff([]byte(c.left), []byte(c.right))

			// then
			if err != nil {
				t.Errorf("failed to compare sides, got error: %s", err)
			}
			if r.Result != c.report.Result {
				t.Errorf("wrong result: expected: %s, got: %s", c.report.Result, r.Result)
			}
			if len(r.Insights) != len(c.report.Insights) {
				t.Errorf("wrong number of insights, expected: %d, got: %d", len(c.report.Insights), len(r.Insights))
			} else {
				for i, v := range c.report.Insights {
					if v != r.Insights[i] {
						t.Errorf("wrong insight at position %d, expected: %v, got: %v", i, v, r.Insights[i])
					}
				}
			}
		})
	}
}
