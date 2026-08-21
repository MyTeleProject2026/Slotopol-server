package api

import (
	"encoding/xml"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/MyTeleProject2026/Slotopol-server/config"
)

type Allocation struct {
	ID            uint64    `xorm:"id" json:"id" yaml:"id" xml:"id,attr"`
	TransactionID string    `xorm:"transaction_id" json:"transaction_id" yaml:"transaction_id" xml:"transaction_id"`
	ClubID        uint64    `xorm:"club_id" json:"club_id" yaml:"club_id" xml:"club_id"`
	Amount        float64   `xorm:"amount" json:"amount" yaml:"amount" xml:"amount"`
	Type          string    `xorm:"type" json:"type" yaml:"type" xml:"type"`
	Status        string    `xorm:"status" json:"status" yaml:"status" xml:"status"`
	CreatedAt     time.Time `xorm:"created_at" json:"created_at" yaml:"created_at" xml:"created_at"`
	ApprovedAt    *time.Time `xorm:"approved_at" json:"approved_at,omitempty" yaml:"approved_at,omitempty" xml:"approved_at,omitempty"`
	Note          string    `xorm:"note" json:"note,omitempty" yaml:"note,omitempty" xml:"note,omitempty"`
}

func ApiAllocationCreate(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		ClubID  uint64   `json:"club_id" yaml:"club_id" xml:"club_id" binding:"required"`
		Amount  float64  `json:"amount" yaml:"amount" xml:"amount" binding:"required"`
		Note    string   `json:"note,omitempty" yaml:"note,omitempty" xml:"note,omitempty"`
	}
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}
	admin, al := MustAdmin(c, 0)
	if admin == nil || al&ALmaster == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}
	if _, ok := Clubs.Get(arg.ClubID); !ok {
		Ret404(c, 0, ErrNoClub)
		return
	}
	txID := "ALLOC-" + time.Now().Format("20060102150405") + "-" + util.RandString(8)
	alloc := &Allocation{
		TransactionID: txID,
		ClubID:        arg.ClubID,
		Amount:        arg.Amount,
		Type:          "ALLOCATE",
		Status:        "PENDING",
		Note:          arg.Note,
	}
	if _, err := XormStorage.Insert(alloc); err != nil {
		Ret500(c, 0, err)
		return
	}
	RetOk(c, alloc)
}

func ApiAllocationApprove(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		ID      uint64   `json:"id" yaml:"id" xml:"id" binding:"required"`
	}
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}
	admin, al := MustAdmin(c, 0)
	if admin == nil || al&ALmaster == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}
	var alloc Allocation
	if has, err := XormStorage.ID(arg.ID).Get(&alloc); err != nil || !has {
		Ret404(c, 0, ErrNotFound)
		return
	}
	if alloc.Status != "PENDING" {
		Ret400(c, 0, errors.New("allocation not pending"))
		return
	}
	now := time.Now()
	alloc.Status = "APPROVED"
	alloc.ApprovedAt = &now
	session := XormStorage.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		Ret500(c, 0, err)
		return
	}
	if _, err := session.ID(alloc.ID).Update(&alloc); err != nil {
		session.Rollback()
		Ret500(c, 0, err)
		return
	}
	// Increase club bank
	if _, err := session.Exec("UPDATE club SET bank=bank+? WHERE cid=?", alloc.Amount, alloc.ClubID); err != nil {
		session.Rollback()
		Ret500(c, 0, err)
		return
	}
	if err := session.Commit(); err != nil {
		Ret500(c, 0, err)
		return
	}
	Ret204(c)
}

func ApiAllocationList(c *gin.Context) {
	admin, al := MustAdmin(c, 0)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}
	var allocs []Allocation
	if err := XormStorage.Desc("created_at").Find(&allocs); err != nil {
		Ret500(c, 0, err)
		return
	}
	RetOk(c, gin.H{"allocations": allocs})
}
