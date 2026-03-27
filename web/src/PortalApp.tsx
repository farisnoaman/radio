import { Admin, Resource, CustomRoutes } from 'react-admin';
import { Route } from 'react-router-dom';
import { portalAuthProvider } from './providers/portalAuthProvider';
import { dataProvider } from './providers/dataProvider';
import { i18nProvider } from './i18n';
import { theme, darkTheme } from './theme';
import { InvoiceList, InvoiceShow } from './resources/invoices';
import { CustomLayout, CustomError } from './components';
import UserDashboard from './pages/UserDashboard';
import { LoginPage } from './pages/LoginPage';
import MyDevices from './pages/MyDevices';
import VoucherRedeem from './pages/VoucherRedeem';
import NotificationPreferences from './pages/NotificationPreferences';
import AlertHistory from './pages/AlertHistory';

const PortalApp = () => (
  <Admin
    title="portal.title"
    authProvider={portalAuthProvider}
    dataProvider={dataProvider}
    i18nProvider={i18nProvider}
    theme={theme}
    darkTheme={darkTheme}
    layout={CustomLayout}
    dashboard={UserDashboard}
    loginPage={LoginPage}
    error={CustomError}
    requireAuth
  >
    <Resource
      name="radius/invoices"
      list={InvoiceList}
      show={InvoiceShow}
    />
    <CustomRoutes>
      <Route path="/portal/devices" element={<MyDevices />} />
      <Route path="/portal/vouchers/redeem" element={<VoucherRedeem />} />
      <Route path="/portal/preferences/notifications" element={<NotificationPreferences />} />
      <Route path="/portal/alerts/history" element={<AlertHistory />} />
    </CustomRoutes>
  </Admin>
);

export default PortalApp;
