import BusinessIcon from '@mui/icons-material/Business';
import AssessmentIcon from '@mui/icons-material/Assessment';
import DomainOutlinedIcon from '@mui/icons-material/DomainOutlined';
import HowToRegIcon from '@mui/icons-material/HowToReg';
import StorageOutlinedIcon from '@mui/icons-material/StorageOutlined';
import MonitorIcon from '@mui/icons-material/Monitor';
import SettingsOutlinedIcon from '@mui/icons-material/SettingsOutlined';
import { Box, useTheme } from '@mui/material';
import { MenuItemLink, MenuProps, useTranslate, useLocale } from 'react-admin';

const RTL_LANGUAGES = ['ar', 'he', 'fa', 'ur'];

const platformMenuItems = [
  { to: '/platform/', labelKey: 'menu.platform_dashboard', icon: <BusinessIcon /> },
  { to: '/platform/monitoring', labelKey: 'menu.monitoring', icon: <MonitorIcon /> },
  { to: '/platform/monitoring/devices', labelKey: 'menu.device_health', icon: <AssessmentIcon /> },
  { to: '/platform/quotas', labelKey: 'menu.quotas', icon: <AssessmentIcon /> },
  { to: '/platform/providers', labelKey: 'menu.providers', icon: <DomainOutlinedIcon /> },
  { to: '/platform/providers/registrations', labelKey: 'menu.provider_registrations', icon: <HowToRegIcon /> },
  { to: '/platform/provider/backup', labelKey: 'menu.backups', icon: <StorageOutlinedIcon /> },
  { to: '/platform/settings', labelKey: 'menu.platform_settings', icon: <SettingsOutlinedIcon /> },
];

export const PlatformMenu = ({ dense, onMenuClick, logout }: MenuProps) => {
  const currentYear = new Date().getFullYear();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
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

      <Box sx={footerSx}>
        <Box component="div" sx={{ fontWeight: 600, mb: 0.5 }}>Platform Admin</Box>
        <Box component="div">© {currentYear} ALL RIGHTS RESERVED</Box>
        {logout && <Box sx={{ mt: 2 }}>{logout}</Box>}
      </Box>
    </Box>
  );
};
