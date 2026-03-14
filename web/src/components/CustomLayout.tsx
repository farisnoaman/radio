import type { SxProps, Theme } from '@mui/material';
import { Layout, LayoutProps, Sidebar, TitlePortal, useLocale } from 'react-admin';
import { Box, Typography, useMediaQuery } from '@mui/material';

import { CustomAppBar } from './CustomAppBar';
import { CustomMenu } from './CustomMenu';

type CustomLayoutProps = LayoutProps & { sx?: SxProps<Theme> };

const RTL_LANGUAGES = ['ar', 'he', 'fa', 'ur'];

const CustomSidebar = (props: any) => {
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

export const CustomLayout = ({ sx, children, ...rest }: CustomLayoutProps & { children?: React.ReactNode }) => {
  const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
  const locale = useLocale();
  const isRTL = RTL_LANGUAGES.includes(locale || '');

  return (
    <Layout
      {...rest}
      appBar={CustomAppBar}
      menu={CustomMenu}
      sidebar={CustomSidebar}
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
            direction: isRTL ? 'rtl' : 'ltr',
          },
        },
        ...(Array.isArray(sx) ? sx : sx ? [sx] : []),
      ]}
    >
      <Box dir={isRTL ? 'rtl' : 'ltr'} sx={{ minHeight: '100vh' }}>
        {isSmall && (
          <Box sx={{
            position: 'absolute',
            top: 16,
            ...(isRTL ? { right: 16 } : { left: 16 }),
            zIndex: 10,
            maxWidth: 'calc(100% - 130px)',
            pointerEvents: 'none'
          }}>
            <Typography component="div" sx={{
              fontWeight: 'bold',
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              pointerEvents: 'auto',
              color: 'text.primary',
              fontSize: '1.25rem',
              lineHeight: 1.6
            }}>
              <TitlePortal />
            </Typography>
          </Box>
        )}
        {children}
      </Box>
    </Layout>
  );
};
