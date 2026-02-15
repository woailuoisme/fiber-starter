package command

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"fiber-starter/config"
	"fiber-starter/database/seeders"

	"github.com/fatih/color"
	"github.com/golang-migrate/migrate/v4"

	// 引入 postgres 驱动
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// 引入 file 驱动
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "数据库迁移管理",
	Long:  `管理数据库迁移，包括运行迁移、回滚迁移等操作`,
}

// migrateRunCmd represents the migrate:run command
var migrateRunCmd = &cobra.Command{
	Use:   "run",
	Short: "运行所有待执行的数据库迁移",
	Long: `运行所有待执行的数据库迁移。
这个命令会执行所有尚未运行的迁移，将数据库结构更新到最新状态。`,
	Run: func(_ *cobra.Command, _ []string) {
		runMigrations()
	},
}

// migrateRollbackCmd represents the migrate:rollback command
var migrateRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "回滚最后一次数据库迁移",
	Long: `回滚最后一次执行的数据库迁移。
这个命令会撤销最后一次迁移操作，将数据库恢复到迁移前的状态。`,
	Run: func(_ *cobra.Command, _ []string) {
		rollbackMigrations()
	},
}

// migrateResetCmd represents the migrate:reset command
var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "重置数据库（删除所有表并重新运行迁移）",
	Long: `重置数据库，删除所有表并重新运行所有迁移。
警告：这个操作会删除所有数据，请谨慎使用！`,
	Run: func(_ *cobra.Command, _ []string) {
		resetDatabase()
	},
}

// migrateFreshCmd represents the migrate:fresh command
var migrateFreshCmd = &cobra.Command{
	Use:   "fresh",
	Short: "清空数据库并重新运行迁移和种子数据",
	Long: `清空数据库，删除所有表并重新运行所有迁移，然后运行种子数据。
警告：这个操作会删除所有数据，请谨慎使用！`,
	Run: func(_ *cobra.Command, _ []string) {
		freshDatabase()
	},
}

// migrateStatusCmd represents the migrate:status command
var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "显示迁移状态",
	Long:  `显示数据库迁移的当前状态，包括已运行和待运行的迁移。`,
	Run: func(_ *cobra.Command, _ []string) {
		showMigrationStatus()
	},
}

// seedCmd represents the seed command
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "数据库种子数据管理",
	Long:  `管理数据库种子数据，包括运行种子数据、清除种子数据等操作`,
}

// seedRunCmd represents the seed:run command
var seedRunCmd = &cobra.Command{
	Use:   "run",
	Short: "运行所有种子数据",
	Long: `运行所有种子数据，填充数据库的初始数据。
这个命令会执行所有种子数据的创建操作。`,
	Run: func(_ *cobra.Command, _ []string) {
		runSeeds()
	},
}

// seedRunRandomCmd represents the seed:run:random command
var seedRunRandomCmd = &cobra.Command{
	Use:   "random [count]",
	Short: "生成指定数量的随机测试数据",
	Long: `生成指定数量的随机测试数据。
如果不指定数量，默认生成 10 条记录。`,
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		count := 10 // 默认数量
		if len(args) > 0 {
			_, _ = fmt.Sscanf(args[0], "%d", &count)
		}
		runRandomSeeds(count)
	},
}

// seedClearCmd represents the seed:clear command
var seedClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "清除所有种子数据",
	Long: `清除所有种子数据，删除由种子数据创建的记录。
这个操作会删除所有种子数据，但不会删除表结构。`,
	Run: func(_ *cobra.Command, _ []string) {
		clearSeeds()
	},
}

// seedRefreshCmd represents the seed:refresh command
var seedRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "刷新种子数据（清除后重新运行）",
	Long: `刷新种子数据，先清除所有种子数据，然后重新运行。
这个命令会先清除现有的种子数据，然后重新填充。`,
	Run: func(_ *cobra.Command, _ []string) {
		refreshSeeds()
	},
}

