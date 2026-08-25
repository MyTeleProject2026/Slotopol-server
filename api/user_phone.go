package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func normalizePhoneIdentity(value string) string {
	digits := make([]byte, 0, len(value))
	for _, r := range strings.TrimSpace(value) {
		if r >= '0' && r <= '9' {
			digits = append(digits, byte(r))
		}
	}
	if len(digits) == 0 {
		return ""
	}
	if strings.HasPrefix(string(digits), "00") {
		digits = digits[2:]
	}
	if len(digits) > 0 && digits[0] == '0' {
		digits = append([]byte("95"), digits[1:]...)
	}
	return string(digits)
}

func newPhoneSecret() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "slotopol-phone-provisioned-secret"
	}
	return hex.EncodeToString(buf)
}

// ApiUserPhone returns an existing Slotopol UID for a phone number or creates
// a new Slotopol player identity when the phone is not registered yet.
// This endpoint is intended for authenticated platform integrations such as
// N999Bet and deliberately does not use email as the player identity.
func ApiUserPhone(c *gin.Context) {
	var arg struct {
		Phone string `json:"phone" form:"phone" xml:"phone" yaml:"phone"`
		Name  string `json:"name" form:"name" xml:"name" yaml:"name"`
	}
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	phone := normalizePhoneIdentity(arg.Phone)
	if len(phone) < 8 || len(phone) > 15 {
		RetErr(c, http.StatusBadRequest, 0, ErrSmallKey)
		return
	}

	admin, al := MustAdmin(c, 0)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	var user User
	if ok, err := XormStorage.Where("phone = ?", phone).Get(&user); err != nil {
		Ret500(c, 0, err)
		return
	} else if ok {
		RetOk(c, gin.H{"success": true, "exists": true, "uid": user.UID, "phone": phone, "name": user.Name})
		return
	}

	name := strings.TrimSpace(arg.Name)
	if name == "" {
		name = "Player-" + phone
	}

	// The legacy Slotopol schema requires a non-null unique email. It is kept
	// only as an internal compatibility value; N999Bet identity matching is
	// exclusively phone-based and this endpoint never accepts an email.
	internalEmail := phone + "@phone.slotopol.internal"
	user = User{
		Email:  internalEmail,
		Phone:  phone,
		Secret: newPhoneSecret(),
		Name:   name,
		Status: UFactivated,
		GAL:    ALmember,
	}

	if _, err := XormStorage.Insert(&user); err != nil {
		// Handle a concurrent phone provisioning request safely by re-reading.
		var existing User
		if ok, readErr := XormStorage.Where("phone = ?", phone).Get(&existing); readErr == nil && ok {
			RetOk(c, gin.H{"success": true, "exists": true, "uid": existing.UID, "phone": phone, "name": existing.Name})
			return
		}
		Ret500(c, 0, err)
		return
	}

	user.Init()
	Users.Set(user.UID, &user)
	RetOk(c, gin.H{"success": true, "exists": false, "uid": user.UID, "phone": phone, "name": user.Name})
}
