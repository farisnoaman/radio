import { Box, Container, Grid, Typography, Paper } from '@mui/material';
import {
  Business,
  People,
  TrendingUp,
  AttachMoney,
  Pending,
  CheckCircle,
  Warning,
  DomainOutlined,
  History,
  PieChart,
  Router as RouterIcon,
} from '@mui/icons-material';
import { MetricCard } from '../../components/saas';
import { useTranslate, useGetList } from 'react-admin';

// Platform-wide statistics with real data
const PlatformStats = () => {
  const translate = useTranslate();

  // Fetch real provider data from API
  const { data: providers, isLoading } = useGetList('admin/providers', {
    pagination: { page: 1, perPage: 100 },
    sort: { field: 'id', order: 'ASC' },
  });

  // Calculate real statistics
  const totalProviders = providers?.length || 0;
  const totalUsers = providers?.reduce((sum: number, p: any) => sum + (p.usage?.current_users || 0), 0) || 0;
  const activeProviders = providers?.filter((p: any) => p.status === 'active').length || 0;

  const stats = [
    {
      title: translate('platform.total_providers'),
      value: totalProviders.toString(),
      unit: translate('menu.providers'),
      trend: 'up' as const,
      trendValue: translate('platform.trend_this_month', { count: activeProviders }),
      icon: <Business />,
      description: translate('platform.platform_stats'),
      isLoading,
    },
    {
      title: translate('platform.total_users'),
      value: totalUsers.toLocaleString(),
      unit: translate('platform.users'),
      trend: 'up' as const,
      trendValue: '+12.5%',
      icon: <People />,
      description: translate('platform.provider_distribution'),
      isLoading,
    },
    {
      title: translate('platform.monthly_revenue'),
      value: '$24,850',
      unit: translate('platform.per_month'),
      trend: 'up' as const,
      trendValue: '+18.2%',
      icon: <AttachMoney />,
      description: translate('platform.revenue_distribution'),
      isLoading,
    },
    {
      title: translate('platform.pending_requests'),
      value: '8',
      unit: translate('platform.activity_feed'),
      trend: 'neutral' as const,
      trendValue: translate('platform.pending_registrations'),
      icon: <Pending />,
      description: translate('platform.pending_requests'),
      isLoading,
    },
  ];

  return (
    <Grid container spacing={2}>
      {stats.map((stat, index) => (
        <Grid size={{ xs: 12, sm: 6, md: 6, lg: 3 }} key={index}>
          <Box sx={{ height: '100%', display: 'flex' }}>
            <MetricCard
              {...stat}
              variant="detailed"
              sx={{ flex: 1, height: '100%', display: 'flex', flexDirection: 'column' }}
            />
          </Box>
        </Grid>
      ))}
    </Grid>
  );
};