// dbCmd represents the database command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "数据库操作管理",
	Long:  `提供数据库相关的各种操作，包括迁移、种子数据管理等`,
}

// dbSetupCmd represents the db:setup command
var dbSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "设置数据库（运行迁移和种子数据）",
	Long: `设置数据库，运行所有迁移并填充种子数据。
这个命令会依次执行迁移和种子数据操作，完成数据库的初始化。`,
	Run: func(_ *cobra.Command, _ []string) {
		setupDatabase()
	},
}

// getMigrate 获取 migrate 实例
func getMigrate() (*migrate.Migrate, error) {
	// 初始化配置
	if err := config.Init(); err != nil {
		return nil, fmt.Errorf("初始化配置失败: %w", err)
	}

	// 获取数据库配置
	dbConfig := &config.GlobalConfig.Database
	defaultConn := dbConfig.Default
	connConfig, exists := dbConfig.Connections[defaultConn]
	if !exists {
		return nil, fmt.Errorf("数据库连接配置 '%s' 不存在", defaultConn)
	}

	// 构建数据库连接字符串
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		connConfig.Username,
		connConfig.Password,
		connConfig.Host,
		connConfig.Port,
		connConfig.Database,
		connConfig.SSLMode,
	)

	// 创建 migrate 实例
	m, err := migrate.New(
		"file://database/migrations/sql",
		databaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("创建迁移实例失败: %w", err)
	}

	return m, nil
}

// runMigrations 运行数据库迁移
func runMigrations() {
	color.Cyan("正在运行数据库迁移...")

	m, err := getMigrate()
	if err != nil {
		color.Red("获取迁移实例失败: %v", err)
		os.Exit(1)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		color.Red("迁移失败: %v", err)
		os.Exit(1)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		color.Yellow("没有待执行的迁移")
	} else {
		color.Green("数据库迁移完成")
	}
}

// rollbackMigrations 回滚数据库迁移
func rollbackMigrations() {
	color.Cyan("正在回滚数据库迁移...")

	m, err := getMigrate()
	if err != nil {
		color.Red("获取迁移实例失败: %v", err)
		os.Exit(1)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		color.Red("回滚失败: %v", err)
		os.Exit(1)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		color.Yellow("没有可回滚的迁移")
	} else {
		color.Green("数据库迁移回滚完成")
	}
}

// resetDatabase 重置数据库
func resetDatabase() {
	// 确认操作
	fmt.Print("警告：这将删除所有数据！确定要继续吗？(y/N): ")
	var response string
	_, _ = fmt.Scanln(&response)

	if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
		color.Yellow("操作已取消")
		return
	}

	color.Cyan("正在重置数据库...")

	m, err := getMigrate()
	if err != nil {
		_, _ = color.New(color.FgRed).Printf("获取迁移实例失败: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_, _ = m.Close()
	}()

	// 回滚所有迁移
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		_, _ = color.New(color.FgRed).Printf("回滚失败: %v\n", err)
		os.Exit(1)
	}

	// 重新运行所有迁移
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		_, _ = color.New(color.FgRed).Printf("迁移失败: %v\n", err)
		os.Exit(1)
	}

	color.Green("数据库重置完成")
}

// freshDatabase 清空数据库并重新运行迁移和种子数据
func freshDatabase() {
	// 确认操作
	fmt.Print("警告：这将删除所有数据！确定要继续吗？(y/N): ")
	var response string
	_, _ = fmt.Scanln(&response)

	if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
		color.Yellow("操作已取消")
		return
	}

	color.Cyan("正在清空数据库...")

	m, err := getMigrate()
	if err != nil {
		color.Red("获取迁移实例失败: %v", err)
		os.Exit(1)
	}
	defer func() {
		_, _ = m.Close()
	}()

	// 回滚所有迁移
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		color.Red("回滚失败: %v", err)
		os.Exit(1)
	}

	// 重新运行所有迁移
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		color.Red("迁移失败: %v", err)
		os.Exit(1)
	}

	color.Cyan("正在运行种子数据...")
	err = seeders.RunAllSeeders()
	if err != nil {
		color.Red("运行种子数据失败: %v", err)
		os.Exit(1)
	}

	color.Green("数据库清空并重新初始化完成")
}

