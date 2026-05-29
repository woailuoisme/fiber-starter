package tests

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"lfiber/internal/common/jobs"
	"lfiber/internal/providers"
	queueContracts "lfiber/internal/providers/queue/contracts"
	"lfiber/pkg/excel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

// ----------------------------------------------------
// 1. Slice 导入导出测试定义
// ----------------------------------------------------

type mockSliceRow struct {
	ID        int       `excel:"ID"`
	Name      string    `excel:"Username"`
	Age       int       `excel:"Age"`
	Active    bool      `excel:"Active"`
	CreatedAt time.Time `excel:"CreatedAt"`
}

type testSliceExport struct {
	rows []mockSliceRow
}

func (e *testSliceExport) FromSlice() interface{} {
	return e.rows
}

func (e *testSliceExport) Headings() []string {
	return []string{"ID", "Username", "Age", "Active", "CreatedAt"}
}

type testSliceImport struct {
	dest []mockSliceRow
}

func (i *testSliceImport) ToSlice() interface{} {
	return &i.dest
}

func (i *testSliceImport) HeadingRow() int {
	return 1
}

// ----------------------------------------------------
// 2. Query/Model 数据库导入导出测试定义
// ----------------------------------------------------

type TestExcelUser struct {
	bun.BaseModel `bun:"table:test_excel_users,alias:u"`

	ID    int    `bun:"id,pk,autoincrement" excel:"ID"`
	Name  string `bun:"name,notnull" excel:"Username"`
	Email string `bun:"email" excel:"Email"`
}

type testQueryExport struct{}

func (e *testQueryExport) FromQuery(ctx context.Context) (*bun.SelectQuery, error) {
	connection := providers.App().Connection
	db, err := connection.BunDB()
	if err != nil {
		return nil, err
	}
	return db.NewSelect().Model((*TestExcelUser)(nil)).Order("id ASC"), nil
}

func (e *testQueryExport) Headings() []string {
	return []string{"ID", "Username", "Email"}
}

type testModelImport struct{}

func (i *testModelImport) ToModel(row []string) (interface{}, error) {
	if len(row) < 3 {
		return nil, nil
	}
	id, _ := strconv.Atoi(row[0])
	return &TestExcelUser{
		ID:    id,
		Name:  row[1],
		Email: row[2],
	}, nil
}

func (i *testModelImport) HeadingRow() int {
	return 1
}

func (i *testModelImport) BatchSize() int {
	return 2 // 使用较小 Batch 验证多次冲刷
}

// ----------------------------------------------------
// 测试主入口
// ----------------------------------------------------

func TestExcel_MemorySliceExportAndImport(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	// 1. 初始化待导出数据
	exportObj := &testSliceExport{
		rows: []mockSliceRow{
			{ID: 1, Name: "ZhangSan", Age: 18, Active: true, CreatedAt: now},
			{ID: 2, Name: "LiSi", Age: 22, Active: false, CreatedAt: now.Add(time.Hour)},
		},
	}

	buf := new(bytes.Buffer)
	err := excel.Export(context.Background(), exportObj, buf)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.Bytes())

	// 2. 测试导入读取
	importObj := &testSliceImport{}
	err = excel.Import(context.Background(), importObj, buf)
	require.NoError(t, err)

	require.Len(t, importObj.dest, 2)
	assert.Equal(t, 1, importObj.dest[0].ID)
	assert.Equal(t, "ZhangSan", importObj.dest[0].Name)
	assert.Equal(t, 18, importObj.dest[0].Age)
	assert.True(t, importObj.dest[0].Active)
	assert.True(t, importObj.dest[0].CreatedAt.Equal(now))

	assert.Equal(t, 2, importObj.dest[1].ID)
	assert.Equal(t, "LiSi", importObj.dest[1].Name)
	assert.Equal(t, 22, importObj.dest[1].Age)
	assert.False(t, importObj.dest[1].Active)
	assert.True(t, importObj.dest[1].CreatedAt.Equal(now.Add(time.Hour)))
}

