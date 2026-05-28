# 1. 使用扁平单包（Flat Feature Package）结构组织业务特性

为了规避 Go 语言在包级别的循环导入（import cycle）限制并简化目录层级，我们决定所有的业务特性（Feature）内部文件均采用扁平单包结构，同属于同一个 package。严禁在特性目录下划分 `controllers`、`services`、`repositories` 等嵌套子包；若特性复杂度增长，应优先采用水平拆分出新 Feature 的方式进行解耦。
