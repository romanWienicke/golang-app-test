package rest

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
)

func Test_payload(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		valid   bool
	}{
		{
			name:    "Valid payload",
			payload: "{\"key\": \"value\"}",
			valid:   true,
		},
		{
			name:    "Missing key",
			payload: "{}",
			valid:   false,
		},
		{
			name:    "Nil payload",
			payload: "",
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Payload

			var errs error

			errs = errors.Join(json.Unmarshal([]byte(tt.payload), &d))

			validator := validator.New()
			errs = errors.Join(errs, validator.Struct(d))
			if (errs == nil) != tt.valid {
				t.Errorf("Expected validity: %v, got error: %v", tt.valid, errs)
			}
		})
	}
}
