package domain

import (
	"time"
)

// AgentWallet Stores agent balance
type AgentWallet struct {
	AgentID   int64     `json:"agent_id,string" gorm:"primaryKey;autoIncrement:false" form:"agent_id"`
	Balance   float64   `json:"balance" form:"balance"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AgentWallet) TableName() string {
	return "agent_wallet"
}

// WalletLog Immutable ledger of transactions
type WalletLog struct {
	ID          int64     `json:"id,string" form:"id"`
	AgentID     int64     `json:"agent_id,string" form:"agent_id"`
	Type        string    `json:"type" form:"type"`     // deposit, purchase, refund
	Amount      float64   `json:"amount" form:"amount"` // Positive for deposit, negative for purchase
	Balance     float64   `json:"balance" form:"balance"` // Balance after transaction
	ReferenceID string    `json:"reference_id" form:"reference_id"` // Voucher Batch ID or Payment Ref
	Remark      string    `json:"remark" form:"remark"`
	CreatedAt   time.Time `json:"created_at"`
}

func (WalletLog) TableName() string {
	return "wallet_log"
}
