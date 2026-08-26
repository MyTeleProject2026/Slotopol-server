package api

import "github.com/gin-gonic/gin"

// RegisterClubConfigurationRoutes wires the country/currency profile and
// provider-credit APIs into the authenticated platform API. Player phone
// identity remains registered by SetupRouter so Gin receives each route once.
func RegisterClubConfigurationRoutes(r *gin.Engine) {
	ra := r.Group("/", Auth(true))

	ra.POST("/admin/club/profile", ApiClubProfileSet)
	ra.GET("/admin/club/profile", ApiClubProfileGet)

	ra.POST("/admin/country-game-profile", ApiCountryGameProfileSet)
	ra.GET("/admin/country-game-profiles", ApiCountryGameProfileList)

	ra.GET("/admin/club/currency-balance", ApiClubCurrencyBalanceGet)
	ra.POST("/admin/club/currency-balance/adjust", ApiClubCurrencyBalanceAdjust)
}
