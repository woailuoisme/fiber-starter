package logviewer

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"lfiber/configs"
	exceptions "lfiber/internal/common/exceptions"
	helpers "lfiber/internal/support"

	"github.com/gofiber/fiber/v3"
)

const (
	defaultPageLimit = 100
	maxPageLimit     = 1000
)

// Register 挂载日志查看大盘的路由及其所需的 API 端点
// 并集成安全所需的 Basic Auth 校验中间件
func Register(router fiber.Router, config configs.LogViewerConfig, logsDir string) {
	if !config.Enabled {
		return
	}

	if strings.TrimSpace(config.Path) != "" {
		logsDir = config.Path
	}

	basicAuth := basicAuthMiddleware(config)

	uiHandler := func(c fiber.Ctx) error {
		if !strings.HasSuffix(c.Path(), "/") {
			return c.Redirect().Status(fiber.StatusPermanentRedirect).To(c.Path() + "/")
		}

		htmlBytes, err := assetsFS.ReadFile("assets/index.html")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.Send(htmlBytes)
	}
	router.Get("/", basicAuth, uiHandler)
	router.Get("", basicAuth, uiHandler)

	router.Get("/assets/vendor/tw-animate.1.4.0.css", func(c fiber.Ctx) error {
		cssBytes, err := assetsFS.ReadFile("assets/vendor/tw-animate.1.4.0.css")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		c.Set(fiber.HeaderContentType, "text/css; charset=utf-8")
		return c.Send(cssBytes)
	})

	router.Get("/assets/vendor/fontawesome.7.2.0.min.css", func(c fiber.Ctx) error {
		cssBytes, err := assetsFS.ReadFile("assets/vendor/fontawesome.7.2.0.min.css")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		c.Set(fiber.HeaderContentType, "text/css; charset=utf-8")
		return c.Send(cssBytes)
	})

	router.Get("/assets/vendor/fontawesome-solid.7.2.0.min.css", func(c fiber.Ctx) error {
		cssBytes, err := assetsFS.ReadFile("assets/vendor/fontawesome-solid.7.2.0.min.css")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		c.Set(fiber.HeaderContentType, "text/css; charset=utf-8")
		return c.Send(cssBytes)
	})

	router.Get("/assets/vendor/fa-solid-900.7.2.0.woff2", func(c fiber.Ctx) error {
		fontBytes, err := assetsFS.ReadFile("assets/vendor/fa-solid-900.7.2.0.woff2")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		c.Set(fiber.HeaderContentType, "font/woff2")
		return c.Send(fontBytes)
	})

	router.Get("/assets/vendor/tailwind-browser.4.3.0.global.js", func(c fiber.Ctx) error {
		jsBytes, err := assetsFS.ReadFile("assets/vendor/tailwind-browser.4.3.0.global.js")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		c.Set(fiber.HeaderContentType, "application/javascript; charset=utf-8")
		return c.Send(jsBytes)
	})

	router.Get("/assets/vendor/alpine.3.15.12.min.js", func(c fiber.Ctx) error {
		jsBytes, err := assetsFS.ReadFile("assets/vendor/alpine.3.15.12.min.js")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		c.Set(fiber.HeaderContentType, "application/javascript; charset=utf-8")
		return c.Send(jsBytes)
	})

	api := router.Group("/api", basicAuth)

	api.Get("/files", func(c fiber.Ctx) error {
		files, err := listLogFiles(logsDir)
		if err != nil {
			return err
		}
		return helpers.HandleSuccess(c, "Log files fetched successfully", fiber.Map{"files": files})
	})

	api.Get("/entries", func(c fiber.Ctx) error {
		fileName := c.Query("file")
		if fileName == "" {
			return exceptions.NewBadRequestException("file parameter is required")
		}

		filePath, err := resolveLogFilePath(logsDir, fileName)
		if err != nil {
			return err
		}

		page, _ := strconv.Atoi(c.Query("page", "1"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(c.Query("limit", strconv.Itoa(defaultPageLimit)))
		if limit < 1 || limit > maxPageLimit {
			limit = defaultPageLimit
		}

		search := c.Query("search")
		level := c.Query("level")

		entries, totalMatches, stats, err := readLogFileBackward(filePath, page, limit, search, level)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return exceptions.NewNotFoundException("log file not found")
			}
			return err
		}

		return helpers.HandleSuccess(c, "Log entries fetched successfully", fiber.Map{
			"entries":       entries,
			"total_matches": totalMatches,
			"stats":         stats,
		})
	})

	api.Delete("/files/:name", func(c fiber.Ctx) error {
		if !config.AllowDelete {
			return exceptions.NewAuthorizationException("log deletion is disabled")
		}

		filePath, err := resolveLogFilePath(logsDir, c.Params("name"))
		if err != nil {
			return err
		}
		if err := os.Remove(filePath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return exceptions.NewNotFoundException("file not found")
			}
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	api.Get("/files/:name/download", func(c fiber.Ctx) error {
		filePath, err := resolveLogFilePath(logsDir, c.Params("name"))
		if err != nil {
			return err
		}
		if _, err := os.Stat(filePath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return exceptions.NewNotFoundException("file not found")
			}
			return err
		}

		return c.Download(filePath)
	})
}

func basicAuthMiddleware(config configs.LogViewerConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get(fiber.HeaderAuthorization)
		user, pass, ok := parseBasicAuth(authHeader)
		if !ok || !validBasicAuth(user, pass, config.Username, config.Password) {
			c.Set(fiber.HeaderWWWAuthenticate, `Basic realm="Log Viewer"`)
			return helpers.HandleUnauthorized(c, "Unauthorized")
		}
		return c.Next()
	}
}

