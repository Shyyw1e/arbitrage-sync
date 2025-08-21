package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/domain"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
)

var GlobalOrderCache = NewOrderCache()

type OrderCacheKey struct {
	Source domain.Source
	Pair   domain.Pair
	Side   domain.OrderSide
}

func (k OrderCacheKey) String() string {
	return fmt.Sprintf("%s|%s|%s", k.Source, k.Pair, k.Side)
}

type cacheEntry struct {
	Orders     []*domain.Order
	UpdatedAt  time.Time
	IsUpdating bool
}

type OrderCache struct {
	mu   sync.RWMutex
	data map[string]*cacheEntry
}

func NewOrderCache() *OrderCache {
	return &OrderCache{
		data: make(map[string]*cacheEntry),
		mu:   sync.RWMutex{},
	}
}

func (c *OrderCache) GetOrFetch(
	key OrderCacheKey,
	fetchFunc func() ([]*domain.Order, error),
) ([]*domain.Order, error) {
	keyHash := key.String()

	// Чтение из кэша
	c.mu.RLock()
	entry, exists := c.data[keyHash]
	c.mu.RUnlock()

	if exists {
		// Если кэш свежий
		if time.Since(entry.UpdatedAt) < 60*time.Second {
			logger.Log.Info("Got cache")
			return entry.Orders, nil
		}

		// Если кэш обновляется – ждём
		if entry.IsUpdating {
			for {
				time.Sleep(100 * time.Millisecond)

				c.mu.RLock()
				isDone := !entry.IsUpdating
				orders := entry.Orders
				c.mu.RUnlock()

				if isDone {
					logger.Log.Info("Waited for cache")
					return orders, nil
				}
			}
		}
	}

	// Обновление кэша
	c.mu.Lock()
	entry, exists = c.data[keyHash]
	if !exists {
		entry = &cacheEntry{}
		c.data[keyHash] = entry
	}
	entry.IsUpdating = true
	c.mu.Unlock()

	orders, err := fetchFunc()
	if err != nil {
		logger.Log.Errorf("failed to fetch: %v", err)

		c.mu.Lock()
		entry.IsUpdating = false
		c.mu.Unlock()

		return nil, err
	}

	c.mu.Lock()
	entry.Orders = orders
	entry.UpdatedAt = time.Now()
	entry.IsUpdating = false
	c.mu.Unlock()

	logger.Log.Info("Updated cache")
	return orders, nil
}
