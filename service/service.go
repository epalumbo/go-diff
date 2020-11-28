package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/ehpalumbo/go-diff/domain"
)

// DiffService is the service layer implementation.
// It validates the diff side payloads and stores them for later comparison.
// It performs diffs based on previously saved payloads.
type DiffService struct {
	differ     Differ
	repository DiffRepository
}

// Differ is the contract of the diffing logic
type Differ interface {
	Diff([]byte, []byte) (domain.DiffReport, error)
}

// DiffRepository is the contract of the persistence layer
type DiffRepository interface {
	SaveDataSide(ID string, side string, data []byte) error
	GetDataSidesByID(ID string) (map[string][]byte, error)
}

// NewDiffService can be used by client code to obtain a DiffService
func NewDiffService(d Differ, rep DiffRepository) DiffService {
	return DiffService{d, rep}
}

// Save a DiffPayload for comparison
func (ds DiffService) Save(p domain.DiffPayload) error {
	if !validID(p.ID) {
		return domain.IllegalDiffPayloadError("cannot save payload without ID")
	}
	b, err := base64.StdEncoding.DecodeString(p.Value)
	if err != nil {
		return domain.IllegalDiffPayloadError("payload value is not in base64")
	}
	err = ds.repository.SaveDataSide(p.ID, p.Side.String(), b)
	if err != nil {
		return errors.New("cannot save payload: " + err.Error())
	}
	return nil
}

// GetDiffReport returns a report of the comparison with result
// and insights of the differences
func (ds DiffService) GetDiffReport(ID string) (domain.DiffReport, error) {

	var r domain.DiffReport

	if !validID(ID) {
		return r, domain.DiffNotFoundError{ID: ID}
	}

	data, err := ds.repository.GetDataSidesByID(ID)
	if err != nil {
		return r, fmt.Errorf("cannot get resource %s from storage: %v", ID, err)
	}

	left, okLeft := data[domain.LeftSide.String()]
	right, okRight := data[domain.RightSide.String()]
	if !okLeft && !okRight {
		return r, domain.DiffNotFoundError{ID: ID}
	}

	return ds.differ.Diff(nilToEmpty(left), nilToEmpty(right))
}

func validID(ID string) bool {
	return len(strings.TrimSpace(ID)) > 0
}

func nilToEmpty(b []byte) []byte {
	if b == nil {
		return []byte("")
	}
	return b
}
