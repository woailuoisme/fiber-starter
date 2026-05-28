package authorization

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"lfiber/configs"
	exceptions "lfiber/internal/common/exceptions"
	middleware "lfiber/internal/common/middleware"
	authorizationContracts "lfiber/internal/providers/authorization/contracts"

	casbinv2 "github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	fibercasbin "github.com/gofiber/contrib/v3/casbin"
	"github.com/gofiber/fiber/v3"
)

type Service struct {
	middleware *fibercasbin.Middleware
}

const (
	defaultModelFile  = "configs/casbin/model.conf"
	defaultPolicyFile = "configs/casbin/policy.csv"
)

// Register initializes the Casbin authorizer from local model and policy files.
func Register(cfg configs.AuthorizationConfig) (authorizationContracts.Authorizer, error) {
	if cfg.ModelFile == "" {
		cfg.ModelFile = defaultModelFile
	}
	if cfg.PolicyFile == "" {
		cfg.PolicyFile = defaultPolicyFile
	}

	modelFile, err := resolveLocalPath(cfg.ModelFile)
	if err != nil {
		return nil, fmt.Errorf("resolve authorization model file: %w", err)
	}
	policyFile, err := resolveLocalPath(cfg.PolicyFile)
	if err != nil {
		return nil, fmt.Errorf("resolve authorization policy file: %w", err)
	}

	enforcer, err := casbinv2.NewEnforcer(modelFile, fileadapter.NewAdapter(policyFile))
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}

	return &Service{
		middleware: fibercasbin.New(fibercasbin.Config{
			Enforcer:     enforcer,
			Lookup:       SubjectFromJWT,
			Unauthorized: func(_ fiber.Ctx) error { return exceptions.NewAuthenticationException("Unauthenticated") },
			Forbidden:    func(_ fiber.Ctx) error { return exceptions.NewAuthorizationException("Forbidden") },
		}),
	}, nil
}

// RequirePermissions returns middleware that requires every named permission.
func (s *Service) RequirePermissions(permissions ...string) fiber.Handler {
	if s == nil || s.middleware == nil {
		return func(_ fiber.Ctx) error {
			return exceptions.NewAuthorizationException("Authorization service unavailable")
		}
	}
	return s.middleware.RequiresPermissions(permissions, fibercasbin.WithValidationRule(fibercasbin.MatchAllRule))
}

// Subject returns the current Casbin subject for a Fiber request.
func (s *Service) Subject(c fiber.Ctx) string {
	return SubjectFromJWT(c)
}

// SubjectFromJWT maps parsed JWT claims to the Casbin subject key.
func SubjectFromJWT(c fiber.Ctx) string {
	claims := middleware.GetCurrentUser(c)
	if claims == nil || claims.UserID <= 0 {
		return ""
	}
	return fmt.Sprintf("user:%d", claims.UserID)
}

func resolveLocalPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path is empty")
	}
	if filepath.IsAbs(path) {
		if err := ensureFile(path); err != nil {
			return "", err
		}
		return path, nil
	}

	candidate := path
	for i := 0; i <= 3; i++ {
		if err := ensureFile(candidate); err == nil {
			return candidate, nil
		}
		candidate = filepath.Join("..", candidate)
	}

	return "", fmt.Errorf("%s does not exist", path)
}

func ensureFile(path string) error {
	info, err := os.Stat(path) //nolint:gosec // local config path controlled by deployment config
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	return nil
}
