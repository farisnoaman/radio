# Platform Routing Reorganization Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Reorganize all platform-level management pages under `/platform` route with separate Admin instance, layout, and menu to create clear separation between platform-level and provider-level operations.

**Architecture:** Create separate React Admin instance for platform routes (`/platform/*`) with its own layout (PlatformLayout) and menu (PlatformMenu). Platform admin manages multi-tenant operations (providers, quotas, registrations, backups, monitoring) while main admin handles provider-specific operations (RADIUS, users, products, billing).

**Tech Stack:** React Admin v4, React Router v6, TypeScript, Material UI, echo framework (Go backend)

---

## Task 1: Create PlatformLayout Component

**Files:**
- Create: `web/src/components/PlatformLayout.tsx`

**Step 1: Create the PlatformLayout component**

```tsx
import { Layout, LayoutProps } from 'react-admin';
import { Box } from '@mui/material';
import { PlatformMenu } from './PlatformMenu';

export const PlatformLayout = (props: LayoutProps) => (
  <Layout
    {...props}
    menu={PlatformMenu}
    sx={{
      '& .RaLayout-content': {
        bgcolor: 'background.default',
      },
    }}
  >
    {props.children}
  </Layout>
);
```

**Step 2: Verify file creation**

Run: `ls -la web/src/components/PlatformLayout.tsx`
Expected: File exists

**Step 3: Commit**

```bash
git add web/src/components/PlatformLayout.tsx
git commit -m "feat: add PlatformLayout component for platform admin section"
```

---

## Task 2: Create PlatformMenu Component

**Files:**
- Create: `web/src/components/PlatformMenu.tsx`

**Step 1: Create the PlatformMenu component**

```tsx
import BusinessIcon from '@mui/icons-material/Business';
import AssessmentIcon from '@mui/icons-material/Assessment';
import DomainOutlinedIcon from '@mui/icons-material/DomainOutlined';
import HowToRegIcon from '@mui/icons-material/HowToReg';
import StorageOutlinedIcon from '@mui/icons-material/StorageOutlined';
import MonitorIcon from '@mui/icons-material/Monitor';
import { Box, useTheme } from '@mui/material';
import { MenuItemLink, MenuProps, useTranslate, useLocale } from 'react-admin';

const RTL_LANGUAGES = ['ar', 'he', 'fa', 'ur'];

const platformMenuItems = [
  { to: '/platform/dashboard', labelKey: 'menu.platform_dashboard', icon: <BusinessIcon /> },
  { to: '/platform/monitoring', labelKey: 'menu.monitoring', icon: <MonitorIcon /> },
  { to: '/platform/monitoring/devices', labelKey: 'menu.device_health', icon: <AssessmentIcon /> },
  { to: '/platform/quotas', labelKey: 'menu.quotas', icon: <AssessmentIcon /> },
  { to: '/platform/providers', labelKey: 'menu.providers', icon: <DomainOutlinedIcon /> },
  { to: '/platform/registrations', labelKey: 'menu.provider_registrations', icon: <HowToRegIcon /> },
  { to: '/platform/backups', labelKey: 'menu.backups', icon: <StorageOutlinedIcon /> },
];

export const PlatformMenu = ({ dense, onMenuClick, logout }: MenuProps) => {
  const currentYear = new Date().getFullYear();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const translate = useTranslate();
  const locale = useLocale();
  const isRTL = RTL_LANGUAGES.includes(locale || '');

  return (
    <Box
      dir={isRTL ? 'rtl' : 'ltr'}
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        backgroundColor: isDark ? '#1e293b' : '#0f766e', // Teal color to distinguish from main admin
        color: '#ffffff',
        pt: 0,
        transition: 'background-color 0.3s ease',
      }}
    >
      <Box sx={{ flexGrow: 1, overflowY: 'auto', pt: 1, marginTop: 2 }}>
        {platformMenuItems.map((item) => (
          <MenuItemLink
            key={item.to}
            to={item.to}
            primaryText={translate(item.labelKey)}
            leftIcon={item.icon}
            dense={dense}
            onClick={onMenuClick}
            sx={{
              textAlign: isRTL ? 'right' : 'left',
              '& .RaMenuItemLink-icon': {
                minWidth: isRTL ? 'auto' : undefined,
                ml: isRTL ? 1 : 0,
                mr: isRTL ? 0 : undefined,
              },
            }}
          />
        ))}
      </Box>

      <Box
        sx={{
          borderTop: '1px solid rgba(255, 255, 255, 0.1)',
          textAlign: 'center',
          px: 2,
          py: 3,
          fontSize: 12,
          color: 'rgba(255, 255, 255, 0.6)',
          transition: 'all 0.3s ease',
        }}
      >
        <div style={{ fontWeight: 600, marginBottom: 4 }}>Platform Admin</div>
        <div>© {currentYear} ALL RIGHTS RESERVED</div>
        {logout && <Box sx={{ mt: 2 }}>{logout}</Box>}
      </Box>
    </Box>
  );
};
```

