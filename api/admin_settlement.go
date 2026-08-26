package api

import (
    "errors"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
)

// ApiAdminSettlementRecent exposes real spin settlement records to the
// privileged admin console. It is read-only: settlement outcomes are never
// modified through this endpoint.
func ApiAdminSettlementRecent(c *gin.Context) {
    admin, al := MustAdmin(c, 0)
    if admin == nil || al&ALadmin == 0 {
        Ret403(c, 0, ErrNoAccess)
        return
    }

    limit := 100
    if raw := c.Query("limit"); raw != "" {
        parsed, err := strconv.Atoi(raw)
        if err != nil || parsed < 1 || parsed > 500 {
            Ret400(c, 0, errors.New("limit must be between 1 and 500"))
            return
        }
        limit = parsed
    }

    query := XormSpinlog.Desc("ctime")
    if gid := c.Query("gid"); gid != "" {
        query = query.Where("gid=?", gid)
    }
    if after := c.Query("after"); after != "" {
        t, err := time.Parse(time.RFC3339, after)
        if err != nil {
            Ret400(c, 0, errors.New("after must be RFC3339"))
            return
        }
        query = query.Where("ctime>=?", t)
    }

    rows := make([]Spinlog, 0, limit)
    if err := query.Limit(limit).Find(&rows); err != nil {
        Ret500(c, 0, err)
        return
    }

    var totalGain float64
    var wins uint64
    for _, row := range rows {
        totalGain += row.Gain
        if row.Gain > 0 {
            wins++
        }
    }

    recordAdminAudit(c, 0, "settlement.view", "spinlog", "read-only settlement records")
    RetOk(c, gin.H{
        "records": rows,
        "count": len(rows),
        "wins": wins,
        "total_gain": totalGain,
    })
}
