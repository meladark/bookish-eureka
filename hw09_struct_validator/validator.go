package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
	ErrUnsupportedType        = errors.New("unsupported type")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for _, e := range v {
		sb.WriteString(fmt.Sprintf("%s: %s\n", e.Field, e.Err))
	}
	return sb.String()
}

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("%w: expected struct", ErrUnsupportedType)
	}
	var validationErrs ValidationErrors
	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		fieldValue := val.Field(i)
		validateTag, ok := field.Tag.Lookup("validate")
		if !ok {
			continue
		}
		err := validateField(field.Name, fieldValue, validateTag)
		if err != nil {
			var verrs ValidationErrors
			if errors.As(err, &verrs) {
				validationErrs = append(validationErrs, verrs...)
			} else {
				return err
			}
		}
	}
	if len(validationErrs) > 0 {
		return validationErrs
	}
	return nil
}

func validateField(fieldName string, val reflect.Value, tag string) error {
	if tag == "nested" && val.Kind() == reflect.Struct {
		return Validate(val.Interface())
	}
	validators := strings.Split(tag, "|")
	switch val.Kind() {
	case reflect.String:
		return validateStringField(fieldName, val.String(), validators)
	case reflect.Int:
		return validateIntField(fieldName, int(val.Int()), validators)
	case reflect.Slice:
		return validateSliceField(fieldName, val, validators)
	case reflect.Invalid,
		reflect.Bool,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Struct,
		reflect.UnsafePointer:
		return ErrUnsupportedType
	default:
		return ErrUnsupportedType
	}
}

func validateStringField(fieldName, str string, validators []string) error {
	var errs ValidationErrors
	for _, rule := range validators {
		if err := validateString(fieldName, str, rule); err != nil {
			errs = append(errs, *err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateIntField(fieldName string, num int, validators []string) error {
	var errs ValidationErrors
	for _, rule := range validators {
		if err := validateInt(fieldName, num, rule); err != nil {
			errs = append(errs, *err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateSliceField(fieldName string, val reflect.Value, validators []string) error {
	elemKind := val.Type().Elem().Kind()
	var errs ValidationErrors
	switch elemKind {
	case reflect.String:
		for i := 0; i < val.Len(); i++ {
			str := val.Index(i).String()
			for _, rule := range validators {
				if err := validateString(fieldName, str, rule); err != nil {
					errs = append(errs, *err)
				}
			}
		}
	case reflect.Int:
		for i := 0; i < val.Len(); i++ {
			num := int(val.Index(i).Int())
			for _, rule := range validators {
				if err := validateInt(fieldName, num, rule); err != nil {
					errs = append(errs, *err)
				}
			}
		}
	case reflect.Invalid,
		reflect.Bool,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice,
		reflect.Struct,
		reflect.UnsafePointer:
		return ErrUnsupportedType
	default:
		return ErrUnsupportedType
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateString(field, value, rule string) *ValidationError {
	switch {
	case strings.HasPrefix(rule, "len:"):
		{
			expectedLen, err := strconv.Atoi(strings.TrimPrefix(rule, "len:"))
			if err != nil {
				return &ValidationError{Field: field, Err: ErrInvalidValidatorSyntax}
			}
			if len(value) != expectedLen {
				return &ValidationError{Field: field, Err: fmt.Errorf("length must be %d", expectedLen)}
			}
		}
	case strings.HasPrefix(rule, "regexp:"):
		{
			pattern := strings.TrimPrefix(rule, "regexp:")
			re, err := regexp.Compile(pattern)
			if err != nil {
				return &ValidationError{Field: field, Err: ErrInvalidValidatorSyntax}
			}
			if !re.MatchString(value) {
				return &ValidationError{Field: field, Err: fmt.Errorf("does not match regexp %s", pattern)}
			}
		}
	case strings.HasPrefix(rule, "in:"):
		{
			options := strings.Split(strings.TrimPrefix(rule, "in:"), ",")
			found := false
			for _, opt := range options {
				if value == opt {
					found = true
					break
				}
			}
			if !found {
				return &ValidationError{Field: field, Err: fmt.Errorf("must be one of [%s]", strings.Join(options, ", "))}
			}
		}
	default:
		return nil
	}
	return nil
}

func validateInt(field string, value int, rule string) *ValidationError {
	switch {
	case strings.HasPrefix(rule, "min:"):
		{
			minValue, err := strconv.Atoi(strings.TrimPrefix(rule, "min:"))
			if err != nil {
				return &ValidationError{Field: field, Err: ErrInvalidValidatorSyntax}
			}
			if value < minValue {
				return &ValidationError{Field: field, Err: fmt.Errorf("must be >= %d", minValue)}
			}
		}
	case strings.HasPrefix(rule, "max:"):
		{
			maxValue, err := strconv.Atoi(strings.TrimPrefix(rule, "max:"))
			if err != nil {
				return &ValidationError{Field: field, Err: ErrInvalidValidatorSyntax}
			}
			if value > maxValue {
				return &ValidationError{Field: field, Err: fmt.Errorf("must be <= %d", maxValue)}
			}
		}
	case strings.HasPrefix(rule, "in:"):
		{
			options := strings.Split(strings.TrimPrefix(rule, "in:"), ",")
			found := false
			for _, opt := range options {
				optVal, err := strconv.Atoi(opt)
				if err != nil {
					return &ValidationError{Field: field, Err: ErrInvalidValidatorSyntax}
				}
				if value == optVal {
					found = true
					break
				}
			}
			if !found {
				return &ValidationError{Field: field, Err: fmt.Errorf("must be one of [%s]", strings.Join(options, ", "))}
			}
		}
	}
	return nil
}
