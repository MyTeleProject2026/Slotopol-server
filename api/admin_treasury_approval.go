package api

import (
    "fmt"
    "math"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
)

// TreasuryApproval is a maker-checker request for a high-impact currency transfer.
type TreasuryApproval struct {
    ID uint64 `xorm:"pk autoincr" json:"id"`
    CTime time.Time `xorm:"created 'ctime' notnull default CURRENT_TIMESTAMP" json:"ctime"`
    UTime time.Time `xorm:"updated 'utime' notnull default CURRENT_TIMESTAMP" json:"utime"`
    RequestedBy uint64 `xorm:"notnull index" json:"requested_by"`
    ApprovedBy uint64 `xorm:"notnull default 0 index" json:"approved_by"`
    FromClubID uint64 `xorm:"notnull index" json:"from_club_id"`
    ToClubID uint64 `xorm:"notnull index" json:"to_club_id"`
    FromCurrency string `xorm:"varchar(12) notnull" json:"from_currency"`
    ToCurrency string `xorm:"varchar(12) notnull" json:"to_currency"`
    FromAmount float64 `xorm:"notnull" json:"from_amount"`
    Rate float64 `xorm:"notnull" json:"rate"`
    Reference string `xorm:"varchar(128) notnull" json:"reference"`
    Status string `xorm:"varchar(16) notnull index" json:"status"`
    DecisionNote string `xorm:"varchar(255) notnull default ''" json:"decision_note"`
    LedgerID uint64 `xorm:"notnull default 0" json:"ledger_id"`
}
func (TreasuryApproval) TableName() string { return "treasury_approvals" }
func ensureTreasuryApprovals() error { return XormStorage.Sync(new(TreasuryApproval)) }

func ApiAdminTreasuryTransferRequest(c *gin.Context) {
    admin, al := GetAdmin(c, 0)
    if admin == nil || al&ALmaster == 0 { Ret403(c,0,ErrNoAccess); return }
    var a TreasuryApproval
    if err := c.ShouldBindJSON(&a); err != nil { Ret400(c,0,err); return }
    a.FromCurrency=strings.ToUpper(strings.TrimSpace(a.FromCurrency)); a.ToCurrency=strings.ToUpper(strings.TrimSpace(a.ToCurrency)); a.Reference=strings.TrimSpace(a.Reference)
    if a.FromClubID==0 || a.ToClubID==0 || a.FromClubID==a.ToClubID || a.FromAmount<=0 || a.Rate<=0 || math.IsNaN(a.Rate) || math.IsInf(a.Rate,0) || len(a.FromCurrency)<3 || len(a.ToCurrency)<3 { Ret400(c,0,fmt.Errorf("invalid treasury request")); return }
    if _,ok:=Clubs.Get(a.FromClubID); !ok { Ret404(c,0,ErrNoClub); return }; if _,ok:=Clubs.Get(a.ToClubID); !ok { Ret404(c,0,ErrNoClub); return }
    if a.Reference=="" { a.Reference=fmt.Sprintf("approval-%d",time.Now().UnixNano()) }
    a.RequestedBy=admin.UID; a.Status="pending"; a.ApprovedBy=0; a.LedgerID=0
    if err:=ensureTreasuryApprovals(); err!=nil { Ret500(c,0,err); return }
    if _,err:=XormStorage.InsertOne(&a); err!=nil { Ret500(c,0,err); return }
    recordAdminAudit(c,a.FromClubID,"treasury.transfer.request","treasury_approvals",fmt.Sprintf("id=%d to_club=%d reference=%s",a.ID,a.ToClubID,a.Reference))
    RetOk(c,gin.H{"approval":a})
}

func ApiAdminTreasuryApprovalList(c *gin.Context) {
    _,al:=GetAdmin(c,0); if al&ALadmin==0 { Ret403(c,0,ErrNoAccess); return }
    if err:=ensureTreasuryApprovals(); err!=nil { Ret500(c,0,err); return }
    var rows []TreasuryApproval; q:=XormStorage.Desc("id"); if s:=strings.TrimSpace(c.Query("status")); s!="" { q=q.Where("status=?",s) }
    if err:=q.Limit(250).Find(&rows); err!=nil { Ret500(c,0,err); return }; RetOk(c,gin.H{"approvals":rows})
}

