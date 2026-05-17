package Contracts

import validator "github.com/go-playground/validator/v10"

// ReplacerFunc matches Laravel's rule replacer concept.
type ReplacerFunc func(field string, param string, value any) string

// ConditionFunc determines whether a conditional rule should be applied.
type ConditionFunc func(data any) bool

// AfterFunc runs after validation completes.
type AfterFunc func(Validator)

// Factory mirrors Laravel's validation factory contract.
type Factory interface {
	// Make creates a new validator instance.
	Make(data any, rules map[string]string, messages map[string]string, attributes map[string]string) Validator

	// Extend registers a custom validation rule.
	Extend(rule string, extension validator.Func, message string) error

	// ExtendImplicit registers a custom implicit validation rule.
	ExtendImplicit(rule string, extension validator.Func, message string) error

	// Replacer registers a custom message replacer.
	Replacer(rule string, replacer ReplacerFunc)
}