func validBasicAuth(user, pass, expectedUser, expectedPass string) bool {
	expectedUser = strings.TrimSpace(expectedUser)
	expectedPass = strings.TrimSpace(expectedPass)
	if expectedUser == "" || expectedPass == "" {
		return false
	}

	userHash := sha256.Sum256([]byte(user))
	expectedUserHash := sha256.Sum256([]byte(expectedUser))
	passHash := sha256.Sum256([]byte(pass))
	expectedPassHash := sha256.Sum256([]byte(expectedPass))

	return subtle.ConstantTimeCompare(userHash[:], expectedUserHash[:]) == 1 &&
		subtle.ConstantTimeCompare(passHash[:], expectedPassHash[:]) == 1
}

func resolveLogFilePath(logsDir, fileName string) (string, error) {
	cleanName := filepath.Clean(fileName)
	if cleanName == "." || cleanName == ".." || strings.Contains(cleanName, "/") || strings.Contains(cleanName, "\\") {
		return "", exceptions.NewBadRequestException("invalid file name")
	}
	if !strings.HasSuffix(cleanName, ".log") {
		return "", exceptions.NewBadRequestException("invalid file name")
	}

	root, err := filepath.Abs(logsDir)
	if err != nil {
		return "", err
	}
	filePath, err := filepath.Abs(filepath.Join(root, cleanName))
	if err != nil {
		return "", err
	}
	if filepath.Dir(filePath) != root {
		return "", exceptions.NewBadRequestException("invalid file name")
	}

	return filePath, nil
}

// parseBasicAuth 解析 HTTP 请求头的 Basic 鉴权凭证
// 支持在 Fiber 剥离了 BasicAuth 中间件的 Fiber v3 引擎下稳定兼容
func parseBasicAuth(authHeader string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(authHeader) < len(prefix) || !strings.EqualFold(authHeader[:len(prefix)], prefix) {
		return "", "", false
	}
	c, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):])
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return "", "", false
	}
	return cs[:s], cs[s+1:], true
}

func Export_validBasicAuth(user, pass, expectedUser, expectedPass string) bool {
	return validBasicAuth(user, pass, expectedUser, expectedPass)
}
