//go:build android && arm64

package sherpa_onnx

// #cgo LDFLAGS: -L ${SRCDIR}/lib/arm64-v8a -lsherpa-onnx-c-api -lonnxruntime -Wl,-rpath,${SRCDIR}/lib/arm64-v8a
import "C"
