import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { createAppTheme } from '../theme';
import ar from '../i18n/ar';
import enUS from '../i18n/en-US';
import zhCN from '../i18n/zh-CN';

const translations = {
  'ar': ar,
  'en-US': enUS,
  'zh-CN': zhCN,
};

type Locale = keyof typeof translations;

interface I18nContextType {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  translate: (key: string, params?: Record<string, any>) => string;
}

const I18nContext = createContext<I18nContextType | undefined>(undefined);

export const LandingI18nProvider = ({ children }: { children: ReactNode }) => {
  const [locale, setLocale] = useState<Locale>('ar');

  // Set document direction when locale changes
  useEffect(() => {
    const direction = locale === 'ar' ? 'rtl' : 'ltr';
    document.documentElement.dir = direction;
    document.documentElement.lang = locale;
  }, [locale]);

  const translate = (key: string, params?: Record<string, any>): string => {
    const keys = key.split('.');
    let value: any = translations[locale];

    for (const k of keys) {
      value = value?.[k];
    }

    if (typeof value === 'string') {
      // Replace parameters like {name} with actual values
      if (params) {
        return value.replace(/\{(\w+)\}/g, (match, param) => params[param] || match);
      }
      return value;
    }

    return key; // Return key if translation not found
  };

  // Create theme with proper direction and Readex Pro font
  const direction = locale === 'ar' ? 'rtl' : 'ltr';
  const theme = createAppTheme('light', direction);

  return (
    <I18nContext.Provider value={{ locale, setLocale, translate }}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </ThemeProvider>
    </I18nContext.Provider>
  );
};

export const useLandingTranslate = () => {
  const context = useContext(I18nContext);
  if (!context) {
    throw new Error('useLandingTranslate must be used within LandingI18nProvider');
  }
  return context;
};
