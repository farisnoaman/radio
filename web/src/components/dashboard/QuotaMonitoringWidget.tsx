import {
  Box,
  Card,
  CardContent,
  Typography,
  Stack,
  Chip,
} from '@mui/material';
import {
  Warning,
  Error as ErrorIcon,
} from '@mui/icons-material';
import { useTranslate } from 'react-admin';
import { useGetList } from 'react-admin';

export const QuotaMonitoringWidget = () => {
  const translate = useTranslate();
  const { data, isLoading } = useGetList('quotas', {
    pagination: { page: 1, perPage: 50 },
    sort: { field: 'id', order: 'ASC' },
  });

  if (isLoading) return null;

  // Find providers with quota issues
  const criticalProviders = data?.filter((p: any) => {
    const userPercent = p.utilization?.users_percent || 0;
    const sessionPercent = p.utilization?.sessions_percent || 0;
    return userPercent >= 100 || sessionPercent >= 100;
  }) || [];

  const warningProviders = data?.filter((p: any) => {
    const userPercent = p.utilization?.users_percent || 0;
    const sessionPercent = p.utilization?.sessions_percent || 0;
    return (userPercent >= 80 && userPercent < 100) || (sessionPercent >= 80 && sessionPercent < 100);
  }) || [];

  if (criticalProviders.length === 0 && warningProviders.length === 0) {
    return null;
  }

  return (
    <Card
      sx={{
        background: 'linear-gradient(135deg, rgba(239, 68, 68, 0.05) 0%, rgba(245, 158, 11, 0.05) 100%)',
        border: '1px solid rgba(239, 68, 68, 0.2)',
        borderRadius: 2,
      }}
    >
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
          {criticalProviders.length > 0 ? (
            <ErrorIcon sx={{ color: '#ef4444', fontSize: 24 }} />
          ) : (
            <Warning sx={{ color: '#f59e0b', fontSize: 24 }} />
          )}
          <Typography variant="h6" sx={{ fontWeight: 600 }}>
            {translate('quota.alerts')}
          </Typography>
        </Box>

        <Stack spacing={2}>
          {criticalProviders.slice(0, 3).map((provider: any) => (
            <Box
              key={provider.id}
              sx={{
                p: 2,
                borderRadius: 1,
                bgcolor: 'rgba(239, 68, 68, 0.1)',
                border: '1px solid rgba(239, 68, 68, 0.3)',
              }}
            >
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2" sx={{ fontWeight: 600 }}>
                  {provider.provider_name}
                </Typography>
                <Chip
                  label={translate('quota.quota_exceeded')}
                  size="small"
                  sx={{
                    bgcolor: '#ef4444',
                    color: 'white',
                    fontWeight: 600,
                  }}
                />
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                  {translate('quota.users')}: {provider.usage?.current_users} / {provider.quota?.max_users}
                </Typography>
                <Typography variant="caption" sx={{ color: '#ef4444', fontWeight: 600 }}>
                  {((provider.usage?.current_users / provider.quota?.max_users) * 100).toFixed(0)}%
                </Typography>
              </Box>
            </Box>
          ))}

          {warningProviders.slice(0, 2).map((provider: any) => (
            <Box
              key={provider.id}
              sx={{
                p: 2,
                borderRadius: 1,
                bgcolor: 'rgba(245, 158, 11, 0.1)',
                border: '1px solid rgba(245, 158, 11, 0.3)',
              }}
            >
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2" sx={{ fontWeight: 600 }}>
                  {provider.provider_name}
                </Typography>
                <Chip
                  label={translate('quota.approaching_limit')}
                  size="small"
                  sx={{
                    bgcolor: '#f59e0b',
                    color: 'white',
                    fontWeight: 600,
                  }}
                />
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                  {translate('quota.users')}: {provider.usage?.current_users} / {provider.quota?.max_users}
                </Typography>
                <Typography variant="caption" sx={{ color: '#f59e0b', fontWeight: 600 }}>
                  {((provider.usage?.current_users / provider.quota?.max_users) * 100).toFixed(0)}%
                </Typography>
              </Box>
            </Box>
          ))}
        </Stack>

        <Box sx={{ mt: 2, pt: 2, borderTop: '1px solid rgba(0,0,0,0.1)' }}>
          <Typography variant="caption" sx={{ color: 'text.secondary' }}>
            {translate('quota.last_updated')}: {new Date().toLocaleString()}
          </Typography>
        </Box>
      </CardContent>
    </Card>
  );
};

QuotaMonitoringWidget.displayName = 'QuotaMonitoringWidget';
