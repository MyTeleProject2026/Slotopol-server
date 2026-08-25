package api

import "github.com/gin-gonic/gin"

// RegisterClubConfigurationRoutes wires the country/currency profile,
// provider-credit and phone-based player identity APIs into the authenticated
// platform API. These endpoints remain separate from customer-wallet accounting.
func RegisterClubConfigurationRoutes(r *gin.Engine) {
	ra := r.Group("/", Auth(true))

	ra.POST("/admin/club/profile", ApiClubProfileSet)
	ra.GET("/admin/club/profile", ApiClubProfileGet)

	ra.POST("/admin/country-game-profile", ApiCountryGameProfileSet)
	ra.GET("/admin/country-game-profiles", ApiCountryGameProfileList)

	ra.GET("/admin/club/currency-balance", ApiClubCurrencyBalanceGet)
	ra.POST("/admin/club/currency-balance/adjust", ApiClubCurrencyBalanceAdjust)

	// N999Bet player provisioning uses phone as the authoritative identity.
	ra.GET("/user/phone", ApiUserPhone)
	ra.POST("/user/phone", ApiUserPhone)
}
