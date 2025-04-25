package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		// Создали кэш с 3 элементами, потом 4 заменяет 1
		c := NewCache(3)
		c.Set("key1", "value1")
		c.Set("key2", "value2")
		c.Set("key3", "value3")
		value1, ok1 := c.Get("key1")
		value2, ok2 := c.Get("key2")
		value3, ok3 := c.Get("key3")
		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.True(t, ok3)
		assert.Equal(t, "value1", value1)
		assert.Equal(t, "value2", value2)
		assert.Equal(t, "value3", value3)
		c.Set("key4", "value4")
		value1, ok1 = c.Get("key1")
		value2, ok2 = c.Get("key2")
		value3, ok3 = c.Get("key3")
		value4, ok4 := c.Get("key4")
		assert.False(t, ok1)
		assert.True(t, ok2)
		assert.True(t, ok3)
		assert.True(t, ok4)
		assert.Nil(t, value1)
		assert.Equal(t, "value2", value2)
		assert.Equal(t, "value3", value3)
		assert.Equal(t, "value4", value4)

		// Тест, что будет удален именно последний, после работы
		c.Clear()
		c.Set("key1", "value1")
		c.Set("key2", "value2")
		c.Set("key3", "value3")
		// поднял вверх 1 и 3
		value1, ok1 = c.Get("key1")
		value2, ok2 = c.Get("key3")
		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.Equal(t, "value1", value1)
		assert.Equal(t, "value3", value2)
		// самый последний элемент 2, он будет удален
		c.Set("key4", "value4")
		value1, ok1 = c.Get("key2")
		assert.False(t, ok1)
		assert.Nil(t, value1)
		value1, ok1 = c.Get("key4")
		assert.True(t, ok1)
		assert.Equal(t, "value4", value1)
	})
}

func TestCacheMultithreading(_ *testing.T) {
	// Remove me if task with asterisk completed.

	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()
	wg.Wait()
}
