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

import { Box, useTheme } from '@mui/material';
import { MenuItemLink, MenuProps, useGetIdentity, useTranslate, useLocale } from 'react-admin';

const RTL_LANGUAGES = ['ar', 'he', 'fa', 'ur'];

const menuItems = [
  { to: '/', labelKey: 'menu.dashboard', icon: <DashboardOutlinedIcon />, permissions: ['super', 'admin', 'user'] },
  { to: '/network/servers', labelKey: 'menu.servers', icon: <StorageOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/network/nodes', labelKey: 'menu.network_nodes', icon: <AccountTreeOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/network/nas', labelKey: 'menu.nas_devices', icon: <RouterOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/cpes', labelKey: 'menu.cpe_devices', icon: <RouterOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/users', labelKey: 'menu.radius_users', icon: <PeopleAltOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/profiles', labelKey: 'menu.radius_profiles', icon: <SettingsSuggestOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/online', labelKey: 'menu.online_sessions', icon: <SensorsOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/accounting', labelKey: 'menu.accounting', icon: <ReceiptLongOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/invoices', labelKey: 'menu.invoices', icon: <ReceiptLongOutlinedIcon />, permissions: ['super', 'admin', 'user'] },
  { to: '/system/config', labelKey: 'menu.system_config', icon: <SettingsOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/system/operators', labelKey: 'menu.operators', icon: <AdminPanelSettingsOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/system/logs', labelKey: 'menu.system_logs', icon: <HistoryOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/products', labelKey: 'menu.products', icon: <Inventory2OutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/agents', labelKey: 'menu.agents', icon: <SupportAgentOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/financial/performance', labelKey: 'menu.financial_performance', icon: <AssessmentOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/settings/tunnel', labelKey: 'menu.tunnel_settings', icon: <VpnKeyOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/system/maintenance', labelKey: 'menu.maintenance', icon: <ConstructionIcon />, permissions: ['super', 'admin'] },
  { to: '/voucher-batches', labelKey: 'menu.vouchers', icon: <ConfirmationNumberOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/voucher-printing', labelKey: 'menu.print_vouchers', icon: <PrintIcon />, permissions: ['super', 'admin'] },
  { to: '/portal/devices', labelKey: 'portal.my_devices', icon: <DevicesIcon />, permissions: ['user'] },
  { to: '/portal/vouchers/redeem', labelKey: 'portal.redeem_voucher', icon: <ConfirmationNumberOutlinedIcon />, permissions: ['user'] },
];

export const CustomMenu = ({ dense, onMenuClick, logout }: MenuProps) => {
  const currentYear = new Date().getFullYear();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { data: identity } = useGetIdentity();
  const translate = useTranslate();
  const locale = useLocale();
  const isRTL = RTL_LANGUAGES.includes(locale || '');

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
        <div style={{ fontWeight: 600, marginBottom: 4 }}>TOUGHRADIUS v9</div>
        <div>© {currentYear} ALL RIGHTS RESERVED</div>
        {logout && <Box sx={{ mt: 2 }}>{logout}</Box>}
      </Box>
    </Box>
  );
};