**Step 2: Verify file creation**

Run: `ls -la web/src/components/PlatformMenu.tsx`
Expected: File exists

**Step 3: Commit**

```bash
git add web/src/components/PlatformMenu.tsx
git commit -m "feat: add PlatformMenu component with platform navigation"
```

---

## Task 3: Add Missing Translation Keys

**Files:**
- Modify: `web/src/i18n/en-US.ts`
- Modify: `web/src/i18n/ar.ts`

**Step 1: Add English translations**

Find the `menu` section in `web/src/i18n/en-US.ts` and add:
```typescript
monitoring: 'Monitoring',
device_health: 'Device Health',
backups: 'Backup Management',
```

**Step 2: Add Arabic translations**

Find the `menu` section in `web/src/i18n/ar.ts` and add:
```typescript
monitoring: 'المراقبة',
device_health: 'صحة الأجهزة',
backups: 'إدارة النسخ الاحتياطية',
```

**Step 3: Test translation loading**

Run: `npm run dev`
Expected: Dev server starts without translation errors

**Step 4: Commit**

```bash
git add web/src/i18n/en-US.ts web/src/i18n/ar.ts
git commit -m "feat: add missing translation keys for platform menu"
```

---

## Task 4: Reorganize App.tsx - Create PlatformAdmin Section

**Files:**
- Modify: `web/src/App.tsx:113-330`

**Step 1: Update imports to include platform components**

Add these imports at the top with other platform imports:
```tsx
import { PlatformLayout } from './components/PlatformLayout';
import { PlatformMenu } from './components/PlatformMenu';
```

**Step 2: Create PlatformAdmin component**

Add this new component before the main `App` component:
```tsx
const PlatformAdmin = () => (
  <Admin
    dataProvider={dataProvider}
    authProvider={authProvider}
    i18nProvider={i18nProvider}
    dashboard={PlatformDashboard}
    loginPage={LoginPage}
    title="Platform Admin"
    theme={theme}
    darkTheme={darkTheme}
    defaultTheme="light"
    layout={PlatformLayout}
    loading={CustomLoading}
    error={CustomError}
    requireAuth
  >
    {/* Platform Monitoring */}
    <Resource
      name="monitoring/devices"
      list={DeviceHealthList}
      options={{ label: 'Device Health' }}
    />

    {/* Platform Quota Management */}
    <Resource
      name="quotas"
      list={QuotaList}
      show={QuotaShow}
      edit={QuotaEdit}
      options={{ label: 'Resource Quotas' }}
    />

    {/* Provider Registration Management */}
    <Resource
      name="providers/registrations"
      list={ProviderRegistrationList}
      options={{ label: 'Provider Registrations' }}
    />

    {/* Provider Management */}
    <Resource
      name="providers"
      list={ProviderList}
      show={ProviderShow}
      create={ProviderCreate}
      options={{ label: 'Providers' }}
    />

    {/* Platform Backup Management */}
    <Resource
      name="provider/backup"
      list={BackupList}
      create={BackupCreate}
      options={{ label: 'Backup Management' }}
    />

    {/* Platform Custom Routes */}
    <CustomRoutes>
      <Route path="/platform/dashboard" element={<PlatformDashboard />} />
      <Route path="/platform/monitoring" element={<MonitoringDashboard />} />
    </CustomRoutes>
  </Admin>
);
```

**Step 3: Update App routing to use PlatformAdmin**

