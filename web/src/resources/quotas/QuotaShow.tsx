import {
  Show,
  SimpleShowLayout,
  TopToolbar,
  ListButton,
  useRecordContext,
} from 'react-admin';
import {
  Box,
  Card,
  CardContent,
  Typography,
  LinearProgress,
  Grid,
  useMediaQuery,
  useTheme,
  Button,
} from '@mui/material';
import {
  People,
  Router,
  Storage,
  TrendingUp,
  Warning,
  CheckCircle,
  Edit,
} from '@mui/icons-material';
import { useTranslate } from 'react-admin';
import { useNavigate } from 'react-router-dom';

const QuotaShowActions = () => {
  const translate = useTranslate();
  const record = useRecordContext();
  const navigate = useNavigate();

  return (
    <TopToolbar>
      <ListButton />
      <Button
        variant="contained"
        size="small"
        startIcon={<Edit />}
        onClick={() => navigate(`/platform/quotas/${record?.id}/edit`)}
      >
        {translate('quota.edit_quota')}
      </Button>
    </TopToolbar>
  );
};

const QuotaUsageCard = ({
  title,
  current,
  max,
  icon,
  color,
}: any) => {
  const translate = useTranslate();
  const percentage = max > 0 ? (current / max) * 100 : 0;

  const getStatusColor = (percent: number) => {
    if (percent >= 100) return '#ef4444';
    if (percent >= 80) return '#f59e0b';
    return '#10b981';
  };

  return (
    <Card
      sx={{
        background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(248,250,252,0.98) 100%)',
        border: '1px solid rgba(148, 163, 184, 0.1)',
        borderRadius: 2,
        height: '100%',
      }}
    >
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
          {icon}
          <Typography variant="h6" sx={{ fontWeight: 600 }}>
            {title}
          </Typography>
        </Box>

        <Box sx={{ mb: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('quota.current')}
            </Typography>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('quota.maximum')}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
            <Typography variant="h5" sx={{ fontWeight: 700, color }}>
              {current.toLocaleString()}
            </Typography>
            <Typography variant="h5" sx={{ fontWeight: 600, color: 'text.secondary' }}>
              {max.toLocaleString()}
            </Typography>
          </Box>
        </Box>

        <Box>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('quota.percent_used')}
            </Typography>
            <Typography
              variant="body2"
              sx={{ color: getStatusColor(percentage), fontWeight: 600 }}
            >
              {percentage.toFixed(1)}%
            </Typography>
          </Box>
          <LinearProgress
            variant="determinate"
            value={Math.min(percentage, 100)}
            sx={{
              height: 8,
              borderRadius: 4,
              backgroundColor: 'rgba(0,0,0,0.1)',
              '& .MuiLinearProgress-bar': {
                backgroundColor: getStatusColor(percentage),
              },
            }}
          />
        </Box>

        {percentage >= 80 && (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 2 }}>
            {percentage >= 100 ? (
              <Warning sx={{ color: '#ef4444', fontSize: 20 }} />
            ) : (
              <CheckCircle sx={{ color: '#f59e0b', fontSize: 20 }} />
            )}
            <Typography
              variant="body2"
              sx={{
                color: percentage >= 100 ? '#ef4444' : '#f59e0b',
                fontWeight: 600,
              }}
            >
              {percentage >= 100
                ? translate('quota.quota_exceeded')
                : translate('quota.approaching_limit')}
            </Typography>
          </Box>
        )}
      </CardContent>
    </Card>
  );
};

