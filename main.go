package main

import (
	"log"
	"net/http"
)

func main() {
	go startRIPListener(":520") // RIPv2の受信処理開始 (ポート520)
	go startRIPBroadcaster()    // RIPv2の送信処理開始

	// HTTP APIサーバーの起動
	http.HandleFunc("/routes", routesHandler)       // API経由の経路情報
	http.HandleFunc("/rip-routes", ripRoutesHandler) // RIP Updateからの経路情報

	log.Println("API server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