Replace the current routing structure with:
```tsx
const App = () => (
  <BrowserRouter>
    <Routes>
      {/* Public Landing Page - Accessible without authentication */}
      <Route
        path="/landing/*"
        element={
          <LandingI18nProvider>
            <LandingPage />
          </LandingI18nProvider>
        />

      {/* Platform Admin - Super Admin Only */}
      <Route path="/platform/*" element={<PlatformAdmin />} />

      {/* Main Admin Application - Provider Operations */}
      <Route path="/*" element={
        <Admin
          dataProvider={dataProvider}
          authProvider={authProvider}
          i18nProvider={i18nProvider}
          dashboard={Dashboard}
          loginPage={LoginPage}
          title="Radio v0.01"
          theme={theme}
          darkTheme={darkTheme}
          defaultTheme="light"
          layout={CustomLayout}
          loading={CustomLoading}
          error={CustomError}
          requireAuth
        >
    {/* RADIUS 用户管理 */}
    <Resource
      name="radius/users"
      list={RadiusUserList}
      edit={RadiusUserEdit}
      create={RadiusUserCreate}
      show={RadiusUserShow}
    />

    {/* 在线会话 */}
    <Resource
      name="radius/online"
      list={OnlineSessionList}
      show={OnlineSessionShow}
    />

    {/* 计费记录 */}
    <Resource
      name="radius/accounting"
      list={AccountingList}
      show={AccountingShow}
    />

    {/* Invoices */}
    <Resource
      name="radius/invoices"
      list={InvoiceList}
      show={InvoiceShow}
      options={{ label: 'Invoices' }}
    />

    {/* RADIUS 配置 */}
    <Resource
      name="radius/profiles"
      list={RadiusProfileList}
      edit={RadiusProfileEdit}
      create={RadiusProfileCreate}
      show={RadiusProfileShow}
    />

    {/* NAS 设备管理 */}
    <Resource
      name="network/nas"
      list={NASList}
      edit={NASEdit}
      create={NASCreate}
      show={NASShow}
    />

    {/* 网络节点 */}
    <Resource
      name="network/nodes"
      list={NodeList}
      edit={NodeEdit}
      create={NodeCreate}
      show={NodeShow}
    />

    {/* Servers */}
    <Resource
      name="network/servers"
      list={ServerList}
      edit={ServerEdit}
      create={ServerCreate}
    />

    {/* 操作员管理 */}
    <Resource
      name="system/operators"
      list={OperatorList}
      edit={OperatorEdit}
      create={OperatorCreate}
      show={OperatorShow}
    />

    {/* Products */}
    <Resource
      name="products"
      list={ProductList}
      create={ProductCreate}
      edit={ProductEdit}
      show={ProductShow}
    />

    {/* Vouchers */}
    <Resource
      name="voucher-batches"
      list={VoucherBatchList}
      create={VoucherBatchCreate}
    />
    <Resource
      name="vouchers"
      list={VoucherList}
    />

    <Resource
      name="voucher-bundles"
      list={VoucherBundleList}
      create={VoucherBundleCreate}
    />

    {/* Agents */}
    <Resource
      name="agents"
      list={AgentList}
      create={AgentCreate}
      edit={AgentEdit}
      show={AgentShow}
    />

    {/* System Logs */}
    <Resource
      name="system/logs"
      list={SystemLogList}
    />

    <Resource name="cpes" options={{ label: 'CPE Devices' }} list={CpeList} icon={RouterIcon} />

    {/* Phase 5A: Billing - Provider Level */}
    <Resource
      name="billing/invoices"
      list={BillingInvoiceList}
      show={BillingInvoiceShow}
      options={{ label: 'Provider Billing' }}
    />

    {/* Custom Routes (Authenticated) */}
    <CustomRoutes>
      {/* Monitoring Dashboard - Keep accessible in main admin too */}
      <Route path="/monitoring/dashboard" element={<MonitoringDashboard />} />

      {/* Existing Routes */}
      <Route path="/account/settings" element={<AccountSettings />} />
      <Route path="/system/config" element={<SystemConfigPage />} />
      <Route path="/settings/tunnel" element={<TunnelSettings />} />
      <Route path="/system/maintenance" element={<SystemSettings />} />
      <Route path="/financial/performance" element={<FinancialPerformance />} />
      <Route path="/voucher-printing" element={<VoucherPrintingPage />} />
    </CustomRoutes>
  </Admin>
      } />
    </Routes>
  </BrowserRouter>
);
```

**Step 4: Verify TypeScript compilation**

Run: `cd web && npm run build`
Expected: No TypeScript errors

**Step 5: Commit**

```bash
git add web/src/App.tsx
git commit -m "feat: create separate PlatformAdmin instance for platform routes"
```

---

## Task 5: Remove Platform Items from CustomMenu

**Files:**
- Modify: `web/src/components/CustomMenu.tsx:30-59`

**Step 1: Remove platform menu items**

Remove these lines from the menuItems array:
```tsx
{ to: '/platform/dashboard', labelKey: 'menu.platform_dashboard', icon: <BusinessIcon />, permissions: ['super'] },
{ to: '/providers', labelKey: 'menu.providers', icon: <DomainOutlinedIcon />, permissions: ['super'] },
{ to: '/quotas', labelKey: 'menu.quotas', icon: <PieChartIcon />, permissions: ['super'] },
{ to: '/providers/registrations', labelKey: 'menu.provider_registrations', icon: <HowToRegIcon />, permissions: ['super'] },
```

**Step 2: Add platform entry point**

Add this to menuItems (after dashboard item):
```tsx
{ to: '/platform/dashboard', labelKey: 'menu.platform_admin', icon: <BusinessIcon />, permissions: ['super'] },
```

**Step 3: Test menu renders correctly**

Run: `npm run dev`
Expected: Menu shows "Platform Admin" instead of individual platform items

**Step 4: Commit**

```bash
git add web/src/components/CustomMenu.tsx
git commit -m "refactor: remove platform items from main menu, add platform admin entry"
```

---

