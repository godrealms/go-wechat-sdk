package isv

import (
	"context"
	"sync"
	"time"
)

// AuthorizerTokens 打包单个被授权企业的凭证信息。
type AuthorizerTokens struct {
	CorpID            string
	PermanentCode     string
	CorpAccessToken   string
	CorpTokenExpireAt time.Time
}

// Store 负责持久化 ISV 认证流程中的各类凭证。
//
// 所有方法的第一 key 都是 suiteID,允许多 Client 共享同一 Store。
type Store interface {
	GetSuiteTicket(ctx context.Context, suiteID string) (string, error)
	PutSuiteTicket(ctx context.Context, suiteID, ticket string) error

	GetSuiteToken(ctx context.Context, suiteID string) (token string, expiresAt time.Time, err error)
	PutSuiteToken(ctx context.Context, suiteID, token string, expiresAt time.Time) error

	GetProviderToken(ctx context.Context, suiteID string) (token string, expiresAt time.Time, err error)
	PutProviderToken(ctx context.Context, suiteID, token string, expiresAt time.Time) error

	GetAuthorizer(ctx context.Context, suiteID, corpID string) (*AuthorizerTokens, error)
	PutAuthorizer(ctx context.Context, suiteID, corpID string, tokens *AuthorizerTokens) error
	DeleteAuthorizer(ctx context.Context, suiteID, corpID string) error
	ListAuthorizers(ctx context.Context, suiteID string) ([]string, error)
}

// ---- MemoryStore ----

type tokenEntry struct {
	value     string
	expiresAt time.Time
}

// MemoryStore 是 Store 的进程内默认实现,线程安全。
type MemoryStore struct {
	mu           sync.RWMutex
	suiteTickets map[string]string                       // suiteID → ticket
	suiteTokens  map[string]tokenEntry                   // suiteID → token
	providerToks map[string]tokenEntry                   // suiteID → provider_token
	authorizers  map[string]map[string]*AuthorizerTokens // suiteID → corpID → tokens
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		suiteTickets: make(map[string]string),
		suiteTokens:  make(map[string]tokenEntry),
		providerToks: make(map[string]tokenEntry),
		authorizers:  make(map[string]map[string]*AuthorizerTokens),
	}
}

func (m *MemoryStore) GetSuiteTicket(_ context.Context, suiteID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.suiteTickets[suiteID]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (m *MemoryStore) PutSuiteTicket(_ context.Context, suiteID, ticket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.suiteTickets[suiteID] = ticket
	return nil
}

func (m *MemoryStore) GetSuiteToken(_ context.Context, suiteID string) (string, time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.suiteTokens[suiteID]
	if !ok {
		return "", time.Time{}, ErrNotFound
	}
	return e.value, e.expiresAt, nil
}

func (m *MemoryStore) PutSuiteToken(_ context.Context, suiteID, token string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.suiteTokens[suiteID] = tokenEntry{value: token, expiresAt: expiresAt}
	return nil
}

func (m *MemoryStore) GetProviderToken(_ context.Context, suiteID string) (string, time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.providerToks[suiteID]
	if !ok {
		return "", time.Time{}, ErrNotFound
	}
	return e.value, e.expiresAt, nil
}

func (m *MemoryStore) PutProviderToken(_ context.Context, suiteID, token string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providerToks[suiteID] = tokenEntry{value: token, expiresAt: expiresAt}
	return nil
}

func (m *MemoryStore) GetAuthorizer(_ context.Context, suiteID, corpID string) (*AuthorizerTokens, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inner, ok := m.authorizers[suiteID]
	if !ok {
		return nil, ErrNotFound
	}
	v, ok := inner[corpID]
	if !ok {
		return nil, ErrNotFound
	}
	// 复制一份避免调用方修改污染内部状态
	cp := *v
	return &cp, nil
}

func (m *MemoryStore) PutAuthorizer(_ context.Context, suiteID, corpID string, tokens *AuthorizerTokens) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	inner, ok := m.authorizers[suiteID]
	if !ok {
		inner = make(map[string]*AuthorizerTokens)
		m.authorizers[suiteID] = inner
	}
	cp := *tokens
	inner[corpID] = &cp
	return nil
}

func (m *MemoryStore) DeleteAuthorizer(_ context.Context, suiteID, corpID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if inner, ok := m.authorizers[suiteID]; ok {
		delete(inner, corpID)
	}
	return nil
}

func (m *MemoryStore) ListAuthorizers(_ context.Context, suiteID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inner, ok := m.authorizers[suiteID]
	if !ok {
		return nil, nil
	}
	out := make([]string, 0, len(inner))
	for k := range inner {
		out = append(out, k)
	}
	return out, nil
}
