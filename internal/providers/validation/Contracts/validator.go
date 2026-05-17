package Contracts

// MessageBag mirrors Laravel's MessageBag concept in a Go-friendly form.
type MessageBag map[string][]string

// NewMessageBag creates a defensive copy of validation messages.
func NewMessageBag(messages map[string][]string) MessageBag {
	bag := make(MessageBag, len(messages))
	for field, errs := range messages {
		cloned := make([]string, len(errs))
		copy(cloned, errs)
		bag[field] = cloned
	}
	return bag
}

// Has reports whether a field has any validation errors.
func (b MessageBag) Has(field string) bool {
	_, ok := b[field]
	return ok
}

// First returns the first error message for a field.
func (b MessageBag) First(field string) string {
	errs := b[field]
	if len(errs) == 0 {
		return ""
	}
	return errs[0]
}

// All flattens all error messages into a single slice.
func (b MessageBag) All() []string {
	out := make([]string, 0)
	for _, errs := range b {
		out = append(out, errs...)
	}
	return out
}

// Map returns a defensive copy of the underlying map.
func (b MessageBag) Map() map[string][]string {
	return NewMessageBag(b)
}

// Validator mirrors Laravel's validator contract.
type Validator interface {
	// Validate runs the validation rules against the instance data.
	Validate() error

	// Validated returns the validated attributes after a successful run.
	Validated() map[string]any

	// Fails reports whether validation failed.
	Fails() bool

	// Failed returns the failed rule map.
	Failed() map[string]map[string]any

	// Sometimes conditionally appends rules to a field.
	Sometimes(attribute string, rules string, callback ConditionFunc) Validator

	// After registers a callback that runs after validation.
	After(callback AfterFunc) Validator

	// Errors returns the validation message bag.
	Errors() MessageBag
}
