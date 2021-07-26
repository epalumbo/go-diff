package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ehpalumbo/go-diff/api"
	"github.com/ehpalumbo/go-diff/repository/fake"
)

type DiffResponseBody struct {
	Result   string `json:"result" binding:"required"`
	Insights []struct {
		Offset uint `json:"offset"`
		Length uint `json:"length"`
	} `json:"insights"`
}

var app api.Application

func TestMain(m *testing.M) {
	app = RunApplication(fake.NewFakeDiffRepository())
	c := m.Run()
	os.Exit(c)
}

func TestDiffNotEqual(t *testing.T) {

	upload(t, "1", "left", "R29sYW5n") // "golang"

	upload(t, "1", "right", "R29sb25n") // "golong"

	diff := diff(t, "1")

	if diff.Result != "NOT_EQUAL" {
		t.Errorf("got wrong result: %s", diff.Result)
	}
	if len(diff.Insights) != 1 {
		t.Errorf("got wrong number of insights: %d", len(diff.Insights))
	}
	i := diff.Insights[0]
	if i.Offset != 3 || i.Length != 1 {
		t.Errorf("got wrong insight: %v", i)
	}

}

func TestDiffEqual(t *testing.T) {

	upload(t, "2", "left", "R29sYW5n")

	upload(t, "2", "right", "R29sYW5n")

	diff := diff(t, "2")

	if diff.Result != "EQUAL" {
		t.Errorf("got wrong result: %s", diff.Result)
	}
	if diff.Insights != nil {
		t.Errorf("got insights for non-diffable set: %v", diff.Insights)
	}

}

func TestDiffWithSizeMismatch(t *testing.T) {

	upload(t, "3", "left", "R29sYW5n")
	upload(t, "3", "right", "R29sYW5kIHJvY2tz")

	diff := diff(t, "3")

	if diff.Result != "SIZE_MISMATCH" {
		t.Errorf("got wrong result: %s", diff.Result)
	}

}

func TestDiffWithMissingSide(t *testing.T) {

	upload(t, "4", "left", "R29sYW5n")

	diff := diff(t, "4")

	if diff.Result != "SIZE_MISMATCH" {
		t.Errorf("got wrong result: %s", diff.Result)
	}

}

func TestMissingDiff(t *testing.T) {

	r := performGET(t, "5")

	if r.StatusCode != 404 {
		t.Fatalf("accepted missing diff ID, got status: %d", r.StatusCode)
	}

}

func TestDiffWithOverride(t *testing.T) {

	upload(t, "6", "left", "R29sYW5n")
	upload(t, "6", "right", "R29sYW5kIHJvY2tz")

	upload(t, "6", "right", "R29sYW5n")

	diff := diff(t, "6")

	if diff.Result != "EQUAL" {
		t.Errorf("got wrong result: %s", diff.Result)
	}

}

func TestPayloadRejected(t *testing.T) {

	r := performPOST(t, "7", "left", []byte("not/base64"))

	if r.StatusCode != 400 {
		t.Fatalf("accepted invalid data, got status: %d, body: %v", r.StatusCode, r.Body)
	}

}

func performPOST(t *testing.T, ID, side string, p []byte) events.APIGatewayProxyResponse {
	return app.Handle(events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		PathParameters: map[string]string{
			"id":   ID,
			"side": side,
		},
		Body: string(p),
	})
}

func upload(t *testing.T, ID string, side string, data string) {
	p, err := json.Marshal(map[string]string{
		"data": data,
	})
	if err != nil {
		t.Fatal("cannot serialize test payload", err)
	}

	r := performPOST(t, ID, side, p)

	if r.StatusCode != 204 {
		t.Errorf("POST %s/%s, got wrong status code: %d, body: %v", ID, side, r.StatusCode, r.Body)
	}
}

func performGET(t *testing.T, ID string) events.APIGatewayProxyResponse {
	return app.Handle(events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		PathParameters: map[string]string{
			"id": ID,
		},
	})
}

func diff(t *testing.T, ID string) (body DiffResponseBody) {
	r := performGET(t, ID)

	if r.StatusCode != 200 {
		t.Fatalf("GET %s, got wrong status code: %d, body: %v", ID, r.StatusCode, r.Body)
	}

	if err := json.Unmarshal([]byte(r.Body), &body); err != nil {
		t.Fatal("cannot parse diff response body", err)
	}
	return
}