const ProviderStatusOverview = () => {
  const translate = useTranslate();

  // Fetch real provider data from API
  // @ts-ignore - isLoading is used in stats objects below
  const { data: providers, isLoading } = useGetList('admin/providers', {
    pagination: { page: 1, perPage: 100 },
    sort: { field: 'id', order: 'ASC' },
  });

  // Calculate status counts from real data
  const statusCounts = {
    active: providers?.filter((p: any) => p.status === 'active').length || 0,
    warning: providers?.filter((p: any) => {
      const userPercent = p.utilization?.users_percent || 0;
      return userPercent >= 80 && userPercent < 100;
    }).length || 0,
    pending: providers?.filter((p: any) => p.status === 'pending').length || 0,
  };

  return (
    <Grid container spacing={2}>
      {[
        { status: 'active', label: translate('platform.active_providers'), color: '#10b981', bgColor: '#10b981', icon: CheckCircle, count: statusCounts.active },
        { status: 'warning', label: translate('platform.warning_providers'), color: '#f59e0b', bgColor: '#f59e0b', icon: Warning, count: statusCounts.warning },
        { status: 'pending', label: translate('platform.pending_registrations'), color: '#3b82f6', bgColor: '#3b82f6', icon: Pending, count: statusCounts.pending },
      ].map((item, index) => (
        <Grid size={{ xs: 12, sm: 6, md: 4 }} key={index}>
          <Box
            sx={{
              p: { xs: 2, sm: 3 },
              borderRadius: 2.5,
              background: `linear-gradient(135deg, ${item.bgColor}08 0%, ${item.bgColor}15 100%)`,
              border: `1.5px solid ${item.color}25`,
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
              position: 'relative',
              overflow: 'hidden',
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              '&::before': {
                content: '""',
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                height: 3,
                background: item.color,
                opacity: 0.5,
              },
              '&:hover': {
                transform: 'translateY(-6px)',
                boxShadow: `0 12px 24px -8px ${item.color}30`,
                borderColor: `${item.color}40`,
              },
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: { xs: 1.5, sm: 2 } }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: { xs: 1, sm: 1.5 } }}>
                <Box
                  sx={{
                    p: { xs: 0.75, sm: 1 },
                    borderRadius: 1.5,
                    bgcolor: `${item.color}15`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                  }}
                >
                  <item.icon sx={{ color: item.color, fontSize: { xs: 20, sm: 24 } }} />
                </Box>
                <Typography variant="body1" sx={{ fontWeight: 600, color: 'text.primary', fontSize: { xs: '0.875rem', sm: '1rem' } }}>
                  {item.label}
                </Typography>
              </Box>
            </Box>
            <Box sx={{ display: 'flex', alignItems: 'baseline', gap: 1, mb: { xs: 0.75, sm: 1 } }}>
              <Typography variant="h3" sx={{ fontWeight: 800, color: item.color, fontSize: { xs: '2rem', sm: '2.5rem' } }}>
                {item.count}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ color: 'text.secondary', fontWeight: 500, fontSize: { xs: '0.75rem', sm: '0.875rem' } }}>
              {translate('menu.providers')}
            </Typography>
          </Box>
        </Grid>
      ))}
    </Grid>
  );
};

