package domain

import (
	"time"
)

// RADIUS related models

// RadiusProfile RADIUS billing profile
type RadiusProfile struct {
	ID             int64     `json:"id,string" form:"id"`                                  // Primary key ID
	NodeId         int64     `gorm:"index" json:"node_id,string" form:"node_id"`          // Node ID
	TenantID       int64     `gorm:"index" json:"tenant_id" form:"tenant_id"`             // Tenant/Provider ID

	Name           string    `json:"name" form:"name"`                         // Profile name
	Status         string    `gorm:"index" json:"status" form:"status"`        // Profile status: 0=disabled 1=enabled
	AddrPool       string    `json:"addr_pool" form:"addr_pool"`               // Address pool
	ActiveNum      int       `json:"active_num" form:"active_num"`             // Concurrent sessions
	UpRate         int       `json:"up_rate" form:"up_rate"`                   // Upload rate in Kb
	DownRate       int       `json:"down_rate" form:"down_rate"`               // Download rate in Kb
	DataQuota      int64     `json:"data_quota" form:"data_quota"`             // Data quota in MB (0 = unlimited)
	Domain         string    `json:"domain" form:"domain"`                     // Domain, corresponds to NAS device domain attribute, e.g., Huawei domain_code
	IPv6PrefixPool string    `json:"ipv6_prefix_pool" form:"ipv6_prefix_pool"` // IPv6 prefix pool name for NAS-side allocation
	BindMac        int       `json:"bind_mac" form:"bind_mac"`                 // Bind MAC
	BindVlan       int       `json:"bind_vlan" form:"bind_vlan"`               // Bind VLAN
	Remark         string    `json:"remark" form:"remark"`                     // Remark
	CreatedAt      time.Time `json:"created_at" form:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" form:"updated_at"`
}

// TableName Specify table name
func (RadiusProfile) TableName() string {
	return "radius_profile"
}

