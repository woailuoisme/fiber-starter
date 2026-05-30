package realtime

import "github.com/redis/go-redis/v9"

// Config 定义了实时通信模块所需要的全部配置参数，与全局配置结构解耦
type Config struct {
	Enabled           bool          `json:"enabled"`
	AppID             string        `json:"app_id"`
	AppKey            string        `json:"app_key"`
	AppSecret         string        `json:"app_secret"`
	Path              string        `json:"path"`
	SSEPath           string        `json:"sse_path"`
	BusMode           string        `json:"bus_mode"` // "redis" 或 "memory"
	RedisPrefix       string        `json:"redis_prefix"`
	WriteQueueSize    int           `json:"write_queue_size"`
	MaxMessageSize    int           `json:"max_message_size"`
	HeartbeatInterval int           `json:"heartbeat_interval"`
	PresenceTTL       int           `json:"presence_ttl"`
	URL               string        `json:"url"`
	ClientURL         string        `json:"client_url"`
	ClientSSEURL      string        `json:"client_sse_url"`
	APIKey            string        `json:"api_key"`
	Secret            string        `json:"secret"`
	RedisClient       *redis.Client `json:"-"` // 可选：从外部注入复用的 Redis 客户端实例
}
