package domain_test

import (
	"testing"

	"github.com/ehpalumbo/go-diff/domain"
)

func TestDiffResultConstants(t *testing.T) {
	expected := map[domain.DiffResult]string{
		domain.Equal:        "EQUAL",
		domain.NotEqual:     "NOT_EQUAL",
		domain.SizeMismatch: "SIZE_MISMATCH",
	}
	for r, v := range expected {
		actual := r.String()
		if actual != v {
			t.Errorf("%s NOK, expected %s, got %s", r, v, actual)
		}
	}
}

func TestCreateDiffNotFoundErrorOutOfID(t *testing.T) {
	err := domain.DiffNotFoundError{"1"}
	if err.Error() != "diff not found for ID: 1" {
		t.Error("DiffNotFoundError does not generate expected error message")
	}
}
