//go:generate mockgen -destination mocks/repository_mock.go -package=mocks . DiffRepository
package service_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ehpalumbo/go-diff/domain"
	"github.com/ehpalumbo/go-diff/service"

	"github.com/ehpalumbo/go-diff/service/mocks"

	"github.com/golang/mock/gomock"
)

var repMock *mocks.MockDiffRepository

var svc service.DiffService

func setUp(t *testing.T) func() {
	ctrl := gomock.NewController(t)
	repMock = mocks.NewMockDiffRepository(ctrl)
	svc = service.NewDiffService(domain.NewDifferImpl(), repMock)
	return func() {
		ctrl.Finish()
	}
}

func TestServiceSavesValidDiffSide(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	cases := []struct {
		name    string
		side    domain.DiffSide
		decoded string
		encoded string
	}{
		{
			name:    "valid base64 input data",
			side:    domain.LeftSide,
			decoded: "Go go go!",
			encoded: "R28gZ28gZ28h",
		},
		{
			name:    "empty base64 input data",
			side:    domain.LeftSide,
			decoded: "",
			encoded: "",
		},
		{
			name:    "right side also accepted",
			side:    domain.RightSide,
			decoded: "Go go go!",
			encoded: "R28gZ28gZ28h",
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			// given
			repMock.EXPECT().SaveDataSide("1", c.side.String(), []byte(c.decoded)).Return(nil)

			p := domain.DiffPayload{
				ID:    "1",
				Side:  c.side,
				Value: c.encoded,
			}

			// when
			err := svc.Save(p)

			// then
			if err != nil {
				t.Errorf("failed to accept valid payload, got: %v", err)
			}
		})

	}

}

func TestServiceRejectsInvalidInputs(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	cases := []struct {
		name    string
		ID      string
		encoded string
		message string
	}{
		{
			name:    "invalid base64 input data",
			ID:      "1",
			encoded: "abc-/-xyz",
			message: "payload value is not in base64",
		},
		{
			name:    "empty ID must not be accepted",
			ID:      "",
			encoded: "R28gZ28gZ28h",
			message: "cannot save payload without ID",
		},
		{
			name:    "blank ID must not be accepted",
			ID:      " ",
			encoded: "R28gZ28gZ28h",
			message: "cannot save payload without ID",
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			// given
			p := domain.DiffPayload{
				ID:    c.ID,
				Side:  domain.LeftSide,
				Value: c.encoded,
			}

			// when
			err := svc.Save(p)

			// then
			if err == nil {
				t.Errorf("accepted invalid payload, expected: %s", c.message)
			}
			if perr, ok := err.(domain.IllegalDiffPayloadError); ok {
				if perr.Error() != c.message {
					t.Errorf("wrong error message, expected: %s, got: %v", c.message, err)
				}
			} else {
				t.Errorf("wrong error type, got: %v", err)
			}
		})

	}

}

func TestServiceRejectsPayloadIfRepositorySaveOperationFailed(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()
	// given
	repMock.EXPECT().SaveDataSide("1", "left", []byte("Go go go!")).Return(errors.New("oops"))

	p := domain.DiffPayload{
		ID:    "1",
		Side:  domain.LeftSide,
		Value: "R28gZ28gZ28h",
	}

	// when
	err := svc.Save(p)

	// then
	if err == nil {
		t.Error("did not propagate errors from repository")
	}
	if err.Error() != "cannot save payload: oops" {
		t.Errorf("wrong error message, got: %v", err)
	}
}

