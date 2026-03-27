import { MenuItem, ListItemIcon, ListItemText } from '@mui/material';
import { useSetLocale, useLocaleState, useTranslate } from 'react-admin';
import LanguageIcon from '@mui/icons-material/Language';

export const LanguageSwitcher = () => {
  const setLocale = useSetLocale();
  const [locale] = useLocaleState();
  const translate = useTranslate();

  // Available locales with cycle order
  const locales = ['ar', 'en-US', 'zh-CN'];

  // Language display names (native names)
  const languageNames: Record<string, string> = {
    'ar': 'العربية',
    'en-US': 'English',
    'zh-CN': '简体中文',
  };

  const handleLanguageChange = () => {
    // Find current locale index
    const currentIndex = locales.indexOf(locale);
    // Move to next locale (wrap around to 0 if at end)
    const nextIndex = (currentIndex + 1) % locales.length;
    const newLocale = locales[nextIndex];
    setLocale(newLocale);

    // Update localStorage to persist language preference
    localStorage.setItem('locale', newLocale);
  };

  // Get current language display name
  const currentLanguageName = languageNames[locale] || locale;

  return (
    <MenuItem onClick={handleLanguageChange}>
      <ListItemIcon>
        <LanguageIcon fontSize="small" />
      </ListItemIcon>
      <ListItemText>
        {translate('appbar.switch_language', { _: 'Switch Language' })}: {currentLanguageName}
      </ListItemText>
    </MenuItem>
  );
};
