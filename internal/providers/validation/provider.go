package validation

import (
	"reflect"
	"strings"
	"sync"

	"lfiber/configs"
	validationContracts "lfiber/internal/providers/validation/contracts"

	"github.com/go-playground/validator/v10"
)

type Service struct {
	validator *validator.Validate
	mu        sync.RWMutex
	replacers map[string]validationContracts.ReplacerFunc
}

// RegisterValidation initializes and returns the validation service factory.
func RegisterValidation(cfg *configs.Config) (validationContracts.Factory, error) {
	v := validator.New()

	// Set tag name function to use "json" tags for field names in errors
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("json")
		if tag == "" || tag == "-" {
			return fld.Name
		}
		name := strings.Split(tag, ",")[0]
		if name == "" {
			return fld.Name
		}
		return name
	})

	svc := &Service{validator: v}

	// Register default custom rules
	svc.registerLaravelAliases()
	svc.registerCustomValidations()

	return svc, nil
}

// Make creates a new validator instance.
func (s *Service) Make(data any, rules map[string]string, messages map[string]string, attributes map[string]string) validationContracts.Validator {
	s.mu.RLock()
	replacers := cloneReplacerMap(s.replacers)
	s.mu.RUnlock()

	return newInstance(s.validator, data, rules, messages, attributes, replacers)
}

// Extend registers a custom validation rule.
func (s *Service) Extend(rule string, extension validator.Func, message string) error {
	_ = message
	return s.validator.RegisterValidation(rule, extension)
}

// ExtendImplicit registers a custom implicit validation rule.
func (s *Service) ExtendImplicit(rule string, extension validator.Func, message string) error {
	_ = message
	return s.validator.RegisterValidation(rule, extension, true)
}

// Replacer registers a message replacer.
func (s *Service) Replacer(rule string, replacer validationContracts.ReplacerFunc) {
	if rule == "" || replacer == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.replacers == nil {
		s.replacers = make(map[string]validationContracts.ReplacerFunc)
	}

	s.replacers[rule] = replacer
}

// Validate is kept for compatibility with existing call sites.
func (s *Service) Validate(data any) error {
	return s.validator.Struct(data)
}
