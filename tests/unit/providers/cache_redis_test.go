package providers_test

import (
	"testing"
	"time"

	cacheDrivers "fiber-starter/internal/providers/cache/drivers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisStore_NilReceiverPaths(t *testing.T) {
	var store *cacheDrivers.RedisStore

	key := "missing"

	val, err := store.Get(key)
	require.Error(t, err)
	assert.Empty(t, val)

	bytesVal, err := store.GetBytes(key)
	require.Error(t, err)
	assert.Nil(t, bytesVal)

	var decoded map[string]interface{}
	require.Error(t, store.GetJSON(key, &decoded))

	require.Error(t, store.Set(key, "value", time.Minute))
	require.Error(t, store.Put(key, "value", time.Minute))

	added, err := store.Add(key, "value", time.Minute)
	require.Error(t, err)
	assert.False(t, added)

	require.Error(t, store.Forever(key, "value"))

	require.NoError(t, store.Delete(key))
	require.NoError(t, store.Forget(key))

	require.NoError(t, store.DeletePattern("*"))
	require.NoError(t, store.Flush())

	exists, err := store.Exists(key)
	require.NoError(t, err)
	assert.False(t, exists)

	has, err := store.Has(key)
	require.NoError(t, err)
	assert.False(t, has)

	pulled, err := store.Pull(key)
	require.Error(t, err)
	assert.Empty(t, pulled)

	_, err = store.TTL(key)
	require.Error(t, err)
	require.Error(t, store.Expire(key, time.Minute))
	_, err = store.Increment(key)
	require.Error(t, err)
	_, err = store.Decrement(key)
	require.Error(t, err)

	require.NoError(t, store.Close())
}
