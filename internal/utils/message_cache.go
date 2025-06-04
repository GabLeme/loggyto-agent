package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type MessageCache struct {
	entries map[string]time.Time
	ttl     time.Duration
	mu      sync.Mutex
}

func NewMessageCache(ttl time.Duration) *MessageCache {
	return &MessageCache{
		entries: make(map[string]time.Time),
		ttl:     ttl,
	}
}

func (mc *MessageCache) ShouldProcess(message string) bool {
	hash := hashMessage(message)

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Limpeza de mensagens expiradas
	now := time.Now()
	for key, ts := range mc.entries {
		if now.Sub(ts) > mc.ttl {
			delete(mc.entries, key)
		}
	}

	// Se já existe, ignorar
	if _, exists := mc.entries[hash]; exists {
		return false
	}

	mc.entries[hash] = now
	return true
}

func hashMessage(message string) string {
	h := sha256.Sum256([]byte(message))
	return hex.EncodeToString(h[:])
}
