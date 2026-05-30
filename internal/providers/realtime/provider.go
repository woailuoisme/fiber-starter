package realtime

import (
	"errors"
	"fmt"
	"reflect"

	"lfiber/configs"
	realtimeContracts "lfiber/internal/providers/realtime/contracts"
	helpers "lfiber/internal/support"
	"lfiber/pkg/realtime"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// zapLoggerBridge 把应用 zap logger 适配为 realtime.Logger 接口
type zapLoggerBridge struct{}

func (zapLoggerBridge) Info(msg string, fields ...any) {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			zapFields = append(zapFields, zap.Any(fmt.Sprint(fields[i]), fields[i+1]))
		}
	}
	helpers.Info(msg, zapFields...)
}

func (zapLoggerBridge) Warn(msg string, fields ...any) {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			zapFields = append(zapFields, zap.Any(fmt.Sprint(fields[i]), fields[i+1]))
		}
	}
	helpers.Warn(msg, zapFields...)
}

func (zapLoggerBridge) Error(msg string, fields ...any) {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			zapFields = append(zapFields, zap.Any(fmt.Sprint(fields[i]), fields[i+1]))
		}
	}
	helpers.Error(msg, zapFields...)
}

// RegisterRealtime 将全局配置、默认连接及鉴权适配，实例化并装载到应用容器中
func RegisterRealtime(cfg *configs.Config) (realtimeContracts.Manager, error) {
	if !cfg.WebSocket.Enabled {
		return nil, nil
	}

	pkgCfg := &realtime.Config{
		Enabled:           cfg.WebSocket.Enabled,
		AppID:             cfg.WebSocket.AppID,
		AppKey:            cfg.WebSocket.AppKey,
		AppSecret:         cfg.WebSocket.AppSecret,
		Path:              cfg.WebSocket.Path,
		SSEPath:           cfg.WebSocket.SSEPath,
		BusMode:           cfg.WebSocket.BusMode,
		RedisPrefix:       cfg.WebSocket.RedisPrefix,
		WriteQueueSize:    cfg.WebSocket.WriteQueueSize,
		MaxMessageSize:    cfg.WebSocket.MaxMessageSize,
		HeartbeatInterval: cfg.WebSocket.HeartbeatInterval,
		PresenceTTL:       cfg.WebSocket.PresenceTTL,
		URL:               cfg.WebSocket.URL,
		ClientURL:         cfg.WebSocket.ClientURL,
		ClientSSEURL:      cfg.WebSocket.ClientSSEURL,
		APIKey:            cfg.WebSocket.APIKey,
		Secret:            cfg.WebSocket.Secret,
	}

	logger := zapLoggerBridge{}
	manager := realtime.NewManager(pkgCfg, logger)

	// 注入 HTTP Broadcasting Auth 广播路由的鉴权提取逻辑
	manager.SetAuthUserResolver(func(c fiber.Ctx) (realtime.User, error) {
		localUser := c.Locals("user")
		id, name, email, err := extractUserFields(localUser)
		if err != nil {
			return realtime.User{}, errors.New("unauthorized")
		}
		return realtime.User{
			ID: id,
			Info: map[string]any{
				"id":    id,
				"email": email,
				"name":  name,
			},
		}, nil
	})

	return manager, nil
}

// extractUserFields 通过反射动态抽取用户结构体字段，以消除对 features/user 包的物理依赖，切断循环导入
func extractUserFields(val any) (id string, name string, email string, err error) {
	if val == nil {
		return "", "", "", errors.New("user object is nil")
	}
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", "", "", errors.New("user object must be a struct pointer or struct")
	}

	// 动态解析 ID
	idField := v.FieldByName("ID")
	if idField.IsValid() {
		if idField.Kind() == reflect.Int64 {
			id = fmt.Sprintf("%d", idField.Int())
		} else if idField.Kind() == reflect.String {
			id = idField.String()
		}
	}

	// 动态解析 Name
	nameField := v.FieldByName("Name")
	if nameField.IsValid() && nameField.Kind() == reflect.String {
		name = nameField.String()
	}

	// 动态解析 Email
	emailField := v.FieldByName("Email")
	if emailField.IsValid() && emailField.Kind() == reflect.String {
		email = emailField.String()
	}

	if id == "" {
		return "", "", "", errors.New("failed to extract ID from user structure")
	}
	return id, name, email, nil
}
