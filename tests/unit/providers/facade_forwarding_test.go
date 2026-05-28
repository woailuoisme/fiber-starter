package providers_test

import (
	"context"
	"testing"
	"time"

	"lfiber/configs"
	models "lfiber/internal/features/user"
	artisanContracts "lfiber/internal/providers/artisan/contracts"
	auth "lfiber/internal/providers/auth"
	authContracts "lfiber/internal/providers/auth/contracts"
	authorizationContracts "lfiber/internal/providers/authorization/contracts"
	cache "lfiber/internal/providers/cache"
	cacheContracts "lfiber/internal/providers/cache/contracts"
	configContracts "lfiber/internal/providers/config/contracts"
	databaseContracts "lfiber/internal/providers/database/contracts"
	hashContracts "lfiber/internal/providers/hash/contracts"
	i18nContracts "lfiber/internal/providers/i18n/contracts"
	loggingContracts "lfiber/internal/providers/logging/contracts"
	mail "lfiber/internal/providers/mail"
	mailContracts "lfiber/internal/providers/mail/contracts"
	mailDrivers "lfiber/internal/providers/mail/drivers"
	notificationContracts "lfiber/internal/providers/notification/contracts"
	queue "lfiber/internal/providers/queue"
	queueContracts "lfiber/internal/providers/queue/contracts"
	ratelimiterContracts "lfiber/internal/providers/ratelimiter/contracts"
	scheduleContracts "lfiber/internal/providers/schedule/contracts"
	searchContracts "lfiber/internal/providers/search/contracts"
	storage "lfiber/internal/providers/storage"
	storageContracts "lfiber/internal/providers/storage/contracts"
	"lfiber/internal/support/appctx"
	"lfiber/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthFacade_ForwardsToGuard(t *testing.T) {
	guard := &fakeAuthGuard{
		user:     &models.User{ID: 42},
		check:    true,
		guest:    false,
		validate: true,
		attempt:  true,
	}
	manager := &fakeAuthManager{guard: guard}
	app := &fakeApp{auth: manager}
	appctx.Set(app)
	defer appctx.Clear(app)

	c, fiberApp := testkit.AcquireCtx(t)
	defer fiberApp.ReleaseCtx(c)

	assert.Same(t, guard, auth.Guard())
	assert.Same(t, guard, auth.Guard("api"))
	assert.Same(t, guard.user, auth.User(c))
	assert.Equal(t, int64(42), auth.Id(c))
	assert.True(t, auth.Check(c))
	auth.SetUser(c, guard.user)
	assert.True(t, auth.Attempt(c, map[string]string{"email": "user@example.com"}))
	assert.True(t, auth.Validate(map[string]string{"email": "user@example.com"}))
	require.NoError(t, auth.Logout(c))

	require.GreaterOrEqual(t, len(manager.requests), 2)
	assert.Equal(t, []string{"jwt", "api"}, manager.requests[:2])
	assert.Equal(t, 1, guard.setUserCalls)
	assert.Equal(t, 1, guard.logoutCalls)
}

