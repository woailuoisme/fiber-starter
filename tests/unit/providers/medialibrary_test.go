package providers_test

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"lfiber/internal/providers/database"
	queueContracts "lfiber/internal/providers/queue/contracts"
	"lfiber/internal/providers/storage"
	"lfiber/pkg/medialibrary"
	"lfiber/tests/internal/testkit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureQueue struct {
	jobs []queueContracts.Job
}

func (q *captureQueue) Push(job queueContracts.Job) error {
	q.jobs = append(q.jobs, job)
	return nil
}

func (q *captureQueue) Size(queue ...string) (int64, error) { return int64(len(q.jobs)), nil }
func (q *captureQueue) PushOn(_ string, job queueContracts.Job) error {
	return q.Push(job)
}
func (q *captureQueue) Later(_ time.Duration, job queueContracts.Job) error { return q.Push(job) }
func (q *captureQueue) LaterOn(_ string, _ time.Duration, job queueContracts.Job) error {
	return q.Push(job)
}

func (q *captureQueue) Bulk(jobs []queueContracts.Job, queue ...string) error {
	q.jobs = append(q.jobs, jobs...)
	return nil
}
func (q *captureQueue) ProcessAt(_ time.Time, job queueContracts.Job) error { return q.Push(job) }
func (q *captureQueue) Register(job queueContracts.Job)                     {}
func (q *captureQueue) StartWorker(queue ...string) error                   { return nil }
func (q *captureQueue) RunWorker(queue ...string) error                     { return nil }
func (q *captureQueue) StopWorker() error                                   { return nil }
func (q *captureQueue) InspectQueues() ([]queueContracts.QueueStatus, error) {
	return nil, nil
}

func (q *captureQueue) ListFailed(page, pageSize int) ([]queueContracts.FailedJob, error) {
	return nil, nil
}
func (q *captureQueue) RetryFailed(id string) error  { return nil }
func (q *captureQueue) DeleteFailed(id string) error { return nil }
func (q *captureQueue) Flush(queue string) error     { return nil }
func (q *captureQueue) HealthCheck() error           { return nil }
func (q *captureQueue) Close() error                 { return nil }
func (q *captureQueue) SetConcurrency(num int)       {}
func (q *captureQueue) GetConcurrency() int          { return 1 }

// TestProduct 模拟实现 HasMedia 接口的模型实体
type TestProduct struct {
	ID   int64
	Name string
}

func (p *TestProduct) GetModelType() string {
	return "products"
}

func (p *TestProduct) GetModelID() string {
	return fmt.Sprintf("%d", p.ID)
}

func (p *TestProduct) RegisterMediaCollections(reg *medialibrary.CollectionRegistry) {
	// 注册 gallery 集合并绑定限制与 conversions
	reg.Add("gallery").
		AcceptMimes("image/png", "image/jpeg").
		MaxSize(500*1024).                                      // 500KB
		RegisterConversion("thumb", 80, 80, true).              // 80x80 居中裁剪
		RegisterConversion("thumb_webp", 50, 50, false, "webp") // 50x50 自适应转为 webp

	// 注册 avatar 集合为 SingleFile 覆盖模式
	reg.Add("avatar").
		SingleFileMode().
		AcceptMimes("image/png")
}

