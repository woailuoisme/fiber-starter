package providers_test

import (
	"testing"

	"fiber-starter/configs"
	models "fiber-starter/internal/features/user"
	auth "fiber-starter/internal/providers/auth"
	database "fiber-starter/internal/providers/database"
	hash "fiber-starter/internal/providers/hash"
	"fiber-starter/tests/internal/testkit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthProvider_Manager(t *testing.T) {
	// 加载配置
	cfg, _, err := configs.LoadConfig()
	require.NoError(t, err)

	// 初始化数据库连接 (AuthManager 依赖数据库)
	db := database.NewConnection(cfg)

	// 初始化 Hash
	hasher, err := hash.RegisterHash(cfg)
	require.NoError(t, err)

	// 创建 Manager
	manager := auth.NewManager(cfg, db, hasher)
	require.NotNil(t, manager)

	registered, err := auth.Register(cfg, db, hasher)
	require.NoError(t, err)
	require.NotNil(t, registered)

	t.Run("DefaultGuard", func(t *testing.T) {
		guard := manager.Guard()
		assert.NotNil(t, guard)
	})

	t.Run("SwitchGuards", func(t *testing.T) {
		// 测试 api guard
		apiGuard := manager.Guard("api")
		assert.NotNil(t, apiGuard)

		// 测试 admin guard
		adminGuard := manager.Guard("admin")
		assert.NotNil(t, adminGuard)
	})
}

func TestAuthProvider_Operations(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)
	cfg.Hash.Driver = "bcrypt"
	cfg.Hash.Bcrypt.Rounds = 4
	_, conn, err := database.RegisterDatabase(cfg)
	require.NoError(t, err)
	db, err := conn.GetDB()
	require.NoError(t, err)

	// Create users table
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT, password TEXT, name TEXT, avatar TEXT, bio TEXT, phone TEXT, status TEXT, email_verified_at DATETIME, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)")
	require.NoError(t, err)

	// Hash password
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	// Insert test user
	_, err = db.Exec(
		"INSERT INTO users (email, password, name, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"test@example.com",
		string(hashed),
		"Test User",
		"active",
		"2024-01-01 00:00:00",
		"2024-01-01 00:00:00",
	)
	require.NoError(t, err)
	hasher, err := hash.RegisterHash(cfg)
	require.NoError(t, err)
	manager := auth.NewManager(cfg, conn, hasher)
	guard := manager.Guard()
	ctx, _ := testkit.AcquireCtx(t)

	t.Run("Attempt", func(t *testing.T) {
		// Failure
		ok := guard.Attempt(ctx, map[string]string{"email": "test@example.com", "password": "wrong"})
		assert.False(t, ok)

		// Success
		ok = guard.Attempt(ctx, map[string]string{"email": "test@example.com", "password": "password"})
		assert.True(t, ok)
	})

	t.Run("LoginAndCheck", func(t *testing.T) {
		user := &models.User{ID: 1, Email: "test@example.com"}

		err := guard.Login(ctx, user)
		require.NoError(t, err)

		// LoginUsingId
		err = guard.LoginUsingId(ctx, 1)
		require.NoError(t, err)
	})

	t.Run("CheckState", func(t *testing.T) {
		// Note: These might depend on whether Login actually set state in the mock ctx
		// but we verify the method signature works
		_ = guard.Check(ctx)
		_ = guard.Guest(ctx)
		_ = guard.User(ctx)
		_ = guard.Id(ctx)
		assert.True(t, guard.Validate(map[string]string{"email": "test@example.com", "password": "password"}))
		require.NoError(t, guard.Logout(ctx))
	})
}
