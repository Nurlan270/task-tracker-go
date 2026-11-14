package validator

import (
	"context"
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"

	gpv "github.com/go-playground/validator/v10"
)

// MessageFunc carries failed field's info as well as its formatted field name [fname].
// fname is equal to field.StructField() converted from PascalCase to sentence case.
// e.g. from "SomeFieldName" to "Some field name".
type MessageFunc func(f gpv.FieldError, fname string) string

type Validate struct {
	*gpv.Validate
	messageMap map[string]MessageFunc
}

const defaultTagName string = "json"

func New(opts ...gpv.Option) *Validate {
	v := gpv.New(opts...)

	validator := &Validate{
		Validate:   v,
		messageMap: messages,
	}

	//	Register default tag name func
	validator.RegisterTagName(defaultTagName)

	return validator
}

func (v *Validate) Struct(s any) error {
	err := v.StructCtx(context.Background(), s)
	if err == nil {
		return nil
	}

	var vErrs gpv.ValidationErrors

	if !errors.As(err, &vErrs) {
		return err
	}

	errorMap := make(map[string]string, len(vErrs))
	for _, vErr := range vErrs {
		errorMap[vErr.Field()] = v.getMessage(vErr)
	}

	return &ErrValidation{
		errors: errorMap,
	}
}

// RegisterTagName registers tag name function to get tag value when calling FieldError.Field().
// By default it will use "json" as tag name
//
// For example:
//
// RegisterTagName("vbody")
//
//	 type Request struct {
//		 PassConfirm string `vbody:"password_confirmation" validate:"required"`
//	 }
//
// Will result:
//
// "password_confirmation": "PassConfirm is required."
func (v *Validate) RegisterTagName(tagName string) {
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if tagName != defaultTagName {
			return fld.Tag.Get(tagName)
		}

		return strings.SplitN(fld.Tag.Get(tagName), ",", 2)[0]
	})
}

type CustomRule struct {
	//	Tag associated with validation function
	Tag string

	//	Validation function
	Fn gpv.Func
}

// RegisterCustomRules registers passed custom validation rules.
//
// It will panic if there's an error while registering rule
//
//	NOTES:
//	- if the tag name already exists, the previous validation function will be replaced.
//	- this method is not thread-safe it is intended that these all be registered prior to any validation
func (v *Validate) RegisterCustomRules(rules ...CustomRule) {
	for _, rule := range rules {
		if err := v.RegisterValidation(rule.Tag, rule.Fn); err != nil {
			panic(fmt.Sprintf("failed to register custom validation rule %q: %s", rule.Tag, err))
		}
	}
}

// RegisterCustomMessages registers passed custom validation messages.
//
//	NOTES:
//	- if the tag name already exists, the previous validation function will be replaced.
//	- this method is not thread-safe it is intended that these all be registered prior to any validation
func (v *Validate) RegisterCustomMessages(messages map[string]MessageFunc) {
	for tag, fn := range messages {
		v.messageMap[tag] = fn
	}
}

// getMessage returns message associated with tag,
// if it does not find matching tag it will fallback to default message
func (v *Validate) getMessage(f gpv.FieldError) string {
	for tag, msg := range v.messageMap {
		if tag == f.Tag() {
			return msg(f, formatFieldName(f.StructField()))
		}
	}

	return v.messageMap[messageDefault](f, formatFieldName(f.StructField()))
}

// formatFieldName formats field's name
// from "SomeFieldName" to "Some field name".
func formatFieldName(name string) string {
	s := strcase.ToDelimited(name, ' ')
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
