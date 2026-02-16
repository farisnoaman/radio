# ToughRadius Capabilities Report

## 1. Mikrotik & ISP Management
**Supported.**
-   **Authentication & Accounting**: Full support for standard RADIUS protocol.
-   **Mikrotik Optimization**: Includes `MikrotikAcceptEnhancer` which automatically maps user profile speeds (Up/Down Rate) to Mikrotik's `Mikrotik-Rate-Limit` attribute.
-   **Session Management**: The system supports standard `Disconnect-Request` (CoA) to kick online users directly from the admin dashboard.
-   **Vendor Dictionary**: Includes built-in `dictionary.mikrotik` for resolving vendor-specific attributes.

## 2. Billing System
**Partial / Usage-Only.**
-   **AAA Logic**: It is a core AAA (Authentication, Authorization, Accounting) system. It tracks **Time** and **Data Usage**.
-   **Prepaid/Expiration**: Supports `ExpireTime` for prepaid-style access control.
-   **No Financials**: It is **NOT** a billing platform. It does **not** manage:
    -   Invoices / Receipts
    -   Payments / Gateways
    -   Account Balances (Money)
    -   Tax / Currency

## 3. Bulk User Creation
**Not Supported.**
-   **Single Creation**: The API and Web UI only support creating users one by one.
-   **No Import**: There is no built-in feature to import users from CSV/Excel.
-   **Workaround**: You would need to write a script to loop through a CSV and call the `/api/v1/users` endpoint to create users in bulk.