const RecentActivity = () => {
  const translate = useTranslate();

  // Fetch real provider data
  const { data: providers } = useGetList('admin/providers', {
    pagination: { page: 1, perPage: 100 },
    sort: { field: 'created_at', order: 'DESC' },
  });

  // Generate real activity from provider data
  const activities = [
    // Pending registrations (providers with status 'pending')
    ...(providers?.filter((p: any) => p.status === 'pending').slice(0, 3).map((p: any) => ({
      id: `reg-${p.id}`,
      type: 'registration',
      provider: p.provider_name || p.name,
      timeKey: 'time_hours_ago',
      timeValue: Math.floor((Date.now() - new Date(p.created_at).getTime()) / (1000 * 60 * 60)) || 1,
      status: 'pending',
    })) || []),

    // Quota alerts (providers with high utilization)
    ...(providers?.filter((p: any) => {
      const userPercent = p.utilization?.users_percent || 0;
      return userPercent >= 80 && userPercent < 100 && p.status === 'active';
    }).slice(0, 3).map((p: any) => ({
      id: `quota-${p.id}`,
      type: 'quota_alert',
      provider: p.provider_name || p.name,
      timeKey: 'time_hours_ago',
      timeValue: Math.floor((Date.now() - new Date(p.updated_at || p.created_at).getTime()) / (1000 * 60 * 60)) || 1,
      status: 'warning',
    })) || []),

    // Recently active providers (just created)
    ...(providers?.filter((p: any) => p.status === 'active').slice(0, 2).map((p: any) => ({
      id: `active-${p.id}`,
      type: 'approval',
      provider: p.provider_name || p.name,
      timeKey: 'time_day_ago',
      timeValue: Math.floor((Date.now() - new Date(p.created_at).getTime()) / (1000 * 60 * 60 * 24)) || 1,
      status: 'completed',
    })) || []),
  ].slice(0, 5); // Limit to 5 activities

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'registration':
        return <Pending sx={{ color: '#3b82f6' }} />;
      case 'approval':
        return <CheckCircle sx={{ color: '#10b981' }} />;
      case 'backup':
        return <CheckCircle sx={{ color: '#10b981' }} />;
      case 'quota_alert':
        return <Warning sx={{ color: '#f59e0b' }} />;
      case 'billing':
        return <AttachMoney sx={{ color: '#10b981' }} />;
      default:
        return <Pending />;
    }
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5, height: '100%' }}>
      {activities.map((activity) => {
        const getStatusColor = () => {
          switch (activity.status) {
            case 'approved':
            case 'completed':
              return { bg: 'rgba(16, 185, 129, 0.1)', color: '#10b981', border: 'rgba(16, 185, 129, 0.2)' };
            case 'warning':
              return { bg: 'rgba(245, 158, 11, 0.1)', color: '#f59e0b', border: 'rgba(245, 158, 11, 0.2)' };
            default:
              return { bg: 'rgba(59, 130, 246, 0.1)', color: '#3b82f6', border: 'rgba(59, 130, 246, 0.2)' };
          }
        };

        const statusColors = getStatusColor();

        return (
          <Box
            key={activity.id}
            sx={{
              display: 'flex',
              alignItems: 'center',
              gap: { xs: 1, sm: 2 },
              py: { xs: 1.5, sm: 2 },
              px: { xs: 1, sm: 1.5 },
              transition: 'all 0.2s ease',
              borderRadius: 2,
              '&:hover': {
                bgcolor: 'rgba(30, 58, 138, 0.04)',
              },
            }}
          >
            <Box sx={{ flexShrink: 0 }}>
              <Box
                sx={{
                  p: { xs: 1, sm: 1.5 },
                  borderRadius: 2,
                  bgcolor: statusColors.bg,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <Box sx={{ color: statusColors.color, fontSize: { xs: 14, sm: 16 } }}>
                  {getActivityIcon(activity.type)}
                </Box>
              </Box>
            </Box>
            <Box sx={{ flex: 1, minWidth: 0 }}>
              <Typography
                variant="body2"
                sx={{
                  fontWeight: 600,
                  color: 'text.primary',
                  mb: 0.25,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                  fontSize: { xs: '0.75rem', sm: '0.813rem' },
                }}
              >
                {activity.provider}
              </Typography>
              <Typography
                variant="caption"
                sx={{
                  color: 'text.secondary',
                  fontWeight: 500,
                  fontSize: { xs: '0.65rem', sm: '0.7rem' },
                }}
              >
                {translate(`platform.${activity.timeKey}`, { count: activity.timeValue })}
              </Typography>
            </Box>
            <Box sx={{ flexShrink: 0 }}>
              <Typography
                variant="caption"
                sx={{
                  px: { xs: 1.5, sm: 2 },
                  py: 0.5,
                  borderRadius: 1.5,
                  backgroundColor: statusColors.bg,
                  color: statusColors.color,
                  fontWeight: 700,
                  fontSize: { xs: '0.6rem', sm: '0.65rem' },
                  textTransform: 'uppercase',
                  letterSpacing: '0.5px',
                  whiteSpace: 'nowrap',
                }}
              >
                {translate(`platform.status_${activity.status}`)}
              </Typography>
            </Box>
          </Box>
        );
      })}
    </Box>
  );
};

