package port

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	validatorV10 "github.com/go-playground/validator/v10"
)

var validator = validatorV10.New()

func validateRequest(r *http.Request, target interface{}) error {
	if r.Method != http.MethodPost {
		return errors.New("invalid request method: only POST allowed")
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return errors.New("invalid Content-Type: must be application/json")
	}
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(target)
	if err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := validator.Struct(target); err != nil {
		return formatValidationErrors(err)
	}

	return nil
}

// formatValidationErrors converts validator.ValidationErrors to a readable message - its not perfect and expose internal structure
func formatValidationErrors(err error) error {
	if validationErrors, ok := err.(validatorV10.ValidationErrors); ok {
		var errMsgs []string
		for _, ve := range validationErrors {
			errMsgs = append(errMsgs,
				fmt.Sprintf("field '%s' failed on '%s' tag (value: %v)",
					ve.Field(), ve.Tag(), ve.Value()))
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errMsgs, "; "))
	}

	return err
}
