package providers_test

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"fiber-starter/configs"
	cacheDrivers "fiber-starter/internal/providers/cache/Drivers"

	"github.com/stretchr/testify/require"
)

func TestRedisStore_BasicPathsWithDockerRedis(t *testing.T) {
	containerName := fmt.Sprintf("fiber-starter-redis-%d", time.Now().UnixNano())

	run := exec.Command("docker", "run", "-d", "--rm", "-P", "--name", containerName, "redis:7-alpine")
	var runOut bytes.Buffer
	run.Stdout = &runOut
	run.Stderr = &runOut
	require.NoErrorf(t, run.Run(), "docker run failed: %s", runOut.String())
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-f", containerName).Run()
	})

	require.Eventually(t, func() bool {
		ping := exec.Command("docker", "exec", containerName, "redis-cli", "ping")
		return ping.Run() == nil
	}, 20*time.Second, 200*time.Millisecond)

	portCmd := exec.Command("docker", "port", containerName, "6379/tcp")
	portOut, err := portCmd.Output()
	require.NoError(t, err)
	hostPort := parseDockerPort(string(portOut))
	require.NotEmpty(t, hostPort)

	cfg := &configs.Config{
		Redis: configs.RedisConfig{
			Host: "127.0.0.1",
			Port: hostPort,
		},
		Cache: configs.CacheConfig{
			Prefix: "cache:",
		},
	}

	store := cacheDrivers.NewRedisStore(cfg)
	require.NotNil(t, store)

	require.NoError(t, store.Set("key", map[string]string{"name": "fiber"}, time.Minute))

	got, err := store.Get("key")
	require.NoError(t, err)
	require.NotEmpty(t, got)

	bytesValue, err := store.GetBytes("key")
	require.NoError(t, err)
	require.NotEmpty(t, bytesValue)

	var decoded map[string]string
	require.NoError(t, store.GetJSON("key", &decoded))
	require.Equal(t, "fiber", decoded["name"])

	exists, err := store.Exists("key")
	require.NoError(t, err)
	require.True(t, exists)

	has, err := store.Has("key")
	require.NoError(t, err)
	require.True(t, has)

	added, err := store.Add("other", "value", time.Minute)
	require.NoError(t, err)
	require.True(t, added)

	require.NoError(t, store.Forever("forever", "value"))
	pulled, err := store.Pull("other")
	require.NoError(t, err)
	require.Equal(t, "value", pulled)

	require.Error(t, store.DeletePattern("*"))
	require.NoError(t, store.Delete("key"))
	require.NoError(t, store.Flush())
	_, err = store.TTL("key")
	require.Error(t, err)
	require.NoError(t, store.Expire("forever", time.Minute))
	_, err = store.Increment("counter")
	require.Error(t, err)
	_, err = store.Decrement("counter")
	require.Error(t, err)
	require.NoError(t, store.Close())
}

func parseDockerPort(output string) string {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return ""
	}

	lines := strings.Split(trimmed, "\n")
	last := lines[len(lines)-1]
	parts := strings.Split(last, ":")
	if len(parts) == 0 {
		return ""
	}

	port := parts[len(parts)-1]
	port = strings.TrimSpace(port)
	port = strings.TrimSuffix(port, "/tcp")
	_, _ = strconv.Atoi(port)
	return port
}

func TestRedisStore_DockerPortParser(t *testing.T) {
	require.Equal(t, "6379", parseDockerPort("0.0.0.0:6379/tcp\n"))
	require.Empty(t, parseDockerPort(""))
}
