package api

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/MyTeleProject2026/Slotopol-server/config"
	"github.com/schwarzlichtbezirk/go-disk-usage"
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

// ApiDiskUsage returns disk usage information.
func ApiDiskUsage(c *gin.Context) {
	var path = "."
	if Cfg.CfgPath != "" {
		path = Cfg.CfgPath
	}
	usage, err := diskusage.New(path)
	if err != nil {
		Ret500(c, 0, err)
		return
	}
	RetOk(c, gin.H{
		"path":      path,
		"total":     usage.Size(),
		"free":      usage.Free(),
		"available": usage.Available(),
		"used":      usage.Used(),
		"usage":     usage.Usage(),
	})
}