// executeApprovedTreasuryTransfer atomically moves provider credit and marks the
// approval executed in the same transaction, making retries unable to duplicate
// the transfer once status leaves pending.
func executeApprovedTreasuryTransfer(a *TreasuryApproval, approver uint64) error {
    if err:=ensureClubCurrencyBalances(); err!=nil { return err }
    if err:=ensureClubCurrencyLedger(); err!=nil { return err }
    toAmount:=a.FromAmount*a.Rate
    if toAmount<=0 || math.IsNaN(toAmount) || math.IsInf(toAmount,0) { return fmt.Errorf("invalid converted amount") }
    s:=XormStorage.NewSession(); defer s.Close()
    if err:=s.Begin(); err!=nil { return err }
    // Claim the request only if still pending. This is the idempotency guard.
    res,err:=s.Exec("UPDATE treasury_approvals SET status=?, approved_by=?, utime=? WHERE id=? AND status=?", "executing", approver, time.Now(), a.ID, "pending")
    if err!=nil { _=s.Rollback(); return err }
    n,err:=res.RowsAffected(); if err!=nil { _=s.Rollback(); return err }; if n!=1 { _=s.Rollback(); return fmt.Errorf("approval already processed") }
    if err:=debitClubCurrencyBalance(s,a.FromClubID,a.FromCurrency,a.FromAmount); err!=nil { _=s.Rollback(); return err }
    if err:=creditClubCurrencyBalance(s,a.ToClubID,a.ToCurrency,toAmount); err!=nil { _=s.Rollback(); return err }
    entry:=&ClubCurrencyLedger{UID:approver,FromClubID:a.FromClubID,ToClubID:a.ToClubID,FromCurrency:a.FromCurrency,ToCurrency:a.ToCurrency,FromAmount:a.FromAmount,ToAmount:toAmount,Rate:a.Rate,Reference:a.Reference}
    if _,err:=s.InsertOne(entry); err!=nil { _=s.Rollback(); return err }
    if _,err:=s.Exec("UPDATE treasury_approvals SET status=?, ledger_id=?, utime=? WHERE id=? AND status=?", "executed", entry.ID, time.Now(), a.ID, "executing"); err!=nil { _=s.Rollback(); return err }
    if err:=s.Commit(); err!=nil { return err }
    a.Status="executed"; a.ApprovedBy=approver; a.LedgerID=entry.ID; a.UTime=time.Now()
    return nil
}

func ApiAdminTreasuryApprovalDecision(c *gin.Context) {
    admin,al:=GetAdmin(c,0); if admin==nil || al&ALadmin==0 { Ret403(c,0,ErrNoAccess); return }
    var arg struct { ID uint64 `json:"id" binding:"required"`; Approve bool `json:"approve"`; Note string `json:"note"` }
    if err:=c.ShouldBindJSON(&arg); err!=nil { Ret400(c,0,err); return }; if err:=ensureTreasuryApprovals(); err!=nil { Ret500(c,0,err); return }
    var a TreasuryApproval; ok,err:=XormStorage.ID(arg.ID).Get(&a); if err!=nil { Ret500(c,0,err); return }; if !ok { Ret404(c,0,fmt.Errorf("approval not found")); return }
    if a.Status!="pending" { Ret400(c,0,fmt.Errorf("approval is not pending")); return }; if a.RequestedBy==admin.UID { Ret403(c,0,fmt.Errorf("requester cannot approve own transfer")); return }
    a.ApprovedBy=admin.UID; a.DecisionNote=strings.TrimSpace(arg.Note)
    if !arg.Approve { a.Status="rejected"; a.UTime=time.Now(); if _,err:=XormStorage.ID(a.ID).Cols("approved_by","decision_note","status","utime").Update(&a); err!=nil { Ret500(c,0,err); return }; recordAdminAudit(c,a.FromClubID,"treasury.transfer.reject","treasury_approvals",fmt.Sprintf("id=%d",a.ID)); RetOk(c,gin.H{"approval":a}); return }
    if err:=executeApprovedTreasuryTransfer(&a,admin.UID); err!=nil { Ret400(c,0,err); return }
    if _,err:=XormStorage.ID(a.ID).Cols("decision_note").Update(&a); err!=nil { Ret500(c,0,err); return }
    recordAdminAudit(c,a.FromClubID,"treasury.transfer.execute-approved","treasury_approvals",fmt.Sprintf("id=%d ledger_id=%d reference=%s",a.ID,a.LedgerID,a.Reference))
    RetOk(c,gin.H{"approval":a})
}
