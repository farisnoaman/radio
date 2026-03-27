import { IconButton, Menu, MenuItem, ListItemIcon, ListItemText, Box } from '@mui/material';
import LanguageIcon from '@mui/icons-material/Language';
import { useState } from 'react';
import { useLandingTranslate } from '../contexts/LandingI18nProvider';

type Locale = 'ar' | 'en-US' | 'zh-CN';

export const LandingLanguageSwitcher = () => {
  const { locale, setLocale } = useLandingTranslate();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const locales: Locale[] = ['ar', 'en-US', 'zh-CN'];

  const languageNames: Record<Locale, string> = {
    'ar': 'العربية',
    'en-US': 'English',
    'zh-CN': '简体中文',
  };

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLanguageChange = (newLocale: Locale) => {
    setLocale(newLocale);
    handleClose();
  };

  // Get current language display name
  const currentLanguageName = languageNames[locale] || locale;

  return (
    <>
      <IconButton
        onClick={handleClick}
        size="small"
        sx={{
          color: 'white',
          bgcolor: 'rgba(255, 255, 255, 0.1)',
          '&:hover': {
            bgcolor: 'rgba(255, 255, 255, 0.2)',
          },
        }}
      >
        <LanguageIcon fontSize="small" />
        <Box component="span" sx={{ ml: 1, fontSize: '0.875rem' }}>
          {currentLanguageName}
        </Box>
      </IconButton>
      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        onClick={handleClose}
      >
        {locales.map((loc) => (
          <MenuItem
            key={loc}
            selected={loc === locale}
            onClick={() => handleLanguageChange(loc)}
          >
            <ListItemIcon>
              <LanguageIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText>{languageNames[loc]}</ListItemText>
          </MenuItem>
        ))}
      </Menu>
    </>
  );
};
