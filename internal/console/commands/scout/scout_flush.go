package scout

import (
	"fmt"
	"strings"

	"lfiber/internal/console/commands/commandutil"
	"lfiber/internal/console/ui"
	"lfiber/pkg/search"

	"github.com/spf13/cobra"
)

// ScoutFlushCommand 清空指定的 Searchable 模型的检索索引数据
func ScoutFlushCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scout:flush [model]",
		Short:   "Flush all records of the given model from the search index",
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

			// 检查模型是否注册
			model, _, ok := search.Get(modelName)
			if !ok {
				return fmt.Errorf("model %s is not registered in search registry (registered: %s)", modelName, strings.Join(search.Names(), ", "))
			}

			ui.Info(cmd.OutOrStdout(), "Flushing all documents from search index %s...", model.SearchableIndex())

			_, err = search.DeleteAllDocuments(model.SearchableIndex())
			if err != nil {
				return fmt.Errorf("failed to flush index: %w", err)
			}

			ui.Success(cmd.OutOrStdout(), "Successfully flushed all documents of model %s", modelName)
			return nil
		},
	}

	return cmd
}
