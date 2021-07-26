//go:generate mockgen -destination mocks/service_mock.go -package=mocks . DiffService
package api_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ehpalumbo/go-diff/api"
	"github.com/ehpalumbo/go-diff/api/mocks"
	"github.com/ehpalumbo/go-diff/domain"
	"github.com/golang/mock/gomock"
)

var svcMock *mocks.MockDiffService

var app api.Application

func setUp(t *testing.T) func() {
	ctrl := gomock.NewController(t)
	svcMock = mocks.NewMockDiffService(ctrl)
	app = api.NewApplication(svcMock)
	return func() {
		ctrl.Finish()
	}
}

func TestSaveRejectsIllegalDiffSide(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	// given
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		PathParameters: map[string]string{
			"id":   "1",
			"side": "wrong", // illegal
		},
		Body: `{"data": "abc"}`,
	}

	// when
	res := app.Handle(req)

	// then
	if res.StatusCode != 404 {
		t.Error("accepted wrong side in URI")
	}
}

func TestSaveRejectsRequestWithoutProperJSON(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	// given
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		PathParameters: map[string]string{
			"id":   "1",
			"side": "left",
		},
		Body: "invalid input, not JSON", // illegal
	}

	// when
	res := app.Handle(req)

	// then
	if res.StatusCode != 400 {
		t.Error("accepted invalid JSON")
	}

	var body struct {
		ID     string `json:"id"`
		Reason string `json:"reason"`
	}
	json.Unmarshal([]byte(res.Body), &body)
	if body.ID != "1" {
		t.Errorf("wrong ID in error response: %s", body.ID)
	}
	if body.Reason != "invalid body" {
		t.Errorf("wrong error reason: %s", body.Reason)
	}
}

func TestSaveRejectsRequestWithoutDataInPayload(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	// given
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		PathParameters: map[string]string{
			"id":   "1",
			"side": "left",
		},
		Body: "{}", // illegal
	}

	// when
	res := app.Handle(req)

	// then
	if res.StatusCode != 400 {
		t.Error("accepted JSON without data")
	}

	var body struct {
		Reason string `json:"reason"`
	}
	json.Unmarshal([]byte(res.Body), &body)
	if body.Reason != "missing data" {
		t.Errorf("wrong error reason: %s", body.Reason)
	}
}

func TestSaveFailsWith500WhenServiceSaveFails(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	// given
	expectedPayload := domain.DiffPayload{
		ID:    "1",
		Side:  domain.DiffSide("left"),
		Value: "abc",
	}
	svcMock.EXPECT().Save(expectedPayload).Return(errors.New("oops"))

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		PathParameters: map[string]string{
			"id":   "1",
			"side": "left",
		},
		Body: `{"data": "abc"}`,
	}

	// when
	res := app.Handle(req)

	// then
	if res.StatusCode != 500 {
		t.Error("did not fail with 500 when service operation failed")
	}

	var body struct {
		ID     string `json:"id"`
		Reason string `json:"reason"`
		Cause  string `json:"cause"`
	}
	if err := json.Unmarshal([]byte(res.Body), &body); err != nil {
		t.Errorf("return response does not match expected JSON response, got: %s", res.Body)
	}
	if body.ID != "1" {
		t.Errorf("wrong ID in error response: %s", body.ID)
	}
	if body.Reason != "save operation failed" {
		t.Errorf("wrong error reason: %s", body.Reason)
	}
	if body.Cause != "oops" {
		t.Errorf("wrong error cause: %s", body.Cause)
	}
}

func TestSaveSuccess(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	// given
	expectedPayload := domain.DiffPayload{
		ID:    "1",
		Side:  domain.DiffSide("right"),
		Value: "abc",
	}
	svcMock.EXPECT().Save(expectedPayload).Return(nil)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		PathParameters: map[string]string{
			"id":   "1",
			"side": "right",
		},
		Body: `{"data": "abc"}`,
	}

	// when
	res := app.Handle(req)

	// then
	if res.StatusCode != 204 {
		t.Errorf("failed with status %v", res.StatusCode)
	}
}

