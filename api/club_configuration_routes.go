package api

import "github.com/gin-gonic/gin"

// RegisterClubConfigurationRoutes wires the country/currency profile and
// provider-credit APIs into the authenticated platform API. These endpoints
// are intentionally kept separate from customer-wallet accounting: the
// balances here are provider credit owned by the club relationship.
func RegisterClubConfigurationRoutes(r *gin.Engine) {
	ra := r.Group("/", Auth(true))

	ra.POST("/admin/club/profile", ApiClubProfileSet)
	ra.GET("/admin/club/profile", ApiClubProfileGet)

	ra.POST("/admin/country-game-profile", ApiCountryGameProfileSet)
	ra.GET("/admin/country-game-profiles", ApiCountryGameProfileList)

	ra.GET("/admin/club/currency-balance", ApiClubCurrencyBalanceGet)
	ra.POST("/admin/club/currency-balance/adjust", ApiClubCurrencyBalanceAdjust)
}
