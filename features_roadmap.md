# Features Roadmap: Next-Gen ISP Billing & Management

This roadmap outlines the path to transforming ToughRADIUS into a top-tier ISP billing and management solution, competing with industry leaders like Advanced Radius, RLRadius, and Splynx.

## 🚀 Immediate Enhancements (Phase 1)

### 1. User Self-Care Status Page
**Goal:** Empower users to track their own usage without contacting support.
- [ ] **Real-time Dashboard:** Display session time remaining, data used (Upload/Download), and expiration date.
- [ ] **Visual Gauges:** Circular progress bars for data and time quotas.
- [ ] **One-Click Renewal:** Integration with payment gateways for instant voucher purchase/renewal.
- [ ] **Implementation:** Mikrotik `status.html` with reactive JS and API integration.

### 2. Flexible Product Configuration (Validity Units)
**Goal:** Simplify plan creation for administrators.
- [ ] **Smart Validity Units:** Input duration in Hours, Days, or Months (automatically converted to seconds).
- [ ] **Bandwidth on Demand (Turbo):** Allow users to purchase temporary speed boosts (e.g., "1 Hour Turbo Mode").
- [ ] **Time Banks:** Allow unused time to rollover or be banked for later use.

### 3. Advanced Voucher Printing
**Goal:** Professional, customizable branding.
- [ ] **WYSIWYG Template Editor:** Drag-and-drop editor for voucher design.
- [ ] **Bulk Export:** Export batches to PDF, CSV, or Excel with custom layouts.
- [ ] **Dealer Branding:** Allow resellers to add their own logos to vouchers.

---

## 🛠 Core System Evolution (Phase 2)

### 4. Reseller & Agent Management
**Goal:** Scale sales operations through a multi-tier agent network.
- [ ] **Commission System:** Automated commission calculation and wallet system.
- [ ] **Sub-Resellers:** Master agents can create and manage their own sub-agents.
- [ ] **Inventory Management:** Track voucher stock transfer from admin to agents.

### 5. Fair Usage Policy (FUP) & QoS
**Goal:** Ensure network stability and fair resource distribution.
- [ ] **Dynamic Throttling:** Automatically reduce speed (COA) after a user exceeds daily data thresholds.
- [ ] **Burst Profiles:** Configure Mikrotik Burst attributes directly from the plan settings.
- [ ] **Time-Based Policies:** Different speeds for Peak vs. Off-Peak hours.

### 6. Network Monitoring & Alerts
**Goal:** Proactive network maintenance.
- [ ] **NAS Status Monitoring:** Real-time up/down status of Mikrotik routers via SNMP/Ping.
- [ ] **Signal Strength Logs:** Collect and graph user signal strength (RSSI) history to debug connection issues.
- [ ] **Telegram/WhatsApp Alerts:** Notify admins of outages or suspicious activity.

---

## 🔮 Future Innovation (Phase 3)

### 7. AI-Driven Insights
- [ ] **Churn Prediction:** Identify users likely to leave based on usage patterns.
- [ ] **Revenue Forecasting:**  Predict next month's revenue based on active subscriptions.

### 8. Mobile App (Flutter/React Native)
- [ ] **White-label App:** A branded app for ISPs to offer their customers (User: Recharge, Support, Status; Admin: Monitoring).
