package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "strings"
    "sync"
    "sync/atomic"
    "time"

    "github.com/gin-gonic/gin"
)

type OperationalLog struct {
    ID uint64 `xorm:"pk autoincr" json:"id"`
    Timestamp time.Time `xorm:"created index notnull" json:"timestamp"`
    Level string `xorm:"varchar(12) index notnull" json:"level"`
    Service string `xorm:"varchar(48) index notnull" json:"service"`
    EventType string `xorm:"varchar(96) index notnull" json:"event_type"`
    RequestID string `xorm:"varchar(128) index" json:"request_id,omitempty"`
    UID uint64 `xorm:"index" json:"uid,omitempty"`
    CID uint64 `xorm:"index" json:"club_id,omitempty"`
    GameID string `xorm:"varchar(128) index" json:"game_id,omitempty"`
    Method string `xorm:"varchar(12)" json:"method,omitempty"`
    Endpoint string `xorm:"varchar(255) index" json:"endpoint,omitempty"`
    Status int `xorm:"index" json:"status,omitempty"`
    DurationMS int64 `json:"duration_ms,omitempty"`
    RemoteIP string `xorm:"varchar(64)" json:"remote_ip,omitempty"`
    Message string `xorm:"text" json:"message,omitempty"`
    Error string `xorm:"text" json:"error,omitempty"`
    Metadata string `xorm:"text" json:"metadata,omitempty"`
}
func (OperationalLog) TableName() string { return "operational_logs" }

var operationHub = struct { sync.RWMutex; clients map[chan OperationalLog]struct{} }{clients: make(map[chan OperationalLog]struct{})}
var operationRequestSeq uint64
func ensureOperationalLogs() error { return XormStorage.Sync(new(OperationalLog)) }
func operationRequestID(c *gin.Context) string { if v:=strings.TrimSpace(c.GetHeader("X-Request-ID")); v!="" { return v }; return fmt.Sprintf("slotopol-%d-%d", time.Now().UnixNano(), atomic.AddUint64(&operationRequestSeq,1)) }
func publishOperationalLog(e OperationalLog) { operationHub.RLock(); defer operationHub.RUnlock(); for ch:=range operationHub.clients { select { case ch<-e: default: } } }
func writeOperationalLog(e OperationalLog) { if XormStorage!=nil { if ensureOperationalLogs()==nil { _,_=XormStorage.InsertOne(&e) } }; publishOperationalLog(e) }

// OperationalLogMiddleware records completed HTTP requests without storing secrets or bodies.
func OperationalLogMiddleware() gin.HandlerFunc { return func(c *gin.Context) {
    started:=time.Now(); rid:=operationRequestID(c); c.Header("X-Request-ID",rid); c.Set("slotopol.request_id",rid); c.Next()
    status:=c.Writer.Status(); level,eventType:="INFO","api.request"; if status>=500 { level,eventType="ERROR","api.error" } else if status>=400 { level,eventType="WARN","api.error" }
    msg:=http.StatusText(status); if msg=="" { msg="HTTP response" }
    e:=OperationalLog{Timestamp:time.Now().UTC(),Level:level,Service:"slotopol-server",EventType:eventType,RequestID:rid,Method:c.Request.Method,Endpoint:c.FullPath(),Status:status,DurationMS:time.Since(started).Milliseconds(),RemoteIP:c.ClientIP(),Message:msg}
    if e.Endpoint=="" { e.Endpoint=c.Request.URL.Path }; if status>=400 { e.Error=msg }
    if u,ok:=c.Get(userKey); ok { if user,ok:=u.(*User); ok && user!=nil { e.UID=user.UID } }
    if cid:=c.Query("cid"); cid!="" { if n,err:=strconv.ParseUint(cid,10,64); err==nil { e.CID=n } }
    writeOperationalLog(e)
} }

func ApiAdminOperationsSummary(c *gin.Context) { _,al:=GetAdmin(c,0); if al&ALadmin==0 { Ret403(c,0,ErrNoAccess); return }; if err:=ensureOperationalLogs(); err!=nil { Ret500(c,0,err); return }
    total,_:=XormStorage.Count(new(OperationalLog)); errors,_:=XormStorage.Where("level=?","ERROR").Count(new(OperationalLog)); warnings,_:=XormStorage.Where("level=?","WARN").Count(new(OperationalLog)); var latest OperationalLog; has,_:=XormStorage.Desc("id").Get(&latest); var last any; if has { last=latest }; RetOk(c,gin.H{"total":total,"errors":errors,"warnings":warnings,"latest":last}) }
func ApiAdminOperationsLogs(c *gin.Context) { _,al:=GetAdmin(c,0); if al&ALadmin==0 { Ret403(c,0,ErrNoAccess); return }; if err:=ensureOperationalLogs(); err!=nil { Ret500(c,0,err); return }; limit:=100; if v,err:=strconv.Atoi(c.Query("limit")); err==nil && v>0 && v<=500 { limit=v }; q:=XormStorage.Desc("id"); if level:=strings.TrimSpace(c.Query("level")); level!="" { q=q.Where("level=?",strings.ToUpper(level)) }; if et:=strings.TrimSpace(c.Query("event_type")); et!="" { q=q.Where("event_type=?",et) }; var rows []OperationalLog; if err:=q.Limit(limit).Find(&rows); err!=nil { Ret500(c,0,err); return }; RetOk(c,gin.H{"events":rows}) }
func ApiAdminOperationsStream(c *gin.Context) { _,al:=GetAdmin(c,0); if al&ALadmin==0 { Ret403(c,0,ErrNoAccess); return }; if err:=ensureOperationalLogs(); err!=nil { Ret500(c,0,err); return }; c.Header("Content-Type","text/event-stream"); c.Header("Cache-Control","no-cache"); c.Header("Connection","keep-alive"); c.Header("X-Accel-Buffering","no")
    ch:=make(chan OperationalLog,32); operationHub.Lock(); operationHub.clients[ch]=struct{}{}; operationHub.Unlock(); defer func(){ operationHub.Lock(); delete(operationHub.clients,ch); close(ch); operationHub.Unlock() }(); flusher,ok:=c.Writer.(http.Flusher); if !ok { Ret500(c,0,fmt.Errorf("streaming is not supported by this server")); return }; _,_=c.Writer.WriteString("event: ready\ndata: {\"service\":\"slotopol-server\"}\n\n"); flusher.Flush(); ticker:=time.NewTicker(15*time.Second); defer ticker.Stop(); for { select { case <-c.Request.Context().Done(): return; case e:=<-ch: data,_:=json.Marshal(e); _,_=fmt.Fprintf(c.Writer,"id: %d\nevent: log\ndata: %s\n\n",e.ID,data); flusher.Flush(); case <-ticker.C: _,_=c.Writer.WriteString(": heartbeat\n\n"); flusher.Flush() } } }
