package i18n

import (
	"fiber-starter/internal/platform/helpers"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/gofiber/fiber/v3"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
)

// Translator Translator
type Translator struct {
	localizer *i18n.Localizer
	lang      string
	trans     ut.Translator
}

// contextKey Key for storing translator in Fiber context
const contextKey = "translator"

var (
	uni      *ut.UniversalTranslator
	validate *validator.Validate
)

func init() {
	en := en.New()
	zh := zh.New()
	uni = ut.New(en, en, zh)
	validate = validator.New()
}

// NewTranslator Create translator
func NewTranslator(lang string) *Translator {
	localizer := GetLocalizer(lang)

	// Get translator for corresponding language
	trans, found := uni.GetTranslator(lang)
	if !found {
		trans, _ = uni.GetTranslator("en")
	}

	// Register validator translations
	switch lang {
	case "zh-CN", "zh":
		_ = zh_translations.RegisterDefaultTranslations(validate, trans)
	default:
		_ = en_translations.RegisterDefaultTranslations(validate, trans)
	}

	return &Translator{
		localizer: localizer,
		lang:      lang,
		trans:     trans,
	}
}

// T Translate message (simple version)
// If translation doesn't exist, return messageID itself
func (t *Translator) T(messageID string) string {
	if t.localizer == nil {
		helpers.Warn("Localizer not initialized", zap.String("messageID", messageID))
		return messageID
	}

	translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
		DefaultMessage: &i18n.Message{
			ID: messageID,
		},
	})

	if err != nil {
		helpers.Warn("Translation failed",
			zap.String("messageID", messageID),
			zap.String("language", t.lang),
			zap.Error(err))
		return messageID
	}

	return translation
}

// TWithData Translate message (with variable substitution)
// data is a map containing variables to replace
func (t *Translator) TWithData(messageID string, data map[string]interface{}) string {
	if t.localizer == nil {
		helpers.Warn("Localizer not initialized", zap.String("messageID", messageID))
		return messageID
	}

	translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
		DefaultMessage: &i18n.Message{
			ID: messageID,
		},
	})

	if err != nil {
		helpers.Warn("Translation failed",
			zap.String("messageID", messageID),
			zap.String("language", t.lang),
			zap.Error(err))
		return messageID
	}

	return translation
}

// ValidateAndTranslate Validate and translate errors
func (t *Translator) ValidateAndTranslate(err error) map[string]string {
	errs := make(map[string]string)
	if err == nil {
		return errs
	}

	validatorErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		errs["error"] = err.Error()
		return errs
	}

	for _, e := range validatorErrs {
		errs[e.Field()] = e.Translate(t.trans)
	}

	return errs
}

// GetLanguage Get current language
func (t *Translator) GetLanguage() string {
	return t.lang
}

// GetFromContext Get translator from Fiber context
func GetFromContext(c fiber.Ctx) *Translator {
	translator := c.Locals(contextKey)
	if translator == nil {
		// If no translator in context, create a default one
		helpers.Warn("Translator not found in context, using default language")
		return NewTranslator(DefaultLanguage)
	}

	if t, ok := translator.(*Translator); ok {
		return t
	}

	helpers.Warn("Translator type in context is incorrect, using default language")
	return NewTranslator(DefaultLanguage)
}

// SetToContext Set translator to context
func SetToContext(c fiber.Ctx, t *Translator) {
	c.Locals(contextKey, t)
}

// MustT Translate message, panic if fails (for scenarios where success is required)
func (t *Translator) MustT(messageID string) string {
	translation := t.T(messageID)
	if translation == messageID {
		panic("Translation failed: " + messageID)
	}
	return translation
}

// TDefault Translate message, return default value if fails
func (t *Translator) TDefault(messageID string, defaultValue string) string {
	if t.localizer == nil {
		return defaultValue
	}

	translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
		DefaultMessage: &i18n.Message{
			ID:    messageID,
			Other: defaultValue,
		},
	})

	if err != nil {
		return defaultValue
	}

	return translation
}

// Exists Check if translation key exists
func (t *Translator) Exists(messageID string) bool {
	if t.localizer == nil {
		return false
	}

	_, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})

	return err == nil
}
