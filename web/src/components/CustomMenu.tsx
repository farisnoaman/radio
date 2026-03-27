import DashboardOutlinedIcon from '@mui/icons-material/DashboardOutlined';
import PeopleAltOutlinedIcon from '@mui/icons-material/PeopleAltOutlined';
import SensorsOutlinedIcon from '@mui/icons-material/SensorsOutlined';
import ReceiptLongOutlinedIcon from '@mui/icons-material/ReceiptLongOutlined';
import SettingsSuggestOutlinedIcon from '@mui/icons-material/SettingsSuggestOutlined';
import SettingsOutlinedIcon from '@mui/icons-material/SettingsOutlined';
import RouterOutlinedIcon from '@mui/icons-material/RouterOutlined';
import AccountTreeOutlinedIcon from '@mui/icons-material/AccountTreeOutlined';
import AdminPanelSettingsOutlinedIcon from '@mui/icons-material/AdminPanelSettingsOutlined';
import Inventory2OutlinedIcon from '@mui/icons-material/Inventory2Outlined';
import SupportAgentOutlinedIcon from '@mui/icons-material/SupportAgentOutlined';
import ConfirmationNumberOutlinedIcon from '@mui/icons-material/ConfirmationNumberOutlined';
import AssessmentOutlinedIcon from '@mui/icons-material/AssessmentOutlined';
import HistoryOutlinedIcon from '@mui/icons-material/HistoryOutlined';
import VpnKeyOutlinedIcon from '@mui/icons-material/VpnKeyOutlined';
import ConstructionIcon from '@mui/icons-material/Construction';
import PrintIcon from '@mui/icons-material/Print';
import StorageOutlinedIcon from '@mui/icons-material/StorageOutlined';
import DevicesIcon from '@mui/icons-material/Devices';
import BusinessIcon from '@mui/icons-material/Business';
import NotificationsIcon from '@mui/icons-material/Notifications';
import NotificationsActiveIcon from '@mui/icons-material/NotificationsActive';
import ThermostatIcon from '@mui/icons-material/Thermostat';

import LocationOnIcon from '@mui/icons-material/LocationOn';
import RouterIcon from '@mui/icons-material/Router';
import { Box, useTheme } from '@mui/material';
import { MenuItemLink, MenuProps, useGetIdentity, useTranslate, useLocale } from 'react-admin';

const RTL_LANGUAGES = ['ar', 'he', 'fa', 'ur'];

const menuItems = [
  { to: '/', labelKey: 'menu.dashboard', icon: <DashboardOutlinedIcon />, permissions: ['super', 'admin', 'user', 'agent'] },
  // Platform Administration (Super Admin Only)
  { to: '/platform/', labelKey: 'menu.platform_admin', icon: <BusinessIcon />, permissions: ['super'] },
  { to: '/platform/settings', labelKey: 'menu.platform_settings', icon: <SettingsOutlinedIcon />, permissions: ['super'] },
  // Network Management
  { to: '/network/servers', labelKey: 'menu.servers', icon: <StorageOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/network/nodes', labelKey: 'menu.network_nodes', icon: <AccountTreeOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/network/nas', labelKey: 'menu.nas_devices', icon: <RouterOutlinedIcon />, permissions: ['super', 'admin'] },
  // New devices and locations in network section
  { to: '/network/devices', labelKey: 'menu.network_devices', icon: <RouterIcon />, permissions: ['super', 'admin'] },
  { to: '/network/locations', labelKey: 'menu.network_locations', icon: <LocationOnIcon />, permissions: ['super', 'admin'] },
  { to: '/environment', labelKey: 'menu.env_monitoring', icon: <ThermostatIcon />, permissions: ['super', 'admin'] },
  { to: '/cpes', labelKey: 'menu.cpe_devices', icon: <RouterOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/users', labelKey: 'menu.radius_users', icon: <PeopleAltOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/profiles', labelKey: 'menu.radius_profiles', icon: <SettingsSuggestOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/online', labelKey: 'menu.online_sessions', icon: <SensorsOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/accounting', labelKey: 'menu.accounting', icon: <ReceiptLongOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/invoices', labelKey: 'menu.invoices', icon: <ReceiptLongOutlinedIcon />, permissions: ['super', 'admin', 'user', 'agent'] },
  { to: '/system/config', labelKey: 'menu.system_config', icon: <SettingsOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/system/operators', labelKey: 'menu.operators', icon: <AdminPanelSettingsOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/system/logs', labelKey: 'menu.system_logs', icon: <HistoryOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/products', labelKey: 'menu.products', icon: <Inventory2OutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/agents', labelKey: 'menu.agents', icon: <SupportAgentOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/financial/performance', labelKey: 'menu.financial_performance', icon: <AssessmentOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/reporting', labelKey: 'menu.reporting_dashboard', icon: <AssessmentOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/reporting/notifications', labelKey: 'menu.notification_settings', icon: <NotificationsActiveIcon />, permissions: ['super', 'admin'] },
  { to: '/settings/tunnel', labelKey: 'menu.tunnel_settings', icon: <VpnKeyOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/system/maintenance', labelKey: 'menu.maintenance', icon: <ConstructionIcon />, permissions: ['super', 'admin'] },
  { to: '/voucher-batches', labelKey: 'menu.vouchers', icon: <ConfirmationNumberOutlinedIcon />, permissions: ['super', 'admin', 'agent'] },
  { to: '/voucher-printing', labelKey: 'menu.print_vouchers', icon: <PrintIcon />, permissions: ['super', 'admin', 'agent'] },
  { to: '/portal/devices', labelKey: 'portal.my_devices', icon: <DevicesIcon />, permissions: ['user'] },
  { to: '/portal/vouchers/redeem', labelKey: 'portal.redeem_voucher', icon: <ConfirmationNumberOutlinedIcon />, permissions: ['user'] },
  { to: '/portal/preferences/notifications', labelKey: 'menu.notification_preferences', icon: <NotificationsIcon />, permissions: ['user'] },
  { to: '/portal/alerts/history', labelKey: 'menu.alert_history', icon: <HistoryOutlinedIcon />, permissions: ['user'] },
];

export const CustomMenu = ({ dense, onMenuClick, logout }: MenuProps) => {
  const currentYear = new Date().getFullYear();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { data: identity } = useGetIdentity();
  const translate = useTranslate();
  const locale = useLocale();
  const isRTL = RTL_LANGUAGES.includes(locale || '');

  const footerSx = {
    borderTop: '1px solid rgba(255, 255, 255, 0.1)',
    textAlign: 'center' as const,
    px: 2,
    py: 3,
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.6)',
    transition: 'all 0.3s ease',
  };

  // Filter menu items based on user permissions
  const filteredMenuItems = menuItems.filter(item => {
    if (!item.permissions) return true;
    if (!identity?.level) return false;
    return item.permissions.includes(identity.level);
  });

  return (
    <Box
      dir={isRTL ? 'rtl' : 'ltr'}
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        backgroundColor: isDark ? '#1e293b' : '#1e40af',
        color: '#ffffff',
        pt: 0,
        transition: 'background-color 0.3s ease',
      }}
    >
      <Box sx={{ flexGrow: 1, overflowY: 'auto', pt: 1, marginTop: 2 }}>
        {filteredMenuItems.map((item) => (
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

      <Box sx={footerSx}>
        <Box component="div" sx={{ fontWeight: 600, mb: 0.5 }}>{translate('app.title')}</Box>
        <Box component="div">© {currentYear} ALL RIGHTS RESERVED</Box>
        {logout && <Box sx={{ mt: 2 }}>{logout}</Box>}
      </Box>
    </Box>
  );
};