// RadiusUser RADIUS Authentication account
type RadiusUser struct {
	ID              int64     `json:"id,string" form:"id"`                              // Primary key ID
	TenantID        int64     `gorm:"index" json:"tenant_id" form:"tenant_id"`         // Tenant/Provider ID
	NodeId          int64     `json:"node_id,string" form:"node_id"`                    // Node ID
	ProfileId       int64     `gorm:"index" json:"profile_id,string" form:"profile_id"` // RADIUS profile ID
	Realname        string    `gorm:"index" json:"realname" form:"realname"`                         // Contact name
	Email           string    `json:"email" form:"email"`                               // Email address
	Mobile          string    `gorm:"index" json:"mobile" form:"mobile"`                             // Contact phone

	Address         string    `json:"address" form:"address"`                           // Contact address
	Username        string    `json:"username" gorm:"uniqueIndex" form:"username"`      // Account name
	Password        string    `json:"password" form:"password"`                         // Password
	AddrPool        string    `json:"addr_pool" form:"addr_pool"`                       // Address pool
	ActiveNum       int       `gorm:"index" json:"active_num" form:"active_num"`        // Concurrent sessions
	UpRate          int       `json:"up_rate" form:"up_rate"`                           // Upload rate
	DownRate        int       `json:"down_rate" form:"down_rate"`                       // Download rate
	DataQuota       int64     `json:"data_quota" form:"data_quota"`                     // Data quota in MB (0 = unlimited)
	TimeQuota       int64     `json:"time_quota" form:"time_quota"`                     // Time quota in seconds (0 = unlimited)
	Vlanid1         int       `json:"vlanid1" form:"vlanid1"`                           // VLAN ID 1
	Vlanid2         int       `json:"vlanid2" form:"vlanid2"`                           // VLAN ID 2
	IpAddr          string    `json:"ip_addr" form:"ip_addr"`                           // Static IP
	IpV6Addr        string    `json:"ipv6_addr" form:"ipv6_addr"`                       // Static IPv6 address
	MacAddr         string    `json:"mac_addr" form:"mac_addr"`                         // MAC address
	Domain          string    `json:"domain" form:"domain"`                             // Domain name for vendor-specific features (e.g., Huawei domain)
	IPv6PrefixPool  string    `json:"ipv6_prefix_pool" form:"ipv6_prefix_pool"`         // IPv6 prefix pool name (inherited from profile or user-specific)
	BindVlan        int       `json:"bind_vlan" form:"bind_vlan"`                       // Bind VLAN
	BindMac         int       `json:"bind_mac" form:"bind_mac"`                         // Bind MAC
	ProfileLinkMode int       `json:"profile_link_mode" form:"profile_link_mode"`       // 0=static (snapshot), 1=dynamic (real-time from profile)
	IdleTimeout     int       `json:"idle_timeout" form:"idle_timeout"`                  // Inactivity timeout in seconds
	SessionTimeout  int       `json:"session_timeout" form:"session_timeout"`            // Max session duration in seconds
	ExpireTime      time.Time `gorm:"index" json:"expire_time"`                         // Expiration time
	Status          string    `gorm:"index" json:"status" form:"status"`                // Status: enabled | disabled
	Remark          string    `json:"remark" form:"remark"`                             // Remark

	// Postpaid billing fields.
	// BillingType determines whether the user is billed via prepaid vouchers or postpaid monthly invoices.
	// Default is "prepaid" so existing users are unaffected by the migration.
	BillingType        string    `json:"billing_type" gorm:"size:20;default:'prepaid';index" form:"billing_type"`
	// SubscriptionStatus tracks the postpaid subscription lifecycle.
	// Only relevant when BillingType is "postpaid".
	// Values: "active", "suspended", "canceled". Default empty means N/A for prepaid users.
	SubscriptionStatus string    `json:"subscription_status" gorm:"size:20" form:"subscription_status"`
	// NextBillingDate is the date when the next invoice should be generated.
	// The billing engine cron job checks this field daily.
	NextBillingDate    time.Time `json:"next_billing_date" gorm:"index" form:"next_billing_date"`
	// MonthlyFee is the recurring charge amount for the postpaid subscription.
	// Inherited from the assigned RadiusProfile/Product at subscription time.
	MonthlyFee         float64   `json:"monthly_fee" form:"monthly_fee"`
	// PricePerGb is the cost per Gigabyte of data consumed.
	// Used for consumption-based billing.
	PricePerGb         float64   `json:"price_per_gb" form:"price_per_gb"`

	OnlineCount     int       `json:"online_count" gorm:"-:migration;<-:false"`
	LastOnline      time.Time `json:"last_online"`
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Voucher linkage fields
	// These fields link a RadiusUser to a voucher for batch-level control
	VoucherBatchID int64  `json:"voucher_batch_id,string" gorm:"index" form:"voucher_batch_id"`
	VoucherCode    string `json:"voucher_code" gorm:"index" form:"voucher_code"`

	// Reverse relationships
	// UsageAlerts is the list of alerts sent to this user
	UsageAlerts []*UsageAlert `json:"usage_alerts,omitempty" gorm:"foreignKey:UserID"`
	// NotificationPreference holds this user's notification settings
	NotificationPreference *NotificationPreference `json:"notification_preference,omitempty" gorm:"foreignKey:UserID"`
}

// TableName Specify table name
func (RadiusUser) TableName() string {
	return "radius_user"
}

