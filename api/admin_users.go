package api

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminUserSummary deliberately omits secrets, activation codes and other
// authentication material. It is the server-authoritative read model used by
// Slotopol Admin for user administration.
type AdminUserSummary struct {
	UID    uint64 `json:"uid"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Status uint64 `json:"status"`
}

func ApiAdminUserList(c *gin.Context) {
	_, access := GetAdmin(c, 0)
	if access&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	query := strings.ToLower(strings.TrimSpace(c.Query("q")))
	rows := make([]AdminUserSummary, 0)
	for _, user := range Users.Items() {
		if query != "" && !strings.Contains(strings.ToLower(user.Email), query) && !strings.Contains(strings.ToLower(user.Name), query) {
			continue
		}
		rows = append(rows, AdminUserSummary{
			UID: user.UID, Email: user.Email, Name: user.Name, Status: uint64(user.Status),
		})
	}

	RetOk(c, gin.H{"users": rows, "count": len(rows)})
}
