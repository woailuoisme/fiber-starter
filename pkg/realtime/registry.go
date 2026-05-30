package realtime

import (
	"context"
	"regexp"
	"strings"
	"sync"
)

type channelRoute struct {
	pattern string
	names   []string
	re      *regexp.Regexp
	auth    ChannelAuthorization
}

type channelRegistry struct {
	mu     sync.RWMutex
	routes []channelRoute
}

func newChannelRegistry() *channelRegistry {
	return &channelRegistry{}
}

func (r *channelRegistry) Register(pattern string, auth ChannelAuthorization) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || auth == nil {
		return
	}

	route := compileChannelRoute(pattern, auth)

	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = append(r.routes, route)
}

func (r *channelRegistry) Authorize(ctx context.Context, user User, channel string) error {
	if r == nil {
		return nil
	}

	var matchedRoute channelRoute
	var params map[string]string
	found := false

	// 在读锁区间内仅做正则查找和参数提取以缩短锁持有时间，避免锁内执行业务回调导致潜在死锁
	r.mu.RLock()
	for _, route := range r.routes {
		matches := route.re.FindStringSubmatch(channel)
		if matches != nil {
			params = make(map[string]string, len(route.names))
			for i, name := range route.names {
				params[name] = matches[i+1]
			}
			matchedRoute = route
			found = true
			break
		}
	}
	r.mu.RUnlock()

	if !found {
		return nil
	}

	return matchedRoute.auth(ctx, user, channel, params)
}

func compileChannelRoute(pattern string, auth ChannelAuthorization) channelRoute {
	var names []string
	var out strings.Builder
	out.WriteString("^")

	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]
		switch ch {
		case '{':
			end := strings.IndexByte(pattern[i+1:], '}')
			if end < 0 {
				out.WriteString(regexp.QuoteMeta(pattern[i:]))
				i = len(pattern)
				break
			}
			name := pattern[i+1 : i+1+end]
			names = append(names, name)
			// 支持以点、冒号和斜杠作为分隔符抽取参数
			out.WriteString(`([^\\.:/]+)`)
			i += end + 1
		case '*':
			out.WriteString(`.*`)
		default:
			out.WriteString(regexp.QuoteMeta(string(ch)))
		}
	}

	out.WriteString("$")
	return channelRoute{
		pattern: pattern,
		names:   names,
		re:      regexp.MustCompile(out.String()),
		auth:    auth,
	}
}
