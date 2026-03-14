import { CpeList } from './resources/cpes';
import { Router as RouterIcon } from '@mui/icons-material';
import { Admin, Resource, CustomRoutes } from 'react-admin';
import { Route } from 'react-router-dom';
import { Box, CircularProgress, Typography } from '@mui/material';
import { dataProvider } from './providers/dataProvider';
import { authProvider } from './providers/authProvider';
import { i18nProvider } from './i18n';
import Dashboard from './pages/Dashboard';
import AccountSettings from './pages/AccountSettings';
import { SystemConfigPage } from './pages/SystemConfigPage';
import FinancialPerformance from './pages/FinancialPerformance';
import VoucherPrintingPage from './pages/VoucherPrintingPage';
import { LoginPage } from './pages/LoginPage';
import { TunnelSettings } from './pages/Settings/TunnelSettings';
import { SystemSettings } from './pages/SystemSettings';
import { CustomLayout, CustomError } from './components';
import { theme, darkTheme } from './theme';

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
  NASList,
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

const App = () => (
  <Admin
    dataProvider={dataProvider}
    authProvider={authProvider}
    i18nProvider={i18nProvider}
    dashboard={Dashboard}
    loginPage={LoginPage}
    title="TOUGHRADIUS v9"
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

    {/* 自定义路由 */}
    <CustomRoutes>
      <Route path="/account/settings" element={<AccountSettings />} />
      <Route path="/system/config" element={<SystemConfigPage />} />
      <Route path="/settings/tunnel" element={<TunnelSettings />} />
      <Route path="/system/maintenance" element={<SystemSettings />} />
      <Route path="/financial/performance" element={<FinancialPerformance />} />
      <Route path="/voucher-printing" element={<VoucherPrintingPage />} />
    </CustomRoutes>
  </Admin >
);

export default App;