func TestExcel_DBQueryStreamExportAndModelBatchImport(t *testing.T) {
	ctx := context.Background()

	// 1. 使用临时 SQLite 数据库构建 App 容器
	dbPath := filepath.Join(t.TempDir(), "test_excel.sqlite")
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("DB_SQLITE_DATABASE", dbPath)

	rt, err := providers.Build()
	require.NoError(t, err)
	defer func() { _ = rt.Close() }()

	db, err := rt.Connection.BunDB()
	require.NoError(t, err)
	require.NotNil(t, db)

	// 创建测试表
	_, err = db.NewCreateTable().Model((*TestExcelUser)(nil)).Exec(ctx)
	require.NoError(t, err)

	// 插入种子测试数据
	users := []TestExcelUser{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
		{ID: 3, Name: "Charlie", Email: "charlie@example.com"},
		{ID: 4, Name: "David", Email: "david@example.com"},
		{ID: 5, Name: "Eve", Email: "eve@example.com"},
	}
	_, err = db.NewInsert().Model(&users).Exec(ctx)
	require.NoError(t, err)

	// 2. 流式导出为 Excel
	buf := new(bytes.Buffer)
	err = excel.Export(ctx, &testQueryExport{}, buf)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.Bytes())

	// 3. 清空测试表
	_, err = db.NewDelete().Model((*TestExcelUser)(nil)).Where("1=1").Exec(ctx)
	require.NoError(t, err)

	// 4. 将 Excel 里的数据通过 ToModel + WithBatchInserts 流式批量导入
	err = excel.Import(ctx, &testModelImport{}, buf)
	require.NoError(t, err)

	// 5. 验证是否全部写回数据库且正确
	var dbUsers []TestExcelUser
	err = db.NewSelect().Model(&dbUsers).Order("id ASC").Scan(ctx)
	require.NoError(t, err)

	require.Len(t, dbUsers, 5)
	assert.Equal(t, "Alice", dbUsers[0].Name)
	assert.Equal(t, "alice@example.com", dbUsers[0].Email)
	assert.Equal(t, "Eve", dbUsers[4].Name)
	assert.Equal(t, "eve@example.com", dbUsers[4].Email)
}

// ----------------------------------------------------
// 3. Validation 导入测试定义与用例
// ----------------------------------------------------

type testValidationImport struct {
	dest []mockSliceRow
}

func (i *testValidationImport) ToSlice() interface{} {
	return &i.dest
}

func (i *testValidationImport) HeadingRow() int {
	return 1
}

func (i *testValidationImport) Validate(model any) error {
	row, ok := model.(*mockSliceRow)
	if !ok {
		return fmt.Errorf("unexpected model type: %T", model)
	}
	if row.Age < 18 {
		return fmt.Errorf("age must be at least 18, got %d", row.Age)
	}
	return nil
}

