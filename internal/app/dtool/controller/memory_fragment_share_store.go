package controller

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const memoryFragmentShareTTL = 7 * 24 * time.Hour
const memoryFragmentShareFileName = `.memory_fragment_shares.json`

var memoryFragmentShareStoreRegistry sync.Map

type memoryFragmentShare struct {
	Token      string
	FragmentID string
	ExpireAt   time.Time
}

type memoryFragmentShareStore struct {
	mu       sync.Mutex
	filePath string
	items    map[string]memoryFragmentShare
	loaded   bool
}

type memoryFragmentShareRecord struct {
	Token      string `json:"token"`
	FragmentID string `json:"fragment_id"`
	ExpireAt   int64  `json:"expire_at"`
}

func newMemoryFragmentShareStore(root string) *memoryFragmentShareStore {
	return &memoryFragmentShareStore{
		filePath: filepath.Join(strings.TrimSpace(root), memoryFragmentShareFileName),
		items:    map[string]memoryFragmentShare{},
	}
}

func memoryFragmentShareStoreForRoot(root string) *memoryFragmentShareStore {
	root = strings.TrimSpace(root)
	store, _ := memoryFragmentShareStoreRegistry.LoadOrStore(root, newMemoryFragmentShareStore(root))
	return store.(*memoryFragmentShareStore)
}

func (h *memoryFragmentShareStore) Create(fragmentID string, now time.Time) (memoryFragmentShare, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.loadLocked(); err != nil {
		return memoryFragmentShare{}, err
	}
	h.clearExpiredLocked(now)
	token := h.createUniqueTokenLocked()
	share := memoryFragmentShare{
		Token:      token,
		FragmentID: strings.TrimSpace(fragmentID),
		ExpireAt:   now.Add(memoryFragmentShareTTL),
	}
	h.items[token] = share
	if err := h.saveLocked(); err != nil {
		return memoryFragmentShare{}, err
	}
	return share, nil
}

func (h *memoryFragmentShareStore) Resolve(token string, now time.Time) (memoryFragmentShare, bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.loadLocked(); err != nil {
		return memoryFragmentShare{}, false, err
	}
	token = strings.TrimSpace(token)
	share, ok := h.items[token]
	if !ok {
		return memoryFragmentShare{}, false, nil
	}
	if !now.Before(share.ExpireAt) {
		delete(h.items, token)
		if err := h.saveLocked(); err != nil {
			return memoryFragmentShare{}, false, err
		}
		return memoryFragmentShare{}, false, nil
	}
	return share, true, nil
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

func (h *memoryFragmentShareStore) loadLocked() error {
	if h.loaded {
		return nil
	}
	h.loaded = true
	h.items = map[string]memoryFragmentShare{}
	if strings.TrimSpace(h.filePath) == `` {
		return nil
	}
	body, err := os.ReadFile(h.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	recordList := make([]memoryFragmentShareRecord, 0)
	if err = json.Unmarshal(body, &recordList); err != nil {
		return err
	}
	for _, record := range recordList {
		token := strings.TrimSpace(record.Token)
		fragmentID := strings.TrimSpace(record.FragmentID)
		if token == `` || fragmentID == `` || record.ExpireAt <= 0 {
			continue
		}
		h.items[token] = memoryFragmentShare{
			Token:      token,
			FragmentID: fragmentID,
			ExpireAt:   time.Unix(record.ExpireAt, 0),
		}
	}
	return nil
}

func (h *memoryFragmentShareStore) saveLocked() error {
	if strings.TrimSpace(h.filePath) == `` {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(h.filePath), 0o755); err != nil {
		return err
	}
	recordList := make([]memoryFragmentShareRecord, 0, len(h.items))
	for _, share := range h.items {
		recordList = append(recordList, memoryFragmentShareRecord{
			Token:      share.Token,
			FragmentID: share.FragmentID,
			ExpireAt:   share.ExpireAt.Unix(),
		})
	}
	body, err := json.MarshalIndent(recordList, ``, `  `)
	if err != nil {
		return err
	}
	tmpPath := h.filePath + `.tmp`
	if err = os.WriteFile(tmpPath, body, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, h.filePath)
}

func randomMemoryFragmentShareToken() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		sum := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
		return hex.EncodeToString(sum[:])
	}
	return hex.EncodeToString(buf)
}
