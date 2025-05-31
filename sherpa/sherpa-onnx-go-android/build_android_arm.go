//go:build android && arm

package sherpa_onnx

// #cgo LDFLAGS: -L ${SRCDIR}/lib/armeabi-v7a -lsherpa-onnx-c-api -lonnxruntime -Wl,-rpath,${SRCDIR}/lib/armeabi-v7a
import "C"
