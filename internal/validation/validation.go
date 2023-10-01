package validation

import (
	"net/url"
	"regexp"
)

func IsValidDirectory(path string) bool {
	// TODO: validation of directory path
	return true
}

func IsValidS3Bucket(bucketName string) bool {
	// TODO: validation of s3 bucket name
	return true
}

func IsValidS3Region(region string) bool {
	// TODO: validation of s3 bucket region
	return true
}

func IsValidEmail(email string) bool {
	pattern := "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	match, _ := regexp.MatchString(pattern, email)
	return match
}

func IsValidURL(u string) bool {
	_, err := url.Parse(u)
	return err == nil
}
