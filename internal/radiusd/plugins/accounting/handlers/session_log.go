package handlers

import (
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"layeh.com/radius/rfc2866"
)

type SessionLogHandler struct {
	DB *gorm.DB
}

func NewSessionLogHandler(db *gorm.DB) *SessionLogHandler {
	return &SessionLogHandler{DB: db}
}

func (h *SessionLogHandler) Name() string {
	return "session_log"
}

func (h *SessionLogHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	// Handle Start (1), Stop (2), Interim-Update (3)
	return ctx.StatusType == 1 || ctx.StatusType == 2 || ctx.StatusType == 3
}

func (h *SessionLogHandler) Handle(ctx *accounting.AccountingContext) error {
	// We only primarily care about STOP for the full log, but we can track START too if needed.
	// For "Session History" requirements, usually a closed session (STOP) is the most important.
	// However, to track active sessions turning into history, we might want to log START or just rely on STOP.
	// Let's log STOP events as completed sessions.
	
	if ctx.StatusType == 2 { // STOP
		return h.handleStop(ctx)
	}
	
	return nil
}

func (h *SessionLogHandler) handleStop(ctx *accounting.AccountingContext) error {
	// Extract details and save to SessionLog
	// Note: In a real high-perf scenario, might push to a queue. For now, direct DB write.
	
	val := rfc2866.AcctSessionTime_Get(ctx.Request.Packet)
	input := rfc2866.AcctInputOctets_Get(ctx.Request.Packet)
	output := rfc2866.AcctOutputOctets_Get(ctx.Request.Packet)
	
	log := domain.SessionLog{
		AcctSessionId:   rfc2866.AcctSessionID_GetString(ctx.Request.Packet),
		Username:        ctx.Username,
		NasAddr:         ctx.NASIP,
		NasId:           ctx.NAS.Identifier,
		AcctStopTime:    time.Now(),
		AcctSessionTime: int(val),
		AcctInputTotal:  int64(input),
		AcctOutputTotal: int64(output),
		TerminateCause:  fmt.Sprintf("%v", rfc2866.AcctTerminateCause_Get(ctx.Request.Packet)),
	}

	if err := h.DB.Create(&log).Error; err != nil {
		zap.L().Error("failed to create session log", zap.Error(err))
		return err
	}
	
	return nil
}
