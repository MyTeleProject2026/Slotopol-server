package api

import (
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/MyTeleProject2026/Slotopol-server/config"
)

// ApiPing is a simple health check.
func ApiPing(c *gin.Context) {
	RetOk(c, gin.H{"ping": "pong"})
}

// ApiServInfo returns server version and build info.
func ApiServInfo(c *gin.Context) {
	RetOk(c, gin.H{
		"version":   config.BuildVers,
		"built":     config.BuildTime,
		"goos":      runtime.GOOS,
		"goarch":    runtime.GOARCH,
		"goversion": runtime.Version(),
	})
}

// ApiMemUsage returns memory statistics.
func ApiMemUsage(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	RetOk(c, gin.H{
		"alloc":       m.Alloc,
		"total_alloc": m.TotalAlloc,
		"sys":         m.Sys,
		"num_gc":      m.NumGC,
		"heap_alloc":  m.HeapAlloc,
		"heap_sys":    m.HeapSys,
		"heap_idle":   m.HeapIdle,
		"heap_inuse":  m.HeapInuse,
		"stack_inuse": m.StackInuse,
	})
}

// ApiDiskUsage returns disk usage information (simplified).
func ApiDiskUsage(c *gin.Context) {
	// Simplified – remove external dependency that caused build errors
	RetOk(c, gin.H{
		"message": "disk usage not available in this build",
	})
}
