package lock

import "time"

// Option 锁抢占配置参数闭包
type Option func(*options)

type options struct {
	acquireTimeout time.Duration
	retryInterval  time.Duration
}

// defaultOptions 返回默认配置：不重试，即抢不到立即报错
func defaultOptions() *options {
	return &options{
		acquireTimeout: 0,
		retryInterval:  50 * time.Millisecond,
	}
}

// WithTimeout 设定抢锁总超时时间（在此期间会自旋重试）
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.acquireTimeout = timeout
	}
}

// WithRetryInterval 设定获取锁失败后再次尝试的间隔时间
func WithRetryInterval(interval time.Duration) Option {
	return func(o *options) {
		o.retryInterval = interval
	}
}
