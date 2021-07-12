package validations

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/distribution/distribution/reference"
	"github.com/go-playground/validator"
)

func NewValidator() (*validator.Validate, error) {
	validate := validator.New()

	//add custom validations
	customValidations := map[string]func(fl validator.FieldLevel) bool{
		"alphanumunderscore":            isAlphaNumericUnderscore,
		"is_arn":                        isValidARN,
		"valid_target_type":             isValidTargetType,
		"valid_execute_container_image": isValidExecuteContainerImage,
		"valid_precontainer_image":      isValidPreContainerImage,
		"valid_argument":                isValidArgument,
	}

	for jsonTag, fnName := range customValidations {
		err := validate.RegisterValidation(jsonTag, fnName)
		if err != nil {
			return nil, err
		}
	}

	return validate, nil
}

// Initialize validator and validate struct
func ValidateStruct(s interface{}) error {
	validate, err := NewValidator()
	if err != nil {
		return err
	}
	if err := validate.Struct(s); err != nil {
		return validationErrorMessage("structValidation", err)
	}
	return nil
}

// Initialize validator and validate struct
func ValidateVar(name string, s interface{}, validation string) error {
	validate, err := NewValidator()
	if err != nil {
		return err
	}
	if err := validate.Var(s, validation); err != nil {
		return validationErrorMessage(name, err)
	}
	return nil
}

// ValidateValuer implements validator.CustomTypeFunc
// Vault does not allow dashes
func isAlphaNumericUnderscore(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^([a-zA-Z])[a-zA-Z0-9_]*$`).MatchString(fl.Field().String())
}

func isValidExecuteContainerImage(fl validator.FieldLevel) bool {
	image := fl.Field().MapIndex(reflect.ValueOf("execute_container_image_uri"))
	if image.IsValid() {
		return isValidImageURI(image.String())
	}
	// execute_container_image_uri key missing
	return false
}

func isValidPreContainerImage(fl validator.FieldLevel) bool {
	image := fl.Field().MapIndex(reflect.ValueOf("pre_container_image_uri"))
	if image.IsValid() {
		return isValidImageURI(image.String())
	}

	// pre_container_image_uri is not required
	return true
}

func isValidTargetType(fl validator.FieldLevel) bool {
	return fl.Field().String() == "aws_account"
}

// TODO long term, we should evaluate if hard coding in code is the right approach to
// specifying different argument types vs allowing dynmaic specification and
// interpolation in service/config.yaml
func isValidArgument(fl validator.FieldLevel) bool {
	for _, key := range fl.Field().MapKeys() {
		switch key.String() {
		case "execute", "init":
			return true
		}
	}
	return false
}

func isValidARN(fl validator.FieldLevel) bool {
	return arn.IsARN(fl.Field().String())
}

// Returns true, if the image uri is a valid container image uri
func isValidImageURI(imageURI string) bool {
	_, err := reference.ParseAnyReference(imageURI)
	return err == nil
}

// Custom error messages
func validationErrorMessage(name string, err error) error {
	var validationErrors validator.ValidationErrors
	if ok := errors.As(err, &validationErrors); ok {
		validationError := validationErrors[0]
		switch validationError.Tag() {
		case "is_arn":
			return fmt.Errorf("'%s' value '%v' is not a valid arn", validationError.Tag(), validationError.Value())
		case "valid_target_type":
			return fmt.Errorf("'%s' value '%v' is invalid, types supported:'aws_account'", validationError.Tag(), validationError.Value())
		case "alphanumunderscore":
			return fmt.Errorf("value '%v' is invalid, must only contain alpha numberic underscore characters", validationError.Value())
		case "valid_execute_container_image", "valid_precontainer_image":
			return fmt.Errorf("'%s' value '%v' is an invalid container uri", validationError.Tag(), validationError.Value())
		case "valid_argument":
			return fmt.Errorf("'%s' value '%v' is an invalid argument", validationError.Tag(), validationError.Value())
		default:
			if validationError.Field() == "" {
				return fmt.Errorf("'%s' '%v'", name, validationError.Param())
			}
			return fmt.Errorf("'%s' '%v'", validationError.Tag(), validationError.Field())
		}

	}
	return err
}
