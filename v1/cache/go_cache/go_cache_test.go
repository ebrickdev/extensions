package go_cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ebrickdev/ebrick/cache/store"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewGoCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockGoCacheClientInterface(ctrl)
	// When
	caStore := NewGoCache(client, store.WithCost(8))

	// Then
	assert.IsType(t, new(GoCacheStore), caStore)
	assert.Equal(t, client, caStore.client)
	assert.Equal(t, &store.Options{Cost: 8}, caStore.options)
}

func TestGoCacheGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(cacheValue, true)

	caStore := NewGoCache(client)

	// When
	value, err := caStore.Get(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestGoCacheGetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(nil, false)

	caStore := NewGoCache(client)

	// When
	value, err := caStore.Get(ctx, cacheKey)

	// Then
	assert.Nil(t, value)
	assert.Error(t, err, store.NotFound{})
}

func TestGoCacheGetWithTTL(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().GetWithExpiration(cacheKey).Return(cacheValue, time.Now(), true)

	caStore := NewGoCache(client)

	// When
	value, ttl, err := caStore.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, int64(0), ttl.Milliseconds())
}

func TestGoCacheGetWithTTLWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().GetWithExpiration(cacheKey).Return(nil, time.Now(), false)

	caStore := NewGoCache(client)

	// When
	value, ttl, err := caStore.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Nil(t, value)
	assert.Error(t, err, store.NotFound{})
	assert.Equal(t, 0*time.Second, ttl)
}

func TestGoCacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)

	caStore := NewGoCache(client)

	// When
	err := caStore.Set(ctx, cacheKey, cacheValue, store.WithCost(4))

	// Then
	assert.Nil(t, err)
}

func TestGoCacheSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)

	caStore := NewGoCache(client)

	// When
	err := caStore.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestGoCacheSetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)
	client.EXPECT().Get("gocache_tag_tag1").Return(nil, true)
	cacheKeys := map[string]struct{}{"my-key": {}}
	client.EXPECT().Set("gocache_tag_tag1", cacheKeys, 720*time.Hour)

	caStore := NewGoCache(client)

	// When
	err := caStore.Set(ctx, cacheKey, cacheValue, store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestGoCacheSetWithTagsWhenAlreadyInserted(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)

	cacheKeys := map[string]struct{}{"my-key": {}, "a-second-key": {}}
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, true)

	caStore := NewGoCache(client)

	// When
	err := caStore.Set(ctx, cacheKey, cacheValue, store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestGoCacheDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Delete(cacheKey)

	caStore := NewGoCache(client)

	// When
	err := caStore.Delete(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestGoCacheInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := map[string]struct{}{"a23fdf987h2svc23": {}, "jHG2372x38hf74": {}}

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, true)
	client.EXPECT().Delete("a23fdf987h2svc23")
	client.EXPECT().Delete("jHG2372x38hf74")

	caStore := NewGoCache(client)

	// When
	err := caStore.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestGoCacheInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, false)

	caStore := NewGoCache(client)

	// When
	err := caStore.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestGoCacheClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Flush()

	caStore := NewGoCache(client)

	// When
	err := caStore.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestGoCacheGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockGoCacheClientInterface(ctrl)

	caStore := NewGoCache(client)

	// When - Then
	assert.Equal(t, GoCacheType, caStore.GetType())
}

func TestGoCacheSetTagsConcurrency(t *testing.T) {
	ctx := context.Background()

	client := cache.New(10*time.Second, 30*time.Second)
	caStore := NewGoCache(client)

	for i := 0; i < 200; i++ {
		go func(i int) {
			key := fmt.Sprintf("%d", i)

			err := caStore.Set(
				ctx,
				key,
				[]string{"one", "two"},
				store.WithTags([]string{"tag1", "tag2", "tag3"}),
			)
			assert.Nil(t, err, err)
		}(i)
	}
}

func TestGoCacheInvalidateConcurrency(t *testing.T) {
	ctx := context.Background()

	client := cache.New(10*time.Second, 30*time.Second)
	caStore := NewGoCache(client)

	var tags []string
	for i := 0; i < 200; i++ {
		tags = append(tags, fmt.Sprintf("tag%d", i))
	}

	for i := 0; i < 200; i++ {

		go func(i int) {
			key := fmt.Sprintf("%d", i)

			err := caStore.Set(ctx, key, []string{"one", "two"}, store.WithTags(tags))
			assert.Nil(t, err, err)
		}(i)

		go func(i int) {
			err := caStore.Invalidate(ctx, store.WithInvalidateTags([]string{fmt.Sprintf("tag%d", i)}))
			assert.Nil(t, err, err)
		}(i)

	}
}
