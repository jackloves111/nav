package router

import (
	"compress/gzip"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

// staticCacheMiddleware 为静态资源设置缓存策略：
// Vite 构建产物文件名带内容 hash，可以长期缓存；HTML 入口不缓存。
func staticCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		switch {
		case strings.HasPrefix(p, "/assets/"):
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		case strings.HasPrefix(p, "/presetIcons/"):
			c.Header("Cache-Control", "public, max-age=604800")
		case strings.HasPrefix(p, "/custom/"), strings.HasPrefix(p, "/uploads/"):
			c.Header("Cache-Control", "public, max-age=86400")
		case p == "/" || strings.HasSuffix(p, ".html"):
			c.Header("Cache-Control", "no-cache")
		case p == "/favicon.ico" || p == "/favicon.svg":
			c.Header("Cache-Control", "public, max-age=86400")
		}
		c.Next()
	}
}

// gzipStaticMiddleware 只压缩可压缩的静态文本资源。
// 图片、视频、字体等已压缩格式不参与，避免浪费 CPU。
func gzipStaticMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		if c.Request.Method != http.MethodGet ||
			c.Request.Header.Get("Range") != "" ||
			!strings.Contains(acceptEncoding, "gzip") ||
			strings.Contains(acceptEncoding, "gzip;q=0") ||
			!isCompressibleStaticPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		gz := gzipWriterPool.Get().(*gzip.Writer)
		gz.Reset(c.Writer)

		gw := &gzipResponseWriter{ResponseWriter: c.Writer, gz: gz}
		c.Writer = gw
		c.Header("Vary", "Accept-Encoding")

		defer func() {
			gw.close()
			gzipWriterPool.Put(gz)
		}()

		c.Next()
	}
}

type gzipResponseWriter struct {
	gin.ResponseWriter
	gz          *gzip.Writer
	headerSent  bool
	compressing bool
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.headerSent {
		return
	}
	w.headerSent = true

	// 只压缩正常的 200 响应；304/206/重定向等保持原样。
	if code != http.StatusOK || w.Header().Get("Content-Encoding") != "" {
		w.ResponseWriter.WriteHeader(code)
		return
	}

	// 太小的响应压缩后可能更大，直接跳过。
	if length, err := strconv.Atoi(w.Header().Get("Content-Length")); err == nil && length < 1024 {
		w.ResponseWriter.WriteHeader(code)
		return
	}

	w.Header().Del("Content-Length")
	w.Header().Set("Content-Encoding", "gzip")
	w.compressing = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) WriteHeaderNow() {
	if !w.headerSent {
		w.WriteHeader(http.StatusOK)
	}
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.headerSent {
		w.WriteHeader(http.StatusOK)
	}
	if !w.compressing {
		return w.ResponseWriter.Write(b)
	}
	return w.gz.Write(b)
}

func (w *gzipResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *gzipResponseWriter) close() {
	if w.compressing {
		_ = w.gz.Close()
	}
}

func isCompressibleStaticPath(p string) bool {
	if p == "/" || strings.HasSuffix(p, "/") {
		p += "index.html"
	}
	switch strings.ToLower(path.Ext(p)) {
	case ".js", ".mjs", ".css", ".html", ".json", ".svg", ".txt", ".xml", ".map", ".webmanifest":
		return true
	default:
		return false
	}
}
