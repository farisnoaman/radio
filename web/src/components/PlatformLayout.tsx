import type { SxProps, Theme } from '@mui/material';
import { Layout, LayoutProps, Sidebar, useLocale } from 'react-admin';
import { Box } from '@mui/material';

import { PlatformMenu } from './PlatformMenu';

type PlatformLayoutProps = LayoutProps & { sx?: SxProps<Theme> };

const RTL_LANGUAGES = ['ar', 'he', 'fa', 'ur'];

const PlatformSidebar = (props: any) => {
  const locale = useLocale();
  const isRTL = RTL_LANGUAGES.includes(locale || '');

  return (
    <Sidebar
      {...props}
      sx={{
        // Position sidebar on the correct side depending on locale
        ...(isRTL
          ? {
              '& .MuiDrawer-paper': {
                right: 0,
                left: 'auto',
              },
            }
          : {}),
      }}
    />
  );
};

export const PlatformLayout = ({ sx, children, ...rest }: PlatformLayoutProps & { children?: React.ReactNode }) => {
  const locale = useLocale();
  const isRTL = RTL_LANGUAGES.includes(locale || '');

  return (
    <Layout
      {...rest}
      menu={PlatformMenu}
      sidebar={PlatformSidebar}
      sx={[
        {
          // Fixed top AppBar
          '& .MuiAppBar-root': {
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            zIndex: 1200,
          },
          // App frame: flex container for sidebar + content
          '& .RaLayout-appFrame': {
            marginTop: '48px',
            direction: isRTL ? 'rtl' : 'ltr',
          },
          // Fixed sidebar wrapper
          '& .RaSidebar-fixed': {
            position: 'fixed',
            top: '48px',
            height: 'calc(100vh - 48px)',
            overflowY: 'auto',
            ...(isRTL
              ? { right: 0, left: 'auto' }
              : { left: 0, right: 'auto' }),
          },
          // Content area
          '& .RaLayout-content': {
            position: 'relative',
            padding: { xs: 2, md: 3, lg: 4 },
            minHeight: 'calc(100vh - 48px)',
            transition: 'background-color 0.3s ease',
            bgcolor: 'background.default',
            direction: isRTL ? 'rtl' : 'ltr',
          },
        },
        ...(Array.isArray(sx) ? sx : sx ? [sx] : []),
      ]}
    >
      <Box dir={isRTL ? 'rtl' : 'ltr'} sx={{ minHeight: '100vh' }}>
        {children}
      </Box>
    </Layout>
  );
};
