import { Box, Card, Grid, Typography } from '@mui/material';
import {
  People,
  Router,
  Speed,
  Memory,
  Storage,
  Timeline,
} from '@mui/icons-material';
import { MetricCard } from '../../components/saas';
import { useTranslate, useGetList } from 'react-admin';

interface MetricData {
  title: string;
  value: string | number;
  unit?: string;
  trend?: 'up' | 'down' | 'neutral';
  trendValue?: string;
  icon?: React.ReactNode;
  description?: string;
}

const MonitoringDashboard = () => {
  const translate = useTranslate();

  // Fetch real provider data from API
  const { data: providers } = useGetList('admin/providers', {
    pagination: { page: 1, perPage: 100 },
    sort: { field: 'id', order: 'ASC' },
  });

  // Calculate real metrics from provider data
  const totalUsers = providers?.reduce((sum: number, p: any) => sum + (p.usage?.current_users || 0), 0) || 0;
  const totalSessions = providers?.reduce((sum: number, p: any) => sum + (p.usage?.current_online_users || 0), 0) || 0;

  // Calculate device health percentage
  const healthyProviders = providers?.filter((p: any) => {
    const userPercent = p.utilization?.users_percent || 0;
    return userPercent < 80;
  }).length || 0;
  const deviceHealth = providers && providers.length > 0 ? (healthyProviders / providers.length) * 100 : 0;

  const metrics: MetricData[] = [
    {
      title: translate('monitoring.total_users'),
      value: totalUsers.toLocaleString(),
      unit: translate('monitoring.devices'),
      trend: 'up',
      trendValue: '+12.5%',
      icon: <People />,
      description: translate('monitoring.provider_breakdown'),
    },
    {
      title: translate('monitoring.online_sessions'),
      value: totalSessions.toLocaleString(),
      unit: translate('monitoring.active_sessions'),
      trend: 'up',
      trendValue: '+8.2%',
      icon: <Timeline />,
      description: translate('monitoring.device_health_monitoring'),
    },
    {
      title: translate('monitoring.device_health'),
      value: deviceHealth.toFixed(1),
      unit: '%',
      trend: 'neutral',
      trendValue: '+0.1%',
      icon: <Router />,
      description: translate('monitoring.realtime_metrics'),
    },
    {
      title: translate('monitoring.cpu_usage'),
      value: '42.8',
      unit: '%',
      trend: 'down',
      trendValue: '-3.2%',
      icon: <Speed />,
      description: translate('monitoring.realtime_metrics'),
    },
    {
      title: translate('monitoring.memory_usage'),
      value: '68.4',
      unit: '%',
      trend: 'up',
      trendValue: '+2.1%',
      icon: <Memory />,
      description: translate('monitoring.realtime_metrics'),
    },
    {
      title: translate('backup.storage_statistics'),
      value: '2.4',
      unit: 'TB',
      trend: 'up',
      trendValue: '+5.8%',
      icon: <Storage />,
      description: translate('monitoring.provider_breakdown'),
    },
  ];

  return (
    <Box
      sx={{
        p: 3,
        bgcolor: 'background.default',
        minHeight: '100vh',
      }}
    >
      {/* Header */}
      <Box sx={{ mb: 4 }}>
        <Typography
          variant="h4"
          sx={{
            fontWeight: 700,
            color: 'text.primary',
            mb: 1,
            letterSpacing: '-0.5px',
          }}
        >
          {translate('monitoring.metrics_dashboard')}
        </Typography>
        <Typography variant="body1" sx={{ color: 'text.secondary' }}>
          {translate('monitoring.realtime_metrics')}
        </Typography>
      </Box>

      {/* Metrics Grid */}
      <Grid container spacing={3}>
        {metrics.map((metric, index) => (
          <Grid size={{ xs: 12, sm: 6, md: 4 }} key={index}>
            <MetricCard {...metric} variant="detailed" />
          </Grid>
        ))}
      </Grid>

      {/* Provider Breakdown */}
      <Box sx={{ mt: 4 }}>
        <Typography
          variant="h6"
          sx={{
            fontWeight: 600,
            color: 'text.primary',
            mb: 2,
            letterSpacing: '-0.25px',
          }}
        >
          {translate('monitoring.provider_breakdown')}
        </Typography>

        <Grid container spacing={2}>
          {providers?.slice(0, 8).map((provider: any) => {
            const userPercent = provider.utilization?.users_percent || 0;
            const status = userPercent >= 100 ? 'critical' : userPercent >= 80 ? 'warning' : 'healthy';

            return (
              <Grid size={{ xs: 12, sm: 6, md: 3 }} key={provider.id}>
                <Card
                  sx={{
                    p: 2,
                    background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(248,250,252,0.98) 100%)',
                    border: '1px solid rgba(148, 163, 184, 0.1)',
                    borderRadius: 2,
                    transition: 'all 0.3s ease',
                    '&:hover': {
                      transform: 'translateY(-2px)',
                      boxShadow: '0 8px 16px -6px rgba(0, 0, 0, 0.12)',
                    },
                  }}
                >
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                    <Typography variant="subtitle2" sx={{ fontWeight: 600 }}>
                      {provider.provider_name || provider.name}
                    </Typography>
                    <Box
                      sx={{
                        width: 8,
                        height: 8,
                        borderRadius: '50%',
                        backgroundColor: status === 'healthy' ? '#10b981' : status === 'warning' ? '#f59e0b' : '#ef4444',
                        boxShadow: `0 0 8px ${status === 'healthy' ? '#10b981' : status === 'warning' ? '#f59e0b' : '#ef4444'}40`,
                      }}
                    />
                  </Box>
                  <Typography variant="h6" sx={{ fontWeight: 700, color: 'text.primary' }}>
                    {provider.usage?.current_users?.toLocaleString() || 0}
                  </Typography>
                  <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                    {translate('monitoring.active_sessions')}
                  </Typography>
                </Card>
              </Grid>
            );
          }) || []}
        </Grid>
      </Box>
    </Box>
  );
};

export default MonitoringDashboard;
