# 基于Golang的Sherpa-Onnx语音服务

一个基于WebSocket的语音服务，使用Go语言开发，集成sherpa-onnx引擎。目前支持语音识别(ASR)功能，未来计划支持语音合成(TTS)。

## 功能概览

### 已实现功能
- ✅ 语音识别 (ASR)
    - 基于WebSocket的实时音频流处理
    - 集成Silero VAD进行语音活动检测
    - 支持PCM音频实时转写
    - 低延迟处理

### 开发计划
- 🚧 语音合成 (TTS) - 未开始

## 环境要求

- 依赖模型文件:
    - ASR相关：
        - `model.onnx`: 语音识别模型
        - `tokens.txt`: 词表文件
        - `silero_vad.onnx`: 语音活动检测模型

## 快速开始

目前提供多个平台的二进制文件，开箱即用。