func TestServiceCannotProduceDiffReportIf(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	cases := []struct {
		name     string
		ID       string
		data     map[string][]byte
		err      error
		expected error
	}{
		{
			name:     "missing resource",
			data:     map[string][]byte{},
			expected: domain.DiffNotFoundError{ID: "1"},
		},
		{
			name:     "nil map",
			data:     nil,
			expected: domain.DiffNotFoundError{ID: "1"},
		},
		{
			name: "missing sides",
			data: map[string][]byte{
				"unknown": []byte("hello"),
			},
			expected: domain.DiffNotFoundError{ID: "1"},
		},
		{
			name:     "repository's get operation failed",
			err:      errors.New("oops"),
			expected: fmt.Errorf("cannot get resource %s from storage: %v", "1", "oops"),
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			// given
			repMock.EXPECT().GetDataSidesByID("1").Return(c.data, c.err)

			// when
			_, err := svc.GetDiffReport("1")

			// then
			if err == nil {
				t.Error("did not return error")
			}
			if reflect.TypeOf(err) != reflect.TypeOf(c.expected) {
				t.Errorf("wrong error type, expected: %T, got: %T", c.expected, err)
			}
			if err.Error() != c.expected.Error() {
				t.Errorf("wrong error message, expected: %v, got: %v", c.expected, err)
			}
		})

	}

}

func TestServiceGetDiffReportRejectsIllegalID(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	cases := []struct {
		name     string
		ID       string
		expected error
	}{
		{
			name:     "empty",
			ID:       "",
			expected: domain.DiffNotFoundError{ID: ""},
		},
		{
			name:     "blank",
			ID:       " ",
			expected: domain.DiffNotFoundError{ID: " "},
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			_, err := svc.GetDiffReport(c.ID)

			if err == nil {
				t.Error("did not return error")
			}
			if reflect.TypeOf(err) != reflect.TypeOf(c.expected) {
				t.Errorf("wrong error type, expected: %T, got: %T", c.expected, err)
			}
			if err.Error() != c.expected.Error() {
				t.Errorf("wrong error, expected: %v, got: %v", c.expected, err)
			}
		})

	}

}

func TestServiceProducesDiffReportIf(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	cases := []struct {
		name     string
		data     map[string][]byte
		expected domain.DiffReport
	}{
		{
			name: "equal sides",
			data: map[string][]byte{
				"left":  []byte("hello"),
				"right": []byte("hello"),
			},
			expected: domain.DiffReport{
				Result: domain.Equal,
			},
		},
		{
			name: "non equal sides",
			data: map[string][]byte{
				"left":  []byte("hello"),
				"right": []byte("hallo"),
			},
			expected: domain.DiffReport{
				Result: domain.NotEqual,
				Insights: []domain.DiffInsight{
					{
						Offset: 1,
						Length: 1,
					},
				},
			},
		},
		{
			name: "size mismatch",
			data: map[string][]byte{
				"left":  []byte("hello"),
				"right": []byte("hell"),
			},
			expected: domain.DiffReport{
				Result: domain.SizeMismatch,
			},
		},
		{
			name: "missing right side",
			data: map[string][]byte{
				"left": []byte("hello"),
			},
			expected: domain.DiffReport{
				Result: domain.SizeMismatch,
			},
		},
		{
			name: "missing left side",
			data: map[string][]byte{
				"right": []byte("hello"),
			},
			expected: domain.DiffReport{
				Result: domain.SizeMismatch,
			},
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			// given
			repMock.EXPECT().GetDataSidesByID("1").Return(c.data, nil)

			// when
			r, err := svc.GetDiffReport("1")

			// then
			if err != nil {
				t.Errorf("failed with error: %v", err)
			}
			if r.Result != c.expected.Result {
				t.Errorf("returned wrong result, expected: %s, got: %s", c.expected.Result, r.Result)
			}
			if len(r.Insights) != len(c.expected.Insights) {
				t.Errorf("wrong number of insights, expected: %d, got: %d", len(c.expected.Insights), len(r.Insights))
			} else {
				for i, v := range c.expected.Insights {
					if v != r.Insights[i] {
						t.Errorf("wrong insight at position %d, expected: %v, got: %v", i, v, r.Insights[i])
					}
				}
			}
		})

	}

}
