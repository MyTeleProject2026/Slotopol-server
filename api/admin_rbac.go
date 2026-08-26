package api

import "github.com/gin-gonic/gin"

// AdminPermission describes one authoritative Slotopol access capability.
type AdminPermission struct {
	Key   string `json:"key" yaml:"key" xml:"key"`
	Bit   uint   `json:"bit" yaml:"bit" xml:"bit"`
	Label string `json:"label" yaml:"label" xml:"label"`
	Scope string `json:"scope" yaml:"scope" xml:"scope"`
}

// adminPermissionCatalog is deliberately derived from the server's existing AL
// access model. It is not a second, frontend-only permission system.
var adminPermissionCatalog = []AdminPermission{
	{Key: "member", Bit: uint(ALmember), Label: "Club membership", Scope: "club"},
	{Key: "dealer", Bit: uint(ALdealer), Label: "Game and player gameplay controls", Scope: "club"},
	{Key: "booker", Bit: uint(ALbooker), Label: "Player properties and deposit transfers", Scope: "club"},
	{Key: "master", Bit: uint(ALmaster), Label: "Club bank, fund and deposit controls", Scope: "club"},
	{Key: "admin", Bit: uint(ALadmin), Label: "Administrative access-level management", Scope: "platform"},
}

// ApiAdminRBACMe returns the authenticated administrator's effective global
// access level and the authoritative permission catalog. The response is
// useful to admin clients for hiding controls, while the server remains the
// final authorization boundary for every mutation.
func ApiAdminRBACMe(c *gin.Context) {
	user, access := GetAdmin(c, 0)
	if user == nil {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	RetOk(c, gin.H{
		"uid":          user.UID,
		"access":       uint(access),
		"permissions":  adminPermissionCatalog,
	})
}

// ApiAdminRBACCatalog exposes the server-owned permission definitions to an
// authenticated administrator. It is read-only and contains no secrets.
func ApiAdminRBACCatalog(c *gin.Context) {
	user, access := GetAdmin(c, 0)
	if user == nil || access&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	RetOk(c, gin.H{
		"permissions": adminPermissionCatalog,
		"all":         uint(ALall),
	})
}
