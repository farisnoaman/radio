package domain

import (
	"time"
)

// SessionLog records every RADIUS session lifecycle event
type SessionLog struct {
	ID              int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	AcctSessionId   string    `json:"acct_session_id" gorm:"index;size:64"`
	Username        string    `json:"username" gorm:"index;size:64"`
	NasAddr         string    `json:"nas_addr" gorm:"size:64"`
	NasId           string    `json:"nas_id" gorm:"size:64"`
	FramedIpaddr    string    `json:"framed_ipaddr" gorm:"size:64"`
	MacAddr         string    `json:"mac_addr" gorm:"size:32"`
	AcctStartTime   time.Time `json:"acct_start_time"`
	AcctStopTime    time.Time `json:"acct_stop_time"`
	AcctSessionTime int       `json:"acct_session_time"`
	AcctInputTotal  int64     `json:"acct_input_total"`
	AcctOutputTotal int64     `json:"acct_output_total"`
	TerminateCause  string    `json:"terminate_cause" gorm:"size:32"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (SessionLog) TableName() string {
	return "tr_session_logs"
}

// VoucherCampaign manages bulk voucher generation
type VoucherCampaign struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:128;not null"`
	Prefix      string    `json:"prefix" gorm:"size:16"`
	Length      int       `json:"length"` // Length of code excluding prefix
	Count       int       `json:"count"`  // Number of vouchers to generate
	Value       int64     `json:"value"`  // Monetary value
	PlanId      int64     `json:"plan_id"`
	Status      string    `json:"status" gorm:"size:16;default:'pending';index"` // pending, generating, completed

	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (VoucherCampaign) TableName() string {
	return "tr_voucher_campaigns"
}
