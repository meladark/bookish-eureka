package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}
	App struct {
		Version string `validate:"len:5"`
	}
	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}
	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
	Product struct {
		Version string `validate:"len:5"`
	}
	ErrorsAll struct {
		Version string   `validate:"len:error"`
		Name    string   `validate:"regexp:(abc"`
		Age     int      `validate:"min:kek|max:lol"`
		Date    int      `validate:"in:123,abc,2048"`
		Role    UserRole `validate:"in:256,abc,1024"`
	}
	IntWrapper struct {
		Value int `validate:"min:10|max:20|in:12,14,16"`
	}
	IntSliceWrapper struct {
		Numbers []int `validate:"min:10|max:20|in:12,14,16"`
	}
	BadTypes struct {
		Enabled bool      `validate:"len:10"`
		Scores  []float64 `validate:"min:0|max:10"`
	}
)

func TestValidationErrors_Error(t *testing.T) {
	var errs ValidationErrors
	if got := errs.Error(); got != "" {
		t.Errorf("empty ValidationErrors should return empty string, got %q", got)
	}
	errs = ValidationErrors{
		{Field: "Name", Err: errors.New("too short")},
		{Field: "Age", Err: errors.New("too young")},
	}
	expected := "Name: too short\nAge: too young\n"
	if got := errs.Error(); got != expected {
		t.Errorf("unexpected error string:\n got:\n%q\nwant:\n%q", got, expected)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		{
			name: "valid User",
			in: User{
				ID: "123456789012345678901234567890123456", Age: 30, Email: "test@mail.com",
				Role: "admin", Phones: []string{"12345678901"},
			},
			expectedErr: nil,
		},
		{
			name: "invalid User",
			in: User{
				ID: "short", Age: 17, Email: "invalid-email", Role: "guest", Phones: []string{"123"},
			},
			expectedErr: ValidationErrors{
				{Field: "ID"}, {Field: "Age"}, {Field: "Email"}, {Field: "Role"}, {Field: "Phones"},
			},
		},
		{
			name:        "unsupported type",
			in:          123,
			expectedErr: ErrUnsupportedType,
		},
		{
			name: "nested struct invalid",
			in: struct {
				Meta Product `validate:"nested"`
			}{Meta: Product{Version: "bad"}},
			expectedErr: ValidationErrors{
				{Field: "Version"},
			},
		},
		{
			name: "invalid int value - less than min",
			in: IntWrapper{
				Value: 5,
			},
			expectedErr: ValidationErrors{
				{Field: "Value"},
			},
		},
		{
			name: "invalid int value - more than max",
			in: IntWrapper{
				Value: 25,
			},
			expectedErr: ValidationErrors{
				{Field: "Value"},
			},
		},
		{
			name: "invalid int value - not in set",
			in: IntWrapper{
				Value: 13,
			},
			expectedErr: ValidationErrors{
				{Field: "Value"},
			},
		},
		{
			name:        "unsupported type int",
			in:          42,
			expectedErr: ErrUnsupportedType,
		},
		{
			name: "valid int slice",
			in: IntSliceWrapper{
				Numbers: []int{12, 14, 16},
			},
			expectedErr: nil,
		},
		{
			name: "invalid int slice - element too small",
			in: IntSliceWrapper{
				Numbers: []int{5, 12},
			},
			expectedErr: ValidationErrors{
				{Field: "Numbers"},
			},
		},
		{
			name: "invalid int slice - element too large",
			in: IntSliceWrapper{
				Numbers: []int{12, 25},
			},
			expectedErr: ValidationErrors{
				{Field: "Numbers"},
			},
		},
		{
			name: "invalid int slice - element not in set",
			in: IntSliceWrapper{
				Numbers: []int{12, 13},
			},
			expectedErr: ValidationErrors{
				{Field: "Numbers"},
			},
		},
		{
			name: "unsupported field type bool",
			in: BadTypes{
				Enabled: true,
			},
			expectedErr: ErrUnsupportedType,
		},
		{
			name: "string error len",
			in: ErrorsAll{
				Version: "bad",
			},
			expectedErr: ValidationErrors{
				{Field: "Version"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.in)
			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				return
			}
			if errors.Is(err, ErrUnsupportedType) && !errors.Is(tt.expectedErr, ErrUnsupportedType) {
				t.Errorf("unexpected unsupported type error: %s", err)
				return
			}
			if verrs, ok := GetValidationErrors(err); ok {
				if expectedErrs, ok := GetValidationErrors(tt.expectedErr); ok {
					CheckExpectedErrors(t, verrs, expectedErrs)
				}
			} else {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func GetValidationErrors(err error) (ValidationErrors, bool) {
	var verrs ValidationErrors
	if errors.As(err, &verrs) {
		return verrs, true
	}
	return nil, false
}

func CheckExpectedErrors(t *testing.T, actual, expected ValidationErrors) {
	t.Helper()
	for _, e := range expected {
		found := false
		for _, ve := range actual {
			if ve.Field == e.Field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected error on field %s, but not found", e.Field)
		}
	}
}
