package adminapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

func registerTopologyRoutes() {
	webserver.ApiGET("/network/topology", GetNetworkTopology)
}

// TopologyNode represents a node in the network topology graph
type TopologyNode struct {
	ID        int64           `json:"id,string"`
	Name      string          `json:"name"`
	Type      string          `json:"type"` // "node" or "nas"
	Status    string          `json:"status"`
	Lat       float64         `json:"lat"`
	Lng       float64         `json:"lng"`
	Children  []*TopologyNode `json:"children,omitempty"`
}

func GetNetworkTopology(c echo.Context) error {
	db := GetDB(c)

	var nodes []domain.NetNode
	if err := db.Find(&nodes).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch nodes", err.Error())
	}

	var nasList []domain.NetNas
	if err := db.Find(&nasList).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to fetch NAS devices", err.Error())
	}

	// Build hierarchy
	nodeMap := make(map[int64]*TopologyNode)
	rootNodes := []*TopologyNode{}

	// Convert NetNodes to TopologyNodes
	for _, n := range nodes {
		tn := &TopologyNode{
			ID:     n.ID,
			Name:   n.Name,
			Type:   "node",
			Status: "active", // Nodes are containers, usually active
			Lat:    n.Latitude,
			Lng:    n.Longitude,
			Children: []*TopologyNode{},
		}
		nodeMap[n.ID] = tn
		rootNodes = append(rootNodes, tn)
	}

	// Attach NAS devices to their parent Nodes
	for _, nas := range nasList {
		if parentNode, ok := nodeMap[nas.NodeId]; ok {
			tn := &TopologyNode{
				ID:     nas.ID,
				Name:   nas.Name,
				Type:   "nas",
				Status: nas.Status,
				// NAS inherits location from Node for now, or could have its own if added to specific model
				Lat:    parentNode.Lat, 
				Lng:    parentNode.Lng,
			}
			parentNode.Children = append(parentNode.Children, tn)
		}
	}

	return ok(c, rootNodes)
}
