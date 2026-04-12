package isv

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryStore_SuiteTicket(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, err := s.GetSuiteTicket(ctx, "suite1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if err := s.PutSuiteTicket(ctx, "suite1", "TICKET"); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetSuiteTicket(ctx, "suite1")
	if err != nil || got != "TICKET" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestMemoryStore_SuiteToken(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, _, err := s.GetSuiteToken(ctx, "suite1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	exp := time.Now().Add(time.Hour)
	if err := s.PutSuiteToken(ctx, "suite1", "TOK", exp); err != nil {
		t.Fatal(err)
	}
	tok, gotExp, err := s.GetSuiteToken(ctx, "suite1")
	if err != nil || tok != "TOK" || !gotExp.Equal(exp) {
		t.Fatalf("got %q %v err=%v", tok, gotExp, err)
	}
}

func TestMemoryStore_ProviderToken(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, _, err := s.GetProviderToken(ctx, "suite1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	exp := time.Now().Add(time.Hour)
	if err := s.PutProviderToken(ctx, "suite1", "PTOK", exp); err != nil {
		t.Fatal(err)
	}
	tok, gotExp, err := s.GetProviderToken(ctx, "suite1")
	if err != nil || tok != "PTOK" || !gotExp.Equal(exp) {
		t.Fatalf("got %q %v err=%v", tok, gotExp, err)
	}
}

func TestMemoryStore_Authorizer(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, err := s.GetAuthorizer(ctx, "suite1", "corp1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	tokens := &AuthorizerTokens{
		CorpID:            "corp1",
		PermanentCode:     "PCODE",
		CorpAccessToken:   "CTOK",
		CorpTokenExpireAt: time.Now().Add(time.Hour),
	}
	if err := s.PutAuthorizer(ctx, "suite1", "corp1", tokens); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetAuthorizer(ctx, "suite1", "corp1")
	if err != nil {
		t.Fatal(err)
	}
	if got.PermanentCode != "PCODE" || got.CorpAccessToken != "CTOK" {
		t.Fatalf("got %+v", got)
	}
}

func TestMemoryStore_ListAuthorizers(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	_ = s.PutAuthorizer(ctx, "suite1", "corpA", &AuthorizerTokens{CorpID: "corpA"})
	_ = s.PutAuthorizer(ctx, "suite1", "corpB", &AuthorizerTokens{CorpID: "corpB"})
	_ = s.PutAuthorizer(ctx, "suite2", "corpC", &AuthorizerTokens{CorpID: "corpC"})

	list, err := s.ListAuthorizers(ctx, "suite1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 corps, got %v", list)
	}
}

func TestMemoryStore_DeleteAuthorizer(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	_ = s.PutAuthorizer(ctx, "suite1", "corpA", &AuthorizerTokens{CorpID: "corpA"})
	if err := s.DeleteAuthorizer(ctx, "suite1", "corpA"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetAuthorizer(ctx, "suite1", "corpA"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_Concurrent(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			_ = s.PutSuiteTicket(ctx, "suite1", "T")
			_, _ = s.GetSuiteTicket(ctx, "suite1")
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
