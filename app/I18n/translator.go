package i18n

import (
	"errors"

	helpers "fiber-starter/app/Support"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
	zhtranslations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/gofiber/fiber/v3"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
)

type Translator struct {
	localizer *i18n.Localizer
	lang      string
	trans     ut.Translator
}

const contextKey = "translator"

var uni *ut.UniversalTranslator

func init() {
	enLocale := en.New()
	zhLocale := zh.New()
	uni = ut.New(enLocale, enLocale, zhLocale)
}

func registerValidatorTranslations(lang string, trans ut.Translator) {
	validate := validator.New()
	switch lang {
	case "zh-CN", "zh":
		_ = zhtranslations.RegisterDefaultTranslations(validate, trans)
	default:
		_ = entranslations.RegisterDefaultTranslations(validate, trans)
	}
}

func NewTranslator(lang string) *Translator {
	localizer := GetLocalizer(lang)

	trans, found := uni.GetTranslator(lang)
	if !found {
		trans, _ = uni.GetTranslator("en")
	}

	registerValidatorTranslations(lang, trans)

	return &Translator{
		localizer: localizer,
		lang:      lang,
		trans:     trans,
	}
}

func (t *Translator) localize(messageID string, cfg *i18n.LocalizeConfig) string {
	if t.localizer == nil {
		helpers.Warn("Localizer not initialized", zap.String("messageID", messageID))
		return messageID
	}

	translation, err := t.localizer.Localize(cfg)
	if err != nil {
		helpers.Warn("Translation failed",
			zap.String("messageID", messageID),
			zap.String("language", t.lang),
			zap.Error(err))
		return messageID
	}

	return translation
}

func (t *Translator) T(messageID string) string {
	return t.localize(messageID, &i18n.LocalizeConfig{
		MessageID: messageID,
		DefaultMessage: &i18n.Message{
			ID: messageID,
		},
	})
}

func (t *Translator) TWithData(messageID string, data map[string]interface{}) string {
	return t.localize(messageID, &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
		DefaultMessage: &i18n.Message{
			ID: messageID,
		},
	})
}

func (t *Translator) ValidateAndTranslate(err error) map[string]string {
	errs := make(map[string]string)
	if err == nil {
		return errs
	}

	var validatorErrs validator.ValidationErrors
	ok := errors.As(err, &validatorErrs)
	if !ok {
		errs["error"] = err.Error()
		return errs
	}

	for _, e := range validatorErrs {
		errs[e.Field()] = e.Translate(t.trans)
	}

	return errs
}

func (t *Translator) GetLanguage() string {
	return t.lang
}

func GetFromContext(c fiber.Ctx) *Translator {
	translator := c.Locals(contextKey)
	if translator == nil {
		helpers.Warn("Translator not found in context, using default language")
		return NewTranslator(DefaultLanguage)
	}

	if t, ok := translator.(*Translator); ok {
		return t
	}

	helpers.Warn("Translator type in context is incorrect, using default language")
	return NewTranslator(DefaultLanguage)
}

func SetToContext(c fiber.Ctx, t *Translator) {
	c.Locals(contextKey, t)
}

func (t *Translator) MustT(messageID string) string {
	translation := t.T(messageID)
	if translation == messageID {
		panic("Translation failed: " + messageID)
	}
	return translation
}

func (t *Translator) TDefault(messageID string, defaultValue string) string {
	translation := t.localize(messageID, &i18n.LocalizeConfig{
		MessageID: messageID,
		DefaultMessage: &i18n.Message{
			ID:    messageID,
			Other: defaultValue,
		},
	})
	if translation == messageID {
		return defaultValue
	}
	return translation
}

func (t *Translator) Exists(messageID string) bool {
	if t.localizer == nil {
		return false
	}

	_, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})

	return err == nil
}
