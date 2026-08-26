package api

import (
    "encoding/xml"
    "fmt"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
)

// ClubCurrencyBalance is provider credit only; it is not a customer wallet.
type ClubCurrencyBalance struct { ClubID uint64 `xorm:"pk 'club_id'" json:"club_id"`; Currency string `xorm:"pk 'currency'" json:"currency"`; Balance float64 `xorm:"notnull default 0" json:"balance"`; UpdatedAt time.Time `xorm:"updated" json:"updated_at"` }
func (ClubCurrencyBalance) TableName() string { return "club_currency_balances" }
func ensureClubCurrencyBalances() error { return XormStorage.Sync(new(ClubCurrencyBalance)) }

func ApiClubCurrencyBalanceGet(c *gin.Context) { var a struct{ CID uint64 `form:"cid" binding:"required"`; Currency string `form:"currency"` }; if err:=c.ShouldBindQuery(&a); err!=nil { Ret400(c,0,err); return }; if err:=ensureClubCurrencyBalances(); err!=nil { Ret500(c,0,err); return }; q:=XormStorage.Where("club_id=?",a.CID); if a.Currency!="" { q=q.And("currency=?",strings.ToUpper(strings.TrimSpace(a.Currency))) }; var rows []ClubCurrencyBalance; if err:=q.Find(&rows); err!=nil { Ret500(c,0,err); return }; RetOk(c,gin.H{"balances":rows}) }

func creditClubCurrencyBalance(session *Session, cid uint64, currency string, amount float64) error { currency=strings.ToUpper(strings.TrimSpace(currency)); if currency=="" { return nil }; row:=ClubCurrencyBalance{ClubID:cid,Currency:currency}; has,err:=session.Where("club_id=? AND currency=?",cid,currency).Get(&row); if err!=nil{return err}; if has { _,err=session.Exec("UPDATE club_currency_balances SET balance=balance+?, updated_at=? WHERE club_id=? AND currency=?",amount,time.Now(),cid,currency); return err }; row.Balance=amount; _,err=session.InsertOne(&row); return err }

func ApiClubCurrencyBalanceAdjust(c *gin.Context) { var a struct{ XMLName xml.Name `json:"-"`; ClubID uint64 `json:"club_id" binding:"required"`; Currency string `json:"currency" binding:"required"`; Amount float64 `json:"amount" binding:"required"` }; if err:=c.ShouldBind(&a); err!=nil { Ret400(c,0,err); return }; _,al:=GetAdmin(c,0); if al&ALmaster==0 { Ret403(c,0,ErrNoAccess); return }; if _,ok:=Clubs.Get(a.ClubID); !ok { Ret404(c,0,ErrNoClub); return }; if err:=ensureClubCurrencyBalances(); err!=nil { Ret500(c,0,err); return }; if a.Amount==0 { Ret400(c,0,ErrNoAccess); return }; a.Currency=strings.ToUpper(strings.TrimSpace(a.Currency)); s:=XormStorage.NewSession(); defer s.Close(); if err:=s.Begin(); err!=nil { Ret500(c,0,err); return }; if err:=creditClubCurrencyBalance(s,a.ClubID,a.Currency,a.Amount); err!=nil { _=s.Rollback(); Ret500(c,0,err); return }; if err:=s.Commit(); err!=nil { Ret500(c,0,err); return }; recordAdminAudit(c,a.ClubID,"club.currency-balance.adjust","club_currency_balance",fmt.Sprintf("currency=%s amount=%g",a.Currency,a.Amount)); Ret204(c) }
