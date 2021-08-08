package validations

import (
	"regexp"

	"github.com/asaskevich/govalidator"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/distribution/distribution/reference"
)

// Validate iterates through the provided validation funcs.
func Validate(validations ...func() error) error {
	for _, v := range validations {
		if err := v(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateStruct validates the provided struct.
func ValidateStruct(input interface{}) error {
	customValidators := map[string]govalidator.CustomTypeValidator{
		"alphanumunderscore": isAlphaNumbericUnderscore,
		"gitURI":             isValidGitURI,
	}

	for k, v := range customValidators {
		if _, exists := govalidator.CustomTypeTagMap.Get(k); !exists {
			govalidator.CustomTypeTagMap.Set(k, v)
		}
	}

	_, err := govalidator.ValidateStruct(input)
	return err
}

// isAlphaNumbericUnderscore
func isAlphaNumbericUnderscore(field interface{}, kind interface{}) bool {
	// only handle strings
	switch s := field.(type) {
	case string:
		// Vault does not allow dashes and must start with alpha.
		pattern := `^([a-zA-Z])[a-zA-Z0-9_]*$`
		return regexp.MustCompile(pattern).MatchString(s)
	default:
		panic("unsupported field type for isAlphaNumbericUnderscore2")
	}
}

// IsValidARN determines if the string is a valid AWS ARN.
func IsValidARN(s string) bool {
	return arn.IsARN(s)
}

// IsValidImageURI determines if the image URI is a valid container image URI
// format.
func IsValidImageURI(imageURI string) bool {
	_, err := reference.ParseAnyReference(imageURI)
	return err == nil
}

// isValidGitURI
func isValidGitURI(field interface{}, kind interface{}) bool {
	// only handle strings
	switch s := field.(type) {
	case string:
		pattern := `((git|ssh|https)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`
		return regexp.MustCompile(pattern).MatchString(s)
	default:
		panic("unsupported field type for isValidGitRepository")
	}
}