## Task 6: Update Translation Keys for Menu

**Files:**
- Modify: `web/src/i18n/en-US.ts`
- Modify: `web/src/i18n/ar.ts`

**Step 1: Add platform_admin key to English**

In the menu section, add:
```typescript
platform_admin: 'Platform Admin',
```

**Step 2: Add platform_admin key to Arabic**

In the menu section, add:
```typescript
platform_admin: 'إدارة المنصة',
```

**Step 3: Commit**

```bash
git add web/src/i18n/en-US.ts web/src/i18n/ar.ts
git commit -m "feat: add platform_admin translation key"
```

---

## Task 7: Test All Platform Routes

**Files:**
- No file changes - testing only

**Step 1: Start dev servers**

Run: `cd /home/faris/Documents/lamees/radio && ./start_dev.sh`
Expected: Backend on port 1816, frontend on port 3000

**Step 2: Test platform admin access**

Open: http://localhost:3000/platform/dashboard
Expected: Platform dashboard loads with teal-colored sidebar

**Step 3: Test all platform routes**

Navigate to each route and verify:
- `/platform/dashboard` ✓
- `/platform/monitoring` ✓
- `/platform/monitoring/devices` ✓
- `/platform/quotas` ✓
- `/platform/providers` ✓
- `/platform/registrations` ✓
- `/platform/backups` ✓

**Step 4: Test main admin routes**

Navigate to main routes and verify:
- `/` - Main dashboard ✓
- `/radius/users` ✓
- `/products` ✓
- `/agents` ✓

**Step 5: Test menu navigation**

Expected:
- Main admin menu shows "Platform Admin" link
- Clicking "Platform Admin" navigates to `/platform/dashboard`
- Platform sidebar has teal color, platform-specific menu
- Main admin sidebar has blue color, provider-specific menu

**Step 6: Test permissions**

Log in as non-super admin
Expected: "Platform Admin" menu item not visible, `/platform/*` routes redirect

**Step 7: Commit documentation**

```bash
echo "# Platform Routing Test Results

$(date): Tested all platform and main admin routes
- All platform routes accessible under /platform/*
- Platform admin has separate teal-colored layout
- Main admin retains blue-colored layout
- Permission-based access working correctly
" >> docs/testing/platform-routing-tests.md

git add docs/testing/platform-routing-tests.md
git commit -m "test: document platform routing verification"
```

---

## Task 8: Update Documentation

**Files:**
- Create: `docs/platform-routing.md`

**Step 1: Create documentation**

```markdown
# Platform Routing Architecture

## Overview
Platform-level management pages are organized under `/platform` route with separate Admin instance.

## Platform Routes
- `/platform/dashboard` - Platform dashboard
- `/platform/monitoring` - Platform-wide monitoring metrics
- `/platform/monitoring/devices` - Device health monitoring
- `/platform/quotas` - Resource quota management
- `/platform/providers` - Provider management
- `/platform/registrations` - Provider registration approval
- `/platform/backups` - Backup management

## Main Admin Routes
- All provider-specific operations (RADIUS, billing, users, products, etc.)
- Accessible at root level routes

## Key Differences
- **Platform Admin**: Teal-colored sidebar, multi-tenant management
- **Main Admin**: Blue-colored sidebar, provider-specific operations

## Access Control
- Platform routes require super admin permission
- Provider admins can only access main admin routes relevant to their provider
```

**Step 2: Commit documentation**

```bash
git add docs/platform-routing.md
git commit -m "docs: add platform routing architecture documentation"
```

---

## Task 9: Final Integration Test

**Files:**
- No file changes - integration testing

**Step 1: Stop all processes**

Run: `./stop_dev.sh` or kill all node/go processes

**Step 2: Clean build**

Run:
```bash
cd web
rm -rf node_modules/.vite dist
npm run build
```

Expected: Clean build with no errors

**Step 3: Full restart**

Run: `cd /home/faris/Documents/lamees/radio && ./start_dev.sh`
Expected: Both backend and frontend start successfully

**Step 4: End-to-end test**

1. Login as super admin
2. Navigate to `/platform/dashboard`
3. Click through all platform menu items
4. Navigate back to main admin
5. Verify all main admin routes work
6. Test language switching (English/Arabic)
7. Test dark/light theme toggle

**Step 5: Final commit**

```bash
git add -A
git commit -m "test: complete integration testing of platform routing reorganization"
```

---

## Summary

This implementation plan:
1. Creates separate PlatformLayout and PlatformMenu components
2. Reorganizes App.tsx to have PlatformAdmin and MainAdmin instances
3. Updates CustomMenu to reference platform admin instead of individual items
4. Adds necessary translation keys
5. Tests all routes and permissions
6. Documents the new architecture

**Total estimated time:** 45-60 minutes
**Number of commits:** 9
**Files created:** 3
**Files modified:** 4
