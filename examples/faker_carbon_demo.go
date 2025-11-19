package main

import (
	"fmt"
	"math/rand"

	"github.com/go-faker/faker/v4"
	"github.com/golang-module/carbon/v2"
)

// 演示 faker 和 carbon 库的使用
func main() {
	fmt.Println("=== Faker 库使用示例 ===")

	// 生成随机姓名
	fmt.Printf("姓名: %s\n", faker.Name())

	// 生成随机邮箱
	fmt.Printf("邮箱: %s\n", faker.Email())

	// 生成随机手机号
	fmt.Printf("手机号: %s\n", faker.Phonenumber())

	// 生成随机句子
	fmt.Printf("句子: %s\n", faker.Sentence())

	// 生成随机数字 (使用 Go 标准库)
	fmt.Printf("随机数字 (1-100): %d\n", rand.Intn(100)+1)

	// 生成随机布尔值
	fmt.Printf("随机布尔值: %t\n", rand.Intn(2) == 1)

	// 生成随机用户名
	fmt.Printf("用户名: %s\n", faker.Username())

	// 生成随机URL
	fmt.Printf("URL: %s\n", faker.URL())

	// 生成随机UUID
	fmt.Printf("UUID: %s\n", faker.UUIDHyphenated())

	fmt.Println("\n=== Carbon 库使用示例 ===")

	// 当前时间
	now := carbon.Now()
	fmt.Printf("当前时间: %s\n", now.ToDateTimeString())

	// 格式化时间
	fmt.Printf("格式化时间: %s\n", now.Format("Y-m-d H:i:s"))

	// 时间计算
	fmt.Printf("昨天: %s\n", now.SubDay().ToDateString())
	fmt.Printf("明天: %s\n", now.AddDay().ToDateString())
	fmt.Printf("3天后: %s\n", now.AddDays(3).ToDateString())
	fmt.Printf("1周前: %s\n", now.SubWeek().ToDateString())
	fmt.Printf("2个月前: %s\n", now.SubMonths(2).ToDateString())

	// 时间比较
	fmt.Printf("是否是今天: %t\n", now.IsToday())
	fmt.Printf("是否是周末: %t\n", now.IsWeekend())

	// 时间戳
	fmt.Printf("时间戳: %d\n", now.Timestamp())

	// 解析时间字符串
	parsedTime := carbon.Parse("2024-01-01 12:00:00")
	fmt.Printf("解析时间: %s\n", parsedTime.ToDateTimeString())
	fmt.Printf("是闰年: %t\n", parsedTime.IsLeapYear())

	// 时间差计算
	birthday := carbon.CreateFromDate(1990, 5, 15)
	age := now.DiffInYears(birthday)
	fmt.Printf("年龄: %d 岁\n", age)

	fmt.Println("\n=== 在 Seeder 中的使用示例 ===")

	// 模拟生成用户数据
	for i := 0; i < 5; i++ {
		name := faker.Name()
		email := faker.Email()
		phone := faker.Phonenumber()
		daysAgo := rand.Intn(365) + 1
		createdAt := carbon.Now().SubDays(daysAgo)

		fmt.Printf("用户 %d: %s | %s | %s | 注册时间: %s\n",
			i+1, name, email, phone, createdAt.ToDateString())
	}
}
