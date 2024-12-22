package main

import (
	"log"
	"net/http"
	"flag"
	"os"
        "gopkg.in/yaml.v2"
)

// RipConfig構造体の定義
type RipConfig struct {
	BindIf       string `yaml:"bindif"`
	ReceiveAddr  string `yaml:"receiveaddr"`
	ReceivePort  string `yaml:"receiveport"`
	SendAddrPort string `yaml:"sendaddrport"`
	RestApiPort  string `yaml:"restapiport"`
}

func main() {
	config := loadConfig()	
	//go startRIPService("lo0","224.0.0.9","520","224.0.0.9:520") // RIPv2の受信処理開始 (ポート520)
	go startRIPService(config.BindIf,config.ReceiveAddr,config.ReceivePort,config.SendAddrPort) // RIPv2の受信処理開始 (ポート520)
	//go startRIPBroadcaster("eth0","224.0.0.9:520")    // RIPv2の送信処理開始

	// HTTP APIサーバーの起動
	http.HandleFunc("/routes", routesHandler)       // API経由の経路情報
	http.HandleFunc("/rip-routes", ripRoutesHandler) // RIP Updateからの経路情報

	log.Printf("API server started on port %v",config.RestApiPort)
	log.Fatal(http.ListenAndServe(":"+config.RestApiPort, nil))
}

func loadConfig()(RipConfig) {
	// コマンドライン引数の定義
	configPath := flag.String("c", "", "設定ファイルのパス (YAML形式)")
	flag.Parse()

	// 引数が指定されていない場合の処理
	if *configPath == "" {
		log.Fatal("設定ファイルのパスを -c オプションで指定してください")
	}

	// YAMLファイルを開く
	file, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("ファイルを開けませんでした: %v", err)
	}
	defer file.Close()

	// YAMLをデコードする
	var config RipConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("YAMLのデコードに失敗しました: %v", err)
	}
	return config 
}
