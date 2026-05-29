package make

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"lfiber/internal/console/ui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

//go:embed stubs/*.tmpl
var stubsFS embed.FS

// MakeExportCommand 创建一个生成 Export 结构体的命令
func MakeExportCommand() *cobra.Command {
	var featureFlag string
	var concernsFlag []string

	cmd := &cobra.Command{
		Use:     "make:export [name]",
		Short:   "Create a new export class skeleton",
		GroupID: "app",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			noInteraction, _ := cmd.Flags().GetBool("no-interaction")

			// 检查名称格式，并转为驼峰
			structName := toCamelCase(name)
			if !strings.HasSuffix(structName, "Export") {
				structName += "Export"
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
						Title("Select the associated feature:").
						Options(options...).
						Value(&selectedFeature)
					if err = prompt.Run(); err != nil {
						return err
					}
				}

				// 2. 选择 Concerns
				if len(selectedConcerns) == 0 {
					prompt := huh.NewMultiSelect[string]().
						Title("Select Excel concerns to implement:").
						Options(
							huh.NewOption("FromSlice (Export from slice)", "FromSlice"),
							huh.NewOption("FromQuery (Export from query)", "FromQuery"),
							huh.NewOption("WithHeadings (Custom headings)", "WithHeadings"),
							huh.NewOption("WithMapping (Custom mapping)", "WithMapping"),
							huh.NewOption("ShouldAutoSize (Auto fit columns)", "ShouldAutoSize"),
							huh.NewOption("WithColumnWidths (Custom column widths)", "WithColumnWidths"),
							huh.NewOption("WithQueueNotification (Queue and notify)", "WithQueueNotification"),
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
				selectedConcerns = []string{"FromSlice"}
			}

			// 生成文件
			fileName := toSnakeCase(strings.TrimSuffix(structName, "Export")) + "_export.go"
			dirPath := filepath.Join("internal", "features", selectedFeature, "exports")
			filePath := filepath.Join(dirPath, fileName)

			if err = os.MkdirAll(dirPath, 0o750); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
			}

			// 检查是否已存在
			if _, err = os.Stat(filePath); err == nil {
				return fmt.Errorf("file %s already exists", filePath)
			}

			code, err := generateExportCode(structName, selectedConcerns)
			if err != nil {
				return fmt.Errorf("failed to generate export code: %w", err)
			}

			if err = os.WriteFile(filePath, []byte(code), 0o600); err != nil {
				return fmt.Errorf("failed to write export file: %w", err)
			}

			ui.Success(cmd.OutOrStdout(), "Export skeleton successfully created: %s", filePath)
			return nil
		},
	}

	cmd.Flags().StringVar(&featureFlag, "feature", "", "The feature module where the export will belong")
	cmd.Flags().StringSliceVar(&concernsFlag, "concerns", nil, "The concern interfaces to implement (e.g. FromSlice,WithHeadings)")

	return cmd
}

func getFeaturesList() ([]string, error) {
	entries, err := os.ReadDir(filepath.Join("internal", "features"))
	if err != nil {
		return nil, err
	}
	var features []string
	for _, entry := range entries {
		if entry.IsDir() {
			features = append(features, entry.Name())
		}
	}
	return features, nil
}

type exportTemplateData struct {
	StructName           string
	StructSnakeName      string
	HasSlice             bool
	HasQuery             bool
	HasHeadings          bool
	HasMapping           bool
	HasAutoSize          bool
	HasColumnWidths      bool
	HasQueueNotification bool
}

func generateExportCode(structName string, concerns []string) (string, error) {
	// 读取嵌入的 stub
	tmplContent, err := stubsFS.ReadFile("stubs/export.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("export").Parse(string(tmplContent))
	if err != nil {
		return "", err
	}

	data := exportTemplateData{
		StructName:           structName,
		StructSnakeName:      toSnakeCase(strings.TrimSuffix(structName, "Export")),
		HasSlice:             containsString(concerns, "FromSlice"),
		HasQuery:             containsString(concerns, "FromQuery"),
		HasHeadings:          containsString(concerns, "WithHeadings"),
		HasMapping:           containsString(concerns, "WithMapping"),
		HasAutoSize:          containsString(concerns, "ShouldAutoSize"),
		HasColumnWidths:      containsString(concerns, "WithColumnWidths"),
		HasQueueNotification: containsString(concerns, "WithQueueNotification"),
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func toCamelCase(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	words := strings.Fields(s)
	for i := 0; i < len(words); i++ {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(words[i][:1]) + words[i][1:]
		}
	}
	return strings.Join(words, "")
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
