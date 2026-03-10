package app

import (
	"fmt"
	"time"

	"github.com/go-routeros/routeros"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
)

// SchedServerMonitorTask monitors all servers periodically
// Connects to Mikrotik via API to check health and fetch actual online users
func (a *Application) SchedServerMonitorTask() {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Error("Server monitor task panicked", zap.Any("error", err))
		}
	}()

	var servers []domain.Server
	if err := a.gormDB.Find(&servers).Error; err != nil {
		zap.S().Errorf("Failed to retrieve servers for monitoring: %v", err)
		return
	}

	for _, server := range servers {
		if server.PublicIP == "" || server.RouterStatus == "disabled" {
			continue
		}

		apiPort := "8728" // Default Mikrotik API port
		if server.Ports != "" {
			apiPort = server.Ports
		}

		address := fmt.Sprintf("%s:%s", server.PublicIP, apiPort)
		
		// Attempt to connect to Mikrotik API
		client, err := routeros.Dial(address, server.Username, server.Password)
		
		status := "online"
		var onlineHotspot, onlinePPPoE int

		if err != nil {
			zap.S().Warnf("Failed to connect to Mikrotik router %s at %s: %v", server.Name, address, err)
			status = "offline"
		} else {
			defer client.Close()

			// Fetch Hotspot Active Users
			replyHotspot, err := client.Run("/ip/hotspot/active/print", "=count-only=")
			if err == nil && len(replyHotspot.Re) > 0 {
				if countStr, ok := replyHotspot.Re[0].Map["ret"]; ok {
					fmt.Sscanf(countStr, "%d", &onlineHotspot)
				}
			}

			// Fetch PPPoE Active Users
			replyPPPoE, err := client.Run("/ppp/active/print", "=count-only=")
			if err == nil && len(replyPPPoE.Re) > 0 {
				if countStr, ok := replyPPPoE.Re[0].Map["ret"]; ok {
					fmt.Sscanf(countStr, "%d", &onlinePPPoE)
				}
			}
		}

		// Update database
		err = a.gormDB.Model(&server).Updates(map[string]interface{}{
			"router_status":  status,
			"online_hotspot": onlineHotspot,
			"online_pppoe":   onlinePPPoE,
			"updated_at":     time.Now(),
		}).Error
		
		if err != nil {
			zap.S().Errorf("Failed to update status for server %s: %v", server.Name, err)
		} else {
			zap.S().Debugf("Server %s monitored. Status: %s, Hotspot: %d, PPPoE: %d", server.Name, status, onlineHotspot, onlinePPPoE)
		}
	}
}
