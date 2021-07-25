//go:generate mockgen -destination mocks/s3_mock.go -package=mocks . S3Client
package repository_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ehpalumbo/go-diff/repository"
	"github.com/ehpalumbo/go-diff/repository/mocks"
	"github.com/golang/mock/gomock"
)

func setUp(t *testing.T) (*repository.S3DiffRepository, *mocks.MockS3Client, func()) {
	ctrl := gomock.NewController(t)
	s3 := mocks.NewMockS3Client(ctrl)
	repo := repository.NewS3DiffRepository(s3, "go-diff-bucket")
	return repo, s3, func() {
		ctrl.Finish()
	}
}

// PutObjectInputMatcher
type PutObjectInputMatcher struct {
	bucketName, objectKey, body string
}

func (m *PutObjectInputMatcher) Matches(x interface{}) bool {
	if input, ok := x.(*s3.PutObjectInput); ok {
		if *input.Bucket == m.bucketName && *input.Key == m.objectKey {
			if body, err := ioutil.ReadAll(input.Body); err == nil {
				return string(body) == m.body
			}
		}
	}
	return false
}

func (m *PutObjectInputMatcher) String() string {
	return "PutObjectInput argument matcher"
}

// GetObjectInputMatcher
type GetObjectInputMatcher struct {
	bucketName, objectKey string
}

func (m *GetObjectInputMatcher) Matches(x interface{}) bool {
	if input, ok := x.(*s3.GetObjectInput); ok {
		return *input.Bucket == m.bucketName && *input.Key == m.objectKey
	}
	return false
}

func (m *GetObjectInputMatcher) String() string {
	return "GetObjectInput argument matcher"
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

			repo, client, tearDown := setUp(t)
			defer tearDown()

			putObjectInput := PutObjectInputMatcher{
				bucketName: "go-diff-bucket",
				objectKey:  "diff/1/left",
				body:       c.data,
			}

			client.EXPECT().PutObject(gomock.Any(), &putObjectInput).Return(&s3.PutObjectOutput{}, nil)

			if err := repo.SaveDataSide("1", "left", []byte(c.data)); err != nil {
				t.Errorf("save operation failed, got: %v", err)
			}
		})

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

			repo, _, tearDown := setUp(t)
			defer tearDown()

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
			repo, client, tearDown := setUp(t)
			defer tearDown()

			sides := []string{"left", "right"}
			for _, side := range sides {
				getObjectInput := GetObjectInputMatcher{
					bucketName: "go-diff-bucket",
					objectKey:  fmt.Sprintf("diff/%s/%s", c.ID, side),
				}
				if data, ok := c.hash[side]; ok {
					getObjectOutput := s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewReader([]byte(data))),
					}
					client.EXPECT().GetObject(gomock.Any(), &getObjectInput).Return(&getObjectOutput, nil)
				} else {
					err := types.NoSuchKey{Message: aws.String("not found")}
					client.EXPECT().GetObject(gomock.Any(), &getObjectInput).Return(&s3.GetObjectOutput{}, &err)
				}
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
	repo, client, tearDown := setUp(t)
	defer tearDown()

	getLeftObjectInput := GetObjectInputMatcher{
		bucketName: "go-diff-bucket",
		objectKey:  "diff/1/left",
	}
	getLeftObjectOutput := s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewReader([]byte("hello"))),
	}
	client.EXPECT().GetObject(gomock.Any(), &getLeftObjectInput).Return(&getLeftObjectOutput, nil)

	getRightObjectInput := GetObjectInputMatcher{
		bucketName: "go-diff-bucket",
		objectKey:  "diff/1/right",
	}
	client.EXPECT().GetObject(gomock.Any(), &getRightObjectInput).Return(&s3.GetObjectOutput{}, errors.New("Oops!"))

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
