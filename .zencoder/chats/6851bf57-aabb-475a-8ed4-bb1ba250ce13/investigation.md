# Investigation Report - Lint Errors in `web/src/pages/VoucherPrintingPage.tsx`

## Bug Summary
Multiple lint and diagnostic errors were reported in [./web/src/pages/VoucherPrintingPage.tsx](./web/src/pages/VoucherPrintingPage.tsx):
1.  **no-useless-escape**: Unnecessary escape character `\/` on line 860.
2.  **no-explicit-any**: Use of `any` on lines 190, 228, 305, 306.
3.  **react-hooks/exhaustive-deps**: Missing dependencies in `useEffect` on lines 322 and 334.
4.  **no-inline-styles**: Inline styles used for an `iframe` on line 917 (referenced as 914 in the report).

## Root Cause Analysis
1.  **Line 860**: The developer escaped the forward slash in `</script>` inside a template literal. While often done in script tags to prevent parsing issues, it's flagged as unnecessary here by ESLint.
2.  **any usage**: Developers used `any` for translation options and API response types instead of defining proper interfaces.
3.  **useEffect dependencies**:
    - Line 322: `searchParams` and `notify` are used inside the effect but are missing from the dependency array.
    - Line 334: `products` state is accessed to check if a product is already loaded, but it's not included in the dependency array.
4.  **Inline styles**: The `iframe` uses the `style` prop for CSS, which violates the project's preference for MUI's `sx` prop or styled components.

## Affected Components
- [./web/src/pages/VoucherPrintingPage.tsx](./web/src/pages/VoucherPrintingPage.tsx)

## Proposed Solution
1.  **Fix useless escape**: Change `<\/script>` to `</script>` on line 860.
2.  **Replace `any`**:
    - Define a proper type for translation options (e.g., `Record<string, unknown>`).
    - Use `VoucherBatch[]` and `VoucherTemplate[]` for `apiRequest` generic parameters on lines 305 and 306.
3.  **Update `useEffect` dependencies**:
    - Add `searchParams` and `notify` to the dependency array on line 322.
    - Add `products` to the dependency array on line 334.
4.  **Fix inline styles**: Convert the `style` prop on the `iframe` (line 917) to an `sx` prop using `Box` component.

## Implementation Notes
- **no-useless-escape**: Removed the unnecessary backslash in `</script>` within the print window HTML template.
- **no-explicit-any**:
    - Replaced `any` in `TemplateVars` and `getSampleTemplate` with `Record<string, unknown>`.
    - Specified `VoucherBatch[]` and `VoucherTemplate[]` in `apiRequest` calls within the initial `useEffect`.
- **react-hooks/exhaustive-deps**:
    - Added `searchParams` and `notify` to the mount effect's dependency array.
    - Added `products` to the product fetching effect's dependency array.
- **no-inline-styles**: Wrapped the `iframe` in an MUI `Box` component and moved its styles to the `sx` prop.

## Test Results
- **Lint**: `npx eslint src/pages/VoucherPrintingPage.tsx` passed with 0 errors and 0 warnings.
- **Type Check**: `npx tsc src/pages/VoucherPrintingPage.tsx --noEmit` passed successfully.