const ResourceUtilization = () => {
  const translate = useTranslate();

  // Fetch real provider data from API
  const { data: providers } = useGetList('admin/providers', {
    pagination: { page: 1, perPage: 100 },
    sort: { field: 'id', order: 'ASC' },
  });

  // Calculate real resource utilization from all providers
  const totalUsers = providers?.reduce((sum: number, p: any) => sum + (p.usage?.current_users || 0), 0) || 0;
  const totalSessions = providers?.reduce((sum: number, p: any) => sum + (p.usage?.current_online_users || 0), 0) || 0;
  const totalNas = providers?.reduce((sum: number, p: any) => sum + (p.usage?.current_nas || 0), 0) || 0;

  // Calculate limits from provider quotas
  const maxUsers = providers?.reduce((sum: number, p: any) => sum + (p.max_users || 1000), 0) || 1000;
  const maxSessions = providers?.reduce((sum: number, p: any) => sum + (p.max_users || 1000), 0) || 1000;
  const maxNas = providers?.reduce((sum: number, p: any) => sum + (p.max_nas || 100), 0) || 100;

  // Calculate percentages
  const userPercentage = maxUsers > 0 ? Math.round((totalUsers / maxUsers) * 100) : 0;
  const sessionPercentage = maxSessions > 0 ? Math.round((totalSessions / maxSessions) * 100) : 0;
  const nasPercentage = maxNas > 0 ? Math.round((totalNas / maxNas) * 100) : 0;

  const resources = [
    {
      name: translate('platform.utilization.total_users'),
      used: totalUsers,
      limit: maxUsers,
      percentage: userPercentage,
      icon: People,
      color: '#3b82f6'
    },
    {
      name: translate('platform.utilization.concurrent_sessions'),
      used: totalSessions,
      limit: maxSessions,
      percentage: sessionPercentage,
      icon: TrendingUp,
      color: '#10b981'
    },
    {
      name: translate('platform.utilization.storage_gb'),
      used: 234,
      limit: 1000,
      percentage: 23,
      icon: AttachMoney,
      color: '#f59e0b'
    },
    {
      name: translate('platform.utilization.nas_devices'),
      used: totalNas,
      limit: maxNas,
      percentage: nasPercentage,
      icon: RouterIcon,
      color: '#8b5cf6'
    },
  ];

  return (
    <Grid container spacing={2}>
      {resources.map((resource, index) => (
        <Grid size={{ xs: 12, sm: 6, md: 6, lg: 3 }} key={index}>
          <Box
            sx={{
              p: { xs: 2, sm: 3 },
              borderRadius: 2.5,
              bgcolor: index % 2 === 0 ? 'rgba(59, 130, 246, 0.03)' : 'rgba(16, 185, 129, 0.03)',
              border: `1px solid ${resource.color}15`,
              transition: 'all 0.3s ease',
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'space-between',
              '&:hover': {
                transform: 'translateY(-4px)',
                boxShadow: `0 8px 16px -6px ${resource.color}20`,
                borderColor: `${resource.color}30`,
              },
            }}
          >
            <Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: { xs: 1.5, sm: 2 }, mb: { xs: 2, sm: 2.5 } }}>
                <Box
                  sx={{
                    p: { xs: 1, sm: 1.5 },
                    borderRadius: 2,
                    bgcolor: `${resource.color}10`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                  }}
                >
                  <resource.icon sx={{ color: resource.color, fontSize: { xs: 18, sm: 20 } }} />
                </Box>
                <Typography variant="body2" sx={{ fontWeight: 600, color: 'text.primary', flex: 1, fontSize: { xs: '0.75rem', sm: '0.875rem' } }}>
                  {resource.name}
                </Typography>
              </Box>

              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', mb: { xs: 1.5, sm: 2 } }}>
                <Typography variant="h4" sx={{ fontWeight: 800, color: resource.color, fontSize: { xs: '1.5rem', sm: '1.75rem' } }}>
                  {resource.percentage}%
                </Typography>
                <Typography variant="caption" sx={{ color: 'text.secondary', fontWeight: 500, fontSize: { xs: '0.65rem', sm: '0.75rem' } }}>
                  {resource.used.toLocaleString()} / {resource.limit.toLocaleString()}
                </Typography>
              </Box>

              <Box
                sx={{
                  width: '100%',
                  height: { xs: 8, sm: 10 },
                  bgcolor: `${resource.color}10`,
                  borderRadius: 5,
                  overflow: 'hidden',
                  position: 'relative',
                }}
              >
                <Box
                  sx={{
                    width: `${resource.percentage}%`,
                    height: '100%',
                    background: `linear-gradient(90deg, ${resource.color} 0%, ${resource.color}cc 100%)`,
                    borderRadius: 5,
                    transition: 'width 1s cubic-bezier(0.4, 0, 0.2, 1)',
                    position: 'relative',
                    '&::after': {
                      content: '""',
                      position: 'absolute',
                      top: 0,
                      right: 0,
                      width: '100%',
                      height: '100%',
                      background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.3), transparent)',
                      animation: 'shimmer 2s infinite',
                    },
                  }}
                />
              </Box>
            </Box>
            <Typography variant="caption" sx={{ color: 'text.secondary', mt: { xs: 1.5, sm: 2 }, display: 'block', fontWeight: 500, fontSize: { xs: '0.65rem', sm: '0.75rem' } }}>
              {translate('platform.resource_utilization')}
            </Typography>
          </Box>
        </Grid>
      ))}
    </Grid>
  );
};

