package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

// Config 总配置结构
type Config struct {
	Server   ServerConfig `yaml:"server"`
	ASR      ASRConfig    `yaml:"asr"`
	TTS      TTSConfig    `yaml:"tts"`
	Provider string       `yaml:"provider"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	IP   string `yaml:"ip"`
	Port string `yaml:"port"`
}

// ASRConfig ASR相关配置
type ASRConfig struct {
	VAD VADConfig `yaml:"vad"`
}

type TTSConfig struct {
	SID   int     `yaml:"sid"`
	Speed float32 `yaml:"speed"`
}

// VADConfig VAD配置
type VADConfig struct {
	Threshold           float32 `yaml:"threshold"`
	MinSilenceDuration  float32 `yaml:"minSilenceDuration"`
	MinSpeechDuration   float32 `yaml:"minSpeechDuration"`
	WindowSize          int     `yaml:"windowSize"`
	MaxSpeechDuration   float32 `yaml:"maxSpeechDuration"`
	BufferSizeInSeconds float32 `yaml:"bufferSizeInSeconds"`
}

// LoadConfig 从文件中加载配置
func LoadConfig(filename string) (*Config, error) {
	config := &Config{}

	// 读取配置文件
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	// 解析YAML
	err = yaml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}

	return config, nil
}
