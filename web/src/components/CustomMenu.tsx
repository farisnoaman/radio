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
import CloudUploadOutlinedIcon from '@mui/icons-material/CloudUploadOutlined';
import ConstructionIcon from '@mui/icons-material/Construction';
import { Box, useTheme } from '@mui/material';
import { MenuItemLink, MenuProps, useGetIdentity, useTranslate } from 'react-admin';

const menuItems = [
  { to: '/', labelKey: 'menu.dashboard', icon: <DashboardOutlinedIcon /> }, // Everyone
  { to: '/network/nodes', labelKey: 'menu.network_nodes', icon: <AccountTreeOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/network/nas', labelKey: 'menu.nas_devices', icon: <RouterOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/users', labelKey: 'menu.radius_users', icon: <PeopleAltOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/profiles', labelKey: 'menu.radius_profiles', icon: <SettingsSuggestOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/online', labelKey: 'menu.online_sessions', icon: <SensorsOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/radius/accounting', labelKey: 'menu.accounting', icon: <ReceiptLongOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/system/config', labelKey: 'menu.system_config', icon: <SettingsOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/system/operators', labelKey: 'menu.operators', icon: <AdminPanelSettingsOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/system/logs', labelKey: 'Activity Logs', icon: <HistoryOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/products', labelKey: 'Products', icon: <Inventory2OutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/agents', labelKey: 'Agents', icon: <SupportAgentOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/financial/performance', labelKey: 'Financial Performance', icon: <AssessmentOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/settings/tunnel', labelKey: 'Tunnel Settings', icon: <VpnKeyOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/settings/backup', labelKey: 'Backup Settings', icon: <CloudUploadOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/system/maintenance', labelKey: 'menu.maintenance', icon: <ConstructionIcon />, permissions: ['super', 'admin'] },
  { to: '/voucher-batches', labelKey: 'Vouchers', icon: <ConfirmationNumberOutlinedIcon /> }, // Everyone
];

export const CustomMenu = ({ dense, onMenuClick, logout }: MenuProps) => {
  const currentYear = new Date().getFullYear();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { data: identity } = useGetIdentity();
  const translate = useTranslate();

  // 根据用户权限过滤菜单项
  const filteredMenuItems = menuItems.filter(item => {
    if (!item.permissions) return true; // 无权限限制的菜单项对所有人可见
    if (!identity?.level) return false; // 未登录用户不显示需要权限的菜单
    return item.permissions.includes(identity.level); // 检查用户权限是否在允许列表中
  });

  return (
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        // 侧边栏根据主题使用不同背景色
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
