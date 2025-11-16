package tests

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExample 使用 testify 框架进行简单测试
func TestExample(t *testing.T) {
	// 使用 assert 断言
	assert.Equal(t, 1, 1, "1 应该等于 1")

	// 使用 require 断言
	require.NotEmpty(t, "test", "字符串不应该为空")
}

// ExampleTestSuite 使用 testify 的 suite 进行测试
func ExampleTestSuite(t *testing.T) {
	suite := &ExampleSuite{}
	// 这里可以运行 suite
}

// ExampleSuite 测试套件示例
type ExampleSuite struct {
	// 嵌入测试套件
}

// GinkgoExample 使用 Ginkgo 框架的示例测试
var _ = ginkgo.Describe("Example", func() {
	ginkgo.It("should do something", func() {
		gomega.Expect(1).To(gomega.Equal(1))
	})
})