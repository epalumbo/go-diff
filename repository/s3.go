package repository

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// S3DiffRepository is the AWS S3-backed implementation of the DiffRepository contract
type S3DiffRepository struct {
	client     S3Client
	bucketName string
}

// NewS3DiffRepository creates a new instance of the S3DiffRepository implementation
func NewS3DiffRepository(client S3Client, bucketName string) *S3DiffRepository {
	return &S3DiffRepository{client, bucketName}
}

// SaveDataSide saves data sides to S3
func (r *S3DiffRepository) SaveDataSide(ID string, side string, data []byte) error {
	if len(ID) == 0 {
		return errors.New("cannot save diff side data without ID")
	}
	if len(side) == 0 {
		return errors.New("cannot save diff side data without side")
	}
	request := s3.PutObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(keyOf(ID, side)),
		Body:   bytes.NewReader(data),
	}
	_, err := r.client.PutObject(context.Background(), &request)
	return err
}

type sideData struct {
	side string
	data []byte
}

// GetDataSidesByID gets data sides by ID in parallel from S3
func (r *S3DiffRepository) GetDataSidesByID(ID string) (map[string][]byte, error) {
	m := make(map[string][]byte)

	results := make(chan sideData, 2)
	errors := make(chan error)

	retrieve := func(side string) {
		data, err := r.retrieve(ID, side)
		if err == nil {
			results <- sideData{side, data}
		} else {
			errors <- err
		}
	}

	go retrieve("left")
	go retrieve("right")

	var err error = nil
	for latch := 0; latch < 2; latch++ {
		select {
		case result := <-results:
			if result.data != nil {
				m[result.side] = result.data
			}
		case err = <-errors:
		}
	}

	if err == nil {
		return m, nil
	} else {
		return nil, err
	}
}

func (r *S3DiffRepository) retrieve(ID, side string) ([]byte, error) {
	request := s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(keyOf(ID, side)),
	}
	response, err := r.client.GetObject(context.Background(), &request)
	if err == nil {
		return read(response.Body)
	}
	var notFound *types.NoSuchKey
	if errors.As(err, &notFound) {
		return nil, nil
	}
	return nil, err
}

func read(body io.ReadCloser) ([]byte, error) {
	data, err := ioutil.ReadAll(body)
	if err == nil {
		return data, nil
	}
	return nil, err
}

func keyOf(ID, side string) string {
	return fmt.Sprintf("diff/%s/%s", ID, side)
}
