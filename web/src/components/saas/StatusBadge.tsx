import { Box, SxProps, Theme } from '@mui/material';

interface StatusBadgeProps {
  status: 'online' | 'offline' | 'warning' | 'processing' | 'success' | 'error';
  label?: string;
  size?: 'small' | 'medium' | 'large';
  sx?: SxProps<Theme>;
}

const statusColors = {
  online: '#10b981',
  offline: '#6b7280',
  warning: '#f59e0b',
  processing: '#3b82f6',
  success: '#059669',
  error: '#ef4444',
};

export const StatusBadge = ({ status, label, size = 'medium', sx }: StatusBadgeProps) => {
  const sizes = {
    small: { height: 6, width: 6 },
    medium: { height: 8, width: 8 },
    large: { height: 10, width: 10 },
  };

  return (
    <Box
      sx={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 1,
        ...sx,
      }}
    >
      <Box
        sx={{
          ...sizes[size],
          borderRadius: '50%',
          backgroundColor: statusColors[status],
          boxShadow: `0 0 8px ${statusColors[status]}40`,
          animation: status === 'processing' ? 'pulse 2s ease-in-out infinite' : 'none',
          '@keyframes pulse': {
            '0%, 100%': { opacity: 1 },
            '50%': { opacity: 0.5 },
          },
        }}
      />
      {label && (
        <Box
          sx={{
            fontSize: size === 'small' ? 12 : size === 'large' ? 16 : 14,
            fontWeight: 500,
            color: statusColors[status],
            textTransform: 'capitalize',
          }}
        >
          {label}
        </Box>
      )}
    </Box>
  );
};

StatusBadge.displayName = 'StatusBadge';
