package main

import (
	"encoding/json"
	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"log"
)

func main() {
	content := `{
    "id": "d4dd679e-e8a1-52df-84da-c019051f783a",
    "create_time": "2025-06-17T19:02:43+08:00",
    "resource_type": "encrypt-resource",
    "event_type": "TRANSACTION.SUCCESS",
    "summary": "支付成功",
    "resource": {
        "original_type": "transaction",
        "algorithm": "AEAD_AES_256_GCM",
        "ciphertext": "0pkIz1d/W9jvylnXMUcsmlzplbyW/mVb/37TvK1BV4zBzdEicwblm9C7w5e52idaJ4x/w+kmEsoa7qshnpixdj9YdChzoU/cbNCF2XbfYk6HxGAP0CF2ksNthXZr7DT/kyX1NyFUji47C93s6hEvjteqpUmdQ1H4W9rAIv/d92lvlbnacHJIDXZNeDnQrcTVYAHvOm0tum2f3N2Dp20up2D4ok7Qm09qzD+cAzKng+E1EjcKyuNemrxGdp1OG2xwAC4h9qil4YCpFKCFVff6RpG0pldnwH3si08UQdXX4lqzUb9E7NIEGHk1up3Vcy565kNl/rhftTyTSCJ5LmQZ7XQsjvTuUhoJzHoktFokEOAMfUP+iOi2UCBRRFk0vW1raizvC/eqwLh2N3xSBFbQvPy8pgbXXc9D1ARVUBaRkWELzZT76RwGIk5zMBuYpcPTUekxZefyAtmiQr72VQC2AIIhTz0JeFk5g/g3JVGu3kKdSpylt+e/MLtxRWMWMbMS7tfNYnlIkQUWV9KKuJ+XlhhzeZacpeKJY2nFRu4rK59GS8zxrIUXFdOE1MKcJzpf+eXPtJjQ26aAzar2DaA=",
        "associated_data": "transaction",
        "nonce": "TGr5igc6g5Lt"
    }
}`
	notify := &types.Notify{}
	err := json.Unmarshal([]byte(content), &notify)
	if err != nil {
		log.Fatal(err)
	}
	transaction, err := notify.DecryptResource("APIv3Key")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(transaction)
}