export const QuotaShow = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const translate = useTranslate();
  const record = useRecordContext();

  if (!record) return null;

  const quotaData = {
    max_users: record.max_users || 1000,
    max_online_users: record.max_users || 500,
    max_nas: record.max_nas || 100,
    max_storage: 100,
    max_bandwidth: 10,
    max_daily_backups: 5,
    max_auth_per_second: 100,
    max_acct_per_second: 200,
  };

  const usageData = {
    current_users: record.usage?.current_users || 0,
    current_online_users: record.usage?.current_online_users || 0,
    current_nas: record.usage?.current_nas || 0,
    current_storage: 45,
    current_bandwidth: 3.2,
    current_daily_backups: 2,
    current_auth_per_second: 65,
    current_acct_per_second: 150,
  };

  return (
    <Show actions={<QuotaShowActions />}>
      <SimpleShowLayout>
        <Box sx={{ mb: isMobile ? 2 : 4, px: isMobile ? 1 : 0 }}>
          <Typography
            variant={isMobile ? 'h6' : 'h4'}
            sx={{ fontWeight: 700, mb: 1 }}
          >
            {translate('quota.provider_quota')}
          </Typography>
          <Typography
            variant="body2"
            sx={{ color: 'text.secondary', display: { xs: 'none', sm: 'block' } }}
          >
            {record.name} ({record.code})
          </Typography>
        </Box>

        <Grid container spacing={isMobile ? 1.5 : 3}>
          <Grid size={{ xs: 12, sm: 6 }}>
            <QuotaUsageCard
              title={translate('quota.max_users')}
              current={usageData.current_users}
              max={quotaData.max_users}
              icon={<People sx={{ color: '#1e3a8a', fontSize: 28 }} />}
              color="#1e3a8a"
            />
          </Grid>

          <Grid size={{ xs: 12, sm: 6 }}>
            <QuotaUsageCard
              title={translate('quota.max_online_users')}
              current={usageData.current_online_users}
              max={quotaData.max_online_users}
              icon={<TrendingUp sx={{ color: '#059669', fontSize: 28 }} />}
              color="#059669"
            />
          </Grid>

          <Grid size={{ xs: 12, sm: 6 }}>
            <QuotaUsageCard
              title={translate('quota.max_nas')}
              current={usageData.current_nas}
              max={quotaData.max_nas}
              icon={<Router sx={{ color: '#7c3aed', fontSize: 28 }} />}
              color="#7c3aed"
            />
          </Grid>

          <Grid size={{ xs: 12, sm: 6 }}>
            <QuotaUsageCard
              title={translate('quota.max_storage')}
              current={usageData.current_storage}
              max={quotaData.max_storage}
              icon={<Storage sx={{ color: '#db2777', fontSize: 28 }} />}
              color="#db2777"
            />
          </Grid>

          <Grid size={{ xs: 12 }}>
            <Card
              sx={{
                background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(248,250,252,0.98) 100%)',
                border: '1px solid rgba(148, 163, 184, 0.1)',
                borderRadius: 2,
              }}
            >
              <CardContent>
                <Typography
                  variant="h6"
                  sx={{ fontWeight: 600, mb: isMobile ? 2 : 3, fontSize: isMobile ? '0.95rem' : undefined }}
                >
                  {translate('quota.resource_limits')}
                </Typography>

                <Grid container spacing={isMobile ? 1.5 : 2}>
                  {[
                    {
                      label: translate('quota.max_bandwidth'),
                      value: `${usageData.current_bandwidth} / ${quotaData.max_bandwidth} Gbps`,
                    },
                    {
                      label: translate('quota.max_daily_backups'),
                      value: `${usageData.current_daily_backups} / ${quotaData.max_daily_backups}`,
                    },
                    {
                      label: translate('quota.max_auth_per_second'),
                      value: `${usageData.current_auth_per_second} / ${quotaData.max_auth_per_second}`,
                    },
                    {
                      label: translate('quota.max_acct_per_second'),
                      value: `${usageData.current_acct_per_second} / ${quotaData.max_acct_per_second}`,
                    },
                  ].map((item, idx) => (
                    <Grid size={{ xs: 6, sm: 3 }} key={idx}>
                      <Box>
                        <Typography variant="body2" sx={{ color: 'text.secondary', mb: 0.5 }}>
                          {item.label}
                        </Typography>
                        <Typography
                          variant="body2"
                          sx={{ fontWeight: 600, fontSize: isMobile ? '0.8rem' : undefined }}
                        >
                          {item.value}
                        </Typography>
                      </Box>
                    </Grid>
                  ))}
                </Grid>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      </SimpleShowLayout>
    </Show>
  );
};

QuotaShow.displayName = 'QuotaShow';
