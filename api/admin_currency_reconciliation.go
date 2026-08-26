package api

import (
    "math"
    "sort"
    "strings"

    "github.com/gin-gonic/gin"
)

// ApiAdminCurrencyReconciliation provides a read-only control-plane check over
// provider-credit balances and immutable transfer history. It intentionally
// does not alter balances; discrepancies must be resolved by an authorized,
// separately audited operation.
func ApiAdminCurrencyReconciliation(c *gin.Context) {
    admin, al := GetAdmin(c, 0)
    if admin == nil || al&ALadmin == 0 {
        Ret403(c, 0, ErrNoAccess)
        return
    }
    if err := ensureClubCurrencyBalances(); err != nil { Ret500(c, 0, err); return }
    if err := ensureClubCurrencyLedger(); err != nil { Ret500(c, 0, err); return }

    var balances []ClubCurrencyBalance
    if err := XormStorage.Find(&balances); err != nil { Ret500(c, 0, err); return }
    var entries []ClubCurrencyLedger
    if err := XormStorage.Desc("id").Limit(1000).Find(&entries); err != nil { Ret500(c, 0, err); return }

    totals := map[string]float64{}
    negative := make([]ClubCurrencyBalance, 0)
    invalid := make([]ClubCurrencyBalance, 0)
    currencies := map[string]bool{}
    for _, row := range balances {
        code := strings.ToUpper(strings.TrimSpace(row.Currency))
        currencies[code] = true
        totals[code] += row.Balance
        if row.Balance < 0 { negative = append(negative, row) }
        if math.IsNaN(row.Balance) || math.IsInf(row.Balance, 0) { invalid = append(invalid, row) }
    }
    transferTotals := map[string]float64{}
    for _, entry := range entries {
        currencies[entry.FromCurrency] = true
        currencies[entry.ToCurrency] = true
        transferTotals[entry.FromCurrency] += entry.FromAmount
    }
    codes := make([]string, 0, len(currencies))
    for code := range currencies { codes = append(codes, code) }
    sort.Strings(codes)
    rows := make([]gin.H, 0, len(codes))
    for _, code := range codes {
        rows = append(rows, gin.H{"currency": code, "provider_credit_total": totals[code], "ledger_outgoing_total": transferTotals[code]})
    }
    status := "ok"
    if len(negative) > 0 || len(invalid) > 0 { status = "attention" }
    recordAdminAudit(c, 0, "club.currency.reconcile", "club_currency_balance", "read-only provider-credit reconciliation check")
    RetOk(c, gin.H{"status": status, "currencies": rows, "ledger_entries_checked": len(entries), "negative_balances": negative, "invalid_balances": invalid})
}
