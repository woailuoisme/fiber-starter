package providers_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"lfiber/configs"
	auth "lfiber/internal/providers/auth"
	backupProvider "lfiber/internal/providers/backup"
	cache "lfiber/internal/providers/cache"
	configProvider "lfiber/internal/providers/config"
	database "lfiber/internal/providers/database"
	hash "lfiber/internal/providers/hash"
	i18nProvider "lfiber/internal/providers/i18n"
	logging "lfiber/internal/providers/logging"
	mail "lfiber/internal/providers/mail"
	mailContracts "lfiber/internal/providers/mail/contracts"
	notification "lfiber/internal/providers/notification"
	notificationContracts "lfiber/internal/providers/notification/contracts"
	queue "lfiber/internal/providers/queue"
	queueContracts "lfiber/internal/providers/queue/contracts"
	ratelimiter "lfiber/internal/providers/ratelimiter"
	schedule "lfiber/internal/providers/schedule"
	search "lfiber/internal/providers/search"
	storage "lfiber/internal/providers/storage"
	storageDrivers "lfiber/internal/providers/storage/drivers"
	"lfiber/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestProviderEntryPoints_RegisterWrappers(t *testing.T) {
	t.Run("Config", func(t *testing.T) {
		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]interface{}{
			"app.name": "lfiber",
		}, "."), nil))

		repo, err := configProvider.RegisterConfig(k)
		require.NoError(t, err)
		require.NotNil(t, repo)
		assert.Equal(t, "lfiber", repo.GetString("app.name"))
	})

	t.Run("Database", func(t *testing.T) {
		cfg := testkit.NewSQLiteConfig(t)
		_, conn, err := database.RegisterDatabase(cfg)
		require.NoError(t, err)
		require.NotNil(t, conn)
		assert.Equal(t, "sqlite", conn.GetDriverName())
	})

	t.Run("Auth", func(t *testing.T) {
		cfg := testkit.NewSQLiteConfig(t)
		cfg.Hash.Driver = "bcrypt"
		cfg.Hash.Bcrypt.Rounds = 4
		_, conn, err := database.RegisterDatabase(cfg)
		require.NoError(t, err)

		db, err := conn.GetDB()
		require.NoError(t, err)
		_, err = db.Exec(`CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			email TEXT,
			password TEXT,
			name TEXT,
			avatar TEXT,
			bio TEXT,
			phone TEXT,
			status TEXT,
			email_verified_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`)
		require.NoError(t, err)

		hashed, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		require.NoError(t, err)
		_, err = db.Exec(
			`INSERT INTO users (email, password, name, status, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?)`,
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
		manager, err := auth.Register(cfg, conn, hasher)
		require.NoError(t, err)
		require.NotNil(t, manager)

		guard := manager.Guard()
		require.NotNil(t, guard)

		ctx, app := testkit.AcquireCtx(t)
		defer app.ReleaseCtx(ctx)

		assert.True(t, guard.Validate(map[string]string{
			"email":    "test@example.com",
			"password": "password",
		}))
		assert.False(t, guard.Validate(map[string]string{
			"email":    "test@example.com",
			"password": "wrong",
		}))
	})

	t.Run("Cache", func(t *testing.T) {
		cfg := testkit.DefaultConfig()
		cfg.Cache.Driver = "memory"

		_, store, err := cache.RegisterCache(cfg)
		require.NoError(t, err)
		require.NotNil(t, store)

		require.NoError(t, store.Put("entrypoint-cache", "value", time.Minute))
		if waiter, ok := store.(interface{ Wait() }); ok {
			waiter.Wait()
		}
		got, err := store.Get("entrypoint-cache")
		require.NoError(t, err)
		assert.Equal(t, "value", got)
	})

	t.Run("Hash", func(t *testing.T) {
		cfg := &configs.Config{
			Hash: configs.HashConfig{
				Driver: "bcrypt",
				Bcrypt: configs.BcryptHashConfig{Rounds: 10},
			},
		}

		manager, err := hash.RegisterHash(cfg)
		require.NoError(t, err)
		require.NotNil(t, manager)

		hasher := manager
		require.NotNil(t, hasher)

		hashed, err := hasher.Make("secret-password")
		require.NoError(t, err)
		assert.True(t, hasher.Check("secret-password", hashed))
	})

	t.Run("I18n", func(t *testing.T) {
		cfg := &configs.Config{
			I18n: configs.I18nConfig{
				Enabled:         true,
				DefaultLanguage: "en",
				LanguageDir:     testkit.RepoRoot(t) + "/lang",
			},
		}

		_, translator, err := i18nProvider.RegisterI18n(cfg)
		require.NoError(t, err)
		require.NotNil(t, translator)

		ctx, app := testkit.AcquireCtx(t)
		defer app.ReleaseCtx(ctx)

		assert.Equal(t, "en", translator.GetLocale(ctx))
		assert.NotEmpty(t, translator.Trans(ctx, "validation.required", map[string]interface{}{"attribute": "email"}))
	})

	t.Run("Logging", func(t *testing.T) {
		service, err := logging.Register(configs.LoggerConfig{Level: "info", Output: "stdout"})
		require.NoError(t, err)
		require.NotNil(t, service)
		require.NotNil(t, service.Default())
		assert.Equal(t, service.GetZapLogger(), service.Channel("default").GetZapLogger())
	})

	t.Run("Mail", func(t *testing.T) {
		cfg := &configs.Config{
			Mail: configs.MailConfig{
				Enabled:     true,
				Default:     "log",
				FromName:    "lfiber",
				FromAddress: "noreply@example.com",
				ReplyTo:     "",
				APIKey:      "",
				Host:        "localhost",
				Port:        1025,
				Username:    "",
				Password:    "",
				Encryption:  "",
			},
		}

		_, mailer, err := mail.Register(cfg)
		require.NoError(t, err)
		require.NotNil(t, mailer)

		require.NoError(t, mailer.Send(mailer.To("user@example.com").Subject("Subject").Plain("Body")))
	})

	t.Run("Notification", func(t *testing.T) {
		cfg := &configs.Config{
			Mail: configs.MailConfig{
				Enabled:     true,
				Default:     "log",
				FromName:    "lfiber",
				FromAddress: "noreply@example.com",
				Host:        "localhost",
				Port:        1025,
			},
		}

		_, mailer, err := mail.Register(cfg)
		require.NoError(t, err)
		_, dispatcher, err := notification.RegisterNotification(mailer)
		require.NoError(t, err)
		require.NotNil(t, dispatcher)

		err = dispatcher.Send(mailRecipient{}, testMailNotification{})
		require.NoError(t, err)
	})

	t.Run("Queue", func(t *testing.T) {
		cfg := testkit.DefaultConfig()
		cfg.Queue.Enabled = true
		cfg.Queue.Concurrency = 7
		cfg.Redis.Host = "127.0.0.1"
		cfg.Redis.Port = "6379"

		_, q, err := queue.RegisterQueue(cfg)
		require.NoError(t, err)
		require.NotNil(t, q)
		assert.Equal(t, 7, q.GetConcurrency())

		q.Register(&testQueueJob{})
		q.SetConcurrency(11)
		assert.Equal(t, 11, q.GetConcurrency())
	})

	t.Run("RateLimiter", func(t *testing.T) {
		limiter, err := ratelimiter.Register(configs.LimiterConfig{
			Default: "default",
			Strategies: map[string]configs.RateLimitConfig{
				"default": {Max: 2, Window: 60},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, limiter)

		strategy, ok := limiter.Strategy("default")
		require.True(t, ok)
		assert.Equal(t, 2, strategy.Max)

		app := fiber.New()
		app.Use(limiter.Middleware("default"))
		app.Get("/limited", func(c fiber.Ctx) error {
			return c.SendString("ok")
		})

		resp1, err := app.Test(httptest.NewRequest("GET", "/limited", nil))
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp1.StatusCode)

		resp2, err := app.Test(httptest.NewRequest("GET", "/limited", nil))
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp2.StatusCode)

		resp3, err := app.Test(httptest.NewRequest("GET", "/limited", nil))
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusTooManyRequests, resp3.StatusCode)
	})

	t.Run("Schedule", func(t *testing.T) {
		cfg := testkit.DefaultConfig()
		cfg.Redis.Host = "127.0.0.1"
		cfg.Redis.Port = "6379"

		_, scheduler, err := schedule.RegisterSchedule(cfg)
		require.NoError(t, err)
		require.NotNil(t, scheduler)

		event := scheduler.Job(&testQueueJob{}).
			DailyAt("08:30").
			Name("daily-job").
			Description("daily job")
		assert.Equal(t, "30 08 * * *", event.Expression)
		assert.Equal(t, "daily-job", event.NameStr)
		assert.Equal(t, "daily job", event.DescriptionStr)
		assert.Len(t, scheduler.GetEvents(), 1)
	})

	t.Run("Search", func(t *testing.T) {
		cfg := &configs.Config{
			Search: configs.SearchConfig{
				Enabled: true,
				Default: "null",
			},
		}

		_, engine, err := search.Register(cfg)
		require.NoError(t, err)
		require.NotNil(t, engine)

		err = engine.HealthCheck()
		require.NoError(t, err)

		info, err := engine.CreateIndex("test", "id")
		require.NoError(t, err)
		require.NotNil(t, info)
	})

	t.Run("Storage", func(t *testing.T) {
		cfg := &configs.Config{
			Storage: configs.StorageConfig{
				Driver: "local",
				Local: &configs.LocalStorageConfig{
					Root: t.TempDir(),
					URL:  "/storage",
				},
			},
		}

		manager, err := storage.Register(cfg)
		require.NoError(t, err)
		require.NotNil(t, manager)

		disk := manager.Disk("local")
		require.NotNil(t, disk)

		require.NoError(t, disk.Put("entrypoint.txt", []byte("hello")))
		exists, err := disk.Exists("entrypoint.txt")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Backup", func(t *testing.T) {
		cfg := testkit.NewSQLiteConfig(t)
		cfg.App.Name = "lfiber"
		cfg.Storage.Driver = "local"
		cfg.Storage.Local = &configs.LocalStorageConfig{Root: t.TempDir(), URL: "/storage"}
		cfg.Backup.Disk = "local"
		cfg.Backup.Path = "backups"
		cfg.Backup.TempPath = ".cache/backup"
		cfg.Backup.Notifications.Enabled = true
		cfg.Backup.Notifications.Channels = []string{"mail"}

		dbManager, _, err := database.RegisterDatabase(cfg)
		require.NoError(t, err)
		defer func() { _ = dbManager.CloseAll() }()
		storageManager, err := storage.Register(cfg)
		require.NoError(t, err)
		defer func() { _ = storageManager.Close() }()
		_, mailer, err := mail.Register(&configs.Config{Mail: configs.MailConfig{Enabled: true, Default: "log"}})
		require.NoError(t, err)
		_, dispatcher, err := notification.RegisterNotification(mailer)
		require.NoError(t, err)

		service, err := backupProvider.Register(cfg, dbManager, storageManager, dispatcher)
		require.NoError(t, err)
		require.NotNil(t, service)
	})
}

type testQueueJob struct{}

func (j *testQueueJob) Handle(ctx context.Context) error { return nil }
func (j *testQueueJob) TaskName() string                 { return "entrypoint:test-job" }
func (j *testQueueJob) QueueName() string                { return "default" }

type mailRecipient struct{}

func (mailRecipient) RouteNotificationForMail() string { return "user@example.com" }

type testMailNotification struct{}

func (testMailNotification) Via(notifiable interface{}) []string {
	return []string{"mail"}
}

func (testMailNotification) ToMail(notifiable interface{}) interface{} {
	return mail.NewMessage().To("user@example.com").Subject("Notification").Plain("body")
}

var (
	_ queueContracts.Job                     = (*testQueueJob)(nil)
	_ notificationContracts.MailNotification = testMailNotification{}
	_ mailContracts.Message                  = (*mail.Message)(nil)
)

func TestStorageS3Driver_ErrorPath(t *testing.T) {
	cfg := &configs.Config{}
	driver := storageDrivers.NewS3Driver(cfg, "s3")
	assert.Nil(t, driver)
}
