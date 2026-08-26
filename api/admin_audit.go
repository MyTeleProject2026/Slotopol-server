package api

import (
    "time"

    "github.com/gin-gonic/gin"
)

// AdminAudit records privileged administrative mutations. It is intentionally
// append-only from the API: there is no public update/delete endpoint.
type AdminAudit struct {
    ID        uint64    `xorm:"pk autoincr" json:"id"`
    CTime     time.Time `xorm:"created 'ctime' notnull default CURRENT_TIMESTAMP" json:"ctime"`
    UID       uint64    `xorm:"notnull index" json:"uid"`
    CID       uint64    `xorm:"notnull index" json:"cid"`
    Action    string    `xorm:"varchar(96) notnull index" json:"action"`
    Resource  string    `xorm:"varchar(96) notnull" json:"resource"`
    RequestID string    `xorm:"varchar(128)" json:"request_id,omitempty"`
    RemoteIP  string    `xorm:"varchar(64)" json:"remote_ip,omitempty"`
    Details   string    `xorm:"text" json:"details,omitempty"`
}

func (AdminAudit) TableName() string { return "admin_audit" }

func ensureAdminAudit() error { return XormStorage.Sync(new(AdminAudit)) }

func recordAdminAudit(c *gin.Context, cid uint64, action, resource, details string) {
    admin, _ := GetAdmin(c, cid)
    if admin == nil { return }
    _ = ensureAdminAudit()
    requestID := c.GetHeader("X-Request-ID")
    _, _ = XormStorage.InsertOne(&AdminAudit{
        UID: admin.UID, CID: cid, Action: action, Resource: resource,
        RequestID: requestID, RemoteIP: c.ClientIP(), Details: details,
    })
}

func ApiAdminAuditList(c *gin.Context) {
    _, al := GetAdmin(c, 0)
    if al&ALadmin == 0 { Ret403(c, 0, ErrNoAccess); return }
    if err := ensureAdminAudit(); err != nil { Ret500(c, 0, err); return }
    var rows []AdminAudit
    q := XormStorage.Desc("id")
    if cid := c.Query("cid"); cid != "" { q = q.Where("cid=?", cid) }
    if action := c.Query("action"); action != "" { q = q.Where("action=?", action) }
    if err := q.Limit(250).Find(&rows); err != nil { Ret500(c, 0, err); return }
    RetOk(c, gin.H{"events": rows})
}
