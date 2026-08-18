package api

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"xorm.io/xorm"

	"github.com/MyTeleProject2026/Slotopol-server/config"
)

type Session = xorm.Session

// XormStorage is the global database engine
var XormStorage *xorm.Engine

// "Server" field for HTTP headers.
var serverhdr = fmt.Sprintf("slotopol/%s (%s; %s)", config.BuildVers, runtime.GOOS, runtime.GOARCH)

var Offered = []string{
	binding.MIMEJSON,
	binding.MIMEXML,
	binding.MIMEYAML,
	binding.MIMETOML,
}

func Negotiate(c *gin.Context, code int, data any) {
	c.Writer.Header().Add("Server", serverhdr)
	switch c.NegotiateFormat(Offered...) {
	case binding.MIMEJSON:
		c.JSON(code, data)
	case binding.MIMEXML:
		c.XML(code, data)
	case binding.MIMEYAML