func TestMediaLibrary_IntegrationWorkflow(t *testing.T) {
	ctx := context.Background()

	// 1. 初始化测试配置与底层 DB / Storage 服务
	cfg := testkit.NewSQLiteConfig(t)
	dbManager, conn, err := database.RegisterDatabase(cfg)
	require.NoError(t, err)
	defer func() { _ = dbManager.CloseAll() }()

	bunDB, err := conn.BunDB()
	require.NoError(t, err)

	storageManager, err := storage.Register(cfg)
	require.NoError(t, err)
	defer func() { _ = storageManager.Close() }()

	// 2. 动态创建 media 表
	_, err = bunDB.NewCreateTable().Model((*medialibrary.Media)(nil)).Exec(ctx)
	require.NoError(t, err)

	// 3. 实例化 medialibrary 业务服务
	mediaService := medialibrary.NewService(bunDB, storageManager, "local")
	product := &TestProduct{ID: 99, Name: "Awesome Laptop"}

	// 4. 构建测试用真实 PNG 图像载荷数据
	mockImg := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for x := 0; x < 200; x++ {
		for y := 0; y < 200; y++ {
			mockImg.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var imgBuf bytes.Buffer
	err = png.Encode(&imgBuf, mockImg)
	require.NoError(t, err)
	imgData := imgBuf.Bytes()

	// 5. 测试 FromBytes 流式链式上传到 gallery
	customProperties := map[string]any{"alt": "Laptop Front View"}
	media, err := mediaService.AddMedia(ctx, product).
		FromBytes(imgData, "laptop_front.png").
		WithCustomProperties(customProperties).
		ToMediaCollection("gallery")

	require.NoError(t, err)
	assert.NotEmpty(t, media.UUID)
	assert.Equal(t, "products", media.ModelType)
	assert.Equal(t, "99", media.ModelID)
	assert.Equal(t, "gallery", media.CollectionName)
	assert.Equal(t, "laptop_front", media.Name)
	assert.Equal(t, "laptop_front.png", media.FileName)
	assert.Equal(t, "image/png", media.MimeType)
	assert.Equal(t, int64(len(imgData)), media.Size)
	assert.Equal(t, medialibrary.DerivedMediaStatusCompleted, media.ConversionStatus)
	assert.Equal(t, "Laptop Front View", media.CustomProperties["alt"])

	// 6. 验证物理文件在磁盘上的落库情况
	disk := storageManager.Disk("local")
	originalPath := fmt.Sprintf("media/%s/laptop_front.png", media.UUID)
	originalExists, err := disk.Exists(originalPath)
	require.NoError(t, err)
	assert.True(t, originalExists)

	// 验证图片 conversions 转换子图的写入
	assert.Contains(t, media.Manipulations, "thumb")
	thumbMap := media.Manipulations["thumb"].(map[string]any)
	thumbPath := thumbMap["path"].(string)
	assert.Equal(t, medialibrary.DerivedMediaStatusCompleted, thumbMap["status"])
	assert.Equal(t, fmt.Sprintf("media/%s/conversions/laptop_front-thumb.png", media.UUID), thumbPath)
	thumbExists, err := disk.Exists(thumbPath)
	require.NoError(t, err)
	assert.True(t, thumbExists)

	// 验证 WebP 图片格式转换子图的写入与格式后缀
	assert.Contains(t, media.Manipulations, "thumb_webp")
	webpMap := media.Manipulations["thumb_webp"].(map[string]any)
	webpPath := webpMap["path"].(string)
	assert.Equal(t, fmt.Sprintf("media/%s/conversions/laptop_front-thumb_webp.webp", media.UUID), webpPath)
	webpExists, err := disk.Exists(webpPath)
	require.NoError(t, err)
	assert.True(t, webpExists)

	// 7. 验证查询服务 GetMedia 多态调回
	list, err := mediaService.GetMedia(ctx, product, "gallery")
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, media.UUID, list[0].UUID)

	// 8. 验证 URL 获取
	url := mediaService.GetUrl(media)
	assert.Contains(t, url, originalPath)
	thumbUrl := mediaService.GetUrl(media, "thumb")
	assert.Contains(t, thumbUrl, thumbPath)

	// 9. 验证 SingleFile 覆盖模式
	media2, err := mediaService.AddMedia(ctx, product).
		FromBytes(imgData, "laptop_avatar2.png").
		ToMediaCollection("avatar")
	require.NoError(t, err)

	// 再次上传相同 avatar 集合，原有的 media2 应该被自动物理及数据库清理
	media3, err := mediaService.AddMedia(ctx, product).
		FromBytes(imgData, "laptop_avatar3.png").
		ToMediaCollection("avatar")
	require.NoError(t, err)

	// 获取 avatar 集合列表，应该只剩下最新上传的 media3
	avatars, err := mediaService.GetMedia(ctx, product, "avatar")
	require.NoError(t, err)
	assert.Len(t, avatars, 1)
	assert.Equal(t, media3.UUID, avatars[0].UUID)

	// 物理检查 media2 应该已不复存在
	m2Path := fmt.Sprintf("media/%s/laptop_avatar2.png", media2.UUID)
	m2Exists, err := disk.Exists(m2Path)
	require.NoError(t, err)
	assert.False(t, m2Exists)

	// 10. 验证 DeleteMedia 级联清除
	err = mediaService.DeleteMedia(ctx, media)
	require.NoError(t, err)

	// 物理检查原图与缩略图都应该被彻底擦除
	originalExists, err = disk.Exists(originalPath)
	require.NoError(t, err)
	assert.False(t, originalExists)
	thumbExists, err = disk.Exists(thumbPath)
	require.NoError(t, err)
	assert.False(t, thumbExists)

	// 数据库检查记录也已被清除
	listAfterDelete, err := mediaService.GetMedia(ctx, product, "gallery")
	require.NoError(t, err)
	assert.Empty(t, listAfterDelete)
}

func TestMediaLibrary_ValidationConstraints(t *testing.T) {
	ctx := context.Background()
	cfg := testkit.NewSQLiteConfig(t)
	dbManager, conn, err := database.RegisterDatabase(cfg)
	require.NoError(t, err)
	defer func() { _ = dbManager.CloseAll() }()

	bunDB, err := conn.BunDB()
	require.NoError(t, err)

	storageManager, err := storage.Register(cfg)
	require.NoError(t, err)
	defer func() { _ = storageManager.Close() }()

	_, err = bunDB.NewCreateTable().Model((*medialibrary.Media)(nil)).Exec(ctx)
	require.NoError(t, err)

	mediaService := medialibrary.NewService(bunDB, storageManager, "local")
	product := &TestProduct{ID: 100, Name: "Restricted Product"}

	// 1. 制造大文件以测试最大大小限制
	largeData := make([]byte, 600*1024) // 600KB，超过 500KB 限制
	_, err = mediaService.AddMedia(ctx, product).
		FromBytes(largeData, "heavy.png").
		ToMediaCollection("gallery")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds collection max size limit")

	// 2. 传入不支持的 MIME 类型测试限制
	txtData := []byte("This is plain text content")
	_, err = mediaService.AddMedia(ctx, product).
		FromBytes(txtData, "note.txt").
		ToMediaCollection("gallery")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed in collection")
}

func TestMediaLibrary_QueueModeDefersConversions(t *testing.T) {
	ctx := context.Background()
	cfg := testkit.NewSQLiteConfig(t)
	dbManager, conn, err := database.RegisterDatabase(cfg)
	require.NoError(t, err)
	defer func() { _ = dbManager.CloseAll() }()

	bunDB, err := conn.BunDB()
	require.NoError(t, err)
	storageManager, err := storage.Register(cfg)
	require.NoError(t, err)
	defer func() { _ = storageManager.Close() }()
	_, err = bunDB.NewCreateTable().Model((*medialibrary.Media)(nil)).Exec(ctx)
	require.NoError(t, err)

	queue := &captureQueue{}
	mediaService := medialibrary.NewService(
		bunDB,
		storageManager,
		"local",
		medialibrary.WithConversionMode(medialibrary.ConversionModeQueue),
		medialibrary.WithQueue(queue, "media"),
	)
	product := &TestProduct{ID: 101, Name: "Queued Product"}

	mockImg := image.NewRGBA(image.Rect(0, 0, 120, 120))
	var imgBuf bytes.Buffer
	require.NoError(t, png.Encode(&imgBuf, mockImg))

	media, err := mediaService.AddMedia(ctx, product).
		FromBytes(imgBuf.Bytes(), "queued.png").
		ToMediaCollection("gallery")
	require.NoError(t, err)
	assert.Equal(t, medialibrary.DerivedMediaStatusPending, media.ConversionStatus)
	require.Len(t, queue.jobs, 1)

	err = queue.jobs[0].Handle(ctx)
	require.NoError(t, err)

	var stored medialibrary.Media
	err = bunDB.NewSelect().Model(&stored).Where("id = ?", media.ID).Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, medialibrary.DerivedMediaStatusCompleted, stored.ConversionStatus)
	assert.Contains(t, mediaService.GetUrl(&stored, "thumb"), "queued-thumb.png")
}

