import {
  Create,
  SimpleForm,
  TextInput,
  BooleanInput,
  useCreate,
  useNotify,
  useRedirect,
  useRefresh,
  FormDataConsumer,
  useTranslate,
} from 'react-admin';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Grid,
  Alert,
  AlertTitle,
  Stack,
} from '@mui/material';
import { Backup, Lock } from '@mui/icons-material';

export const BackupCreate = () => {
  const [create] = useCreate();
  const notify = useNotify();
  const redirect = useRedirect();
  const refresh = useRefresh();
  const translate = useTranslate();

  const handleSubmit = async (data: any) => {
    try {
      await create(
        'provider/backup',
        { data: { ...data, backup_type: 'manual' } },
        {
          onSuccess: () => {
            notify(translate('backup.success_create'), { type: 'success' });
            redirect('list', 'provider/backup');
            refresh();
          },
          onError: (error: any) => {
            notify(translate('backup.error_create', { error: error.message }), { type: 'error' });
          },
        }
      );
    } catch (error: any) {
      notify(translate('backup.error_create', { error: error.message }), { type: 'error' });
    }
  };

  return (
    <Create redirect="list">
      <SimpleForm onSubmit={handleSubmit}>
        <Box sx={{ p: 3, maxWidth: 800 }}>
          {/* Header */}
          <Box sx={{ mb: 4 }}>
            <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
              {translate('backup.create_manual')}
            </Typography>
            <Typography variant="body1" sx={{ color: 'text.secondary' }}>
              {translate('backup.create_manual_description')}
            </Typography>
          </Box>

          <Alert severity="info" sx={{ mb: 3 }}>
            <AlertTitle>{translate('backup.backup_information')}</AlertTitle>
            <Typography variant="body2">
              {translate('backup.backup_information_details')}
            </Typography>
          </Alert>

          <Grid container spacing={3}>
            {/* Backup Configuration */}
            <Grid size={{ xs: 12 }}>
              <Card
                sx={{
                  background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(248,250,252,0.98) 100%)',
                  border: '1px solid rgba(148, 163, 184, 0.1)',
                  borderRadius: 2,
                }}
              >
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 3 }}>
                    <Backup sx={{ color: '#1e3a8a' }} />
                    <Typography variant="h6" sx={{ fontWeight: 600 }}>
                      {translate('backup.backup_scope')}
                    </Typography>
                  </Box>

                  <Stack spacing={3}>
                    <BooleanInput
                      source="include_user_data"
                      label={translate('backup.include_user_data')}
                      defaultValue={true}
                      helperText={translate('backup.include_user_data_help')}
                    />

                    <BooleanInput
                      source="include_accounting"
                      label={translate('backup.include_accounting')}
                      defaultValue={true}
                      helperText={translate('backup.include_accounting_help')}
                    />

                    <BooleanInput
                      source="include_vouchers"
                      label={translate('backup.include_vouchers')}
                      defaultValue={true}
                      helperText={translate('backup.include_vouchers_help')}
                    />

                    <BooleanInput
                      source="include_nas"
                      label={translate('backup.include_nas')}
                      defaultValue={true}
                      helperText={translate('backup.include_nas_help')}
                    />
                  </Stack>
                </CardContent>
              </Card>
            </Grid>

            {/* Encryption Settings */}
            <Grid size={{ xs: 12 }}>
              <Card
                sx={{
                  background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(248,250,252,0.98) 100%)',
                  border: '1px solid rgba(148, 163, 184, 0.1)',
                  borderRadius: 2,
                }}
              >
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 3 }}>
                    <Lock sx={{ color: '#1e3a8a' }} />
                    <Typography variant="h6" sx={{ fontWeight: 600 }}>
                      {translate('backup.encryption_settings')}
                    </Typography>
                  </Box>

                  <Stack spacing={3}>
                    <BooleanInput
                      source="encryption_enabled"
                      label={translate('backup.enable_encryption')}
                      defaultValue={true}
                      helperText={translate('backup.enable_encryption_help')}
                    />

                    <FormDataConsumer>
                      {({ formData }) =>
                        formData?.encryption_enabled && (
                          <TextInput
                            source="encryption_key"
                            label={translate('backup.encryption_key')}
                            helperText={translate('backup.encryption_key_help')}
                            type="password"
                            fullWidth
                          />
                        )
                      }
                    </FormDataConsumer>
                  </Stack>
                </CardContent>
              </Card>
            </Grid>

            {/* Quota Information */}
            <Grid size={{ xs: 12 }}>
              <Alert severity="warning">
                <AlertTitle>{translate('backup.backup_quota')}</AlertTitle>
                <Typography variant="body2">
                  {translate('backup.backup_quota_warning', { max_backups: 5 })}
                </Typography>
              </Alert>
            </Grid>
          </Grid>
        </Box>
      </SimpleForm>
    </Create>
  );
};

BackupCreate.displayName = 'BackupCreate';
