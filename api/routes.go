package api

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"xorm.io/xorm"

	"github.com/MyTeleProject2026/Slotopol-server/config"
)

type Session = xorm.Session

var XormStorage *xorm.Engine
var XormSpinlog *xorm.Engine

var serverhdr = fmt.Sprintf("slotopol/%s (%s; %s)", config.BuildVers, runtime.GOOS, runtime.GOARCH)

var Offered = []string{binding.MIMEJSON, binding.MIMEXML, binding.MIMEYAML, binding.MIMETOML}

func Negotiate(c *gin.Context, code int, data any) {
	c.Writer.Header().Add("Server", serverhdr)
	switch c.NegotiateFormat(Offered...) {
	case binding.MIMEJSON: c.JSON(code, data)
	case binding.MIMEXML: c.XML(code, data)
	case binding.MIMEYAML: c.YAML(code, data)
	case binding.MIMETOML: c.TOML(code, data)
	default: c.JSON(code, data)
	}
	c.Abort()
}
func RetOk(c *gin.Context, data any) { Negotiate(c, http.StatusOK, data) }
func Ret204(c *gin.Context) { c.Writer.Header().Add("Server", serverhdr); c.Status(http.StatusNoContent) }

type jerr struct { error }
func (err jerr) Unwrap() error { return err.error }
func (err jerr) MarshalJSON() ([]byte, error) { return json.Marshal(err.Error()) }
func (err jerr) MarshalYAML() (any, error) { return err.Error(), nil }
func (err jerr) MarshalXML(e *xml.Encoder, start xml.StartElement) error { return e.EncodeElement(err.Error(), start) }

type ajaxerr struct {
	XMLName xml.Name `json:"-" yaml:"-" xml:"error"`
	What jerr `json:"what" yaml:"what" xml:"what"`
	Code int `json:"code,omitempty" yaml:"code,omitempty" xml:"code,omitempty"`
	UID uint64 `json:"uid,omitempty" yaml:"uid,omitempty,attr"`
}
func (err ajaxerr) Error() string { return fmt.Sprintf("what: %s, code: %d", err.What, err.Code) }
func (err ajaxerr) Unwrap() error { return err.What.error }
func RetErr(c *gin.Context, status, code int, err error) { var uid uint64; if uv, ok := c.Get(userKey); ok { uid = uv.(*User).UID }; Negotiate(c, status, ajaxerr{What:jerr{err}, Code:code, UID:uid}) }
func Ret400(c *gin.Context, code int, err error) { RetErr(c, http.StatusBadRequest, code, err) }
func Ret401(c *gin.Context, code int, err error) { c.Writer.Header().Add("WWW-Authenticate", realmBasic); c.Writer.Header().Add("WWW-Authenticate", realmBearer); RetErr(c, http.StatusUnauthorized, code, err) }
func Ret403(c *gin.Context, code int, err error) { RetErr(c, http.StatusForbidden, code, err) }
func Ret404(c *gin.Context, code int, err error) { RetErr(c, http.StatusNotFound, code, err) }
func Ret500(c *gin.Context, code int, err error) { RetErr(c, http.StatusInternalServerError, code, err) }

func CorsMiddleware(allowedOrigin string) gin.HandlerFunc { return func(c *gin.Context) { if allowedOrigin != "" { c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin) }; c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS"); c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization"); if c.Request.Method == "OPTIONS" { c.AbortWithStatus(204); return }; c.Next() } }

func SetupRouter(r *gin.Engine) {
	adminOrigin := os.Getenv("SLOTOPOL_ADMIN_ORIGIN")
	r.Use(CorsMiddleware(adminOrigin)); r.NoRoute(Handle404); r.NoMethod(Handle405)
	r.Any("/ping", ApiPing); r.GET("/servinfo", ApiServInfo); r.GET("/memusage", ApiMemUsage); r.GET("/diskusage", ApiDiskUsage)
	r.Any("/signis", ApiSignis); r.GET("/sendcode", ApiSendCode); r.GET("/activate", Auth(false), ApiActivate); r.POST("/signup", Auth(false), ApiSignup); r.POST("/signin", ApiSignin); r.Any("/refresh", Auth(true), ApiRefresh)
	r.GET("/game/algs", ApiGameAlgs); r.GET("/game/list", ApiGameList); r.GET("/game/capabilities/:id", ApiGameCapabilities)
	ra := r.Group("/", Auth(true))
	ra.GET("/club/games", ApiClubGameList)
	rg := ra.Group("/game"); rg.POST("/new", ApiGameNew); rg.POST("/join", ApiGameJoin); rg.POST("/info", ApiGameInfo); rg.POST("/rtp/get", ApiGameRtpGet)
	rs := ra.Group("/slot"); rs.POST("/bet/get", ApiSlotBetGet); rs.POST("/bet/set", ApiSlotBetSet); rs.POST("/sel/get", ApiSlotSelGet); rs.POST("/sel/set", ApiSlotSelSet); rs.POST("/mode/set", ApiSlotModeSet); rs.POST("/spin", ApiSlotSpin); rs.POST("/doubleup", ApiSlotDoubleup); rs.POST("/collect", ApiSlotCollect)
	rp := ra.Group("/prop"); rp.POST("/get", ApiPropsGet); rp.POST("/wallet/get", ApiPropsWalletGet); rp.POST("/wallet/add", ApiPropsWalletAdd); rp.POST("/al/get", ApiPropsAlGet); rp.POST("/al/set", ApiPropsAlSet); rp.POST("/rtp/get", ApiPropsRtpGet); rp.POST("/rtp/set", ApiPropsRtpSet)
	ru := ra.Group("/user"); ru.POST("/is", ApiUserIs); ru.POST("/phone", ApiUserPhone); ru.POST("/rename", ApiUserRename); ru.POST("/secret", ApiUserSecret); ru.POST("/delete", ApiUserDelete)
	rc := ra.Group("/club"); rc.POST("/list", ApiClubList); rc.POST("/is", ApiClubIs); rc.POST("/info", ApiClubInfo); rc.POST("/jpfund", ApiClubJpfund); rc.POST("/rename", ApiClubRename); rc.POST("/cashin", ApiClubCashin)
	rcloud := ra.Group("/cloudinary"); rcloud.POST("/upload", ApiUploadImage); rcloud.GET("/images", ApiGetImages); rcloud.DELETE("/image", ApiDeleteImage)
	ra.POST("/admin/allocate", ApiAllocationCreate); ra.POST("/admin/allocation/approve", ApiAllocationApprove); ra.GET("/admin/allocations", ApiAllocationList); ra.POST("/admin/game/permission", ApiGamePermissionSet); ra.POST("/admin/game/permissions/bulk", ApiGamePermissionsBulkSet)
}
