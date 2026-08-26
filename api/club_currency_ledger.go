package api

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ClubCurrencyLedger is an immutable financial movement record for provider
// credit between clubs. It stores both sides of an FX movement so the
// original rate and currencies remain auditable after balances change.
type ClubCurrencyLedger struct {
	ID           uint64    `xorm:"pk autoincr" json:"id"`
	CTime        time.Time `xorm:"created 'ctime' notnull default CURRENT_TIMESTAMP" json:"ctime"`
	UID          uint64    `xorm:"notnull index" json:"uid"`
	FromClubID   uint64    `xorm:"notnull index" json:"from_club_id"`
	ToClubID     uint64    `xorm:"notnull index" json:"to_club_id"`
	FromCurrency string    `xorm:"varchar(12) notnull" json:"from_currency"`
	ToCurrency   string    `xorm:"varchar(12) notnull" json:"to_currency"`
	FromAmount   float64   `xorm:"notnull" json:"from_amount"`
	ToAmount     float64   `xorm:"notnull" json:"to_amount"`
	Rate         float64   `xorm:"notnull" json:"rate"`
	Reference    string    `xorm:"varchar(128) notnull" json:"reference"`
}

func (ClubCurrencyLedger) TableName() string { return "club_currency_ledger" }

func ensureClubCurrencyLedger() error { return XormStorage.Sync(new(ClubCurrencyLedger)) }

func debitClubCurrencyBalance(session *Session, cid uint64, currency string, amount float64) error {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if cid == 0 || currency == "" || amount <= 0 {
		return fmt.Errorf("invalid currency debit")
	}
	result, err := session.Exec(
		"UPDATE club_currency_balances SET balance=balance-?, updated_at=? WHERE club_id=? AND currency=? AND balance>=?",
		amount, time.Now(), cid, currency, amount,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("insufficient club currency balance")
	}
	return nil
}

// ApiAdminClubCurrencyTransfer performs an atomic provider-credit transfer.
// The operation is restricted to ALmaster and always records an immutable
// ledger entry plus an admin audit event. FX is explicit: to_currency may
// differ only when a positive caller-supplied rate is supplied.
func ApiAdminClubCurrencyTransfer(c *gin.Context) {
	admin, al := GetAdmin(c, 0)
	if admin == nil || al&ALmaster == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}
	var arg struct {
		FromClubID   uint64  `json:"from_club_id" binding:"required"`
		ToClubID     uint64  `json:"to_club_id" binding:"required"`
		FromCurrency string  `json:"from_currency" binding:"required"`
		ToCurrency   string  `json:"to_currency" binding:"required"`
		FromAmount   float64 `json:"from_amount" binding:"required"`
		Rate         float64 `json:"rate" binding:"required"`
		Reference    string  `json:"reference"`
	}
	if err := c.ShouldBindJSON(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}
	if arg.FromClubID == arg.ToClubID || arg.FromAmount <= 0 || arg.Rate <= 0 || math.IsNaN(arg.Rate) || math.IsInf(arg.Rate, 0) {
		Ret400(c, 0, fmt.Errorf("invalid transfer parameters"))
		return
	}
	if _, ok := Clubs.Get(arg.FromClubID); !ok {
		Ret404(c, 0, ErrNoClub)
		return
	}
	if _, ok := Clubs.Get(arg.ToClubID); !ok {
		Ret404(c, 0, ErrNoClub)
		return
	}
	arg.FromCurrency = strings.ToUpper(strings.TrimSpace(arg.FromCurrency))
	arg.ToCurrency = strings.ToUpper(strings.TrimSpace(arg.ToCurrency))
	if len(arg.FromCurrency) < 3 || len(arg.FromCurrency) > 12 || len(arg.ToCurrency) < 3 || len(arg.ToCurrency) > 12 {
		Ret400(c, 0, fmt.Errorf("invalid currency code"))
		return
	}
	arg.Reference = strings.TrimSpace(arg.Reference)
	if arg.Reference == "" {
		arg.Reference = fmt.Sprintf("admin-transfer-%d", time.Now().UnixNano())
	}
	toAmount := arg.FromAmount * arg.Rate
	if toAmount <= 0 || math.IsInf(toAmount, 0) || math.IsNaN(toAmount) {
		Ret400(c, 0, fmt.Errorf("invalid converted amount"))
		return
	}
	if err := ensureClubCurrencyBalances(); err != nil {
		Ret500(c, 0, err)
		return
	}
	if err := ensureClubCurrencyLedger(); err != nil {
		Ret500(c, 0, err)
		return
	}

	s := XormStorage.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		Ret500(c, 0, err)
		return
	}
	if err := debitClubCurrencyBalance(s, arg.FromClubID, arg.FromCurrency, arg.FromAmount); err != nil {
		_ = s.Rollback()
		Ret400(c, 0, err)
		return
	}
	if err := creditClubCurrencyBalance(s, arg.ToClubID, arg.ToCurrency, toAmount); err != nil {
		_ = s.Rollback()
		Ret500(c, 0, err)
		return
	}
	entry := &ClubCurrencyLedger{
		UID: admin.UID, FromClubID: arg.FromClubID, ToClubID: arg.ToClubID,
		FromCurrency: arg.FromCurrency, ToCurrency: arg.ToCurrency,
		FromAmount: arg.FromAmount, ToAmount: toAmount, Rate: arg.Rate,
		Reference: arg.Reference,
	}
	if _, err := s.InsertOne(entry); err != nil {
		_ = s.Rollback()
		Ret500(c, 0, err)
		return
	}
	if err := s.Commit(); err != nil {
		Ret500(c, 0, err)
		return
	}

	recordAdminAudit(c, arg.FromClubID, "club.currency-transfer.execute", "club_currency_ledger",
		fmt.Sprintf("to_club=%d from=%s:%g to=%s:%g rate=%g reference=%s", arg.ToClubID, arg.FromCurrency, arg.FromAmount, arg.ToCurrency, toAmount, arg.Rate, arg.Reference))
	RetOk(c, gin.H{"success": true, "ledger_id": entry.ID, "from_amount": arg.FromAmount, "to_amount": toAmount, "rate": arg.Rate, "from_currency": arg.FromCurrency, "to_currency": arg.ToCurrency})
}

func ApiAdminClubCurrencyLedger(c *gin.Context) {
	_, al := GetAdmin(c, 0)
	if al&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}
	if err := ensureClubCurrencyLedger(); err != nil {
		Ret500(c, 0, err)
		return
	}
	limit := 250
	var rows []ClubCurrencyLedger
	q := XormStorage.Desc("id")
	if cid := c.Query("club_id"); cid != "" {
		q = q.Where("from_club_id=? OR to_club_id=?", cid, cid)
	}
	if currency := strings.ToUpper(strings.TrimSpace(c.Query("currency"))); currency != "" {
		q = q.Where("from_currency=? OR to_currency=?", currency, currency)
	}
	if err := q.Limit(limit).Find(&rows); err != nil {
		Ret500(c, 0, err)
		return
	}
	RetOk(c, gin.H{"entries": rows})
}
