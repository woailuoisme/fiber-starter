package i18n

import (
	"fiber-starter/app/helpers"
	"github.com/gofiber/fiber/v3"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
)

// Translator 翻译器
type Translator struct {
	localizer *i18n.Localizer
	lang      string
}

// contextKey 用于在 Fiber 上下文中存储翻译器的键
const contextKey = "translator"

// NewTranslator 创建翻译器
func NewTranslator(lang string) *Translator {
	localizer := GetLocalizer(lang)
	return &Translator{
		localizer: localizer,
		lang:      lang,
	}
}

// T 翻译消息（简单版本）
// 如果翻译不存在，返回 messageID 本身
func (t *Translator) T(messageID string) string {
	if t.localizer == nil {
		helpers.Warn("Localizer 未初始化", zap.String("messageID", messageID))
		return messageID
	}

	translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
		DefaultMessage: &i18n.Message{
			ID: messageID,
		},
	})

	if err != nil {
		helpers.Warn("翻译失败",
			zap.String("messageID", messageID),
			zap.String("language", t.lang),
			zap.Error(err))
		return messageID
	}

	return translation
}

// TWithData 翻译消息（带变量替换）
// data 是一个 map，包含要替换的变量
func (t *Translator) TWithData(messageID string, data map[string]interface{}) string {
	if t.localizer == nil {
		helpers.Warn("Localizer 未初始化", zap.String("messageID", messageID))
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
		helpers.Warn("翻译失败",
			zap.String("messageID", messageID),
			zap.String("language", t.lang),
			zap.Error(err))
		return messageID
	}

	return translation
}

// TPlural 翻译复数消息
// count 用于确定使用哪个复数形式
// data 包含要替换的变量（通常包含 Count）
func (t *Translator) TPlural(messageID string, count int, data map[string]interface{}) string {
	if t.localizer == nil {
		helpers.Warn("Localizer 未初始化", zap.String("messageID", messageID))
		return messageID
	}

	// 确保 data 中包含 Count
	if data == nil {
		data = make(map[string]interface{})
	}
	data["Count"] = count

	translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		PluralCount:  count,
		TemplateData: data,
		DefaultMessage: &i18n.Message{
			ID: messageID,
		},
	})

	if err != nil {
		helpers.Warn("翻译失败",
			zap.String("messageID", messageID),
			zap.String("language", t.lang),
			zap.Int("count", count),
			zap.Error(err))
		return messageID
	}

	return translation
}

// GetLanguage 获取当前语言
func (t *Translator) GetLanguage() string {
	return t.lang
}

// GetFromContext 从 Fiber 上下文获取翻译器
func GetFromContext(c fiber.Ctx) *Translator {
	translator := c.Locals(contextKey)
	if translator == nil {
		// 如果上下文中没有翻译器，创建一个默认的
		helpers.Warn("上下文中未找到翻译器，使用默认语言")
		return NewTranslator(DefaultLanguage)
	}

	if t, ok := translator.(*Translator); ok {
		return t
	}

	helpers.Warn("上下文中的翻译器类型错误，使用默认语言")
	return NewTranslator(DefaultLanguage)
}

// SetToContext 设置翻译器到上下文
func SetToContext(c fiber.Ctx, t *Translator) {
	c.Locals(contextKey, t)
}

// MustT 翻译消息，如果失败则 panic（用于必须成功的场景）
func (t *Translator) MustT(messageID string) string {
	translation := t.T(messageID)
	if translation == messageID {
		panic("翻译失败: " + messageID)
	}
	return translation
}

// TDefault 翻译消息，如果失败则返回默认值
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

// Exists 检查翻译键是否存在
func (t *Translator) Exists(messageID string) bool {
	if t.localizer == nil {
		return false
	}

	_, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})

	return err == nil
}