func TestDiffNotFound(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	// given
	svcMock.EXPECT().GetDiffReport("1").Return(domain.DiffReport{}, domain.DiffNotFoundError{ID: "1"})

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		PathParameters: map[string]string{
			"id": "1",
		},
	}

	// when
	res := app.Handle(req)

	// then
	if res.StatusCode != 404 {
		t.Errorf("accepted ID that should not have been found by service, got status code: %d", res.StatusCode)
	}

	var body struct {
		ID     string `json:"id"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(res.Body), &body); err != nil {
		t.Errorf("returned error response does not fit expected JSON response, got: %s", res.Body)
	}
	if body.ID != "1" {
		t.Errorf("wrong ID in not found response, got: %s", body.ID)
	}
	if body.Reason != "diff not found" {
		t.Errorf("wrong reason in not found response, got: %s", body.Reason)
	}
}

func TestGetDiffFailed(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	// given
	svcMock.EXPECT().GetDiffReport("1").Return(domain.DiffReport{}, errors.New("oops"))

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		PathParameters: map[string]string{
			"id": "1",
		},
	}

	// when
	res := app.Handle(req)

	// then
	if res.StatusCode != 500 {
		t.Errorf("accepted ID that should not have been accepted by service, got status code: %d", res.StatusCode)
	}

	var body struct {
		ID     string `json:"id"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(res.Body), &body); err != nil {
		t.Errorf("returned error response does not fit expected JSON response, got: %s", res.Body)
	}
	if body.ID != "1" {
		t.Errorf("wrong ID in not found response, got: %s", body.ID)
	}
	if body.Reason != "get diff failed" {
		t.Errorf("wrong reason in not found response, got: %s", body.Reason)
	}
}

func TestGetDiffReportSuccess(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	// given
	r := domain.DiffReport{
		Result: domain.NotEqual,
		Insights: []domain.DiffInsight{
			{
				Length: 1,
				Offset: 2,
			},
		},
	}
	svcMock.EXPECT().GetDiffReport("1").Return(r, nil)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		PathParameters: map[string]string{
			"id": "1",
		},
	}

	// when
	res := app.Handle(req)

	// then
	if res.StatusCode != 200 {
		t.Errorf("rejected ID that should have been accepted by service, got status code: %d", res.StatusCode)
	}

	ct := res.Headers["Content-Type"]
	if ct != "application/json" {
		t.Errorf("wrong content type header: %s", ct)
	}

	var body struct {
		Result   string `json:"result"`
		Insights []struct {
			Length uint `json:"length"`
			Offset uint `json:"offset"`
		} `json:"insights"`
	}
	if err := json.Unmarshal([]byte(res.Body), &body); err != nil {
		t.Errorf("returned error response does not fit expected JSON response, got: %s", res.Body)
	}
	if body.Result != "NOT_EQUAL" {
		t.Errorf("wrong result in response, got: %s", body.Result)
	}
	if len(body.Insights) != 1 {
		t.Errorf("expected 1 diff insight, got: %d", len(body.Insights))
	}
	insight := body.Insights[0]
	if insight.Length != 1 {
		t.Errorf("wrong diff insight length, got: %d", insight.Length)
	}
	if insight.Offset != 2 {
		t.Errorf("wrong diff insight offset, got: %d", insight.Offset)
	}
}

func TestGetDiffReportSuccessWithoutInsights(t *testing.T) {
	tearDown := setUp(t)
	defer tearDown()

	// given
	r := domain.DiffReport{
		Result:   domain.SizeMismatch,
		Insights: []domain.DiffInsight{},
	}
	svcMock.EXPECT().GetDiffReport("1").Return(r, nil)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		PathParameters: map[string]string{
			"id": "1",
		},
	}

	// when
	res := app.Handle(req)

	// then
	if res.StatusCode != 200 {
		t.Errorf("rejected ID that should have been accepted by service, got status code: %d", res.StatusCode)
	}

	var body struct {
		Result   string     `json:"result"`
		Insights []struct{} `json:"insights"`
	}
	if err := json.Unmarshal([]byte(res.Body), &body); err != nil {
		t.Errorf("returned error response does not fit expected JSON response, got: %s", res.Body)
	}
	if body.Result != "SIZE_MISMATCH" {
		t.Errorf("wrong result in response, got: %s", body.Result)
	}
	if body.Insights != nil {
		t.Errorf("expected no diff insights, got: %v", body.Insights)
	}
}
