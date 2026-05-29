package scout

import (
	"fmt"
	"reflect"
	"strings"

	"lfiber/internal/console/commands/commandutil"
	"lfiber/internal/console/ui"
	"lfiber/pkg/search"

	"github.com/spf13/cobra"
)

// ScoutImportCommand 导入指定的 Searchable 模型数据到搜索引擎
func ScoutImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scout:import [model]",
		Short:   "Import the given model into the search index",
		GroupID: "system",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelName := args[0]

			// 构建应用 Runtime 以获取连接
			rt, err := commandutil.BuildRuntime()
			if err != nil {
				return err
			}
			defer func() { _ = commandutil.CloseRuntime(rt) }()

			// 检查模型是否已注册
			model, queryFunc, ok := search.Get(modelName)
			if !ok {
				return fmt.Errorf("model %s is not registered in search registry (registered: %s)", modelName, strings.Join(search.Names(), ", "))
			}

			ctx := cmd.Context()
			query, err := queryFunc(ctx)
			if err != nil {
				return fmt.Errorf("failed to build query: %w", err)
			}

			// 查询总行数
			total, err := query.Count(ctx)
			if err != nil {
				return fmt.Errorf("failed to count records: %w", err)
			}

			if total == 0 {
				ui.Warning(cmd.OutOrStdout(), "No records found for model %s", modelName)
				return nil
			}

			ui.Info(cmd.OutOrStdout(), "Importing %d records for model %s...", total, modelName)
			bar := ui.NewProgressBar(cmd.OutOrStdout(), int(total))

			// 反射提取模型类型并分批读取
			modelType := reflect.TypeOf(model)
			if modelType.Kind() == reflect.Pointer {
				modelType = modelType.Elem()
			}

			batchSize := 500
			offset := 0

			for offset < int(total) {
				// 每次构造新的 query 分批拉取
				batchQuery, err := queryFunc(ctx)
				if err != nil {
					return fmt.Errorf("failed to build query: %w", err)
				}

				// 动态构造 []Model 的指针以供 Bun 填充
				sliceType := reflect.SliceOf(modelType)
				slicePtr := reflect.New(sliceType)

				err = batchQuery.Offset(offset).Limit(batchSize).Scan(ctx, slicePtr.Interface())
				if err != nil {
					return fmt.Errorf("failed to scan database batch: %w", err)
				}

				sliceVal := slicePtr.Elem()
				batchLen := sliceVal.Len()
				if batchLen == 0 {
					break
				}

				// 将 Slice 的每个元素转换并投递到 Meilisearch
				docs := make([]any, 0, batchLen)
				for i := 0; i < batchLen; i++ {
					elem := sliceVal.Index(i)

					// 确保断言为 Searchable
					var searchable search.Searchable
					if elem.Kind() == reflect.Struct {
						// 尝试以指针形式断言，或者值类型断言
						if p, ok := elem.Addr().Interface().(search.Searchable); ok {
							searchable = p
						}
					} else if p, ok := elem.Interface().(search.Searchable); ok {
						searchable = p
					}

					if searchable == nil {
						return fmt.Errorf("scanned model at index %d does not implement search.Searchable", offset+i)
					}

					doc := searchable.ToSearchableArray()
					if _, exists := doc["id"]; !exists {
						doc["id"] = searchable.SearchableId()
					}
					docs = append(docs, doc)
				}

				// 批量导入
				_, err = search.AddDocuments(model.SearchableIndex(), docs)
				if err != nil {
					return fmt.Errorf("failed to sync documents to search engine: %w", err)
				}

				bar.Advance(batchLen)
				offset += batchLen
			}

			bar.Finish()
			ui.Success(cmd.OutOrStdout(), "Successfully imported model %s into search index", modelName)
			return nil
		},
	}

	return cmd
}
