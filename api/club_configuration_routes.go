package api

import "github.com/gin-gonic/gin"

// RegisterClubConfigurationRoutes wires the country/currency profile,
// provider-credit, audit, settlement monitoring and RBAC APIs into the
// authenticated platform API.
func RegisterClubConfigurationRoutes(r *gin.Engine) {
	ra := r.Group("/", Auth(true))

	ra.POST("/admin/club/profile", ApiClubProfileSet)
	ra.GET("/admin/club/profile", ApiClubProfileGet)

	ra.POST("/admin/country-game-profile", ApiCountryGameProfileSet)
	ra.GET("/admin/country-game-profiles", ApiCountryGameProfileList)

	ra.GET("/admin/club/currency-balance", ApiClubCurrencyBalanceGet)
	ra.POST("/admin/club/currency-balance/adjust", ApiClubCurrencyBalanceAdjust)

	ra.GET("/admin/audit", ApiAdminAuditList)
	ra.GET("/admin/settlement/recent", ApiAdminSettlementRecent)

	ra.GET("/admin/rbac/me", ApiAdminRBACMe)
	ra.GET("/admin/rbac/catalog", ApiAdminRBACCatalog)
}
