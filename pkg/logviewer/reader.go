package logviewer

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

// listLogFiles 检索 logsDir 目录下所有以 .log 结尾的文件列表
// 并在内存中将它们按照修改时间从新到旧（按文件名倒序）进行排序
func listLogFiles(logsDir string) ([]LogFileInfo, error) {
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []LogFileInfo{}, nil
		}
		return nil, err
	}

	var files []LogFileInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, LogFileInfo{
			Name:     entry.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	// 按文件名降序排序（符合我们日志文件带日期命名的倒序排列）
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].Name < files[j].Name {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	return files, nil
}

// readLogFileBackward 实现日志大文件的逆向（从文件末尾向前）流式分块 Seeking 读取
// 核心设计：
// 1. chunkSize 设为 64KB，每次往前 Seek 偏移量，只在内存中解析这一块数据，解决大日志文件（GB 级）一次性载入发生 OOM 的问题。
// 2. 在 JSON 解析前，利用极其低廉的 strings.Contains 进行 level/search 快速过滤。只有文本匹配成功的行，才触发高开销的 json.Unmarshal，从而数倍提升海量日志检索性能。
func readLogFileBackward(filePath string, page, limit int, search, level string) ([]LogEntry, int, Stats, error) {
	// #nosec G304
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, Stats{}, err
	}
	defer func() {
		_ = file.Close()
	}()

	stat, err := file.Stat()
	if err != nil {
		return nil, 0, Stats{}, err
	}

	fileSize := stat.Size()
	var entries []LogEntry
	var totalMatches int
	var stats Stats

	if fileSize == 0 {
		return entries, 0, stats, nil
	}

	chunkSize := int64(64 * 1024)
	offset := fileSize
	var leftover []byte

	startIdx := (page - 1) * limit
	level = strings.ToUpper(strings.TrimSpace(level))
	search = strings.ToLower(strings.TrimSpace(search))

	for offset > 0 {
		readSize := chunkSize
		if offset < chunkSize {
			readSize = offset
		}
		offset -= readSize

		_, err := file.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, 0, Stats{}, err
		}

		buf := make([]byte, readSize)
		_, err = file.Read(buf)
		if err != nil {
			return nil, 0, Stats{}, err
		}

		if len(leftover) > 0 {
			buf = append(buf, leftover...)
		}

		lines := splitLines(buf)
		if offset > 0 {
			// 首行有可能在分块边界处被截断，故保留作为 leftover 留给下一次往前 Seek 时拼接
			leftover = lines[0]
			lines = lines[1:]
		} else {
			leftover = nil
		}

		// 倒序处理当前分块内的行，确保最新的日志（在文件最尾部）先输出
		for i := len(lines) - 1; i >= 0; i-- {
			lineBytes := lines[i]
			if len(lineBytes) == 0 {
				continue
			}

			lineStr := string(lineBytes)

			// 快速匹配 level 类别，用于大盘汇总统计
			lineLower := strings.ToLower(lineStr)
			var currentLevel string
			switch {
			case strings.Contains(lineLower, `"level":"info"`):
				stats.Info++
				stats.Total++
				currentLevel = "INFO"
			case strings.Contains(lineLower, `"level":"warn"`):
				stats.Warn++
				stats.Total++
				currentLevel = "WARN"
			case strings.Contains(lineLower, `"level":"error"`):
				stats.Error++
				stats.Total++
				currentLevel = "ERROR"
			case strings.Contains(lineLower, `"level":"fatal"`):
				stats.Fatal++
				stats.Total++
				currentLevel = "FATAL"
			case strings.Contains(lineLower, `"level":"debug"`):
				stats.Debug++
				stats.Total++
				currentLevel = "DEBUG"
			default:
				// 非 JSON 日志做默认 INFO 处理
				stats.Info++
				stats.Total++
				currentLevel = "INFO"
			}

			// 进行前置过滤，不命中则跳过后面的反序列化流程
			if level != "" && currentLevel != level {
				continue
			}
			if search != "" && !strings.Contains(lineLower, search) {
				continue
			}

			totalMatches++

			// 仅对匹配且属于当前分页区间的记录做 parseLogLine 解析
			if totalMatches > startIdx && len(entries) < limit {
				entries = append(entries, parseLogLine(lineStr, currentLevel))
			}
		}
	}

	return entries, totalMatches, stats, nil
}

// splitLines 按换行符分割分块数据，为 Seeking 算法拼接做支持
func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

// parseLogLine 对行日志进行解包
// 如果是 JSON 格式，则智能映射常用时间、Level、Message，剩下的字段自动聚合放入 Context
// 如果是非 JSON，则做优雅降级文本包裹
func parseLogLine(line string, fastDetectedLevel string) LogEntry {
	var rawMap map[string]any
	err := json.Unmarshal([]byte(line), &rawMap)
	if err != nil {
		// 非 JSON 格式退避包装
		return LogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			Level:     fastDetectedLevel,
			Caller:    "-",
			Message:   line,
		}
	}

	entry := LogEntry{
		Level: fastDetectedLevel,
	}

	// 1. 提取时间戳
	if ts, ok := rawMap["timestamp"].(string); ok {
		entry.Timestamp = ts
	} else if ts, ok := rawMap["ts"].(string); ok {
		entry.Timestamp = ts
	} else if ts, ok := rawMap["time"].(string); ok {
		entry.Timestamp = ts
	}

	// 2. 提取细粒度的 Level 覆盖
	if lvl, ok := rawMap["level"].(string); ok {
		entry.Level = strings.ToUpper(lvl)
	}

	// 3. 提取调用位置
	if caller, ok := rawMap["caller"].(string); ok {
		entry.Caller = caller
	}

	// 4. 提取核心日志内容
	if msg, ok := rawMap["message"].(string); ok {
		entry.Message = msg
	} else if msg, ok := rawMap["msg"].(string); ok {
		entry.Message = msg
	}

	// 5. 将剩余的所有额外元数据字段收集到 context 中返回给前端展示
	contextMap := make(map[string]any)
	for k, v := range rawMap {
		if k == "timestamp" || k == "ts" || k == "time" ||
			k == "level" || k == "caller" ||
			k == "message" || k == "msg" {
			continue
		}
		contextMap[k] = v
	}

	if len(contextMap) > 0 {
		entry.Context = contextMap
	}

	return entry
}

// Export_readLogFileBackward 提供给单元测试作为包外反射，保证核心 Seeking 读取算法被精准测试覆盖
func Export_readLogFileBackward(filePath string, page, limit int, search, level string) ([]LogEntry, int, Stats, error) {
	return readLogFileBackward(filePath, page, limit, search, level)
}
