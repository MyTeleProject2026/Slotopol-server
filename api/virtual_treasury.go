package api

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ApiAdminVirtualTreasuryMintTransfer creates virtual provider credit from the
// Slotopol-server treasury and credits a club in that club's configured
// country currency. There is no external payment rail and no finite source
// wallet. The same amount is also added to the club's gameplay bank so the
// recharge is usable by the existing game/wallet flow.
func ApiAdminVirtualTreasuryMintTransfer(c *gin.Context) {
	admin, access := GetAdmin(c, 0)
	if admin == nil || access&ALmaster == 0 { Ret403(c, 0, ErrNoAccess); return }

	var arg struct { ToClubID uint64 `json:"to_club_id" binding:"required"`; Amount float64 `json:"amount" binding:"required"`; Reference string `json:"reference"` }
	if err := c.ShouldBindJSON(&arg); err != nil { Ret400(c, 0, err); return }
	if arg.Amount <= 0 || math.IsNaN(arg.Amount) || math.IsInf(arg.Amount, 0) { Ret400(c, 0, fmt.Errorf("amount must be a positive finite number")); return }
	if _, ok := Clubs.Get(arg.ToClubID); !ok { Ret404(c, 0, ErrNoClub); return }
	if err := ensureClubProfiles(); err != nil { Ret500(c, 0, err); return }

	var profile ClubProfile
	has, err := XormStorage.Where("club_id=?", arg.ToClubID).Get(&profile)
	if err != nil { Ret500(c, 0, err); return }
	if !has || strings.TrimSpace(profile.Currency) == "" || strings.TrimSpace(profile.CountryCode) == "" { Ret400(c, 0, fmt.Errorf("target club must have a country and currency profile before virtual treasury funding")); return }

	currency := strings.ToUpper(strings.TrimSpace(profile.Currency)); arg.Reference = strings.TrimSpace(arg.Reference); if arg.Reference == "" { arg.Reference = fmt.Sprintf("virtual-treasury-%d", time.Now().UnixNano()) }
	if err := ensureClubCurrencyBalances(); err != nil { Ret500(c, 0, err); return }; if err := ensureClubCurrencyLedger(); err != nil { Ret500(c, 0, err); return }

	s := XormStorage.NewSession(); defer s.Close(); if err := s.Begin(); err != nil { Ret500(c, 0, err); return }
	if err := creditClubCurrencyBalance(s, arg.ToClubID, currency, arg.Amount); err != nil { _ = s.Rollback(); Ret500(c, 0, err); return }
	if _, err := s.Exec("UPDATE club SET bank=bank+? WHERE cid=?", arg.Amount, arg.ToClubID); err != nil { _ = s.Rollback(); Ret500(c, 0, err); return }

	entry := &ClubCurrencyLedger{UID:admin.UID, FromClubID:0, ToClubID:arg.ToClubID, FromCurrency:currency, ToCurrency:currency, FromAmount:arg.Amount, ToAmount:arg.Amount, Rate:1, Reference:arg.Reference}
	if _, err := s.InsertOne(entry); err != nil { _ = s.Rollback(); Ret500(c, 0, err); return }
	if err := s.Commit(); err != nil { Ret500(c, 0, err); return }

	recordAdminAudit(c, arg.ToClubID, "virtual-treasury.mint-transfer", "club_currency_ledger", fmt.Sprintf("country=%s currency=%s amount=%g reference=%s", profile.CountryCode, currency, arg.Amount, arg.Reference))
	var balance ClubCurrencyBalance; _, _ = XormStorage.Where("club_id=? AND currency=?", arg.ToClubID, currency).Get(&balance)
	RetOk(c, gin.H{"success":true,"source":"slotopol-server-virtual-treasury","country_code":profile.CountryCode,"currency":currency,"amount":arg.Amount,"club_id":arg.ToClubID,"balance":balance.Balance,"ledger_id":entry.ID,"reference":arg.Reference})
}
