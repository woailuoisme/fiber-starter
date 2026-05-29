package make

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"lfiber/internal/console/ui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// MakeImportCommand 创建一个生成 Import 结构体的命令
func MakeImportCommand() *cobra.Command {
	var featureFlag string
	var concernsFlag []string

	cmd := &cobra.Command{
		Use:     "make:import [name]",
		Short:   "Create a new import class skeleton",
		GroupID: "app",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			noInteraction, _ := cmd.Flags().GetBool("no-interaction")

			// 检查名称格式，并转为驼峰
			structName := toCamelCase(name)
			if !strings.HasSuffix(structName, "Import") {
				structName += "Import"
			}

			// 获取 features 列表
			features, err := getFeaturesList()
			if err != nil {
				return fmt.Errorf("failed to read features directory: %w", err)
			}
			if len(features) == 0 {
				return fmt.Errorf("no feature modules found in internal/features/")
			}

			selectedFeature := featureFlag
			selectedConcerns := concernsFlag

			// 开启交互模式
			if !noInteraction && (selectedFeature == "" || len(selectedConcerns) == 0) {
				// 1. 选择 Feature
				if selectedFeature == "" {
					var options []huh.Option[string]
					for _, f := range features {
						options = append(options, huh.NewOption(f, f))
					}
					prompt := huh.NewSelect[string]().
						Title("选择关联的 Feature 模块:").
						Options(options...).
						Value(&selectedFeature)
					if err = prompt.Run(); err != nil {
						return err
					}
				}

				// 2. 选择 Concerns
				if len(selectedConcerns) == 0 {
					prompt := huh.NewMultiSelect[string]().
						Title("选择需要实现的 Concern 接口:").
						Options(
							huh.NewOption("ToSlice (读取至切片中)", "ToSlice"),
							huh.NewOption("ToModel (逐行映射为数据库 Model)", "ToModel"),
							huh.NewOption("OnRow (通用的逐行读取回调)", "OnRow"),
							huh.NewOption("WithHeadingRow (定义标题行)", "WithHeadingRow"),
							huh.NewOption("WithValidation (添加数据校验)", "WithValidation"),
							huh.NewOption("WithBatchInserts (批量入库)", "WithBatchInserts"),
							huh.NewOption("WithQueueNotification (启用异步队列处理通知)", "WithQueueNotification"),
						).
						Value(&selectedConcerns)
					if err = prompt.Run(); err != nil {
						return err
					}
				}
			}

			// 回退默认值
			if selectedFeature == "" {
				selectedFeature = features[0]
			}
			if len(selectedConcerns) == 0 {
				selectedConcerns = []string{"ToSlice"}
			}

			// 生成文件
			fileName := toSnakeCase(strings.TrimSuffix(structName, "Import")) + "_import.go"
			dirPath := filepath.Join("internal", "features", selectedFeature, "imports")
			filePath := filepath.Join(dirPath, fileName)

			if err = os.MkdirAll(dirPath, 0o750); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
			}

			// 检查是否已存在
			if _, err = os.Stat(filePath); err == nil {
				return fmt.Errorf("file %s already exists", filePath)
			}

			code, err := generateImportCode(structName, selectedConcerns)
			if err != nil {
				return fmt.Errorf("failed to generate import code: %w", err)
			}

			if err = os.WriteFile(filePath, []byte(code), 0o600); err != nil {
				return fmt.Errorf("failed to write import file: %w", err)
			}

			ui.Success(cmd.OutOrStdout(), "Import skeleton successfully created: %s", filePath)
			return nil
		},
	}

	cmd.Flags().StringVar(&featureFlag, "feature", "", "The feature module where the import will belong")
	cmd.Flags().StringSliceVar(&concernsFlag, "concerns", nil, "The concern interfaces to implement (e.g. ToSlice,WithHeadingRow)")

	return cmd
}

type importTemplateData struct {
	StructName           string
	StructSnakeName      string
	HasSlice             bool
	HasModel             bool
	HasOnRow             bool
	HasHeadingRow        bool
	HasValidation        bool
	HasBatchInserts      bool
	HasQueueNotification bool
}

func generateImportCode(structName string, concerns []string) (string, error) {
	// 读取嵌入 of stub
	tmplContent, err := stubsFS.ReadFile("stubs/import.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("import").Parse(string(tmplContent))
	if err != nil {
		return "", err
	}

	data := importTemplateData{
		StructName:           structName,
		StructSnakeName:      toSnakeCase(strings.TrimSuffix(structName, "Import")),
		HasSlice:             containsString(concerns, "ToSlice"),
		HasModel:             containsString(concerns, "ToModel"),
		HasOnRow:             containsString(concerns, "OnRow"),
		HasHeadingRow:        containsString(concerns, "WithHeadingRow"),
		HasValidation:        containsString(concerns, "WithValidation"),
		HasBatchInserts:      containsString(concerns, "WithBatchInserts"),
		HasQueueNotification: containsString(concerns, "WithQueueNotification"),
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
