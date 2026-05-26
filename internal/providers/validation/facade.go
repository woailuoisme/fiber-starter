package validation

import (
	"lfiber/internal/providers/validation/contracts"
	"lfiber/internal/support/appctx"

	"github.com/go-playground/validator/v10"
)

// factory returns the validation factory instance from the container.
func factory() contracts.Factory {
	if app := appctx.App(); app != nil {
		return app.ValidationService()
	}
	return nil
}

// Make creates a new validator instance.
func Make(data any, rules map[string]string, messages map[string]string, attributes map[string]string) contracts.Validator {
	if f := factory(); f != nil {
		return f.Make(data, rules, messages, attributes)
	}
	return nil
}

// Extend registers a custom validation rule.
func Extend(rule string, extension validator.Func, message string) error {
	if f := factory(); f != nil {
		return f.Extend(rule, extension, message)
	}
	return nil
}

// ExtendImplicit registers a custom implicit validation rule.
func ExtendImplicit(rule string, extension validator.Func, message string) error {
	if f := factory(); f != nil {
		return f.ExtendImplicit(rule, extension, message)
	}
	return nil
}

// Replacer registers a custom message replacer.
func Replacer(rule string, replacer contracts.ReplacerFunc) {
	if f := factory(); f != nil {
		f.Replacer(rule, replacer)
	}
}
