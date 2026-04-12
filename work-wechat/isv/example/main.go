//go:build ignore
// +build ignore

// Package main is a compile-only demo for work-wechat/isv. It demonstrates the
// shape of a typical ISV service provider integration. It does not actually
// reach the network; all calls are commented-out or behind a bogus config.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/godrealms/go-wechat-sdk/work-wechat/isv"
)

func main() {
	cfg := isv.Config{
		SuiteID:        "your_suite_id",
		SuiteSecret:    "your_suite_secret",
		ProviderCorpID: "your_provider_corpid",
		ProviderSecret: "your_provider_secret",
		Token:          "your_callback_token",
		EncodingAESKey: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ",
	}

	client, err := isv.NewClient(cfg,
		isv.WithStore(isv.NewMemoryStore()),
	)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/wecom/callback", func(w http.ResponseWriter, r *http.Request) {
		ev, err := client.ParseNotify(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch e := ev.(type) {
		case *isv.SuiteTicketEvent:
			log.Printf("suite_ticket persisted: %s", e.SuiteTicket)
		case *isv.CreateAuthEvent:
			log.Printf("new auth, auth_code=%s", e.AuthCode)
			resp, err := client.GetPermanentCode(r.Context(), e.AuthCode)
			if err != nil {
				log.Printf("get permanent code: %v", err)
				return
			}
			log.Printf("corp %s authorized", resp.AuthCorpInfo.CorpName)
		case *isv.CancelAuthEvent:
			log.Printf("cancel auth corp=%s", e.AuthCorpID)
		default:
			log.Printf("event: %T", ev)
		}
		_, _ = w.Write([]byte("success"))
	})

	// Below code is commented out; demonstrates usage only.
	_ = func(ctx context.Context) {
		preAuth, _ := client.GetPreAuthCode(ctx)
		url := client.AuthorizeURL(preAuth.PreAuthCode, "https://your.callback/ret", "state")
		fmt.Println(url)

		corpClient := client.CorpClient("wxcorp1")
		_, _ = corpClient.AccessToken(ctx)
		_ = client.RefreshAll(ctx)
	}

	log.Println("demo wired; run as production service to enable")
}
