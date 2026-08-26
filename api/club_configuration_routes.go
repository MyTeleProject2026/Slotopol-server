package api

import "github.com/gin-gonic/gin"

// RegisterClubConfigurationRoutes wires country/currency configuration,
// provider-credit governance, audit, settlement monitoring and RBAC APIs into
// the authenticated platform API.
func RegisterClubConfigurationRoutes(r *gin.Engine) {
	ra := r.Group("/", Auth(true))

	ra.POST("/admin/club/profile", ApiClubProfileSet)
	ra.GET("/admin/club/profile", ApiClubProfileGet)

	ra.POST("/admin/country-game-profile", ApiCountryGameProfileSet)
	ra.GET("/admin/country-game-profiles", ApiCountryGameProfileList)

	ra.GET("/admin/club/currency-balance", ApiClubCurrencyBalanceGet)
	ra.POST("/admin/club/currency-balance/adjust", ApiClubCurrencyBalanceAdjust)
	ra.GET("/admin/club/currency-balances", ApiAdminCurrencyOverview)
	ra.POST("/admin/club/currency-transfer", ApiAdminClubCurrencyTransfer)
	ra.GET("/admin/club/currency-ledger", ApiAdminClubCurrencyLedger)
	ra.GET("/admin/club/currency-reconciliation", ApiAdminCurrencyReconciliation)

	// Virtual Slotopol-server treasury: unlimited demo/provider credit is
	// minted server-side and credited in the target club's configured country
	// currency. No external payment rail is involved.
	ra.POST("/admin/treasury/mint-transfer", ApiAdminVirtualTreasuryMintTransfer)
	ra.POST("/admin/treasury/transfer-request", ApiAdminTreasuryTransferRequest)
	ra.GET("/admin/treasury/approvals", ApiAdminTreasuryApprovalList)
	ra.POST("/admin/treasury/approval-decision", ApiAdminTreasuryApprovalDecision)

	ra.GET("/admin/audit", ApiAdminAuditList)
	ra.GET("/admin/settlement/recent", ApiAdminSettlementRecent)

	ra.GET("/admin/rbac/me", ApiAdminRBACMe)
	ra.GET("/admin/rbac/catalog", ApiAdminRBACCatalog)
}