// showMigrationStatus 显示迁移状态
func showMigrationStatus() {
	color.Cyan("正在检查迁移状态...")

	m, err := getMigrate()
	if err != nil {
		color.Red("获取迁移实例失败: %v", err)
		os.Exit(1)
	}
	defer func() {
		_, _ = m.Close()
	}()

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		color.Red("获取迁移版本失败: %v", err)
		os.Exit(1)
	}

	if errors.Is(err, migrate.ErrNilVersion) {
		color.Yellow("数据库尚未运行任何迁移")
	} else {
		color.Green("当前迁移版本: %d", version)
		if dirty {
			color.Red("警告：数据库处于脏状态（dirty），可能需要手动修复")
		}
	}
}

// runSeeds 运行种子数据
func runSeeds() {
	color.Cyan("正在运行种子数据...")

	err := seeders.RunAllSeeders()
	if err != nil {
		color.Red("种子数据运行失败: %v", err)
		os.Exit(1)
	}

	color.Green("种子数据运行完成")
}

// runRandomSeeds 生成随机测试数据
func runRandomSeeds(count int) {
	color.Cyan("正在生成 %d 条随机测试数据...", count)

	err := seeders.RunRandomSeeders(count)
	if err != nil {
		color.Red("生成随机数据失败: %v", err)
		os.Exit(1)
	}

	color.Green("成功生成 %d 条随机测试数据", count)
}

// clearSeeds 清除种子数据
func clearSeeds() {
	color.Cyan("正在清除种子数据...")

	err := seeders.ClearAllSeeders()
	if err != nil {
		color.Red("清除种子数据失败: %v", err)
		os.Exit(1)
	}

	color.Green("种子数据清除完成")
}

// refreshSeeds 刷新种子数据
func refreshSeeds() {
	color.Cyan("正在刷新种子数据...")

	err := seeders.RefreshAllSeeders()
	if err != nil {
		color.Red("刷新种子数据失败: %v", err)
		os.Exit(1)
	}

	color.Green("种子数据刷新完成")
}

// setupDatabase 设置数据库
func setupDatabase() {
	color.Cyan("正在设置数据库...")

	// 运行迁移
	color.Yellow("步骤 1/2: 运行数据库迁移")
	m, err := getMigrate()
	if err != nil {
		color.Red("获取迁移实例失败: %v", err)
		os.Exit(1)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		color.Red("迁移失败: %v", err)
		os.Exit(1)
	}

	// 运行种子数据
	color.Yellow("步骤 2/2: 运行种子数据")
	err = seeders.RunAllSeeders()
	if err != nil {
		color.Red("种子数据运行失败: %v", err)
		os.Exit(1)
	}

	color.Green("数据库设置完成")
}

// init 初始化命令
func init() {
	// 添加迁移子命令
	migrateCmd.AddCommand(migrateRunCmd)
	migrateCmd.AddCommand(migrateRollbackCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	migrateCmd.AddCommand(migrateFreshCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	// 添加种子数据子命令
	seedCmd.AddCommand(seedRunCmd)
	seedCmd.AddCommand(seedRunRandomCmd)
	seedCmd.AddCommand(seedClearCmd)
	seedCmd.AddCommand(seedRefreshCmd)

	// 添加数据库子命令
	dbCmd.AddCommand(migrateCmd)
	dbCmd.AddCommand(seedCmd)
	dbCmd.AddCommand(dbSetupCmd)

	// 将数据库命令添加到根命令
	rootCmd.AddCommand(dbCmd)

	// 也可以直接添加迁移和种子命令到根命令（向后兼容）
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(seedCmd)
}
