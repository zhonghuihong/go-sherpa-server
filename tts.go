package main

import (
	"bytes"
	"encoding/binary"
	"github.com/gorilla/websocket"
	sherpa "go-sherpa-server/sherpa/sherpa-onnx-go/sherpa_onnx"
	"log"
	"math"
	"net/http"
	"os"
)

func Float32ToPCM16(samples []float32) []byte {
	buf := new(bytes.Buffer)
	for _, sample := range samples {
		// Clamp & convert float32 (-1.0 ~ 1.0) to int16 (-32768 ~ 32767)
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}
		i16 := int16(sample * math.MaxInt16)
		binary.Write(buf, binary.LittleEndian, i16)
	}
	return buf.Bytes()
}

func PCMToWavBytes(pcm []byte, sampleRate int, numChannels int, bitsPerSample int) []byte {
	blockAlign := numChannels * bitsPerSample / 8
	byteRate := sampleRate * blockAlign
	subchunk2Size := len(pcm)
	chunkSize := 36 + subchunk2Size

	buf := new(bytes.Buffer)

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(chunkSize))
	buf.WriteString("WAVE")

	// fmt subchunk
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))            // Subchunk1Size
	binary.Write(buf, binary.LittleEndian, uint16(1))             // AudioFormat (1 = PCM)
	binary.Write(buf, binary.LittleEndian, uint16(numChannels))   // NumChannels
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))    // SampleRate
	binary.Write(buf, binary.LittleEndian, uint32(byteRate))      // ByteRate
	binary.Write(buf, binary.LittleEndian, uint16(blockAlign))    // BlockAlign
	binary.Write(buf, binary.LittleEndian, uint16(bitsPerSample)) // BitsPerSample

	// data subchunk
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(subchunk2Size))
	buf.Write(pcm)

	return buf.Bytes()
}

func Float32ToWav(samples []float32, sampleRate int) []byte {
	pcm := Float32ToPCM16(samples)
	return PCMToWavBytes(pcm, sampleRate, 1, 16)
}

var tts *sherpa.OfflineTts

func ResampleLinearFloat32(input []float32, inRate, outRate int) []float32 {
	ratio := float64(outRate) / float64(inRate)
	outLen := int(float64(len(input)) * ratio)
	output := make([]float32, outLen)

	for i := 0; i < outLen; i++ {
		pos := float64(i) / ratio
		index := int(pos)
		if index >= len(input)-1 {
			output[i] = input[len(input)-1]
		} else {
			frac := float32(pos - float64(index))
			output[i] = input[index]*(1-frac) + input[index+1]*frac
		}
	}
	return output
}

func initTTS() {
	ttsConfig := sherpa.OfflineTtsConfig{}
	_, err := os.Stat("./tts")
	if err != nil {
		return
	}
	ttsConfig.Model.Vits.Model = "./tts/model.onnx"
	ttsConfig.Model.Vits.Lexicon = "./tts/lexicon.txt"
	ttsConfig.Model.Vits.Tokens = "./tts/tokens.txt"
	ttsConfig.Model.Vits.DictDir = "./tts/dict"
	ttsConfig.Model.Vits.LengthScale = 1.0
	ttsConfig.Model.Vits.NoiseScale = 0.667
	ttsConfig.Model.Vits.NoiseScaleW = 0.8
	ttsConfig.RuleFsts = "./tts/number.fst"
	ttsConfig.Model.Provider = config.Provider
	tts = sherpa.NewOfflineTts(&ttsConfig)
}

func handleTTSWebSocket(w http.ResponseWriter, r *http.Request) {
	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("升级连接失败: %v", err)
		return
	}
	log.Printf("新的WebSocket连接建立")

	defer func() {
		conn.Close()
		log.Println("连接关闭")
	}()

	for {
		// 读取客户端发送的消息
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if messageType == websocket.TextMessage {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("捕获到 panic：", r)
					}
				}()
				generatedAudio := tts.Generate(string(p), config.TTS.SID, config.TTS.Speed)
				conn.WriteMessage(websocket.BinaryMessage, Float32ToWav(ResampleLinearFloat32(generatedAudio.Samples, generatedAudio.SampleRate, 24000), 16000))
			}()
		}
	}

}
