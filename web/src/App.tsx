import { CpeList } from './resources/cpes';
import { Router as RouterIcon } from '@mui/icons-material';
// New device and location resources
import { DeviceList, DeviceShow } from './pages/Devices';
import { LocationList, LocationCreate, LocationEdit } from './pages/Locations';
import { NasTemplateList, NasTemplateCreate, NasTemplateEdit } from './resources/nasTemplates';
import { LandingI18nProvider } from './contexts/LandingI18nProvider';
import { Admin, Resource, CustomRoutes } from 'react-admin';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Box, CircularProgress, Typography } from '@mui/material';
import { dataProvider } from './providers/dataProvider';
import { authProvider } from './providers/authProvider';
import { i18nProvider } from './i18n';
import Dashboard from './pages/Dashboard';
import EnvironmentMonitoring from './pages/EnvironmentMonitoring';
import AccountSettings from './pages/AccountSettings';
import { SystemConfigPage } from './pages/SystemConfigPage';
import FinancialPerformance from './pages/FinancialPerformance';
import VoucherPrintingPage from './pages/VoucherPrintingPage';
import { LoginPage } from './pages/LoginPage';
import { TunnelSettings } from './pages/Settings/TunnelSettings';
import { SystemSettings } from './pages/SystemSettings';
import { CustomLayout, CustomError } from './components';
import { PlatformLayout } from './components/PlatformLayout';
import { theme, darkTheme } from './theme';
// Import new Phase 4 & 5 resources
import { DeviceHealthList, MonitoringDashboard } from './resources/monitoring';
import { InvoiceList as BillingInvoiceList, InvoiceShow as BillingInvoiceShow } from './resources/billing';
import { BackupList, BackupCreate } from './resources/backups';
// Import platform management pages
import { LandingPage } from './pages/Landing/LandingPage';
import { PlatformDashboard, PlatformSettings, ReportingDashboard, NotificationSettings } from './pages/Platform';
import { ProviderRegistrationList } from './resources/platformSettings';
import { ProviderList, ProviderShow, ProviderCreate, ProviderEdit } from './resources/providers';
import { QuotaList, QuotaShow, QuotaEdit } from './resources/quotas';

// 自定义加载组件，避免闪烁
const CustomLoading = () => (
  <Box
    sx={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '100vh',
      backgroundColor: '#f8fafc',
      gap: 2,
    }}
  >
    <CircularProgress size={40} sx={{ color: '#2563eb' }} />
    <Typography variant="body1" color="text.secondary" sx={{ color: '#64748b' }}>
      正在加载...
    </Typography>
  </Box>
);

// 导入资源组件
import {
  RadiusUserList,
  RadiusUserEdit,
  RadiusUserCreate,
  RadiusUserShow,
} from './resources/radiusUsers';
import { InvoiceList, InvoiceShow } from './resources/invoices';
import { OnlineSessionList, OnlineSessionShow } from './resources/onlineSessions';
import { AccountingList, AccountingShow } from './resources/accounting';
import {
  RadiusProfileList,
  RadiusProfileEdit,
  RadiusProfileCreate,
  RadiusProfileShow,
} from './resources/radiusProfiles';
import {
  NASWithTabs,
  NASEdit,
  NASCreate,
  NASShow,
} from './resources/nas';
import {
  NodeList,
  NodeEdit,
  NodeCreate,
  NodeShow,
} from './resources/nodes';
import {
  OperatorList,
  OperatorEdit,
  OperatorCreate,
  OperatorShow,
} from './resources/operators';
import {
  ProductList,
  ProductCreate,
  ProductEdit,
  ProductShow,
} from './resources/products';
import {
  VoucherBatchList,
  VoucherBatchCreate,
  VoucherList,
} from './resources/vouchers';
import {
  AgentList,
  AgentCreate,
  AgentEdit,
  AgentShow,
} from './resources/agents';
import {
  ServerList,
  ServerCreate,
  ServerEdit,
} from './resources/servers';
import {
  VoucherBundleList,
  VoucherBundleCreate,
} from './resources/voucherBundles';
import { SystemLogList } from './resources/systemLogs';

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

    {/* Network Devices & Locations */}
    <Resource
      name="network/devices"
      list={DeviceList}
      show={DeviceShow}
    />
    <Resource
      name="network/locations"
      list={LocationList}
      create={LocationCreate}
      edit={LocationEdit}
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
      edit={ProviderEdit}
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
      <Route path="/monitoring" element={<MonitoringDashboard />} />
      <Route path="/settings" element={<PlatformSettings />} />
    </CustomRoutes>
  </Admin>
);

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
        }
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
      list={NASWithTabs}
      edit={NASEdit}
      create={NASCreate}
      show={NASShow}
    />

    {/* NAS Templates - accessed via tab */}
    <Resource
      name="network/nas-templates"
      list={NasTemplateList}
      create={NasTemplateCreate}
      edit={NasTemplateEdit}
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

    {/* Network Devices & Locations */}
    <Resource
      name="network/devices"
      list={DeviceList}
      show={DeviceShow}
    />
    <Resource
      name="network/locations"
      list={LocationList}
      create={LocationCreate}
      edit={LocationEdit}
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
      <Route path="/reporting" element={<ReportingDashboard />} />
      <Route path="/reporting/notifications" element={<NotificationSettings />} />
      <Route path="/environment" element={<EnvironmentMonitoring />} />
    </CustomRoutes>
  </Admin>
      } />
    </Routes>
  </BrowserRouter>
);

export default App;
