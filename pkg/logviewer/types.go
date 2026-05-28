package logviewer

// LogEntry 代表从日志文件中解析出来的单条记录
// 使用 map[string]any 存储 Context，以便完美兼容 JSON 格式的扩展字段，无需固定 Schema
type LogEntry struct {
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Caller    string         `json:"caller"`
	Message   string         `json:"message"`
	Context   map[string]any `json:"context,omitempty"`
}

// LogFileInfo 代表日志文件的基础元数据
// 在前端做文件列表的初始展示和大小换算
type LogFileInfo struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

// Stats 承载单文件内的级别汇总计数
// 用于直观在大盘上反馈当前系统健康度（ERROR 与 WARN 的分布）
type Stats struct {
	Total int `json:"total"`
	Info  int `json:"info"`
	Warn  int `json:"warn"`
	Error int `json:"error"`
	Fatal int `json:"fatal"`
	Debug int `json:"debug"`
}
