package middleware

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator interface {
	ValidateStruct(v interface{}) error
}

type tagValidator struct {
	validate *validator.Validate
}

func NewTagValidator() Validator {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(jsonTag)
	return &tagValidator{validate: v}
}

func (t *tagValidator) ValidateStruct(val interface{}) error {
	if t.validate == nil || val == nil {
		return nil
	}
	return t.validate.StructCtx(context.Background(), val)
}

func NewValidatorMiddleware(v Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := contextWithValidator(r.Context(), v)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func jsonTag(fld reflect.StructField) string {
	tag := fld.Tag.Get("json")
	if tag == "" {
		return fld.Name
	}
	name := strings.Split(tag, ",")[0]
	if name == "" || name == "-" {
		return fld.Name
	}
	return name
}
