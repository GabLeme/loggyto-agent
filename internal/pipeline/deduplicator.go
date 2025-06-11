package pipeline

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

type Deduplicator struct {
	cache map[string]time.Time
	ttl   time.Duration
	mu    sync.Mutex
}

func NewDeduplicator(ttl time.Duration) *Deduplicator {
	return &Deduplicator{
		cache: make(map[string]time.Time),
		ttl:   ttl,
	}
}

func (d *Deduplicator) ShouldProcess(message string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	hash := normalizeForDedup(message)
	now := time.Now()

	if exp, found := d.cache[hash]; found {
		if now.Sub(exp) < d.ttl {
			return false
		}
	}

	d.cache[hash] = now
	d.cleanup()
	return true
}

func (d *Deduplicator) cleanup() {
	now := time.Now()
	for k, t := range d.cache {
		if now.Sub(t) > d.ttl {
			delete(d.cache, k)
		}
	}
}

func normalizeForDedup(msg string) string {
	msg = strings.ToLower(strings.TrimSpace(msg))
	msg = removeTimestamp(msg)
	msg = regexp.MustCompile(`\s+`).ReplaceAllString(msg, " ")
	return msg
}

func removeTimestamp(s string) string {
	timestampPattern := regexp.MustCompile(`\[?\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2}(\.\d+)?(Z| UTC)?\]?`)
	return timestampPattern.ReplaceAllString(s, "")
}
