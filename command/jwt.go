package command

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成安全密钥",
	Long:  `生成各种类型的安全密钥，包括 JWT 密钥等`,
}

// jwtCmd represents the jwt command
var jwtCmd = &cobra.Command{
	Use:   "jwt",
	Short: "生成并替换 JWT 密钥",
	Long: `生成一个新的安全 JWT 密钥并自动替换 .env 文件中的 JWT_SECRET 值。
这个命令会生成一个 32 字节的随机密钥，并将其更新到 .env 文件中。`,
	Run: func(cmd *cobra.Command, args []string) {
		generateAndReplaceJWTSecret()
	},
}

// generateJWTSecret 生成一个安全的 JWT 密钥
func generateJWTSecret() (string, error) {
	// 生成 32 字节的随机数据
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("生成随机密钥失败: %v", err)
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
	content, err := ioutil.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("读取 .env 文件失败: %v", err)
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
	err = ioutil.WriteFile(envFile, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("写入 .env 文件失败: %v", err)
	}

	return nil
}

// generateAndReplaceJWTSecret 生成新的 JWT 密钥并替换文件中的值
func generateAndReplaceJWTSecret() {
	fmt.Println("🔐 正在生成新的 JWT 密钥...")

	// 生成新密钥
	newSecret, err := generateJWTSecret()
	if err != nil {
		fmt.Printf("❌ 生成密钥失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 新的 JWT 密钥已生成: %s\n", newSecret)

	// 更新 .env 文件
	fmt.Println("📝 正在更新 .env 文件...")
	err = updateEnvFile(newSecret)
	if err != nil {
		fmt.Printf("❌ 更新 .env 文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ JWT 密钥已成功更新到 .env 文件")
	fmt.Println("🔄 请重启应用程序以使新密钥生效")
}

func init() {
	// 将 jwt 命令添加到 generate 命令
	generateCmd.AddCommand(jwtCmd)

	// 将 generate 命令添加到根命令
	rootCmd.AddCommand(generateCmd)
}
