import {
  Edit,
  SimpleForm,
  NumberInput,
  useUpdate,
  useNotify,
  useRefresh,
  useRecordContext,
} from 'react-admin';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Alert,
  AlertTitle,
  Stack,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { Business, Router } from '@mui/icons-material';
import { useTranslate } from 'react-admin';
import { useNavigate } from 'react-router-dom';

export const QuotaEdit = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const [update] = useUpdate();
  const notify = useNotify();
  const refresh = useRefresh();
  const translate = useTranslate();
  const navigate = useNavigate();
  const record = useRecordContext();

  const handleSubmit = async (data: any) => {
    try {
      await update('quotas', {
        id: record?.id ?? data.id,
        data: {
          tenant_id: data.tenant_id,
          max_users: data.max_users,
          max_online_users: data.max_online_users,
          max_nas: data.max_nas,
          max_storage: data.max_storage,
          max_bandwidth: data.max_bandwidth,
          max_daily_backups: data.max_daily_backups,
          max_auth_per_second: data.max_auth_per_second,
          max_acct_per_second: data.max_acct_per_second,
        },
        previousData: data,
      });
      notify(translate('quota.quota_updated'), { type: 'success' });
      navigate(`/platform/quotas/${record?.id ?? data.id}/show`);
      refresh();
    } catch (error: any) {
      notify(translate('quota.quota_error'), { type: 'error' });
    }
  };

  return (
    <Edit redirect={() => `/platform/quotas/${record?.id}/show`}>
      <SimpleForm onSubmit={handleSubmit}>
        <Box sx={{ p: isMobile ? 1.5 : 3 }}>
          <Box sx={{ mb: isMobile ? 2 : 4 }}>
            <Typography
              variant={isMobile ? 'h6' : 'h4'}
              sx={{ fontWeight: 700, mb: 1 }}
            >
              {translate('quota.edit_quota')}
            </Typography>
            <Typography
              variant="body2"
              sx={{ color: 'text.secondary', display: { xs: 'none', sm: 'block' } }}
            >
              {translate('quota.manage')}
            </Typography>
          </Box>

          <Alert severity="info" sx={{ mb: isMobile ? 2 : 3, fontSize: isMobile ? '0.75rem' : undefined }}>
            <AlertTitle sx={{ fontSize: isMobile ? '0.8rem' : undefined }}>
              {translate('quota.quota_details')}
            </AlertTitle>
          </Alert>

          <Grid container spacing={isMobile ? 2 : 3}>
            <Grid size={{ xs: 12 }}>
              <Card
                sx={{
                  background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(248,250,252,0.98) 100%)',
                  border: '1px solid rgba(148, 163, 184, 0.1)',
                  borderRadius: 2,
                }}
              >
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: isMobile ? 2 : 3 }}>
                    <Business sx={{ color: '#1e3a8a' }} />
                    <Typography
                      variant="h6"
                      sx={{ fontWeight: 600, fontSize: isMobile ? '0.95rem' : undefined }}
                    >
                      {translate('quota.resource_limits')}
                    </Typography>
                  </Box>

                  <Stack spacing={isMobile ? 2 : 3}>
                    <NumberInput
                      source="max_users"
                      label={translate('quota.max_users')}
                      defaultValue={1000}
                      min={1}
                      fullWidth
                    />
                    <NumberInput
                      source="max_online_users"
                      label={translate('quota.max_online_users')}
                      defaultValue={500}
                      min={1}
                      fullWidth
                    />
                    <NumberInput
                      source="max_nas"
                      label={translate('quota.max_nas')}
                      defaultValue={100}
                      min={1}
                      fullWidth
                    />
                    <NumberInput
                      source="max_storage"
                      label={translate('quota.max_storage')}
                      defaultValue={100}
                      min={1}
                      fullWidth
                    />
                  </Stack>
                </CardContent>
              </Card>
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
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: isMobile ? 2 : 3 }}>
                    <Router sx={{ color: '#7c3aed' }} />
                    <Typography
                      variant="h6"
                      sx={{ fontWeight: 600, fontSize: isMobile ? '0.95rem' : undefined }}
                    >
                      RADIUS {translate('quota.resource_limits')}
                    </Typography>
                  </Box>

                  <Grid container spacing={isMobile ? 1.5 : 2}>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <NumberInput
                        source="max_bandwidth"
                        label={translate('quota.max_bandwidth')}
                        defaultValue={10}
                        min={1}
                        fullWidth
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <NumberInput
                        source="max_daily_backups"
                        label={translate('quota.max_daily_backups')}
                        defaultValue={5}
                        min={1}
                        fullWidth
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <NumberInput
                        source="max_auth_per_second"
                        label={translate('quota.max_auth_per_second')}
                        defaultValue={100}
                        min={1}
                        fullWidth
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <NumberInput
                        source="max_acct_per_second"
                        label={translate('quota.max_acct_per_second')}
                        defaultValue={200}
                        min={1}
                        fullWidth
                      />
                    </Grid>
                  </Grid>
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        </Box>
      </SimpleForm>
    </Edit>
  );
};

QuotaEdit.displayName = 'QuotaEdit';
