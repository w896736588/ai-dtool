package controller

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

const memoryFragmentShareTTL = 24 * time.Hour

type memoryFragmentShare struct {
	Token      string
	FragmentID string
	ExpireAt   time.Time
}

type memoryFragmentShareStore struct {
	mu    sync.Mutex
	items map[string]memoryFragmentShare
}

func newMemoryFragmentShareStore() *memoryFragmentShareStore {
	return &memoryFragmentShareStore{
		items: map[string]memoryFragmentShare{},
	}
}

func (h *memoryFragmentShareStore) Create(fragmentID string, now time.Time) memoryFragmentShare {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clearExpiredLocked(now)
	token := h.createUniqueTokenLocked()
	share := memoryFragmentShare{
		Token:      token,
		FragmentID: strings.TrimSpace(fragmentID),
		ExpireAt:   now.Add(memoryFragmentShareTTL),
	}
	h.items[token] = share
	return share
}

func (h *memoryFragmentShareStore) Resolve(token string, now time.Time) (memoryFragmentShare, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	token = strings.TrimSpace(token)
	share, ok := h.items[token]
	if !ok {
		return memoryFragmentShare{}, false
	}
	if !now.Before(share.ExpireAt) {
		delete(h.items, token)
		return memoryFragmentShare{}, false
	}
	return share, true
}

func (h *memoryFragmentShareStore) clearExpiredLocked(now time.Time) {
	for token, share := range h.items {
		if !now.Before(share.ExpireAt) {
			delete(h.items, token)
		}
	}
}

func (h *memoryFragmentShareStore) createUniqueTokenLocked() string {
	for {
		token := randomMemoryFragmentShareToken()
		if _, exists := h.items[token]; !exists {
			return token
		}
	}
}

func randomMemoryFragmentShareToken() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		sum := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
		return hex.EncodeToString(sum[:])
	}
	return hex.EncodeToString(buf)
}
