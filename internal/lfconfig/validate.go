package lfconfig

import (
	"net/url"
	"regexp"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func Validate(ctx ConfigContext) error {
	var result *multierror.Error

	switch ctx.Type {
	case "local":
		if ctx.Dir == "" {
			result = multierror.Append(result, errors.New("directory path cannot be empty"))
		} else if !isValidDirectory(ctx.Dir) {
			result = multierror.Append(result, errors.Errorf("invalid directory path: %s", ctx.Dir))
		}
	case "s3":
		if ctx.Bucket == "" {
			result = multierror.Append(result, errors.New("S3 bucket name cannot be empty"))
		} else if !isValidS3Bucket(ctx.Bucket) {
			result = multierror.Append(result, errors.Errorf("invalid S3 bucket name: %s", ctx.Bucket))
		}

		if ctx.Region == "" {
			result = multierror.Append(result, errors.New("S3 bucket region cannot be empty"))
		} else if !isValidS3Region(ctx.Region) {
			result = multierror.Append(result, errors.Errorf("invalid S3 bucket region: %s", ctx.Region))
		}
	case "cloud":
		if ctx.Email == "" {
			result = multierror.Append(result, errors.New("email cannot be empty"))
		} else if !isValidEmail(ctx.Email) {
			result = multierror.Append(result, errors.Errorf("invalid email: %s", ctx.Email))
		}
		if ctx.Password == "" {
			result = multierror.Append(result, errors.New("password cannot be empty"))
		}
		if ctx.URL == "" {
			result = multierror.Append(result, errors.New("URL cannot be empty"))
		} else if !isValidURL(ctx.URL) {
			result = multierror.Append(result, errors.Errorf("invalid URL: %s", ctx.URL))
		}
	default:
		return errors.New("invalid context type")
	}

	return result.ErrorOrNil()
}

func isValidDirectory(path string) bool {
	// TODO: validation of directory path
	return true
}

func isValidS3Bucket(bucketName string) bool {
	// TODO: validation of s3 bucket name
	return true
}

func isValidS3Region(region string) bool {
	// TODO: validation of s3 bucket region
	return true
}

func isValidEmail(email string) bool {
	pattern := "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	match, _ := regexp.MatchString(pattern, email)
	return match
}

func isValidURL(u string) bool {
	_, err := url.Parse(u)
	return err == nil
}
