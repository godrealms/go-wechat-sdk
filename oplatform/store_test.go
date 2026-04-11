package oplatform

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestMemoryStore_VerifyTicket(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, err := s.GetVerifyTicket(ctx); !errors.Is(err, ErrNotFound) {
		t.Errorf("empty store should return ErrNotFound, got %v", err)
	}
	if err := s.SetVerifyTicket(ctx, "TICKET1"); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetVerifyTicket(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "TICKET1" {
		t.Errorf("got %q", got)
	}
}

func TestMemoryStore_ComponentToken(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, _, err := s.GetComponentToken(ctx); !errors.Is(err, ErrNotFound) {
		t.Errorf("empty store should return ErrNotFound, got %v", err)
	}
	exp := time.Now().Add(time.Hour)
	if err := s.SetComponentToken(ctx, "CTOK", exp); err != nil {
		t.Fatal(err)
	}
	tok, gotExp, err := s.GetComponentToken(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if tok != "CTOK" {
		t.Errorf("token mismatch: %q", tok)
	}
	if !gotExp.Equal(exp) {
		t.Errorf("expiry mismatch")
	}
}

func TestMemoryStore_AuthorizerCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, err := s.GetAuthorizer(ctx, "wxA"); !errors.Is(err, ErrNotFound) {
		t.Errorf("empty store should return ErrNotFound, got %v", err)
	}
	tokens := AuthorizerTokens{
		AccessToken:  "aA",
		RefreshToken: "rA",
		ExpireAt:     time.Now().Add(time.Hour),
	}
	if err := s.SetAuthorizer(ctx, "wxA", tokens); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetAuthorizer(ctx, "wxA")
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "aA" || got.RefreshToken != "rA" {
		t.Errorf("mismatch: %+v", got)
	}

	ids, err := s.ListAuthorizerAppIDs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "wxA" {
		t.Errorf("list mismatch: %v", ids)
	}

	if err := s.DeleteAuthorizer(ctx, "wxA"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetAuthorizer(ctx, "wxA"); !errors.Is(err, ErrNotFound) {
		t.Errorf("after delete expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = s.SetVerifyTicket(ctx, "t")
			_, _ = s.GetVerifyTicket(ctx)
			_ = s.SetAuthorizer(ctx, "wx", AuthorizerTokens{
				AccessToken: "a", RefreshToken: "r", ExpireAt: time.Now(),
			})
			_, _ = s.GetAuthorizer(ctx, "wx")
			_, _ = s.ListAuthorizerAppIDs(ctx)
		}(i)
	}
	wg.Wait()
}
