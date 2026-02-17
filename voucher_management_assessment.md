# Voucher Management Assessment & Test Plan

## 1. Current State Assessment

### **Strengths**
*   **Batch Management:** Supports creating vouchers in batches with prefixes, varying lengths, and character types (numeric, alpha, mixed).
*   **Lifecycle Management:** Full lifecycle support: Create -> Activate -> Deactivate -> Refund -> Delete (Soft).
*   **Agent Integration:** Integrated with Agent Wallet for prepaid voucher generation.
*   **Redemption Logic:** `RedeemVoucher` handles validation (expiry, status) and automatically creates a Radius User.
*   **Export:** CSV export functionality exists.

### **Weaknesses & Gaps**
*   **Testing:** **Critical Gap.** No unit or integration tests found for `internal/adminapi/vouchers.go`. functionality.
*   **Security:**
    *   No rate limiting on `RedeemVoucher` endpoint (susceptible to brute-force attacks).
    *   Voucher codes are stored in plain text (acceptable for printed vouchers, but hashing could be considered for unused ones if high security is needed).
*   **Usability:**
    *   Export is limited to CSV. PDF export (for printing) is handled on frontend but backend support would be more robust for large batches.
    *   No "Search by Voucher Code" in `ListVoucherBatches` (only in `ListVouchers`).
*   **Performance:**
    *   `BulkDeactivateVouchers` iterates through vouchers in memory to disconnect users. For huge batches (10k+), this might be slow.
*   **Flexibility:**
    *   Voucher validity is defined by Product `ValiditySeconds`. Overriding this per-batch is not currently supported.

---

## 2. Proposed Enhancements

### **Immediate (High Priority)**
1.  **Unit Tests:** Implement comprehensive test suite for all voucher operations.
2.  **Rate Limiting:** Add middleware to limit `RedeemVoucher` attempts per IP.

### **Functional Improvements**
3.  **Advanced Search:** Allow searching for a batch by "Contains Voucher Code".
4.  **Batch Validity Override:** Allow setting a custom validity period for a specific batch that overrides the Product default.
5.  **Activity Log:** detailed history of voucher lifecycle changes (who activated it, when, from which IP).

---

## 3. Implementation Plan: Testing Suite

We will create `internal/adminapi/vouchers_test.go` using the existing test infrastructure (test helpers).

### **Test Scenarios to Cover**

#### **A. Batch Creation**
*   **Success:** Admin creates a standard batch.
*   **Success:** Agent creates a batch (Wallet deduction verified).
*   **Failure:** Agent insufficient funds.
*   **Failure:** Invalid Product ID.

#### **B. Voucher Lifecycle**
*   **Activation:** Bulk activate a batch. Verify status changes to "active".
*   **Deactivation:** Bulk deactivate. Verify status "unused" and online users disconnected.
*   **Redemption (The most critical flow):**
    *   **Success:** Redeem a valid "unused" code. Verify `RadiusUser` created with correct attributes.
    *   **Failure:** Redeem non-existent code.
    *   **Failure:** Redeem already "used" code.
    *   **Failure:** Redeem "expired" code.
    *   **Failure:** Redeem "disabled" (not yet active) code.

#### **C. Refund Logic**
*   **Success:** Refund unused vouchers. Verify Wallet balance increases and vouchers marked "refunded".
*   **Failure:** Try to refund a batch not owned by agent.

#### **D. Export**
*   **Success:** Verify CSV output format and headers.

## 4. Execution Steps

1.  **Scaffold Test File:** Create `internal/adminapi/vouchers_test.go`.
2.  **Setup Test Data:** Use `setupTestDB` (or similar helper) to create Products, Agents, and Wallets.
3.  **Write Tests:** Implement the scenarios above using Go's `testing` package and `httptest`.
4.  **Run & Refine:** Execute `go test ./internal/adminapi/...` and fix any bugs discovered.
