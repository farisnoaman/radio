import {
  Create,
  SimpleForm,
  TextInput,
  NumberInput,
  SelectInput,
  useCreate,
  useNotify,
  useRedirect,
  useRefresh,
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
  Button,
} from '@mui/material';
import { Business, People } from '@mui/icons-material';
import { useTranslate } from 'react-admin';
import { useState } from 'react';
import { AdminCredentialsSection } from '../../components/admin';

export const ProviderCreate = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const [create] = useCreate();
  const notify = useNotify();
  const redirect = useRedirect();
  const refresh = useRefresh();
  const translate = useTranslate();
  const [adminUsername, setAdminUsername] = useState('admin');
  const [adminPassword, setAdminPassword] = useState('123456');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const getErrorMessage = (error: any) => {
    if (error?.body?.code === 'PROVIDER_EXISTS') {
      return 'A provider with this code already exists';
    }
    if (error?.body?.code === 'USERNAME_EXISTS') {
      return 'An operator with this username already exists';
    }
    if (error?.body?.code === 'INVALID_CREDENTIALS') {
      return 'Invalid admin credentials format';
    }
    return error?.message || 'An error occurred while creating the provider';
  };

  const handleSubmit = async (data: any) => {
    setLoading(true);
    setError(null);
    try {
      const providerData = {
        ...data,
        admin_username: adminUsername,
        admin_password: adminPassword,
      };
      await create('providers', { data: providerData });
      notify(translate('provider.create_success'), { type: 'success' });
      redirect('list', 'providers');
      refresh();
    } catch (error: any) {
      const errorMessage = getErrorMessage(error);
      setError(errorMessage);
      notify(translate('provider.error_create', { error: errorMessage }), { type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Create redirect="list">
      <SimpleForm onSubmit={handleSubmit}>
        <Box sx={{ p: isMobile ? 1.5 : 3 }}>
          <Box sx={{ mb: isMobile ? 2 : 4 }}>
            <Typography
              variant={isMobile ? 'h6' : 'h4'}
              sx={{ fontWeight: 700, mb: 1 }}
            >
              {translate('provider.create')}
            </Typography>
            <Typography
              variant="body2"
              sx={{ color: 'text.secondary', display: { xs: 'none', sm: 'block' } }}
            >
              {translate('provider.create_description')}
            </Typography>
          </Box>

          {loading && (
            <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', p: 3 }}>
              <CircularProgress size={24} />
              <Typography sx={{ ml: 2 }}>Creating provider...</Typography>
            </Box>
          )}

          {error && (
            <Alert
              severity="error"
              sx={{ mb: isMobile ? 2 : 3 }}
              action={
                <Button
                  color="inherit"
                  size="small"
                  onClick={() => setError(null)}
                >
                  Dismiss
                </Button>
              }
            >
              {error}
            </Alert>
          )}

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
                      defaultValue="active"
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
                        defaultValue={1000}
                        fullWidth
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <NumberInput
                        source="max_nas"
                        label={translate('provider.max_nas')}
                        defaultValue={100}
                        fullWidth
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 6 }}>
                      <NumberInput
                        source="max_storage"
                        label={translate('provider.max_storage')}
                        defaultValue={100}
                        fullWidth
                      />
                    </Grid>
                  </Grid>
                </CardContent>
              </Card>
            </Grid>

            <Grid size={{ xs: 12 }}>
              <Box sx={{ mt: isMobile ? 2 : 3 }}>
                <AdminCredentialsSection
                  username={adminUsername}
                  password={adminPassword}
                  onUsernameChange={setAdminUsername}
                  onPasswordChange={setAdminPassword}
                />
              </Box>
            </Grid>
          </Grid>
        </Box>
      </SimpleForm>
    </Create>
  );
};

ProviderCreate.displayName = 'ProviderCreate';