func TestCacheFacade_ForwardsToStore(t *testing.T) {
	cfg := testkit.DefaultConfig()
	cfg.Cache.Driver = "memory"
	cfg.Cache.Prefix = "facade:"

	manager := cache.NewManager(cfg)
	store := manager.Store()
	app := &fakeApp{cacheManager: manager, cacheStore: store}
	appctx.Set(app)
	defer appctx.Clear(app)
	defer func() { _ = store.Close() }()

	gotStore, err := cache.GetStore("memory")
	require.NoError(t, err)
	require.NotNil(t, gotStore)
	waitForCache := func() {
		if waiter, ok := gotStore.(interface{ Wait() }); ok {
			waiter.Wait()
		}
	}

	require.NoError(t, cache.Set("value", "alpha", time.Minute))
	waitForCache()
	require.NoError(t, cache.Put("value2", "beta", time.Minute))
	waitForCache()

	added, err := cache.Add("added", "gamma", time.Minute)
	require.NoError(t, err)
	assert.True(t, added)
	waitForCache()

	require.NoError(t, cache.Forever("forever", "delta"))
	waitForCache()

	got, err := cache.Get("value")
	require.NoError(t, err)
	assert.Equal(t, "alpha", got)

	type cachedData struct {
		Name string `json:"name"`
	}
	require.NoError(t, cache.Set("json", cachedData{Name: "fiber"}, time.Minute))
	waitForCache()
	var decoded cachedData
	require.NoError(t, cache.GetJSON("json", &decoded))
	assert.Equal(t, "fiber", decoded.Name)

	exists, err := cache.Exists("value")
	require.NoError(t, err)
	assert.True(t, exists)

	has, err := cache.Has("value")
	require.NoError(t, err)
	assert.True(t, has)

	pulled, err := cache.Pull("value")
	require.NoError(t, err)
	assert.Equal(t, "alpha", pulled)

	require.NoError(t, cache.Delete("value2"))
	require.NoError(t, cache.Forget("added"))
	require.NoError(t, cache.Flush())
	waitForCache()

	_, err = cache.Increment("missing")
	require.Error(t, err)
	_, err = cache.Decrement("missing")
	require.Error(t, err)

	typedStore := gotStore.(interface {
		GetBytes(string) ([]byte, error)
		DeletePattern(string) error
		TTL(string) (time.Duration, error)
		Expire(string, time.Duration) error
	})
	require.NoError(t, typedStore.DeletePattern("*"))
	raw, err := typedStore.GetBytes("missing")
	require.NoError(t, err)
	assert.Empty(t, raw)
	_, err = typedStore.TTL("missing")
	require.Error(t, err)
	require.NoError(t, typedStore.Expire("missing", time.Minute))
}

func TestMailFacade_ForwardsToManagerAndMailer(t *testing.T) {
	message := mail.NewMessage()
	mailer := &fakeMailer{message: message}
	manager := &fakeMailManager{mailer: mailer}
	app := &fakeApp{mailManager: manager, emailService: mailer}
	appctx.Set(app)
	defer appctx.Clear(app)

	assert.Same(t, mailer, mail.Drive("log"))
	assert.Equal(t, []string{"log"}, manager.driveRequests)

	msg := mail.To("user@example.com")
	assert.Same(t, message, msg)
	assert.Equal(t, []string{"user@example.com"}, mailer.toRequests[len(mailer.toRequests)-1])

	chain := mail.NewMessage().
		To("to@example.com").
		Cc("cc@example.com").
		Bcc("bcc@example.com").
		Subject("subject").
		Html("<b>body</b>").
		Plain("plain body").
		Attach("/tmp/file.txt").
		Data(map[string]interface{}{"k": "v"})

	assert.Equal(t, []string{"to@example.com"}, chain.GetTo())
	assert.Equal(t, []string{"cc@example.com"}, chain.GetCc())
	assert.Equal(t, []string{"bcc@example.com"}, chain.GetBcc())
	assert.Equal(t, "subject", chain.GetSubject())
	assert.Equal(t, "plain body", chain.GetBody())
	assert.False(t, chain.IsHtml())
	assert.Equal(t, []string{"/tmp/file.txt"}, chain.GetAttachments())
	assert.Equal(t, map[string]interface{}{"k": "v"}, chain.GetData())

	require.NoError(t, mail.Send(message))
	require.NoError(t, mail.Raw("user@example.com", "subject", "body"))
	require.NoError(t, mail.Close())

	assert.Equal(t, 1, mailer.sendCalls)
	assert.Equal(t, 1, mailer.rawCalls)
	assert.Equal(t, 1, manager.closeCalls)
}