// RadiusOnline
// Radius RadiusOnline Recode
type RadiusOnline struct {
	ID                  int64     `json:"id,string"` // Primary key ID
	TenantID            int64     `gorm:"index" json:"tenant_id"`            // Tenant/Provider ID
	Username            string    `gorm:"index" json:"username"`
	NasId               string    `gorm:"index" json:"nas_id"`
	NasAddr             string    `json:"nas_addr"`
	NasPaddr            string    `json:"nas_paddr"`
	SessionTimeout      int       `json:"session_timeout"`
	FramedIpaddr        string    `gorm:"index" json:"framed_ipaddr"`
	FramedNetmask       string    `json:"framed_netmask"`
	// IPv6 fields (RFC 3162)
	FramedIpv6Prefix     string    `json:"framed_ipv6_prefix"`     // RFC 3162 attribute 97
	FramedIpv6PrefixLen  int       `json:"framed_ipv6_prefix_len"`  // Prefix length
	FramedInterfaceId    string    `json:"framed_interface_id"`     // RFC 3162 attribute 96
	FramedIpv6Address    string    `json:"framed_ipv6_address"`     // Delegated address
	DelegatedIpv6Prefix  string    `json:"delegated_ipv6_prefix"`   // Delegated prefix
	MacAddr             string    `gorm:"index" json:"mac_addr"`

	NasPort             int64     `json:"nas_port,string"`
	NasClass            string    `json:"nas_class"`
	NasPortId           string    `json:"nas_port_id"`
	NasPortType         int       `json:"nas_port_type"`
	ServiceType         int       `json:"service_type"`
	AcctSessionId       string    `gorm:"index" json:"acct_session_id"`
	AcctSessionTime     int       `json:"acct_session_time"`
	AcctInputTotal      int64     `json:"acct_input_total,string"`
	AcctOutputTotal     int64     `json:"acct_output_total,string"`
	AcctInputPackets    int       `json:"acct_input_packets"`
	AcctOutputPackets   int       `json:"acct_output_packets"`
	AcctStartTime       time.Time `gorm:"index" json:"acct_start_time"`
	LastUpdate          time.Time `json:"last_update"`
}

// TableName Specify table name
func (RadiusOnline) TableName() string {
	return "radius_online"
}

// RadiusAccounting
// Radius Accounting Recode
type RadiusAccounting struct {
	ID                  int64     `json:"id,string"` // Primary key ID
	TenantID            int64     `gorm:"index" json:"tenant_id"`            // Tenant/Provider ID
	Username            string    `gorm:"index" json:"username"`
	AcctSessionId       string    `gorm:"index" json:"acct_session_id"`
	NasId               string    `gorm:"index" json:"nas_id"`
	NasAddr             string    `json:"nas_addr"`
	NasPaddr            string    `json:"nas_paddr"`
	SessionTimeout      int       `json:"session_timeout"`
	FramedIpaddr        string    `gorm:"index" json:"framed_ipaddr"`
	FramedNetmask       string    `json:"framed_netmask"`
	// IPv6 fields (RFC 3162)
	FramedIpv6Prefix     string    `json:"framed_ipv6_prefix"`     // RFC 3162 attribute 97
	FramedIpv6PrefixLen  int       `json:"framed_ipv6_prefix_len"`  // Prefix length
	FramedInterfaceId    string    `json:"framed_interface_id"`     // RFC 3162 attribute 96
	FramedIpv6Address    string    `json:"framed_ipv6_address"`     // Delegated address
	DelegatedIpv6Prefix  string    `json:"delegated_ipv6_prefix"`   // Delegated prefix
	MacAddr             string    `gorm:"index" json:"mac_addr"`

	NasPort             int64     `json:"nas_port,string"`
	NasClass            string    `json:"nas_class"`
	NasPortId           string    `json:"nas_port_id"`
	NasPortType         int       `json:"nas_port_type"`
	ServiceType         int       `json:"service_type"`
	AcctSessionTime     int       `json:"acct_session_time"`
	AcctInputTotal      int64     `json:"acct_input_total,string"`
	AcctOutputTotal     int64     `json:"acct_output_total,string"`
	AcctInputPackets    int       `json:"acct_input_packets"`
	AcctOutputPackets   int       `json:"acct_output_packets"`
	LastUpdate          time.Time `json:"last_update"`
	AcctStartTime       time.Time `gorm:"index" json:"acct_start_time"`
	AcctStopTime        time.Time `gorm:"index" json:"acct_stop_time"`
}

// TableName Specify table name
func (RadiusAccounting) TableName() string {
	return "radius_accounting"
}
