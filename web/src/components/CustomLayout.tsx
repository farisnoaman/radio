import type { SxProps, Theme } from '@mui/material';
import { Layout, LayoutProps, TitlePortal } from 'react-admin';
import { Box, Typography, useMediaQuery } from '@mui/material';

import { CustomAppBar } from './CustomAppBar';
import { CustomMenu } from './CustomMenu';

type CustomLayoutProps = LayoutProps & { sx?: SxProps<Theme> };

export const CustomLayout = ({ sx, children, ...rest }: CustomLayoutProps & { children?: React.ReactNode }) => {
  const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
  return (
    <Layout
      {...rest}
      appBar={CustomAppBar}
      menu={CustomMenu}
      sx={[
        {
          // 固定顶部 AppBar（滚动时不隐藏）
          '& .MuiAppBar-root': {
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            zIndex: 1200,
          },
          // 为固定的 AppBar 留出空间
          '& .RaLayout-appFrame': {
            marginTop: '48px',
          },
          // 固定左侧菜单（滚动时不跟随移动）
          '& .RaSidebar-fixed': {
            position: 'fixed',
            top: '48px',
            height: 'calc(100vh - 48px)',
            overflowY: 'auto',
          },
          // 内容区域样式
          '& .RaLayout-content': {
            position: 'relative',
            padding: { xs: 2, md: 3, lg: 4 },
            minHeight: 'calc(100vh - 48px)',
            transition: 'background-color 0.3s ease',
          },
        },
        ...(Array.isArray(sx) ? sx : sx ? [sx] : []),
      ]}
    >
      {isSmall && (
        <Box sx={{
          position: 'absolute',
          top: 16,
          left: 16,
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
    </Layout>
  );
};
