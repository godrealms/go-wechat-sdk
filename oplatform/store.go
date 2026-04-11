package oplatform

import (
	"context"
	"sync"
	"time"
)

// AuthorizerTokens 是单个授权方 (authorizer) 的一组 token。
type AuthorizerTokens struct {
	AccessToken  string
	RefreshToken string
	ExpireAt     time.Time
}

// Store 持久化 component_verify_ticket、component_access_token
// 以及每个 authorizer 的 refresh_token / access_token / expire_at。
//
// SDK 内置 MemoryStore 供测试和本地开发使用；生产应实现 Redis/MySQL
// 版本并通过 WithStore 注入。
//
// Get* 方法在 key 不存在时应返回 ErrNotFound。
type Store interface {
	GetVerifyTicket(ctx context.Context) (string, error)
	SetVerifyTicket(ctx context.Context, ticket string) error

	GetComponentToken(ctx context.Context) (token string, expireAt time.Time, err error)
	SetComponentToken(ctx context.Context, token string, expireAt time.Time) error

	GetAuthorizer(ctx context.Context, appid string) (AuthorizerTokens, error)
	SetAuthorizer(ctx context.Context, appid string, tokens AuthorizerTokens) error
	DeleteAuthorizer(ctx context.Context, appid string) error
	ListAuthorizerAppIDs(ctx context.Context) ([]string, error)
}

// MemoryStore 是 Store 接口的线程安全内存实现。进程重启后所有数据丢失，
// 仅适合测试或本地开发。
type MemoryStore struct {
	mu sync.RWMutex

	verifyTicket string

	componentToken    string
	componentExpireAt time.Time

	authorizers map[string]AuthorizerTokens
}

// NewMemoryStore 构造一个空的内存 Store。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{authorizers: make(map[string]AuthorizerTokens)}
}

func (m *MemoryStore) GetVerifyTicket(ctx context.Context) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.verifyTicket == "" {
		return "", ErrNotFound
	}
	return m.verifyTicket, nil
}

func (m *MemoryStore) SetVerifyTicket(ctx context.Context, ticket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.verifyTicket = ticket
	return nil
}

func (m *MemoryStore) GetComponentToken(ctx context.Context) (string, time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.componentToken == "" {
		return "", time.Time{}, ErrNotFound
	}
	return m.componentToken, m.componentExpireAt, nil
}

func (m *MemoryStore) SetComponentToken(ctx context.Context, token string, expireAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.componentToken = token
	m.componentExpireAt = expireAt
	return nil
}

func (m *MemoryStore) GetAuthorizer(ctx context.Context, appid string) (AuthorizerTokens, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.authorizers[appid]
	if !ok {
		return AuthorizerTokens{}, ErrNotFound
	}
	return t, nil
}

func (m *MemoryStore) SetAuthorizer(ctx context.Context, appid string, tokens AuthorizerTokens) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.authorizers[appid] = tokens
	return nil
}

func (m *MemoryStore) DeleteAuthorizer(ctx context.Context, appid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.authorizers, appid)
	return nil
}

func (m *MemoryStore) ListAuthorizerAppIDs(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.authorizers))
	for k := range m.authorizers {
		out = append(out, k)
	}
	return out, nil
}
