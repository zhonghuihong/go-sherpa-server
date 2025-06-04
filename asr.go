package main

import (
	"github.com/gorilla/websocket"
	sherpa "go-sherpa-server/sherpa/sherpa-onnx-go/sherpa_onnx"
	"log"
	"net/http"
	"strings"
	"time"
)

func bytesToNormalizedPCM(opusBytes []byte) []float32 {
	samples := make([]int16, len(opusBytes)/2)
	for i := 0; i < len(opusBytes); i += 2 {
		samples[i/2] = int16(opusBytes[i]) | int16(opusBytes[i+1])<<8
	}
	normalizedPCM := make([]float32, len(samples))
	for i, sample := range samples {
		normalizedPCM[i] = float32(sample) / 32768.0
	}
	return normalizedPCM
}

const VadChunkSize = 160

func handleASRWebSocket(w http.ResponseWriter, r *http.Request) {
	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("升级连接失败: %v", err)
		return
	}
	connectionAlive := true
	log.Printf("新的WebSocket连接建立")

	vad := sherpa.NewVoiceActivityDetector(&vadConfig, config.ASR.VAD.BufferSizeInSeconds)

	quitChan := make(chan struct{}, 1)
	defer func() {
		connectionAlive = false
		<-quitChan
		sherpa.DeleteVoiceActivityDetector(vad)
		conn.Close()
		log.Println("连接关闭")
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("捕获到 panic：", r)
			}
		}()
		for {
			if !connectionAlive {
				quitChan <- struct{}{}
				break
			}
			for !vad.IsEmpty() {
				speechSegment := vad.Front()
				vad.Pop()
				audio := &sherpa.Wave{}
				audio.Samples = speechSegment.Samples
				audio.SampleRate = vadConfig.SampleRate
				decode(recognizer, audio, conn)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	for {
		// 读取客户端发送的消息
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if messageType == websocket.BinaryMessage {
			normalizedData := bytesToNormalizedPCM(p)
			for i := 0; i < len(normalizedData); i += VadChunkSize {
				end := i + VadChunkSize
				if end > len(normalizedData) {
					end = len(normalizedData)
				}
				chunk := normalizedData[i:end]
				vad.AcceptWaveform(chunk)
			}
		}
	}

}

var recognizer *sherpa.OfflineRecognizer
var vadConfig = sherpa.VadModelConfig{}

func initRecognizer() {
	log.Println("开始加载语音识别模型")
	c := sherpa.OfflineRecognizerConfig{}
	c.FeatConfig.SampleRate = 16000
	c.FeatConfig.FeatureDim = 80
	c.ModelConfig.SenseVoice.Model = "./asr/model.onnx"
	c.ModelConfig.Tokens = "./asr/tokens.txt"
	c.ModelConfig.NumThreads = 2
	c.ModelConfig.Debug = 0
	c.ModelConfig.Provider = config.Provider
	recognizer = sherpa.NewOfflineRecognizer(&c)
	log.Println("语音识别模型加载完毕")
}

func initVadConfig() {
	vadConfig = sherpa.VadModelConfig{}
	vadConfig.SileroVad.Model = "./asr/silero_vad.onnx"
	vadConfig.SileroVad.Threshold = config.ASR.VAD.Threshold
	vadConfig.SileroVad.MinSilenceDuration = config.ASR.VAD.MinSilenceDuration
	vadConfig.SileroVad.MinSpeechDuration = config.ASR.VAD.MinSpeechDuration
	vadConfig.SileroVad.WindowSize = config.ASR.VAD.WindowSize
	vadConfig.SileroVad.MaxSpeechDuration = config.ASR.VAD.MaxSpeechDuration
	vadConfig.SampleRate = 16000
	vadConfig.NumThreads = 1
	vadConfig.Provider = config.Provider
	vadConfig.Debug = 0
}

func decode(recognizer *sherpa.OfflineRecognizer, audio *sherpa.Wave, conn *websocket.Conn) {
	stream := sherpa.NewOfflineStream(recognizer)
	defer sherpa.DeleteOfflineStream(stream)
	stream.AcceptWaveform(audio.SampleRate, audio.Samples)
	recognizer.Decode(stream)
	result := stream.GetResult()
	text := strings.ToLower(result.Text)
	text = strings.Trim(text, " ")
	conn.WriteMessage(websocket.TextMessage, []byte(text))
}
