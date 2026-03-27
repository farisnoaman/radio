import polyglotI18nProvider from 'ra-i18n-polyglot';
import zhCN from './zh-CN';
import enUS from './en-US';
import ar from './ar';

const translations: Record<string, typeof zhCN> = {
  'zh-CN': zhCN,
  'en-US': enUS,
  'ar': ar,
};

// 从 localStorage 获取保存的语言设置，如果没有则使用默认语言
const getDefaultLocale = () => {
  const savedLocale = localStorage.getItem('locale');
  return savedLocale && translations[savedLocale as keyof typeof translations]
    ? savedLocale
    : 'ar';
};

const baseI18nProvider = polyglotI18nProvider(
  (locale) => translations[locale as keyof typeof translations] || translations['ar'],
  getDefaultLocale(), // 使用保存的语言或默认语言
  [
    { locale: 'zh-CN', name: '简体中文' },
    { locale: 'en-US', name: 'English' },
    { locale: 'ar', name: 'العربية' },
  ],
  { allowMissing: true }
);

// 包装 i18nProvider 以在切换语言时保存到 localStorage
export const i18nProvider = {
  ...baseI18nProvider,
  changeLocale: (locale: string) => {
    // 保存语言设置到 localStorage
    localStorage.setItem('locale', locale);
    return baseI18nProvider.changeLocale(locale);
  },
};

// 为公开页面创建的 i18nProvider（强制使用指定语言，不从 localStorage 读取）
export const createPublicI18nProvider = (defaultLocale: string = 'ar') => {
  // 强制使用指定的默认语言，忽略 localStorage
  const publicI18nProvider = polyglotI18nProvider(
    (locale) => translations[locale as keyof typeof translations] || translations[defaultLocale],
    defaultLocale,
    [
      { locale: 'zh-CN', name: '简体中文' },
      { locale: 'en-US', name: 'English' },
      { locale: 'ar', name: 'العربية' },
    ],
    { allowMissing: true }
  );

  return {
    ...publicI18nProvider,
    changeLocale: (locale: string) => {
      // 不保存到 localStorage，保持公开页面独立性
      return publicI18nProvider.changeLocale(locale);
    },
  };
};
