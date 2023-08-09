package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

type s3Storage struct {
	svc    *s3.S3
	bucket string
	key    string
}

var _ FileLike = &s3Storage{}

func NewS3Backend(bucket, key, region string) (*s3Storage, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AWS session")
	}

	svc := s3.New(sess)

	return &s3Storage{
		svc:    svc,
		bucket: bucket,
		key:    key,
	}, nil
}

func (s3b *s3Storage) Load(ctx context.Context, v any) error {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s3b.bucket),
		Key:    aws.String(s3b.key),
	}
	output, err := s3b.svc.GetObjectWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			return nil
		}

		return errors.Wrap(err, "fail to load layers from s3")
	}

	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return errors.Wrap(err, "fail to read data from bucket object")
	}

	err = json.Unmarshal(data, &v)
	if err != nil {
		return errors.Wrap(err, "fail to decode layers definitions from bucket object data")
	}

	return nil
}

func (s3b *s3Storage) Save(ctx context.Context, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "fail to marshal layers to json")
	}

	input := &s3.PutObjectInput{
		Body:   bytes.NewReader(data),
		Bucket: aws.String(s3b.bucket),
		Key:    aws.String(s3b.key),
	}

	_, err = s3b.svc.PutObject(input)
	return errors.Wrap(err, "fail to save layers to s3")
}
