package api

import "github.com/gin-gonic/gin"

// ApiAuthMe returns the authenticated administrator identity used by admin clients.
// It is intentionally backed by the same JWT + User store used by protected APIs.
func ApiAuthMe(c *gin.Context) {
	user := c.MustGet(userKey).(*User)
	role := "member"
	if _, al := GetAdmin(c, 0); al&ALadmin != 0 {
		role = "super_admin"
	}
	RetOk(c, gin.H{"user": gin.H{
		"uid": user.UID,
		"email": user.Email,
		"name": user.Name,
		"role": role,
		"access": uint(user.GAL),
	}})
}
