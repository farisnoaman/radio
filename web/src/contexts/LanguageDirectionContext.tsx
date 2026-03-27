import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { useLocale } from 'react-admin';

type Direction = 'rtl' | 'ltr';

interface LanguageDirectionContextType {
  direction: Direction;
  isRTL: boolean;
  setDirection: (dir: Direction) => void;
}

const LanguageDirectionContext = createContext<LanguageDirectionContextType | undefined>(undefined);

const RTL_LOCALES = ['ar', 'he', 'fa', 'ur'];

export const useLanguageDirection = () => {
  const context = useContext(LanguageDirectionContext);
  if (!context) {
    throw new Error('useLanguageDirection must be used within a LanguageDirectionProvider');
  }
  return context;
};

interface LanguageDirectionProviderProps {
  children: ReactNode;
}

export const LanguageDirectionProvider: React.FC<LanguageDirectionProviderProps> = ({ children }) => {
  const locale = useLocale();
  const [direction, setDirection] = useState<Direction>(() => {
    // Initial direction based on default locale
    return RTL_LOCALES.includes(locale) ? 'rtl' : 'ltr';
  });

  useEffect(() => {
    // Update direction when locale changes
    const newDirection = RTL_LOCALES.includes(locale) ? 'rtl' : 'ltr';
    setDirection(newDirection);

    // Update document dir attribute
    document.documentElement.dir = newDirection;
    document.documentElement.lang = locale;

    // Update document body for complete RTL support
    document.body.dir = newDirection;
  }, [locale]);

  const value = {
    direction,
    isRTL: direction === 'rtl',
    setDirection,
  };

  return (
    <LanguageDirectionContext.Provider value={value}>
      {children}
    </LanguageDirectionContext.Provider>
  );
};
