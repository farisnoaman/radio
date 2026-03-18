import { useLocation, useNavigate } from 'react-router-dom';
import SettingsOutlinedIcon from '@mui/icons-material/SettingsOutlined';
import LanguageIcon from '@mui/icons-material/Language';
import MenuIcon from '@mui/icons-material/Menu';
import MenuOpenIcon from '@mui/icons-material/MenuOpen';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { Box, IconButton, Stack, Tooltip, Typography, useTheme, Menu, MenuItem, ListItemIcon, ListItemText, useMediaQuery, Theme } from '@mui/material';
import { AppBar, AppBarProps, TitlePortal, ToggleThemeButton, useRedirect, useGetIdentity, useSetLocale, useLocaleState, useTranslate, useSidebarState } from 'react-admin';
import { useState } from 'react';

export const CustomAppBar = (props: AppBarProps) => {
  const redirect = useRedirect();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { data: identity } = useGetIdentity();
  const setLocale = useSetLocale();
  const [locale] = useLocaleState();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const translate = useTranslate();
  const [sidebarOpen, setSidebarOpen] = useSidebarState();
  const location = useLocation();
  const navigate = useNavigate();
  const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));

  const handleLanguageClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleLanguageClose = () => {
    setAnchorEl(null);
  };

  const handleLanguageSelect = (newLocale: string) => {
    setLocale(newLocale);
    handleLanguageClose();
  };


  // Determine if we are on a page that should have a back button (not a top-level list page)
  const isSubPage = location.pathname !== '/' && location.pathname.split('/').length > 2;

  const handleBack = () => {
    navigate(-1);
  };

  const handleToggleSidebar = () => {
    setSidebarOpen(!sidebarOpen);
  };

  return (
    <AppBar
      {...props}
      toolbar={false}
      elevation={0}
      alwaysOn={true}
      sx={{
        // 浅色主题使用白色背景，深色主题使用深色背景
        backgroundColor: isDark ? '#1e293b' : '#ffffff',
        color: isDark ? '#f1f5f9' : '#1f2937',
        borderBottom: isDark
          ? '1px solid rgba(148, 163, 184, 0.2)'
          : '1px solid rgba(229, 231, 235, 0.8)',
        boxShadow: isDark
          ? 'none'
          : '0 1px 3px 0 rgba(0, 0, 0, 0.05)',
        transition: 'all 0.3s ease',
        '& #react-admin-title': {
          fontSize: { xs: '1.1rem', sm: '1.25rem' },
          fontWeight: 700,
          whiteSpace: 'nowrap',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          maxWidth: { xs: '140px', sm: '300px' },
          flex: 1,
          textAlign: 'center',
          color: isDark ? '#f1f5f9' : '#1f2937',
          position: 'absolute',
          left: '50%',
          transform: 'translateX(-50%)',
          letterSpacing: '0.3px',
        },
        // 隐藏默认的汉堡菜单按钮，我们自己添加
        '& .RaAppBar-menuButton': {
          display: 'none',
        },
      }}
    >
      {!isSmall && <TitlePortal />}
      <Box
        sx={{
          width: '100%',
          px: 2,
          py: 0.5,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Stack direction="row" spacing={0} alignItems="center">
          {isSubPage ? (
            <Tooltip title={translate('ra.action.back') || 'Back'}>
              <IconButton
                size="medium"
                onClick={handleBack}
                sx={{
                  color: isDark ? '#f1f5f9' : '#6b7280',
                  mr: 1
                }}
              >
                <ArrowBackIcon />
              </IconButton>
            </Tooltip>
          ) : (
            <Tooltip title={sidebarOpen ? translate('appbar.collapse_menu') : translate('appbar.expand_menu')}>
              <IconButton
                size="medium"
                onClick={handleToggleSidebar}
                sx={{
                  color: isDark ? '#f1f5f9' : '#6b7280',
                  mr: 1,
                  transition: 'all 0.2s ease',
                  '&:hover': {
                    backgroundColor: isDark
                      ? 'rgba(255, 255, 255, 0.1)'
                      : 'rgba(0, 0, 0, 0.05)',
                  },
                }}
              >
                {sidebarOpen ? <MenuOpenIcon /> : <MenuIcon />}
              </IconButton>
            </Tooltip>
          )}

          {/* Logo only on larger screens when it's not a subpage */}
          <Box sx={{ display: { xs: isSubPage ? 'none' : 'block', sm: 'block' } }}>
            <Typography
              variant="h6"
              sx={{
                fontSize: 18,
                fontWeight: 700,
                color: isDark ? '#f1f5f9' : '#1f2937',
                letterSpacing: '0.5px',
              }}
            >
              {translate('app.title')}
            </Typography>
          </Box>
        </Stack>

        <Stack direction="row" spacing={1} alignItems="center">
          <Tooltip title={translate('appbar.switch_language')}>
            <IconButton
              size="large"
              onClick={handleLanguageClick}
              sx={{
                color: isDark ? '#f1f5f9' : '#6b7280',
                transition: 'all 0.2s ease',
                '&:hover': {
                  transform: 'scale(1.05)',
                  backgroundColor: isDark
                    ? 'rgba(255, 255, 255, 0.1)'
                    : 'rgba(0, 0, 0, 0.05)',
                },
              }}
            >
              <LanguageIcon />
            </IconButton>
          </Tooltip>
          <Menu
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={handleLanguageClose}
            anchorOrigin={{
              vertical: 'bottom',
              horizontal: 'right',
            }}
            transformOrigin={{
              vertical: 'top',
              horizontal: 'right',
            }}
          >
            <MenuItem
              onClick={() => handleLanguageSelect('zh-CN')}
              selected={locale === 'zh-CN'}
            >
              <ListItemIcon>
                {locale === 'zh-CN' && '✓'}
              </ListItemIcon>
              <ListItemText>{translate('appbar.language.zh_CN')}</ListItemText>
            </MenuItem>
            <MenuItem
              onClick={() => handleLanguageSelect('en-US')}
              selected={locale === 'en-US'}
            >
              <ListItemIcon>
                {locale === 'en-US' && '✓'}
              </ListItemIcon>
              <ListItemText>{translate('appbar.language.en_US')}</ListItemText>
            </MenuItem>
            <MenuItem
              onClick={() => handleLanguageSelect('ar')}
              selected={locale === 'ar'}
            >
              <ListItemIcon>
                {locale === 'ar' && '✓'}
              </ListItemIcon>
              <ListItemText>{translate('appbar.language.ar')}</ListItemText>
            </MenuItem>
          </Menu>

          <Tooltip title={translate('appbar.toggle_theme')}>
            <Box
              sx={{
                '& svg': {
                  fontSize: 22,
                  color: isDark ? '#f1f5f9' : '#6b7280',
                },
                '& button': {
                  color: isDark ? '#f1f5f9' : '#6b7280',
                  transition: 'all 0.2s ease',
                  '&:hover': {
                    transform: 'rotate(180deg)',
                    backgroundColor: isDark
                      ? 'rgba(255, 255, 255, 0.1)'
                      : 'rgba(0, 0, 0, 0.05)',
                  },
                },
              }}
            >
              <ToggleThemeButton />
            </Box>
          </Tooltip>

          {/* 只对超级管理员和管理员显示系统设置按钮 */}
          {identity?.level === 'super' || identity?.level === 'admin' ? (
            <Tooltip title={translate('appbar.system_settings')}>
              <IconButton
                size="large"
                onClick={() => redirect('/system/config')}
                sx={{
                  color: isDark ? '#f1f5f9' : '#6b7280',
                  transition: 'all 0.2s ease',
                  '&:hover': {
                    transform: 'scale(1.05)',
                    backgroundColor: isDark
                      ? 'rgba(255, 255, 255, 0.1)'
                      : 'rgba(0, 0, 0, 0.05)',
                  },
                }}
              >
                <SettingsOutlinedIcon />
              </IconButton>
            </Tooltip>
          ) : null}

        </Stack>
      </Box>
    </AppBar>
  );
};