func TestMailBaseMessage_Builders(t *testing.T) {
	message := &mailDrivers.BaseMessage{}

	message.To("one@example.com").
		Cc("cc@example.com").
		Bcc("bcc@example.com").
		Subject("subject").
		Html("<p>body</p>").
		Attach("/tmp/file.txt").
		Data(map[string]interface{}{"k": "v"})

	assert.Equal(t, []string{"one@example.com"}, message.GetTo())
	assert.Equal(t, []string{"cc@example.com"}, message.GetCc())
	assert.Equal(t, []string{"bcc@example.com"}, message.GetBcc())
	assert.Equal(t, "subject", message.GetSubject())
	assert.Equal(t, "<p>body</p>", message.GetBody())
	assert.True(t, message.IsHtml())
	assert.Equal(t, []string{"/tmp/file.txt"}, message.GetAttachments())
	assert.Equal(t, map[string]interface{}{"k": "v"}, message.GetData())
}

func TestStorageFacade_ForwardsToManagerAndDisk(t *testing.T) {
	disk := &fakeDisk{
		getValue:       []byte("hello"),
		exists:         true,
		url:            "/storage/path.txt",
		temporaryURL:   "/tmp/path.txt",
		size:           5,
		lastModified:   time.Unix(100, 0),
		files:          []string{"files/a.txt"},
		allFiles:       []string{"files/a.txt", "files/b.txt"},
		directories:    []string{"files"},
		allDirectories: []string{"files", "files/sub"},
		visibility:     "public",
	}
	manager := &fakeStorageManager{disk: disk}
	app := &fakeApp{storageManager: manager}
	appctx.Set(app)
	defer appctx.Clear(app)

	gotDisk, err := storage.GetDisk("local")
	require.NoError(t, err)
	require.Same(t, disk, gotDisk)
	assert.Equal(t, []string{"local"}, manager.diskRequests)
	require.Same(t, disk, storage.Drive("local"))

	contents, err := storage.Get("path.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), contents)
	require.NoError(t, storage.Put("path.txt", []byte("hello")))
	require.NoError(t, storage.Delete("path.txt"))
	exists, err := storage.Exists("path.txt")
	require.NoError(t, err)
	assert.True(t, exists)
	missing, err := storage.Missing("missing.txt")
	require.NoError(t, err)
	assert.False(t, missing)
	assert.Equal(t, "/storage/path.txt", storage.Url("path.txt"))
	tmpURL, err := storage.TemporaryUrl("path.txt", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/path.txt", tmpURL)
	size, err := storage.Size("path.txt")
	require.NoError(t, err)
	assert.EqualValues(t, 5, size)
	modified, err := storage.LastModified("path.txt")
	require.NoError(t, err)
	assert.Equal(t, time.Unix(100, 0), modified)
	require.NoError(t, storage.Copy("from.txt", "to.txt"))
	require.NoError(t, storage.Move("from.txt", "to.txt"))
	require.NoError(t, storage.Prepend("path.txt", []byte("pre")))
	require.NoError(t, storage.Append("path.txt", []byte("post")))

	files, err := storage.Files("files", true)
	require.NoError(t, err)
	assert.Equal(t, []string{"files/a.txt"}, files)
	files, err = storage.AllFiles("files")
	require.NoError(t, err)
	assert.Equal(t, []string{"files/a.txt", "files/b.txt"}, files)
	dirs, err := storage.Directories("files")
	require.NoError(t, err)
	assert.Equal(t, []string{"files"}, dirs)
	dirs, err = storage.AllDirectories("files")
	require.NoError(t, err)
	assert.Equal(t, []string{"files", "files/sub"}, dirs)
	require.NoError(t, storage.MakeDirectory("new"))
	require.NoError(t, storage.DeleteDirectory("new"))
	visibility, err := storage.GetVisibility("path.txt")
	require.NoError(t, err)
	assert.Equal(t, "public", visibility)
	require.NoError(t, storage.SetVisibility("path.txt", "private"))
}

func TestQueueFacade_ForwardsToManagerAndService(t *testing.T) {
	service := &fakeQueueService{}
	manager := &fakeQueueManager{queue: service}
	app := &fakeApp{queueManager: manager, queueService: service}
	appctx.Set(app)
	defer appctx.Clear(app)

	assert.Same(t, service, queue.Drive("asynq"))
	assert.Equal(t, []string{"asynq"}, manager.driveRequests)

	job := &dummyJob{}
	require.NoError(t, queue.Push(job))
	require.NoError(t, queue.PushOn("high", job))
	require.NoError(t, queue.Later(5*time.Second, job))
	require.NoError(t, queue.LaterOn("high", 10*time.Second, job))
	require.NoError(t, queue.Bulk([]queueContracts.Job{job}, "low"))
	require.NoError(t, queue.ProcessAt(time.Unix(100, 0), job))
	queue.Register(job)
	require.NoError(t, queue.StartWorker("low"))
	require.NoError(t, queue.RunWorker("low"))
	require.NoError(t, queue.StopWorker())

	statuses, err := queue.ListFailed(1, 20)
	require.NoError(t, err)
	assert.Len(t, statuses, 1)

	require.NoError(t, queue.RetryFailed("failed-1"))
	require.NoError(t, queue.DeleteFailed("failed-1"))
	require.NoError(t, queue.Flush("default"))
	require.NoError(t, queue.Close())

	assert.Equal(t, 1, service.pushCalls)
	assert.Equal(t, "high", service.pushOnQueue)
	assert.Equal(t, "low", service.bulkQueue)
	assert.True(t, service.registered)
	assert.Equal(t, 1, manager.closeCalls)
}

type fakeAuthGuard struct {
	user         *models.User
	check        bool
	guest        bool
	validate     bool
	attempt      bool
	setUserCalls int
	logoutCalls  int
}

func (g *fakeAuthGuard) Check(fiber.Ctx) bool                      { return g.check }
func (g *fakeAuthGuard) Guest(fiber.Ctx) bool                      { return g.guest }
func (g *fakeAuthGuard) User(fiber.Ctx) any                        { return g.user }
func (g *fakeAuthGuard) Id(fiber.Ctx) int64                        { return 42 }
func (g *fakeAuthGuard) SetUser(c fiber.Ctx, user any)             { g.setUserCalls++ }
func (g *fakeAuthGuard) Validate(map[string]string) bool           { return g.validate }
func (g *fakeAuthGuard) Attempt(fiber.Ctx, map[string]string) bool { return g.attempt }
func (g *fakeAuthGuard) Login(c fiber.Ctx, user any) error         { return nil }
func (g *fakeAuthGuard) LoginUsingId(fiber.Ctx, int64) error       { return nil }
func (g *fakeAuthGuard) Logout(fiber.Ctx) error                    { g.logoutCalls++; return nil }

type fakeAuthManager struct {
	guard    authContracts.Guard
	requests []string
}

func (m *fakeAuthManager) Guard(name ...string) authContracts.Guard {
	if len(name) == 0 {
		m.requests = append(m.requests, "jwt")
	} else {
		m.requests = append(m.requests, name[0])
	}
	return m.guard
}

func (m *fakeAuthManager) SetModelCreator(provider string, creator func() any) {}

type fakeMailManager struct {
	mailer        mailContracts.Mailer
	driveRequests []string
	closeCalls    int
}

func (m *fakeMailManager) Drive(name ...string) mailContracts.Mailer {
	if len(name) == 0 {
		m.driveRequests = append(m.driveRequests, "default")
	} else {
		m.driveRequests = append(m.driveRequests, name[0])
	}
	return m.mailer
}

func (m *fakeMailManager) Close() error {
	m.closeCalls++
	return nil
}

type fakeMailer struct {
	message    mailContracts.Message
	toRequests [][]string
	sendCalls  int
	rawCalls   int
	closeCalls int
	sendErr    error
	rawErr     error
	closeErr   error
}

func (m *fakeMailer) To(to ...string) mailContracts.Message {
	m.toRequests = append(m.toRequests, append([]string(nil), to...))
	return m.message
}

func (m *fakeMailer) Send(mailContracts.Message) error {
	m.sendCalls++
	return m.sendErr
}

func (m *fakeMailer) Raw(_, _, _ string) error {
	m.rawCalls++
	return m.rawErr
}

func (m *fakeMailer) Close() error {
	m.closeCalls++
	return m.closeErr
}

func (m *fakeMailer) HealthCheck() error {
	return nil
}

type fakeQueueManager struct {
	queue         queueContracts.Queue
	driveRequests []string
	closeCalls    int
}

func (m *fakeQueueManager) Drive(name ...string) queueContracts.Queue {
	if len(name) == 0 {
		m.driveRequests = append(m.driveRequests, "asynq")
	} else {
		m.driveRequests = append(m.driveRequests, name[0])
	}
	return m.queue
}

func (m *fakeQueueManager) Close() error {
	m.closeCalls++
	return nil
}

type fakeQueueService struct {
	pushCalls   int
	pushOnQueue string
	bulkQueue   string
	registered  bool
}

func (s *fakeQueueService) Push(queueContracts.Job) error { s.pushCalls++; return nil }
func (s *fakeQueueService) Size(...string) (int64, error) { return 0, nil }
func (s *fakeQueueService) PushOn(queue string, _ queueContracts.Job) error {
	s.pushOnQueue = queue
	return nil
}
func (s *fakeQueueService) Later(time.Duration, queueContracts.Job) error           { return nil }
func (s *fakeQueueService) LaterOn(string, time.Duration, queueContracts.Job) error { return nil }
func (s *fakeQueueService) Bulk(_ []queueContracts.Job, queueName ...string) error {
	if len(queueName) > 0 {
		s.bulkQueue = queueName[0]
	}
	return nil
}
func (s *fakeQueueService) ProcessAt(time.Time, queueContracts.Job) error { return nil }
func (s *fakeQueueService) Register(queueContracts.Job)                   { s.registered = true }
func (s *fakeQueueService) StartWorker(...string) error                   { return nil }
func (s *fakeQueueService) RunWorker(...string) error                     { return nil }
func (s *fakeQueueService) StopWorker() error                             { return nil }
func (s *fakeQueueService) InspectQueues() ([]queueContracts.QueueStatus, error) {
	return []queueContracts.QueueStatus{{Name: "default"}}, nil
}

func (s *fakeQueueService) ListFailed(int, int) ([]queueContracts.FailedJob, error) {
	return []queueContracts.FailedJob{{ID: "failed-1"}}, nil
}
func (s *fakeQueueService) RetryFailed(string) error  { return nil }
func (s *fakeQueueService) DeleteFailed(string) error { return nil }
func (s *fakeQueueService) Flush(string) error        { return nil }
func (s *fakeQueueService) Close() error              { return nil }
func (s *fakeQueueService) HealthCheck() error        { return nil }
func (s *fakeQueueService) SetConcurrency(int)        {}
func (s *fakeQueueService) GetConcurrency() int       { return 0 }

type dummyJob struct{}

func (j *dummyJob) Handle(context.Context) error { return nil }
func (j *dummyJob) TaskName() string             { return "dummy-job" }
func (j *dummyJob) QueueName() string            { return "default" }

type fakeApp struct {
	artisan       artisanContracts.Artisan
	auth           authContracts.Manager
	authorization  authorizationContracts.Authorizer
	cacheManager   cacheContracts.Manager
	cacheStore     cacheContracts.Store
	mailManager    mailContracts.Manager
	emailService   mailContracts.Mailer
	queueManager   queueContracts.Manager
	queueService   queueContracts.Queue
	storageManager storageContracts.StorageManager
}

func (a *fakeApp) AppConfig() *configs.Config                    { return nil }
func (a *fakeApp) ArtisanService() artisanContracts.Artisan       { return a.artisan }
func (a *fakeApp) ConfigRepository() configContracts.Repository  { return nil }
func (a *fakeApp) DatabaseManager() databaseContracts.Manager    { return nil }
func (a *fakeApp) ConnectionValue() databaseContracts.Connection { return nil }
func (a *fakeApp) CacheManagerValue() cacheContracts.Manager     { return a.cacheManager }
func (a *fakeApp) CacheStore() cacheContracts.Store              { return a.cacheStore }
func (a *fakeApp) AuthManager() authContracts.Manager            { return a.auth }
func (a *fakeApp) AuthorizationService() authorizationContracts.Authorizer {
	return a.authorization
}
func (a *fakeApp) MailManagerValue() mailContracts.Manager               { return a.mailManager }
func (a *fakeApp) EmailServiceValue() mailContracts.Mailer               { return a.emailService }
func (a *fakeApp) QueueManagerValue() queueContracts.Manager             { return a.queueManager }
func (a *fakeApp) QueueServiceValue() queueContracts.Queue               { return a.queueService }
func (a *fakeApp) ScheduleManagerValue() scheduleContracts.Manager       { return nil }
func (a *fakeApp) ScheduleServiceValue() scheduleContracts.Scheduler     { return nil }
func (a *fakeApp) SearchManagerValue() searchContracts.Manager           { return nil }
func (a *fakeApp) SearchServiceValue() searchContracts.Engine            { return nil }
func (a *fakeApp) StorageValue() storageContracts.StorageManager         { return a.storageManager }
func (a *fakeApp) HashService() hashContracts.Hasher                     { return nil }
func (a *fakeApp) NotificationService() notificationContracts.Dispatcher { return nil }
func (a *fakeApp) TranslatorService() i18nContracts.Translator           { return nil }
func (a *fakeApp) LogService() loggingContracts.Logger                   { return nil }
func (a *fakeApp) RateLimiterService() ratelimiterContracts.Limiter      { return nil }

type fakeStorageManager struct {
	disk         storageContracts.Disk
	diskRequests []string
}

func (m *fakeStorageManager) Disk(name ...string) storageContracts.Disk {
	if len(name) == 0 {
		m.diskRequests = append(m.diskRequests, "default")
	} else {
		m.diskRequests = append(m.diskRequests, name[0])
	}
	return m.disk
}

func (m *fakeStorageManager) Close() error { return nil }

type fakeDisk struct {
	getValue       []byte
	exists         bool
	url            string
	temporaryURL   string
	size           int64
	lastModified   time.Time
	files          []string
	allFiles       []string
	directories    []string
	allDirectories []string
	visibility     string
}

func (d *fakeDisk) Get(string) ([]byte, error)                         { return d.getValue, nil }
func (d *fakeDisk) Put(string, []byte, ...interface{}) error           { return nil }
func (d *fakeDisk) Exists(string) (bool, error)                        { return d.exists, nil }
func (d *fakeDisk) Missing(string) (bool, error)                       { return !d.exists, nil }
func (d *fakeDisk) Url(string) string                                  { return d.url }
func (d *fakeDisk) TemporaryUrl(string, time.Duration) (string, error) { return d.temporaryURL, nil }
func (d *fakeDisk) Size(string) (int64, error)                         { return d.size, nil }
func (d *fakeDisk) LastModified(string) (time.Time, error)             { return d.lastModified, nil }
func (d *fakeDisk) Copy(_, _ string) error                             { return nil }
func (d *fakeDisk) Move(_, _ string) error                             { return nil }
func (d *fakeDisk) Prepend(string, []byte) error                       { return nil }
func (d *fakeDisk) Append(string, []byte) error                        { return nil }
func (d *fakeDisk) Delete(...string) error                             { return nil }
func (d *fakeDisk) Files(string, ...bool) ([]string, error)            { return d.files, nil }
func (d *fakeDisk) AllFiles(string) ([]string, error)                  { return d.allFiles, nil }
func (d *fakeDisk) Directories(string, ...bool) ([]string, error)      { return d.directories, nil }
func (d *fakeDisk) AllDirectories(string) ([]string, error)            { return d.allDirectories, nil }
func (d *fakeDisk) MakeDirectory(string) error                         { return nil }
func (d *fakeDisk) DeleteDirectory(string) error                       { return nil }
func (d *fakeDisk) GetVisibility(string) (string, error)               { return d.visibility, nil }
func (d *fakeDisk) SetVisibility(string, string) error                 { return nil }
func (d *fakeDisk) Reset() error                                       { return nil }
func (d *fakeDisk) Close() error                                       { return nil }
func (d *fakeDisk) HealthCheck() error                                 { return nil }
