package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"lfiber/internal/providers/validation/contracts"

	"github.com/go-playground/validator/v10"
)

type conditionalRule struct {
	attribute string
	rules     string
	callback  contracts.ConditionFunc
}

type Instance struct {
	engine     *validator.Validate
	data       any
	rules      map[string]string
	messages   map[string]string
	attributes map[string]string
	replacers  map[string]contracts.ReplacerFunc

	mu         sync.Mutex
	ran        bool
	err        error
	validated  map[string]any
	failed     map[string]map[string]any
	errors     contracts.MessageBag
	conditions []conditionalRule
	after      []contracts.AfterFunc
}

func newInstance(engine *validator.Validate, data any, rules map[string]string, messages map[string]string, attributes map[string]string, replacers map[string]contracts.ReplacerFunc) *Instance {
	return &Instance{
		engine:     engine,
		data:       data,
		rules:      cloneStringMap(rules),
		messages:   cloneStringMap(messages),
		attributes: cloneStringMap(attributes),
		replacers:  cloneReplacerMap(replacers),
	}
}

func (v *Instance) Validate() error {
	v.mu.Lock()
	if v.ran {
		v.mu.Unlock()
		return v.err
	}

	v.validated = nil
	v.failed = make(map[string]map[string]any)
	v.errors = make(contracts.MessageBag)

	rules := v.effectiveRules()
	if len(rules) == 0 {
		v.err = v.validateBareData()
	} else {
		v.err = v.validateWithRules(rules)
	}

	if v.err == nil {
		if len(rules) == 0 {
			v.validated = normalizeValidatedData(v.data, v.attributes)
		} else {
			v.validated = filterValidatedData(normalizeValidatedData(v.data, v.attributes), rules)
		}
	}

	if err := v.err; err != nil {
		v.populateErrorState(err)
	}

	after := append([]contracts.AfterFunc(nil), v.after...)
	v.ran = true
	v.mu.Unlock()

	for _, callback := range after {
		callback(v)
	}

	return v.err
}

func (v *Instance) Validated() map[string]any {
	_ = v.Validate()
	if v.err != nil {
		return nil
	}

	return cloneAnyMap(v.validated)
}

func (v *Instance) Fails() bool {
	return v.Validate() != nil
}

func (v *Instance) Failed() map[string]map[string]any {
	_ = v.Validate()
	return cloneNestedMap(v.failed)
}

func (v *Instance) Sometimes(attribute string, rules string, callback contracts.ConditionFunc) contracts.Validator {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.conditions = append(v.conditions, conditionalRule{
		attribute: attribute,
		rules:     rules,
		callback:  callback,
	})

	v.ran = false
	v.err = nil
	v.validated = nil
	v.failed = nil
	v.errors = nil

	return v
}

func (v *Instance) After(callback contracts.AfterFunc) contracts.Validator {
	if callback == nil {
		return v
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	v.after = append(v.after, callback)
	v.ran = false
	v.err = nil
	v.validated = nil
	v.failed = nil
	v.errors = nil

	return v
}

func (v *Instance) Errors() contracts.MessageBag {
	_ = v.Validate()
	return v.errors.Map()
}

func (v *Instance) validateBareData() error {
	if v.data == nil {
		return nil
	}

	switch data := v.data.(type) {
	case map[string]any:
		return nil
	default:
		rv := reflect.ValueOf(data)
		if !rv.IsValid() {
			return nil
		}
		if rv.Kind() == reflect.Pointer {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}
		if rv.Kind() == reflect.Struct {
			return v.engine.Struct(v.data)
		}
	}

	return nil
}

func (v *Instance) validateWithRules(rules map[string]string) error {
	values := normalizeValidatedData(v.data, v.attributes)
	if len(values) == 0 {
		values = map[string]any{}
	}

	var errs validator.ValidationErrors
	for field, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		value := values[field]
		if err := v.engine.Var(value, rule); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				errs = append(errs, validationErrors...)
				continue
			}
			return err
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (v *Instance) effectiveRules() map[string]string {
	rules := cloneStringMap(v.rules)
	if len(v.conditions) == 0 {
		return rules
	}

	for _, condition := range v.conditions {
		if condition.callback == nil || !condition.callback(v.data) {
			continue
		}

		current := strings.TrimSpace(rules[condition.attribute])
		if current == "" {
			rules[condition.attribute] = strings.TrimSpace(condition.rules)
			continue
		}

		extra := strings.TrimSpace(condition.rules)
		if extra == "" {
			continue
		}
		rules[condition.attribute] = current + "," + extra
	}

	return rules
}

func (v *Instance) populateErrorState(err error) {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		v.errors = contracts.MessageBag{
			"error": []string{err.Error()},
		}
		return
	}

	v.failed = make(map[string]map[string]any, len(validationErrors))
	v.errors = make(contracts.MessageBag, len(validationErrors))

	for _, fe := range validationErrors {
		field := v.resolveFieldName(fe.Field())
		message := v.renderErrorMessage(fe, field)
		v.errors[field] = append(v.errors[field], message)

		if _, ok := v.failed[field]; !ok {
			v.failed[field] = make(map[string]any)
		}
		v.failed[field][fe.Tag()] = fe.Param()
	}
}

func (v *Instance) resolveFieldName(field string) string {
	if field == "" {
		return field
	}

	if attr := v.attributes[field]; attr != "" {
		return attr
	}

	return field
}

func (v *Instance) renderErrorMessage(fe validator.FieldError, field string) string {
	tag := fe.Tag()
	if custom := v.messages[field+"."+tag]; custom != "" {
		return custom
	}
	if custom := v.messages[tag]; custom != "" {
		return custom
	}
	if replacer := v.replacers[tag]; replacer != nil {
		return replacer(field, fe.Param(), fe.Value())
	}

	param := fe.Param()
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min", "max", "len":
		return fmt.Sprintf("%s failed validation for '%s' (%s)", field, tag, param)
	default:
		return fmt.Sprintf("%s failed validation for '%s'", field, tag)
	}
}

func normalizeValidatedData(data any, attributes map[string]string) map[string]any {
	switch typed := data.(type) {
	case nil:
		return nil
	case map[string]any:
		return cloneAnyMap(typed)
	case map[string]string:
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			out[key] = value
		}
		return out
	}

	rv := reflect.ValueOf(data)
	if !rv.IsValid() {
		return nil
	}
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	rt := rv.Type()
	out := make(map[string]any, rv.NumField())
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			name = strings.Split(tag, ",")[0]
		}
		if name == "" {
			name = field.Name
		}
		out[name] = rv.Field(i).Interface()
	}

	return out
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}

	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneAnyMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return nil
	}

	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func filterValidatedData(values map[string]any, rules map[string]string) map[string]any {
	if len(values) == 0 || len(rules) == 0 {
		return cloneAnyMap(values)
	}

	out := make(map[string]any, len(rules))
	for field := range rules {
		if value, ok := values[field]; ok {
			out[field] = value
		}
	}

	return out
}

func cloneNestedMap(input map[string]map[string]any) map[string]map[string]any {
	if len(input) == 0 {
		return nil
	}

	out := make(map[string]map[string]any, len(input))
	for field, rules := range input {
		nested := make(map[string]any, len(rules))
		for rule, value := range rules {
			nested[rule] = value
		}
		out[field] = nested
	}
	return out
}

func cloneReplacerMap(input map[string]contracts.ReplacerFunc) map[string]contracts.ReplacerFunc {
	if len(input) == 0 {
		return nil
	}

	out := make(map[string]contracts.ReplacerFunc, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
