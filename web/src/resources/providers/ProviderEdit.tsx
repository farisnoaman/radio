import {
  Edit,
  SimpleForm,
  TextInput,
  NumberInput,
  SelectInput,
  useUpdate,
  useNotify,
  useRedirect,
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
  CircularProgress,
} from '@mui/material';
import { Business, People } from '@mui/icons-material';
import { useTranslate } from 'react-admin';

export const ProviderEdit = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const translate = useTranslate();
  const record = useRecordContext();
  const [update] = useUpdate();
  const notify = useNotify();
  const redirect = useRedirect();
  const refresh = useRefresh();

  const handleSubmit = async (data: any) => {
    if (!record) return;
    try {
      await update('providers', { id: record.id, data });
      notify(translate('provider.update_success'), { type: 'success' });
      redirect('show', 'providers', record.id);
      refresh();
    } catch (error: any) {
      notify(translate('provider.error_create', { error: error.message }), { type: 'error' });
    }
  };

  if (!record) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Edit redirect="show">
      <SimpleForm onSubmit={handleSubmit}>
        <Box sx={{ p: isMobile ? 1.5 : 3 }}>
          <Box sx={{ mb: isMobile ? 2 : 4 }}>
            <Typography
              variant={isMobile ? 'h6' : 'h4'}
              sx={{ fontWeight: 700, mb: 1 }}
            >
              {translate('provider.edit_provider')}
            </Typography>
            <Typography
              variant="body2"
              sx={{ color: 'text.secondary', display: { xs: 'none', sm: 'block' } }}
            >
              {translate('provider.basic_information')}
            </Typography>
          </Box>

          <Alert
            severity="info"
            sx={{ mb: isMobile ? 2 : 3, fontSize: isMobile ? '0.75rem' : undefined }}
          >
            <AlertTitle sx={{ fontSize: isMobile ? '0.8rem' : undefined }}>
              {translate('provider.provider_information')}
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
                      {translate('provider.basic_information')}
                    </Typography>
                  </Box>

                  <Stack spacing={isMobile ? 2 : 3}>
                    <TextInput
                      source="name"
                      label={translate('provider.name')}
                      fullWidth
                      required
                    />
                    <TextInput
                      source="code"
                      label={translate('provider.code')}
                      fullWidth
                      required
                      helperText={translate('provider.code_help')}
                    />
                    <SelectInput
                      source="status"
                      label={translate('provider.status')}
                      choices={[
                        { id: 'active', name: translate('provider.active') },
                        { id: 'suspended', name: translate('provider.suspended') },
                      ]}
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
                    <People sx={{ color: '#1e3a8a' }} />
                    <Typography
                      variant="h6"
                      sx={{ fontWeight: 600, fontSize: isMobile ? '0.95rem' : undefined }}
                    >
                      {translate('platform_settings.default_quotas')}
                    </Typography>
                  </Box>

                  <Grid container spacing={isMobile ? 1.5 : 2}>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <NumberInput
                        source="max_users"
                        label={translate('provider.max_users')}
                        fullWidth
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <NumberInput
                        source="max_nas"
                        label={translate('provider.max_nas')}
                        fullWidth
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <NumberInput
                        source="max_storage"
                        label={translate('provider.max_storage')}
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

ProviderEdit.displayName = 'ProviderEdit';
