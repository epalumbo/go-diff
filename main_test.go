package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

const baseURL = "http://127.0.0.1:8081/v1/diff"

type DiffResponseBody struct {
	Result   string `json:"result" binding:"required"`
	Insights []struct {
		Offset uint `json:"offset"`
		Length uint `json:"length"`
	} `json:"insights"`
}

func TestMain(m *testing.M) {
	dbURL, stopMiniRedis := startMiniRedis()
	shutdown := RunApplication("127.0.0.1:8081", dbURL)
	waitForServerToStart()
	c := m.Run()
	shutdown(context.Background())
	stopMiniRedis()
	os.Exit(c)
}

func startMiniRedis() (string, func()) {
	mini, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	dbURL := "redis://" + mini.Addr()
	return dbURL, func() {
		mini.Close()
	}
}

func waitForServerToStart() {
	ticks := time.Tick(100 * time.Millisecond)
	i := 0
	for range ticks {
		_, err := http.Get(baseURL)
		if err == nil {
			return
		}
		if i > 30 {
			panic(err)
		}
		i++
	}
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
	defer r.Body.Close()

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
	defer r.Body.Close()

	if r.StatusCode != 400 {
		body, _ := ioutil.ReadAll(r.Body)
		t.Fatalf("accepted invalid data, got status: %d, body: %v", r.StatusCode, string(body))
	}

}

func performPOST(t *testing.T, ID, side string, p []byte) *http.Response {
	URI := fmt.Sprintf("%s/%s/%s", baseURL, ID, side)
	r, err := http.Post(URI, "application/json", bytes.NewBuffer(p))
	if err != nil {
		t.Fatal("cannot call upload API", err)
	}
	return r
}

func upload(t *testing.T, ID string, side string, data string) {
	p, err := json.Marshal(map[string]string{
		"data": data,
	})
	if err != nil {
		t.Fatal("cannot serialize test payload", err)
	}

	r := performPOST(t, ID, side, p)
	defer r.Body.Close()

	if r.StatusCode != 204 {
		body, _ := ioutil.ReadAll(r.Body)
		t.Errorf("POST %s/%s, got wrong status code: %d, body: %v", ID, side, r.StatusCode, string(body))
	}
}

func performGET(t *testing.T, ID string) *http.Response {
	URI := fmt.Sprintf("%s/%s", baseURL, ID)
	r, err := http.Get(URI)
	if err != nil {
		t.Fatalf("cannot get diff report %s: %v", ID, err)
	}
	return r
}

func diff(t *testing.T, ID string) (body DiffResponseBody) {
	r := performGET(t, ID)
	defer r.Body.Close()

	if r.StatusCode != 200 {
		body, _ := ioutil.ReadAll(r.Body)
		t.Fatalf("GET %s, got wrong status code: %d, body: %v", ID, r.StatusCode, string(body))
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatal("cannot read diff response body", err)
	}

	if err := json.Unmarshal(b, &body); err != nil {
		t.Fatal("cannot parse diff response body", err)
	}
	return
}
