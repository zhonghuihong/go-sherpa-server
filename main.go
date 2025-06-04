package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var config *Config

var upgrader = websocket.Upgrader{
	// 允许所有CORS请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	config, _ = LoadConfig("config.yml")
	initVadConfig()
	initRecognizer()
	initTTS()
	http.HandleFunc("/asr", handleASRWebSocket)
	http.HandleFunc("/tts", handleTTSWebSocket)
	// 启动服务器
	port := config.Server.IP + ":" + config.Server.Port
	log.Printf("WebSocket服务器启动在端口%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}