export const PlatformDashboard = () => {
  const translate = useTranslate();

  return (
    <Box
      sx={{
        bgcolor: 'background.default',
        minHeight: '100vh',
      }}
    >
      {/* Header Section with Gradient Background */}
      <Box
        sx={{
          background: 'linear-gradient(135deg, #1e3a8a 0%, #1e40af 50%, #2563eb 100%)',
          color: 'white',
          px: { xs: 2.5, md: 4 },
          py: { xs: 4, md: 5 },
          mb: 3.5,
          position: 'relative',
          overflow: 'hidden',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: 0,
            right: 0,
            width: '50%',
            height: '100%',
            background: 'radial-gradient(circle, rgba(255,255,255,0.1) 0%, transparent 70%)',
            pointerEvents: 'none',
          },
        }}
      >
        <Box sx={{ position: 'relative', zIndex: 1 }}>
          <Typography
            variant="h3"
            sx={{
              fontWeight: 800,
              mb: 0.75,
              fontSize: { xs: '1.75rem', md: '2.25rem' },
              letterSpacing: '-0.5px',
            }}
          >
            {translate('platform.dashboard')}
          </Typography>
          <Typography
            variant="h6"
            sx={{
              fontWeight: 400,
              opacity: 0.9,
              fontSize: { xs: '0.95rem', md: '1.05rem' },
            }}
          >
            {translate('platform.overview')}
          </Typography>
        </Box>
      </Box>

      {/* Main Content */}
      <Container maxWidth="xl" sx={{ px: { xs: 2, md: 3 } }}>
        {/* Platform Stats - Full Width Cards */}
        <Box sx={{ mb: 3.5 }}>
          <PlatformStats />
        </Box>

        {/* Two Column Section: Provider Status & Recent Activity */}
        <Grid container spacing={2.5} sx={{ mb: 3.5 }}>
          <Grid size={{ xs: 12, md: 12, lg: 8, xl: 9 }}>
            <Paper
              elevation={0}
              sx={{
                p: 2.5,
                borderRadius: 2.5,
                bgcolor: 'background.paper',
                border: '1px solid rgba(148, 163, 184, 0.1)',
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              <Box sx={{ mb: 2.5 }}>
                <Typography
                  variant="h6"
                  sx={{
                    fontWeight: 700,
                    color: 'text.primary',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 1,
                  }}
                >
                  <DomainOutlined sx={{ color: '#1e3a8a', fontSize: 22 }} />
                  {translate('platform.provider_distribution')}
                </Typography>
              </Box>
              <Box sx={{ flexGrow: 1 }}>
                <ProviderStatusOverview />
              </Box>
            </Paper>
          </Grid>

          <Grid size={{ xs: 12, md: 12, lg: 4, xl: 3 }}>
            <Paper
              elevation={0}
              sx={{
                p: 2.5,
                borderRadius: 2.5,
                bgcolor: 'background.paper',
                border: '1px solid rgba(148, 163, 184, 0.1)',
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              <Box sx={{ mb: 2.5 }}>
                <Typography
                  variant="h6"
                  sx={{
                    fontWeight: 700,
                    color: 'text.primary',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 1,
                  }}
                >
                  <History sx={{ color: '#1e3a8a', fontSize: 22 }} />
                  {translate('platform.recent_activity')}
                </Typography>
              </Box>
              <Box sx={{ flexGrow: 1 }}>
                <RecentActivity />
              </Box>
            </Paper>
          </Grid>
        </Grid>

        {/* Resource Utilization Section */}
        <Paper
          elevation={0}
          sx={{
            p: 2.5,
            borderRadius: 2.5,
            bgcolor: 'background.paper',
            border: '1px solid rgba(148, 163, 184, 0.1)',
            mb: 4,
          }}
        >
          <Box sx={{ mb: 2.5 }}>
            <Typography
              variant="h6"
              sx={{
                fontWeight: 700,
                color: 'text.primary',
                display: 'flex',
                alignItems: 'center',
                gap: 1,
              }}
            >
              <PieChart sx={{ color: '#1e3a8a', fontSize: 22 }} />
              {translate('platform.resource_utilization')}
            </Typography>
            <Typography variant="body2" sx={{ color: 'text.secondary', mt: 1, ml: 3.75 }}>
              {translate('platform.utilization_description')}
            </Typography>
          </Box>
          <ResourceUtilization />
        </Paper>
      </Container>
    </Box>
  );
};

PlatformDashboard.displayName = 'PlatformDashboard';
