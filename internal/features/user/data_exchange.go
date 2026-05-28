package user

import (
	"bytes"
	"context"
	"fmt"
	"io"

	database "lfiber/internal/providers/database/contracts"
	queue "lfiber/internal/providers/queue"
	helpers "lfiber/internal/support"
	"lfiber/internal/support/appctx"

	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

// UserDataExchange handles importing and exporting user data synchronously and asynchronously.
type UserDataExchange interface {
	// Import parses Excel user data from reader and saves to DB.
	Import(ctx context.Context, reader io.Reader) (int, error)
	// Export loads users from DB and writes Excel data to writer.
	Export(ctx context.Context, writer io.Writer) error

	// QueueImport schedules a background job to import users from a stored file.
	QueueImport(ctx context.Context, storagePath string) error
	// QueueExport schedules a background job to export users to a storage path.
	QueueExport(ctx context.Context, storagePath string) error
}

type userDataExchange struct {
	db database.Connection
}

// NewUserDataExchange creates a new UserDataExchange service instance.
func NewUserDataExchange(db database.Connection) UserDataExchange {
	return &userDataExchange{db: db}
}

// Import parses Excel user data and saves to DB within a single transaction.
func (s *userDataExchange) Import(ctx context.Context, reader io.Reader) (int, error) {
	excel := &helpers.Excel{}
	importer := &UserImport{}

	modelsAny, err := excel.Import(reader, importer)
	if err != nil {
		return 0, fmt.Errorf("excel import failed: %w", err)
	}

	if len(modelsAny) == 0 {
		return 0, nil
	}

	users := make([]User, 0, len(modelsAny))
	for _, m := range modelsAny {
		if u, ok := m.(*User); ok && u != nil {
			users = append(users, *u)
		}
	}

	if len(users) == 0 {
		return 0, nil
	}

	bunDB, err := s.db.BunDB()
	if err != nil {
		return 0, err
	}

	err = bunDB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		chunkSize := 1000
		for i := 0; i < len(users); i += chunkSize {
			end := i + chunkSize
			if end > len(users) {
				end = len(users)
			}
			chunk := users[i:end]

			_, err := tx.NewInsert().Model(&chunk).Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to bulk insert users: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return len(users), nil
}

// Export loads users from DB and writes Excel data to writer.
func (s *userDataExchange) Export(ctx context.Context, writer io.Writer) error {
	bunDB, err := s.db.BunDB()
	if err != nil {
		return err
	}

	var users []User
	err = bunDB.NewSelect().Model(&users).Order("created_at DESC").Limit(1000).Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	export := &UserExport{
		Users: users,
	}

	excel := &helpers.Excel{}
	return excel.WriteTo(writer, export)
}

// QueueImport dispatches a background job to import users.
func (s *userDataExchange) QueueImport(ctx context.Context, storagePath string) error {
	job := NewUserImportJob(storagePath)
	return queue.Push(job)
}

// QueueExport dispatches a background job to export users.
func (s *userDataExchange) QueueExport(ctx context.Context, storagePath string) error {
	job := NewUserExportJob(storagePath)
	return queue.Push(job)
}

// ==========================================
// User Data Exchange Queue Jobs
// ==========================================

// UserImportJob handles asynchronous importing of users.
type UserImportJob struct {
	StoragePath string `json:"storage_path"`
}

func NewUserImportJob(storagePath string) *UserImportJob {
	return &UserImportJob{StoragePath: storagePath}
}

func (j *UserImportJob) TaskName() string {
	return "user_import"
}

func (j *UserImportJob) QueueName() string {
	return "default"
}

func (j *UserImportJob) Handle(ctx context.Context) error {
	helpers.Info("Processing UserImportJob", zap.String("path", j.StoragePath))

	data, err := helpers.Get(j.StoragePath)
	if err != nil {
		helpers.Error("Failed to read import file from storage", zap.Error(err), zap.String("path", j.StoragePath))
		return fmt.Errorf("read import file: %w", err)
	}

	defer func() {
		if err := helpers.Delete(j.StoragePath); err != nil {
			helpers.Warn("Failed to delete temp import file", zap.Error(err), zap.String("path", j.StoragePath))
		} else {
			helpers.Info("Successfully cleaned up temp import file", zap.String("path", j.StoragePath))
		}
	}()

	app := appctx.App()
	if app == nil {
		return fmt.Errorf("app context not initialized")
	}

	exchange := NewUserDataExchange(app.ConnectionValue())
	reader := bytes.NewReader(data)
	count, err := exchange.Import(ctx, reader)
	if err != nil {
		helpers.Error("UserImportJob failed", zap.Error(err), zap.String("path", j.StoragePath))
		return err
	}

	helpers.Info("UserImportJob completed successfully", zap.Int("imported_count", count))
	return nil
}

// UserExportJob handles asynchronous exporting of users.
type UserExportJob struct {
	StoragePath string `json:"storage_path"`
}

func NewUserExportJob(storagePath string) *UserExportJob {
	return &UserExportJob{StoragePath: storagePath}
}

func (j *UserExportJob) TaskName() string {
	return "user_export"
}

func (j *UserExportJob) QueueName() string {
	return "default"
}

func (j *UserExportJob) Handle(ctx context.Context) error {
	helpers.Info("Processing UserExportJob", zap.String("path", j.StoragePath))

	app := appctx.App()
	if app == nil {
		return fmt.Errorf("app context not initialized")
	}

	exchange := NewUserDataExchange(app.ConnectionValue())

	var buf bytes.Buffer
	if err := exchange.Export(ctx, &buf); err != nil {
		helpers.Error("UserExportJob failed during export", zap.Error(err), zap.String("path", j.StoragePath))
		return err
	}

	if err := helpers.Put(j.StoragePath, buf.Bytes()); err != nil {
		helpers.Error("Failed to save exported file to storage", zap.Error(err), zap.String("path", j.StoragePath))
		return fmt.Errorf("save export file: %w", err)
	}

	helpers.Info("UserExportJob completed successfully", zap.String("path", j.StoragePath))
	return nil
}

// UserImport 用户导入类
type UserImport struct{}

// Model 将行数据转换为模型
func (i *UserImport) Model(row []string) any {
	if len(row) < 3 {
		return nil
	}

	// 假设 Excel 列顺序为：姓名, 邮箱, 电话
	// 注意：这里跳过了 ID，因为通常导入是新增或通过邮箱匹配
	user := &User{
		Name:  row[0],
		Email: row[1],
	}

	if len(row) > 2 && row[2] != "" {
		phone := row[2]
		user.Phone = &phone
	}

	// 默认状态
	user.Status = UserStatusActive

	return user
}

// UserExport 用户导出类
type UserExport struct {
	Users []User
}

// Collection 获取导出数据
func (e *UserExport) Collection() any {
	return e.Users
}

// Headings 设置表头
func (e *UserExport) Headings() []string {
	return []string{
		"ID",
		"Name",
		"Email",
		"Phone",
		"Status",
		"Created At",
	}
}

// Map 映射每行数据
func (e *UserExport) Map(item any) []any {
	user := item.(User)

	phone := ""
	if user.Phone != nil {
		phone = *user.Phone
	}

	return []any{
		user.ID,
		user.Name,
		user.Email,
		phone,
		user.Status,
		user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
