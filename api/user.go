package api

import (
	"encoding/xml"

	"github.com/gin-gonic/gin"
)

// ApiUserIs checks if a user exists.
func ApiUserIs(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Exists  bool     `json:"exists" yaml:"exists" xml:"exists"`
		Email   string   `json:"email,omitempty" yaml:"email,omitempty" xml:"email,omitempty"`
		Name    string   `json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var admin, al = MustAdmin(c, 0)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	var user *User
	if user, ok = Users.Get(arg.UID); ok {
		ret.Exists = true
		ret.Email = user.Email
		ret.Name = user.Name
	} else {
		ret.Exists = false
	}

	RetOk(c, ret)
}

// ApiUserRename changes a user's name.
func ApiUserRename(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid" binding:"required"`
		Name    string   `json:"name" yaml:"name" xml:"name" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var user *User
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, 0)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	// Update name in memory
	user.Name = arg.Name

	// Update in database
	if _, err = XormStorage.ID(arg.UID).Cols("name").Update(&User{Name: arg.Name}); err != nil {
		Ret500(c, 0, err)
		return
	}

	Ret204(c)
}

// ApiUserSecret changes a user's password.
func ApiUserSecret(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid" binding:"required"`
		Secret  string   `json:"secret" yaml:"secret" xml:"secret" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}
	if len(arg.Secret) < 6 {
		Ret400(c, 0, ErrSmallKey)
		return
	}

	var user *User
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, 0)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	// Update secret in memory
	user.Secret = arg.Secret

	// Update in database
	if _, err = XormStorage.ID(arg.UID).Cols("secret").Update(&User{Secret: arg.Secret}); err != nil {
		Ret500(c, 0, err)
		return
	}

	Ret204(c)
}

// ApiUserDelete removes a user (hard delete).
func ApiUserDelete(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var user *User
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, 0)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}
	_ = user // avoid unused warning

	// Remove from memory
	Users.Delete(arg.UID)

	// Delete from database
	if _, err = XormStorage.ID(arg.UID).Delete(&User{}); err != nil {
		Ret500(c, 0, err)
		return
	}

	Ret204(c)
}
