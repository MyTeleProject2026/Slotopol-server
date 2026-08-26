package api

import "github.com/gin-gonic/gin"

// ApiAdminCurrencyOverview exposes provider-credit balances for every club to
// the platform administrator. It is read-only and never changes balances.
func ApiAdminCurrencyOverview(c *gin.Context) {
	_, al := GetAdmin(c, 0)
	if al&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}
	if err := ensureClubCurrencyBalances(); err != nil {
		Ret500(c, 0, err)
		return
	}

	var rows []ClubCurrencyBalance
	q := XormStorage.Desc("club_id", "currency")
	if currency := c.Query("currency"); currency != "" {
		q = q.Where("currency=?", currency)
	}
	if err := q.Limit(1000).Find(&rows); err != nil {
		Ret500(c, 0, err)
		return
	}

	recordAdminAudit(c, 0, "club.currency-balance.overview", "club_currency_balance", "read-only all-club provider-credit overview")
	RetOk(c, gin.H{"balances": rows, "count": len(rows)})
}
