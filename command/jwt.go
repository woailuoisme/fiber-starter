package command

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// jwtGenerateCmd represents the jwt:generate command
var jwtGenerateCmd = &cobra.Command{
	Use:   "jwt:generate",
	Short: "生成并替换 JWT 密钥",
	Long: `生成一个新的安全 JWT 密钥并自动替换 .env 文件中的 JWT_SECRET 值。
这个命令会生成一个 32 字节的随机密钥，并将其更新到 .env 文件中。`,
	Run: func(_ *cobra.Command, _ []string) {
		generateAndReplaceJWTSecret()
	},
}

// generateJWTSecret 生成一个安全的 JWT 密钥
func generateJWTSecret() (string, error) {
	// 生成 32 字节的随机数据
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("生成随机密钥失败: %w", err)
	}

	// 转换为 base64 字符串
	secret := base64.StdEncoding.EncodeToString(key)
	return secret, nil
}

// updateEnvFile 更新 .env 文件中的 JWT_SECRET
func updateEnvFile(newSecret string) error {
	envFile := ".env"

	// 检查文件是否存在
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fmt.Errorf(".env 文件不存在")
	}

	// 读取文件内容
	content, err := os.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("读取 .env 文件失败: %w", err)
	}

	// 将内容按行分割
	lines := strings.Split(string(content), "\n")

	// 查找并替换 JWT_SECRET 行
	found := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "JWT_SECRET=") {
			lines[i] = fmt.Sprintf("JWT_SECRET=%s", newSecret)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("在 .env 文件中未找到 JWT_SECRET 配置")
	}

	// 将修改后的内容写回文件
	newContent := strings.Join(lines, "\n")
	err = os.WriteFile(envFile, []byte(newContent), 0600)
	if err != nil {
		return fmt.Errorf("写入 .env 文件失败: %w", err)
	}

	return nil
}

// generateAndReplaceJWTSecret 生成新的 JWT 密钥并替换文件中的值
func generateAndReplaceJWTSecret() {
	color.Cyan("正在生成新的 JWT 密钥...")

	// 生成新密钥
	newSecret, err := generateJWTSecret()
	if err != nil {
		color.Red("生成密钥失败: %v", err)
		os.Exit(1)
	}

	color.Green("新的 JWT 密钥已生成: %s", newSecret)

	// 更新 .env 文件
	color.Yellow("正在更新 .env 文件...")
	err = updateEnvFile(newSecret)
	if err != nil {
		color.Red("更新 .env 文件失败: %v", err)
		os.Exit(1)
	}

	color.Green("JWT 密钥已成功更新到 .env 文件")
	color.Yellow("请重启应用程序以使新密钥生效")
}

func init() {
	// 将 jwt:generate 命令添加到根命令
	rootCmd.AddCommand(jwtGenerateCmd)
}
