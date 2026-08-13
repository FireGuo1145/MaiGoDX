package middleware

import (
	"bytes"
	"compress/zlib"
	"io"
	"net/http"
)

// CompressionMiddleware 处理 maimai 客户端与服务端的 zlib 压缩/解压
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 如果请求带有 deflate 编码，进行解压
		if r.Header.Get("Content-Encoding") == "deflate" {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil && len(bodyBytes) > 0 {
				zReader, err := zlib.NewReader(bytes.NewReader(bodyBytes))
				if err == nil {
					defer zReader.Close()
					decompressed, err := io.ReadAll(zReader)
					if err == nil {
						r.Body = io.NopCloser(bytes.NewReader(decompressed))
					}
				}
			}
		}

		// 2. 包装 ResponseWriter 以支持响应 zlib 压缩
		zw := &zlibResponseWriter{
			ResponseWriter: w,
			buf:            new(bytes.Buffer),
		}
		defer zw.Close()

		// 检查客户端是否支持 deflate
		if r.Header.Get("Accept-Encoding") == "deflate" {
			w.Header().Set("Content-Encoding", "deflate")
			next.ServeHTTP(zw, r)
			_ = zw.Flush()
			return
		}

		next.ServeHTTP(w, r)
	})
}

type zlibResponseWriter struct {
	http.ResponseWriter
	buf    *bytes.Buffer
	closed bool
}

func (w *zlibResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *zlibResponseWriter) Flush() error {
	if w.closed {
		return nil
	}
	w.closed = true

	// 写入 zlib 压缩数据
	zWriter := zlib.NewWriter(w.ResponseWriter)
	defer zWriter.Close()
	_, err := zWriter.Write(w.buf.Bytes())
	return err
}

func (w *zlibResponseWriter) Close() {
	_ = w.Flush()
}