func TestMediaLibrary_RejectsDangerousFileName(t *testing.T) {
	ctx := context.Background()
	cfg := testkit.NewSQLiteConfig(t)
	dbManager, conn, err := database.RegisterDatabase(cfg)
	require.NoError(t, err)
	defer func() { _ = dbManager.CloseAll() }()

	bunDB, err := conn.BunDB()
	require.NoError(t, err)
	storageManager, err := storage.Register(cfg)
	require.NoError(t, err)
	defer func() { _ = storageManager.Close() }()
	_, err = bunDB.NewCreateTable().Model((*medialibrary.Media)(nil)).Exec(ctx)
	require.NoError(t, err)

	mediaService := medialibrary.NewService(bunDB, storageManager, "local")
	product := &TestProduct{ID: 102, Name: "Unsafe Product"}
	mockImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var imgBuf bytes.Buffer
	require.NoError(t, png.Encode(&imgBuf, mockImg))

	_, err = mediaService.AddMedia(ctx, product).
		FromBytes(imgBuf.Bytes(), "shell.php.jpg").
		ToMediaCollection("gallery")
	require.ErrorIs(t, err, medialibrary.ErrFileNameNotAllowed)
}

func TestMediaLibrary_RegenerateOnlyMissing(t *testing.T) {
	ctx := context.Background()
	cfg := testkit.NewSQLiteConfig(t)
	dbManager, conn, err := database.RegisterDatabase(cfg)
	require.NoError(t, err)
	defer func() { _ = dbManager.CloseAll() }()

	bunDB, err := conn.BunDB()
	require.NoError(t, err)
	storageManager, err := storage.Register(cfg)
	require.NoError(t, err)
	defer func() { _ = storageManager.Close() }()
	_, err = bunDB.NewCreateTable().Model((*medialibrary.Media)(nil)).Exec(ctx)
	require.NoError(t, err)

	mediaService := medialibrary.NewService(bunDB, storageManager, "local")
	product := &TestProduct{ID: 103, Name: "Regenerate Product"}
	mockImg := image.NewRGBA(image.Rect(0, 0, 120, 120))
	var imgBuf bytes.Buffer
	require.NoError(t, png.Encode(&imgBuf, mockImg))

	media, err := mediaService.AddMedia(ctx, product).
		FromBytes(imgBuf.Bytes(), "regen.png").
		ToMediaCollection("gallery")
	require.NoError(t, err)

	media.Manipulations["thumb"] = map[string]any{
		"width":  80,
		"height": 80,
		"crop":   true,
		"status": medialibrary.DerivedMediaStatusPending,
	}
	media.ConversionStatus = medialibrary.DerivedMediaStatusFailed
	_, err = bunDB.NewUpdate().
		Model(media).
		Column("manipulations", "conversion_status").
		Where("id = ?", media.ID).
		Exec(ctx)
	require.NoError(t, err)

	count, err := mediaService.Regenerate(ctx, medialibrary.RegenerateOptions{OnlyMissing: true, FailedOnly: true})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
