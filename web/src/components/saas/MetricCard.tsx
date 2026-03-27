import { Box, Card, SxProps, Theme, Typography, useTheme } from '@mui/material';
import { TrendingUp, TrendingDown, Remove } from '@mui/icons-material';
import { useTranslate } from 'react-admin';

interface MetricCardProps {
  title: string;
  value: string | number;
  unit?: string;
  trend?: 'up' | 'down' | 'neutral';
  trendValue?: string;
  icon?: React.ReactNode;
  description?: string;
  sx?: SxProps<Theme>;
  variant?: 'default' | 'compact' | 'detailed';
}

export const MetricCard = ({
  title,
  value,
  unit,
  trend,
  trendValue,
  icon,
  description,
  sx,
  variant = 'default',
}: MetricCardProps) => {
  const theme = useTheme();
  const translate = useTranslate();

  const trendConfig = {
    up: { color: '#10b981', icon: TrendingUp, label: translate('platform.trend_up') },
    down: { color: '#ef4444', icon: TrendingDown, label: translate('platform.trend_down') },
    neutral: { color: '#6b7280', icon: Remove, label: translate('platform.trend_stable') },
  };

  const TrendIcon = trend ? trendConfig[trend].icon : null;

  return (
    <Card
      variant="elevation"
      sx={{
        p: variant === 'compact' ? 1.5 : 2,
        background: 'linear-gradient(135deg, rgba(255,255,255,0.9) 0%, rgba(248,250,252,0.95) 100%)',
        backdropFilter: 'blur(10px)',
        border: '1px solid rgba(148, 163, 184, 0.12)',
        borderRadius: 1.5,
        transition: 'all 0.2s ease',
        position: 'relative',
        overflow: 'hidden',
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        width: '100%',
        minWidth: 0,
        '&:hover': {
          transform: 'translateY(-1px)',
          boxShadow: '0 8px 16px -6px rgba(0, 0, 0, 0.12)',
          borderColor: 'rgba(148, 163, 184, 0.18)',
        },
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: '3px',
          background: trend ? trendConfig[trend].color : theme.palette.primary.main,
          opacity: 0,
          transition: 'opacity 0.2s ease',
        },
        '&:hover::before': {
          opacity: 1,
        },
        ...sx,
      }}
    >
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 1.5 }}>
        <Box sx={{ flex: 1, minWidth: 0 }}>
          <Typography
            variant="caption"
            sx={{
              fontWeight: 600,
              color: 'text.secondary',
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
              fontSize: { xs: '0.65rem', sm: '0.7rem' },
              mb: 0.25,
              display: 'block',
            }}
          >
            {title}
          </Typography>
        </Box>
        {icon && (
          <Box
            sx={{
              p: 0.5,
              borderRadius: 1,
              backgroundColor: 'rgba(30, 58, 138, 0.08)',
              color: '#1e3a8a',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
            }}
          >
            {icon}
          </Box>
        )}
      </Box>

      <Box sx={{ mb: 1.5, flexGrow: 1, minWidth: 0 }}>
        <Box sx={{ display: 'flex', alignItems: 'baseline', gap: 0.5, mb: 0.75, flexWrap: 'wrap' }}>
          <Typography
            variant="h5"
            sx={{
              fontWeight: 800,
              color: 'text.primary',
              fontSize: { xs: '1.35rem', sm: variant === 'compact' ? '1.5rem' : variant === 'detailed' ? '1.75rem' : '1.625rem' },
              letterSpacing: '-0.5px',
              lineHeight: 1.2,
            }}
          >
            {value}
          </Typography>
          {unit && (
            <Typography variant="caption" sx={{ color: 'text.secondary', fontSize: { xs: '0.7rem', sm: '0.75rem' }, fontWeight: 500 }}>
              {unit}
            </Typography>
          )}
        </Box>

        {trend && trendValue && (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.25, flexWrap: 'wrap' }}>
            {TrendIcon && (
              <TrendIcon
                sx={{
                  fontSize: 14,
                  color: trendConfig[trend].color,
                  flexShrink: 0,
                }}
              />
            )}
            <Typography
              variant="caption"
              sx={{
                color: trendConfig[trend].color,
                fontWeight: 700,
                fontSize: { xs: '0.7rem', sm: '0.75rem' },
              }}
            >
              {trendValue}
            </Typography>
            <Typography
              variant="caption"
              sx={{
                color: 'text.secondary',
                fontWeight: 500,
                fontSize: { xs: '0.65rem', sm: '0.7rem' },
              }}
            >
              {trendConfig[trend].label}
            </Typography>
          </Box>
        )}
      </Box>

      {description && variant === 'detailed' && (
        <Typography
          variant="caption"
          sx={{
            color: 'text.secondary',
            fontSize: { xs: '0.65rem', sm: '0.7rem' },
            lineHeight: 1.4,
            fontWeight: 500,
          }}
        >
          {description}
        </Typography>
      )}
    </Card>
  );
};

MetricCard.displayName = 'MetricCard';
