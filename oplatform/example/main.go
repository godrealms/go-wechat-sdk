// Example: running a WeChat Open Platform third-party component callback
// server plus a QR login handler. This file must compile; it is not
// exercised in CI. Replace ComponentAppID / AppSecret / Token / AESKey
// with real values before running.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	mini_program "github.com/godrealms/go-wechat-sdk/mini-program"
	"github.com/godrealms/go-wechat-sdk/offiaccount"
	"github.com/godrealms/go-wechat-sdk/oplatform"
)

func main() {
	op, err := oplatform.NewClient(oplatform.Config{
		ComponentAppID:     "wxcompXXXX",
		ComponentAppSecret: "componentsecretXXXX",
		Token:              "callbacktoken",
		EncodingAESKey:     "0123456789ABCDEF0123456789ABCDEF0123456789A", // 43 chars
	})
	if err != nil {
		log.Fatal(err)
	}

	// 1) Component callback endpoint — receives verify_ticket pushes
	//    and authorization events. SDK auto-persists verify_ticket
	//    into the Store.
	http.HandleFunc("/oplatform/callback", func(w http.ResponseWriter, r *http.Request) {
		notify, err := op.ParseNotify(r, nil)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		switch notify.InfoType {
		case "component_verify_ticket":
			// already persisted by the SDK
			log.Printf("verify_ticket refreshed")
		case "authorized":
			if _, err := op.QueryAuth(r.Context(), notify.AuthorizationCode); err != nil {
				log.Printf("QueryAuth failed: %v", err)
			}
		case "updateauthorized":
			log.Printf("authorizer %s updated", notify.AuthorizerAppID)
		case "unauthorized":
			_ = op.Store().DeleteAuthorizer(r.Context(), notify.AuthorizerAppID)
		}
		_, _ = w.Write([]byte("success"))
	})

	// 2) Delegated offiaccount call — token flows through oplatform
	http.HandleFunc("/call/menu/", func(w http.ResponseWriter, r *http.Request) {
		appid := r.URL.Query().Get("appid")
		auth := op.Authorizer(appid)
		off := auth.OffiaccountClient()
		tok, err := off.AccessTokenE(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "authorized token: %s (for %s)", tok, appid)
	})

	// 3) Delegated mini-program call
	http.HandleFunc("/mp/token/", func(w http.ResponseWriter, r *http.Request) {
		appid := r.URL.Query().Get("appid")
		auth := op.Authorizer(appid)
		mp, err := auth.MiniProgramClient()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		tok, err := mp.AccessToken(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "mp token: %s", tok)
	})

	// 4) QR login flow (open-platform website login)
	qr := oplatform.NewQRLoginClient("wxqrXXXX", "qrsecretXXXX")
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		u := qr.AuthorizeURL("https://example.com/login/cb", "snsapi_login", "csrf_token")
		http.Redirect(w, r, u, http.StatusFound)
	})
	http.HandleFunc("/login/cb", func(w http.ResponseWriter, r *http.Request) {
		tok, err := qr.Code2Token(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		info, err := qr.UserInfo(r.Context(), tok.AccessToken, tok.OpenID)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		fmt.Fprintf(w, "welcome %s (%s)", info.Nickname, info.UnionID)
	})

	// Periodic token warm-up
	go func() {
		if err := op.RefreshAll(context.Background()); err != nil {
			log.Printf("RefreshAll: %v", err)
		}
	}()

	// Silence unused-import warning for mini_program / offiaccount
	// if the reader strips this file of the HTTP handlers above.
	_ = mini_program.Config{}
	_ = offiaccount.Config{}

	log.Fatal(http.ListenAndServe(":8080", nil))
}
