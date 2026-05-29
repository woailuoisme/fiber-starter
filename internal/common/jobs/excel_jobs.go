package jobs

import (
	"bytes"
	"context"
	"fmt"

	"lfiber/internal/providers/storage"
	"lfiber/pkg/excel"
)

// HandleQueueExport 在后台任务内部统一执行：导出数据 -> 写入持久存储 -> 触发回调
// 为什么这样做：统一封装通用异步导出处理管道，避免各业务 Job 编写重复的落盘与通知调度逻辑。
func HandleQueueExport(ctx context.Context, export excel.ExportConcern, diskName, path string) error {
	var buf bytes.Buffer
	if err := excel.Export(ctx, export, &buf); err != nil {
		if notify, ok := export.(excel.WithQueueNotification); ok {
			_ = notify.OnQueueFailure(ctx, err)
		}
		return fmt.Errorf("failed to export excel data: %w", err)
	}

	disk := storage.Drive(diskName)
	if err := disk.Put(path, buf.Bytes()); err != nil {
		if notify, ok := export.(excel.WithQueueNotification); ok {
			_ = notify.OnQueueFailure(ctx, err)
		}
		return fmt.Errorf("failed to save excel to storage %s: %w", path, err)
	}

	if notify, ok := export.(excel.WithQueueNotification); ok {
		url := disk.Url(path)
		if err := notify.OnQueueSuccess(ctx, url); err != nil {
			return fmt.Errorf("failed to trigger success notification: %w", err)
		}
	}

	return nil
}

// HandleQueueImport 在后台任务内部统一执行：读取持久存储 -> 执行流式导入 -> 触发回调
// 为什么这样做：统一封装通用异步导入处理管道，保证导入过程使用流式低内存读取。
func HandleQueueImport(ctx context.Context, importObj excel.ImportConcern, diskName, path string) error {
	disk := storage.Drive(diskName)
	content, err := disk.Get(path)
	if err != nil {
		if notify, ok := importObj.(excel.WithQueueNotification); ok {
			_ = notify.OnQueueFailure(ctx, err)
		}
		return fmt.Errorf("failed to get import file from storage %s: %w", path, err)
	}

	reader := bytes.NewReader(content)
	if err := excel.Import(ctx, importObj, reader); err != nil {
		if notify, ok := importObj.(excel.WithQueueNotification); ok {
			_ = notify.OnQueueFailure(ctx, err)
		}
		return fmt.Errorf("failed to import excel data: %w", err)
	}

	if notify, ok := importObj.(excel.WithQueueNotification); ok {
		if err := notify.OnQueueSuccess(ctx, ""); err != nil {
			return fmt.Errorf("failed to trigger success notification: %w", err)
		}
	}

	return nil
}
