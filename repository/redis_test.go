package repository_test

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/ehpalumbo/go-diff/repository"
	"github.com/go-redis/redis/v8"

	"github.com/alicebob/miniredis/v2"
)

var mini *miniredis.Miniredis

var repo *repository.RedisRepository

func TestMain(m *testing.M) {
	tearDown := setUp()
	os.Exit(m.Run())
	tearDown()
}

func setUp() func() {
	var err error
	mini, err = miniredis.Run()
	if err != nil {
		panic(err)
	}
	repo = repository.NewRedisRepository(&redis.Options{Addr: mini.Addr()})
	return func() {
		mini.Close()
	}
}

func TestSaveOperation(t *testing.T) {

	cases := []struct {
		name string
		data string
	}{
		{
			name: "empty data",
			data: "",
		},
		{
			name: "non empty data",
			data: "hello",
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {

			mini.FlushDB()

			if err := repo.SaveDataSide("1", "left", []byte(c.data)); err != nil {
				t.Errorf("save operation failed, got: %v", err)
			}

			v := mini.HGet("diff:1", "left")
			if v != c.data {
				t.Errorf("did not save input data for side, expected: %s, stored: %s", c.data, v)
			}

		})

	}

}

func TestSaveOperationOverride(t *testing.T) {

	// given
	mini.FlushDB()

	if err := repo.SaveDataSide("1", "right", []byte("original")); err != nil {
		t.Errorf("save operation failed, got: %v", err)
	}

	// when
	if err := repo.SaveDataSide("1", "right", []byte("new")); err != nil {
		t.Errorf("save (override) operation failed, got: %v", err)
	}

	// then
	v := mini.HGet("diff:1", "right")
	if v != "new" {
		t.Errorf("did not override input data for side, expected: %s, stored: %s", "hallo", v)
	}

}

func TestRejectedSaveOperation(t *testing.T) {

	cases := []struct {
		name     string
		ID       string
		side     string
		expected error
	}{
		{
			name:     "empty ID",
			side:     "left",
			expected: errors.New("cannot save diff side data without ID"),
		},
		{
			name:     "empty side",
			ID:       "1",
			expected: errors.New("cannot save diff side data without side"),
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {

			err := repo.SaveDataSide(c.ID, c.side, []byte("hello"))

			if err == nil {
				t.Fatal("accepted invalid input")
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

func TestGetOperation(t *testing.T) {

	cases := []struct {
		name string
		ID   string
		hash map[string]string
	}{
		{
			name: "get existing sides for ID",
			ID:   "1",
			hash: map[string]string{
				"left":  "hello",
				"right": "hallo",
			},
		},
		{
			name: "get empty map for absent ID",
			ID:   "2",
			hash: map[string]string{},
		},
		{
			name: "get empty sides when field is present but no contents",
			ID:   "3",
			hash: map[string]string{
				"left": "",
			},
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {

			// given
			mini.FlushDB()

			for k, v := range c.hash {
				mini.HSet("diff:"+c.ID, k, v)
			}

			// when
			ds, err := repo.GetDataSidesByID(c.ID)

			// then
			if err != nil {
				t.Errorf("failed, got: %v", err)
			}

			if len(ds) != len(c.hash) {
				t.Errorf("wrong number of results, expected: %v, got: %v", len(c.hash), len(ds))
			}
			for k, v := range ds {
				if c.hash[k] != string(v) {
					t.Errorf("wrong diff side data for side %s, expected: %s, got: %s", k, c.hash[k], v)
				}
			}

		})

	}

}

func TestGetOperationPropagatesFailure(t *testing.T) {

	// given
	mini.FlushDB()

	mini.Set("diff:1", "something") // not a hash

	// when
	ds, err := repo.GetDataSidesByID("1")

	// then
	if err == nil {
		t.Fatal("should have failed but it did not")
	}
	if ds != nil {
		t.Errorf("failed but returned non-nil map, got: %v", ds)
	}
}