func TestExcel_ImportWithValidation(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	// 1. 初始化两行数据：一行 age >= 18（合法），一行 age < 18（非法）
	exportObj := &testSliceExport{
		rows: []mockSliceRow{
			{ID: 1, Name: "ZhangSan", Age: 18, Active: true, CreatedAt: now},
			{ID: 2, Name: "XiaoMing", Age: 15, Active: false, CreatedAt: now},
		},
	}

	buf := new(bytes.Buffer)
	err := excel.Export(context.Background(), exportObj, buf)
	require.NoError(t, err)

	// 2. 测试带校验的导入：应该报错，且提示校验失败
	importObj := &testValidationImport{}
	err = excel.Import(context.Background(), importObj, buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data validation failed")
	assert.Contains(t, err.Error(), "age must be at least 18")
}

// ----------------------------------------------------
// 4. 队列导出导入 Job 测试定义与用例
// ----------------------------------------------------

type mockQueueExport struct {
	queueContracts.JobMeta
	rows        []mockSliceRow
	successChan chan string
	errChan     chan error
}

func (e *mockQueueExport) FromSlice() interface{} {
	return e.rows
}

func (e *mockQueueExport) Headings() []string {
	return []string{"ID", "Username"}
}

func (e *mockQueueExport) OnQueueSuccess(ctx context.Context, fileUrl string) error {
	e.successChan <- fileUrl
	return nil
}

func (e *mockQueueExport) OnQueueFailure(ctx context.Context, err error) error {
	e.errChan <- err
	return nil
}

func (e *mockQueueExport) Handle(ctx context.Context) error {
	return jobs.HandleQueueExport(ctx, e, "local", "test_exports/queue_export.xlsx")
}

type mockQueueImport struct {
	queueContracts.JobMeta
	dest        []mockSliceRow
	successChan chan bool
	errChan     chan error
	filePath    string
}

func (i *mockQueueImport) ToSlice() interface{} {
	return &i.dest
}

func (i *mockQueueImport) HeadingRow() int {
	return 1
}

func (i *mockQueueImport) OnQueueSuccess(ctx context.Context, fileUrl string) error {
	i.successChan <- true
	return nil
}

func (i *mockQueueImport) OnQueueFailure(ctx context.Context, err error) error {
	i.errChan <- err
	return nil
}

func (i *mockQueueImport) Handle(ctx context.Context) error {
	return jobs.HandleQueueImport(ctx, i, "local", i.filePath)
}

func TestExcel_QueueExportAndImportJobs(t *testing.T) {
	ctx := context.Background()

	// 1. 初始化 Runtime 容器以支持 Storage
	dbPath := filepath.Join(t.TempDir(), "test_excel_queue.sqlite")
	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("DB_SQLITE_DATABASE", dbPath)
	t.Setenv("STORAGE_LOCAL_ROOT", t.TempDir())

	rt, err := providers.Build()
	require.NoError(t, err)
	defer func() { _ = rt.Close() }()

	now := time.Now().Truncate(time.Second)
	successChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// 2. 导出 Job 测试
	exportJob := &mockQueueExport{
		JobMeta: queueContracts.NewJobMeta("export:test", "default"),
		rows: []mockSliceRow{
			{ID: 1, Name: "ZhangSan", Age: 18, Active: true, CreatedAt: now},
		},
		successChan: successChan,
		errChan:     errChan,
	}

	err = exportJob.Handle(ctx)
	require.NoError(t, err)

	// 3. 断言回调被成功触发，并拿到了生成的文件 URL
	select {
	case fileUrl := <-successChan:
		assert.Contains(t, fileUrl, "queue_export.xlsx")
	case err := <-errChan:
		t.Fatalf("export job failed: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for export notification")
	}

	// 验证文件真的写入了 local 存储
	disk := rt.Storage.Disk("local")
	exists, err := disk.Exists("test_exports/queue_export.xlsx")
	require.NoError(t, err)
	assert.True(t, exists)

	// 4. 导入 Job 测试
	importSuccessChan := make(chan bool, 1)
	importErrChan := make(chan error, 1)

	importJob := &mockQueueImport{
		JobMeta:     queueContracts.NewJobMeta("import:test", "default"),
		successChan: importSuccessChan,
		errChan:     importErrChan,
		filePath:    "test_exports/queue_export.xlsx",
	}

	err = importJob.Handle(ctx)
	require.NoError(t, err)

	select {
	case success := <-importSuccessChan:
		assert.True(t, success)
	case err := <-importErrChan:
		t.Fatalf("import job failed: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for import notification")
	}

	// 5. 验证导入结果数据是否正确
	require.Len(t, importJob.dest, 1)
	assert.Equal(t, 1, importJob.dest[0].ID)
	assert.Equal(t, "ZhangSan", importJob.dest[0].Name)
}
