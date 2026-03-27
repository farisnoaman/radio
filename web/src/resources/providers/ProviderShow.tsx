import {
  Show,
  SimpleShowLayout,
  DateField,
  TopToolbar,
  ListButton,
  useRecordContext,
  useRefresh,
} from 'react-admin';
import {
  Box,
  Card,
  CardContent,
  Grid,
  Typography,
  Divider,
  Stack,
  Button,
  LinearProgress,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { Business, People, TrendingUp, OpenInNew, Edit } from '@mui/icons-material';
import { useTranslate } from 'react-admin';
import { useNavigate } from 'react-router-dom';
import { useCallback } from 'react';
import { AdminCredentialsCard } from '../../components/admin';

const ProviderShowActions = () => {
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
        onClick={() => navigate(`/platform/providers/${record?.id}/edit`)}
      >
        {translate('provider.edit')}
      </Button>
    </TopToolbar>
  );
};

const ProviderInfo = () => {
  const record = useRecordContext();
  const translate = useTranslate();

  if (!record) return null;

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
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 3 }}>
          <Business sx={{ color: '#1e3a8a', fontSize: 28 }} />
          <Typography variant="h6" sx={{ fontWeight: 600 }}>
            {translate('provider.title')}
          </Typography>
        </Box>

        <Stack spacing={2}>
          <Box>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('provider.name')}
            </Typography>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              {record.name}
            </Typography>
          </Box>

          <Box>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('provider.code')}
            </Typography>
            <Typography variant="body1" sx={{ fontWeight: 500 }}>
              {record.code}
            </Typography>
          </Box>

          <Box>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('provider.status')}
            </Typography>
            <Typography variant="body1" sx={{ fontWeight: 500, textTransform: 'capitalize' }}>
              {translate(`provider.${record.status}`)}
            </Typography>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
};

const ProviderQuotas = () => {
  const record = useRecordContext();
  const translate = useTranslate();

  if (!record) return null;

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
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 3 }}>
          <People sx={{ color: '#1e3a8a', fontSize: 28 }} />
          <Typography variant="h6" sx={{ fontWeight: 600 }}>
            {translate('platform_settings.default_quotas')}
          </Typography>
        </Box>

        <Stack spacing={3}>
          <Box>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('provider.max_users')}
            </Typography>
            <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
              {record.max_users?.toLocaleString()}
            </Typography>
          </Box>

          <Divider />

          <Box>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('provider.max_nas')}
            </Typography>
            <Typography variant="h5" sx={{ fontWeight: 600, color: 'text.primary' }}>
              {record.max_nas?.toLocaleString()}
            </Typography>
          </Box>

          <Divider />

          <Box>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('provider.max_storage')}
            </Typography>
            <Typography variant="h5" sx={{ fontWeight: 600, color: 'text.primary' }}>
              {record.max_storage || 'N/A'}
            </Typography>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
};

export const ProviderShow = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const translate = useTranslate();
  const refresh = useRefresh();

  const handleRefresh = useCallback(() => {
    refresh();
  }, [refresh]);

  return (
    <Show actions={<ProviderShowActions />}>
      <SimpleShowLayout>
        <Grid container spacing={isMobile ? 1.5 : 3}>
          <Grid size={{ xs: 12, sm: 6 }}>
            <ProviderInfo />
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <ProviderQuotas />
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
                  {translate('provider.created_at')}
                </Typography>
                <DateField source="created_at" showTime />
              </CardContent>
            </Card>
          </Grid>

          <Grid size={{ xs: 12 }}>
            <Box sx={{ mt: 3 }}>
              <AdminCredentialsCard
                onRefresh={handleRefresh}
              />
            </Box>
          </Grid>

          <Grid size={{ xs: 12 }}>
            <Card
              sx={{
                background: 'linear-gradient(135deg, rgba(16, 185, 129, 0.08) 0%, rgba(16, 185, 129, 0.02) 100%)',
                border: '1px solid rgba(16, 185, 129, 0.2)',
                borderRadius: 2,
              }}
            >
              <CardContent>
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 2,
                    mb: isMobile ? 2 : 3,
                    justifyContent: 'space-between',
                    flexDirection: isMobile ? 'column' : 'row',
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <TrendingUp sx={{ color: '#10b981', fontSize: 28 }} />
                    <Typography
                      variant="h6"
                      sx={{ fontWeight: 600, fontSize: isMobile ? '0.95rem' : undefined }}
                    >
                      {translate('quota.current_usage')}
                    </Typography>
                  </Box>
                  <Button
                    startIcon={<OpenInNew />}
                    onClick={() => {}}
                    sx={{ textTransform: 'none' }}
                    size="small"
                  >
                    {translate('quota.view_usage')}
                  </Button>
                </Box>

                <Grid container spacing={isMobile ? 1.5 : 2}>
                  {[
                    { label: translate('quota.users'), value: '850 / 1,000', progress: 85 },
                    { label: translate('quota.online_sessions'), value: '420 / 500', progress: 84 },
                    { label: translate('quota.nas_devices'), value: '75 / 100', progress: 75 },
                    { label: translate('quota.storage'), value: '45 / 100 GB', progress: 45 },
                  ].map((item, idx) => (
                    <Grid size={{ xs: 6, sm: 6, md: 3 }} key={idx}>
                      <Box>
                        <Typography
                          variant="body2"
                          sx={{ color: 'text.secondary', mb: 0.5, fontSize: isMobile ? '0.7rem' : undefined }}
                        >
                          {item.label}
                        </Typography>
                        <Typography
                          variant="body2"
                          sx={{ fontWeight: 700, fontSize: isMobile ? '0.8rem' : undefined }}
                        >
                          {item.value}
                        </Typography>
                        <LinearProgress
                          variant="determinate"
                          value={item.progress}
                          sx={{
                            mt: 1,
                            height: 6,
                            borderRadius: 3,
                          }}
                        />
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

ProviderShow.displayName = 'ProviderShow';
