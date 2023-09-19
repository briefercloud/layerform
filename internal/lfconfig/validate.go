package lfconfig

import (
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/internal/validation"
)

func Validate(ctx ConfigContext) error {
	var result *multierror.Error

	switch ctx.Type {
	case "local":
		if ctx.Dir == "" {
			result = multierror.Append(result, errors.New("directory path cannot be empty"))
		} else if !validation.IsValidDirectory(ctx.Dir) {
			result = multierror.Append(result, errors.Errorf("invalid directory path: %s", ctx.Dir))
		}
	case "s3":
		if ctx.Bucket == "" {
			result = multierror.Append(result, errors.New("S3 bucket name cannot be empty"))
		} else if !validation.IsValidS3Bucket(ctx.Bucket) {
			result = multierror.Append(result, errors.Errorf("invalid S3 bucket name: %s", ctx.Bucket))
		}

		if ctx.Region == "" {
			result = multierror.Append(result, errors.New("S3 bucket region cannot be empty"))
		} else if !validation.IsValidS3Region(ctx.Region) {
			result = multierror.Append(result, errors.Errorf("invalid S3 bucket region: %s", ctx.Region))
		}
	case "cloud":
		if ctx.Email == "" {
			result = multierror.Append(result, errors.New("email cannot be empty"))
		} else if !validation.IsValidEmail(ctx.Email) {
			result = multierror.Append(result, errors.Errorf("invalid email: %s", ctx.Email))
		}
		if ctx.Password == "" {
			result = multierror.Append(result, errors.New("password cannot be empty"))
		}
		if ctx.URL == "" {
			result = multierror.Append(result, errors.New("URL cannot be empty"))
		} else if !validation.IsValidURL(ctx.URL) {
			result = multierror.Append(result, errors.Errorf("invalid URL: %s", ctx.URL))
		}
	default:
		return errors.New("invalid context type")
	}

	return result.ErrorOrNil()
}
