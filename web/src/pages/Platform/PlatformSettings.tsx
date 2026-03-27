import {
  Box,
  Container,
  Typography,
  Card,
  CardContent,
  Grid,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Button,
  Switch,
  FormControlLabel,
  Divider,
  Alert,
  AlertTitle,
  Stack,
  Tabs,
  Tab,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import {
  Save,
  People,
  MonetizationOn,
  Settings as SettingsIcon,
} from '@mui/icons-material';
import { useState } from 'react';
import { useTranslate } from 'react-admin';
import { PricingPlans } from './PricingPlans';

const defaultQuotas = {
  maxUsers: 5000,
  maxOnlineUsers: 1500,
  maxNas: 50,
  maxMikrotikDevices: 50,
  maxStorage: 50,
  maxDailyBackups: 5,
  maxBandwidth: 10,
  maxAuthPerSecond: 100,
  maxAcctPerSecond: 100,
};

const defaultPricing = {
  baseFee: 99,
  includedUsers: 100,
  overageFeePerUser: 1,
  currency: 'USD',
  billingCycle: 'monthly',
};

const defaultSystemSettings = {
  enableRegistration: true,
  requireApproval: true,
  defaultQuotaPlan: 'professional',
  maxProviders: 100,
  enableMonitoring: true,
  enableAutoBackups: true,
  retentionDays: 30,
  supportEmail: 'support@radio.com',
};

interface SettingsSectionProps {
  title: string;
  icon: React.ReactNode;
  children: React.ReactNode;
}

const SettingsSection = ({ title, icon, children }: SettingsSectionProps) => (
  <Card
    sx={{
      mb: 1,
      background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(248,250,252,0.98) 100%)',
      border: '1px solid rgba(148, 163, 184, 0.1)',
      borderRadius: 1,
    }}
  >
    <CardContent>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 3 }}>
        <Box
          sx={{
            p: 1.5,
            borderRadius: 2,
            bgcolor: 'rgba(30, 58, 138, 0.08)',
            color: '#1e3a8a',
            display: 'flex',
          }}
        >
          {icon}
        </Box>
        <Typography variant="h6" sx={{ fontWeight: 600 }}>
          {title}
        </Typography>
      </Box>
      {children}
    </CardContent>
  </Card>
);

export const PlatformSettings = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const translate = useTranslate();
  const [quotas, setQuotas] = useState(defaultQuotas);
  const [pricing, setPricing] = useState(defaultPricing);
  const [systemSettings, setSystemSettings] = useState(defaultSystemSettings);
  const [saved, setSaved] = useState(false);
  const [currentTab, setCurrentTab] = useState(0);

  const handleSave = async () => {
    try {
      await new Promise((resolve) => setTimeout(resolve, 1000));
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (error) {
      console.error('Failed to save settings:', error);
    }
  };

  return (
    <Container maxWidth={false} sx={{ py: isMobile ? 0.5 : 4, px: isMobile ? 0.625 : undefined }}>
      <Box sx={{ mb: isMobile ? 0.5 : 4 }}>
        <Typography
          variant={isMobile ? 'h6' : 'h4'}
          sx={{ fontWeight: 700, mb: 1 }}
        >
          {translate('platform_settings.title')}
        </Typography>
        <Typography
          variant="body2"
          sx={{ color: 'text.secondary', display: { xs: 'none', sm: 'block' } }}
        >
          {translate('platform_settings.subtitle')}
        </Typography>
      </Box>

      {saved && (
        <Alert severity="success" sx={{ mb: 3 }}>
          <AlertTitle>{translate('platform_settings.success_save')}</AlertTitle>
          {translate('platform_settings.success_save_message')}
        </Alert>
      )}

      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
        <Tabs
          value={currentTab}
          onChange={(_, newValue) => setCurrentTab(newValue)}
          variant={isMobile ? 'scrollable' : 'standard'}
          scrollButtons={isMobile ? 'auto' : false}
        >
          <Tab label={translate('platform_settings.tab_general')} />
          <Tab label={translate('platform_settings.tab_pricing')} />
        </Tabs>
      </Box>

      {currentTab === 0 && (
        <Grid container spacing={isMobile ? 2 : 3}>
          <Grid size={{ xs: 12, lg: 8 }}>
            <Stack spacing={isMobile ? 2 : 3}>
              <SettingsSection
                title={translate('platform_settings.quotas_title')}
                icon={<People />}
              >
                <Alert severity="info" sx={{ mb: 3 }}>
                  <AlertTitle>{translate('platform_settings.quotas_alert_title')}</AlertTitle>
                  {translate('platform_settings.quotas_alert_message')}
                </Alert>
                <Grid container spacing={2}>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      fullWidth
                      label={translate('platform_settings.max_users')}
                      type="number"
                      value={quotas.maxUsers}
                      onChange={(e) => setQuotas({ ...quotas, maxUsers: parseInt(e.target.value) })}
                      helperText={translate('platform_settings.max_users_help')}
                      size={isMobile ? 'small' : 'medium'}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      fullWidth
                      label={translate('platform_settings.max_concurrent_sessions')}
                      type="number"
                      value={quotas.maxOnlineUsers}
                      onChange={(e) => setQuotas({ ...quotas, maxOnlineUsers: parseInt(e.target.value) })}
                      helperText={translate('platform_settings.max_concurrent_sessions_help')}
                      size={isMobile ? 'small' : 'medium'}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      fullWidth
                      label={translate('platform_settings.max_nas_devices')}
                      type="number"
                      value={quotas.maxNas}
                      onChange={(e) => setQuotas({ ...quotas, maxNas: parseInt(e.target.value) })}
                      helperText={translate('platform_settings.max_nas_devices_help')}
                      size={isMobile ? 'small' : 'medium'}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      fullWidth
                      label={translate('platform_settings.max_mikrotik_devices')}
                      type="number"
                      value={quotas.maxMikrotikDevices}
                      onChange={(e) => setQuotas({ ...quotas, maxMikrotikDevices: parseInt(e.target.value) })}
                      helperText={translate('platform_settings.max_mikrotik_devices_help')}
                      size={isMobile ? 'small' : 'medium'}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      fullWidth
                      label={translate('platform_settings.max_storage')}
                      type="number"
                      value={quotas.maxStorage}
                      onChange={(e) => setQuotas({ ...quotas, maxStorage: parseInt(e.target.value) })}
                      helperText={translate('platform_settings.max_storage_help')}
                      size={isMobile ? 'small' : 'medium'}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      fullWidth
                      label={translate('platform_settings.max_daily_backups')}
                      type="number"
                      value={quotas.maxDailyBackups}
                      onChange={(e) => setQuotas({ ...quotas, maxDailyBackups: parseInt(e.target.value) })}
                      helperText={translate('platform_settings.max_daily_backups_help')}
                      size={isMobile ? 'small' : 'medium'}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      fullWidth
                      label={translate('platform_settings.max_bandwidth')}
                      type="number"
                      value={quotas.maxBandwidth}
                      onChange={(e) => setQuotas({ ...quotas, maxBandwidth: parseInt(e.target.value) })}
                      helperText={translate('platform_settings.max_bandwidth_help')}
                      size={isMobile ? 'small' : 'medium'}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      fullWidth
                      label={translate('platform_settings.max_auth_requests')}
                      type="number"
                      value={quotas.maxAuthPerSecond}
                      onChange={(e) => setQuotas({ ...quotas, maxAuthPerSecond: parseInt(e.target.value) })}
                      helperText={translate('platform_settings.max_auth_requests_help')}
                      size={isMobile ? 'small' : 'medium'}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                </Grid>
              </SettingsSection>

              <SettingsSection
                title={translate('platform_settings.pricing_title')}
                icon={<MonetizationOn />}
              >
                <Alert severity="info" sx={{ mb: 3 }}>
                  <AlertTitle>{translate('platform_settings.pricing_alert_title')}</AlertTitle>
                  {translate('platform_settings.pricing_alert_message')}
                </Alert>
                <Grid container spacing={2}>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      fullWidth
                      label={translate('platform_settings.base_fee')}
                      type="number"
                      value={pricing.baseFee}
                      onChange={(e) => setPricing({ ...pricing, baseFee: parseFloat(e.target.value) })}
                      InputProps={{ startAdornment: <Box sx={{ mr: 1 }}>$</Box> }}
                      helperText={translate('platform_settings.base_fee_help')}
                      size={isMobile ? 'small' : 'medium'}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      fullWidth
                      label={translate('platform_settings.included_users')}
                      type="number"
                      value={pricing.includedUsers}
                      onChange={(e) => setPricing({ ...pricing, includedUsers: parseInt(e.target.value) })}
                      helperText={translate('platform_settings.included_users_help')}
                      size={isMobile ? 'small' : 'medium'}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <TextField
                      fullWidth
                      label={translate('platform_settings.overage_fee_user')}
                      type="number"
                      value={pricing.overageFeePerUser}
                      onChange={(e) => setPricing({ ...pricing, overageFeePerUser: parseFloat(e.target.value) })}
                      InputProps={{ startAdornment: <Box sx={{ mr: 1 }}>$</Box> }}
                      helperText={translate('platform_settings.overage_fee_user_help')}
                      size={isMobile ? 'small' : 'medium'}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <FormControl fullWidth size={isMobile ? 'small' : 'medium'}>
                      <InputLabel>{translate('platform_settings.currency')}</InputLabel>
                      <Select
                        label={translate('platform_settings.currency')}
                        value={pricing.currency}
                        onChange={(e) => setPricing({ ...pricing, currency: e.target.value })}
                      >
                        <MenuItem value="USD">USD ($)</MenuItem>
                        <MenuItem value="EUR">EUR (€)</MenuItem>
                        <MenuItem value="GBP">GBP (£)</MenuItem>
                      </Select>
                    </FormControl>
                  </Grid>
                  <Grid size={{ xs: 12 }}>
                    <FormControl fullWidth size={isMobile ? 'small' : 'medium'}>
                      <InputLabel>{translate('platform_settings.billing_cycle')}</InputLabel>
                      <Select
                        label={translate('platform_settings.billing_cycle')}
                        value={pricing.billingCycle}
                        onChange={(e) => setPricing({ ...pricing, billingCycle: e.target.value })}
                      >
                        <MenuItem value="monthly">{translate('platform_settings.monthly')}</MenuItem>
                        <MenuItem value="yearly">{translate('platform_settings.yearly')}</MenuItem>
                      </Select>
                    </FormControl>
                  </Grid>
                </Grid>
              </SettingsSection>

              <SettingsSection
                title={translate('platform_settings.system_title')}
                icon={<SettingsIcon />}
              >
                <Stack spacing={isMobile ? 2 : 3}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={systemSettings.enableRegistration}
                        onChange={(e) =>
                          setSystemSettings({ ...systemSettings, enableRegistration: e.target.checked })
                        }
                      />
                    }
                    label={translate('platform_settings.enable_registration')}
                  />
                  <FormControlLabel
                    control={
                      <Switch
                        checked={systemSettings.requireApproval}
                        onChange={(e) =>
                          setSystemSettings({ ...systemSettings, requireApproval: e.target.checked })
                        }
                      />
                    }
                    label={translate('platform_settings.require_approval')}
                  />
                  <FormControlLabel
                    control={
                      <Switch
                        checked={systemSettings.enableMonitoring}
                        onChange={(e) =>
                          setSystemSettings({ ...systemSettings, enableMonitoring: e.target.checked })
                        }
                      />
                    }
                    label={translate('platform_settings.enable_monitoring')}
                  />
                  <FormControlLabel
                    control={
                      <Switch
                        checked={systemSettings.enableAutoBackups}
                        onChange={(e) =>
                          setSystemSettings({ ...systemSettings, enableAutoBackups: e.target.checked })
                        }
                      />
                    }
                    label={translate('platform_settings.enable_auto_backups')}
                  />
                  <Divider />
                  <TextField
                    fullWidth
                    label={translate('platform_settings.max_providers')}
                    type="number"
                    value={systemSettings.maxProviders}
                    onChange={(e) =>
                      setSystemSettings({ ...systemSettings, maxProviders: parseInt(e.target.value) })
                    }
                    helperText={translate('platform_settings.max_providers_help')}
                    size={isMobile ? 'small' : 'medium'}
                    sx={{ mb: 2 }}
                  />
                  <TextField
                    fullWidth
                    label={translate('platform_settings.retention_days')}
                    type="number"
                    value={systemSettings.retentionDays}
                    onChange={(e) =>
                      setSystemSettings({ ...systemSettings, retentionDays: parseInt(e.target.value) })
                    }
                    helperText={translate('platform_settings.retention_days_help')}
                    size={isMobile ? 'small' : 'medium'}
                    sx={{ mb: 2 }}
                  />
                  <TextField
                    fullWidth
                    label={translate('platform_settings.support_email')}
                    type="email"
                    value={systemSettings.supportEmail}
                    onChange={(e) =>
                      setSystemSettings({ ...systemSettings, supportEmail: e.target.value })
                    }
                    helperText={translate('platform_settings.support_email_help')}
                    size={isMobile ? 'small' : 'medium'}
                  />
                </Stack>
              </SettingsSection>
            </Stack>

            <Box sx={{ mt: 4 }}>
              <Button
                variant="contained"
                size={isMobile ? 'medium' : 'large'}
                startIcon={<Save />}
                onClick={handleSave}
                sx={{
                  px: isMobile ? 3 : 4,
                  py: isMobile ? 1 : 1.5,
                  borderRadius: 2,
                  background: 'linear-gradient(135deg, #1e3a8a 0%, #1e40af 100%)',
                  fontSize: isMobile ? 14 : 16,
                  fontWeight: 600,
                  textTransform: 'none',
                  '&:hover': {
                    background: 'linear-gradient(135deg, #1e40af 0%, #2563eb 100%)',
                  },
                }}
              >
                {translate('platform_settings.save_all')}
              </Button>
            </Box>
          </Grid>

          <Grid size={{ xs: 12, lg: 4 }}>
            <Stack spacing={isMobile ? 2 : 3}>
              <Card
                sx={{
                  background: 'linear-gradient(135deg, rgba(30, 58, 138, 0.05) 0%, rgba(30, 58, 138, 0.02) 100%)',
                  border: '1px solid rgba(30, 58, 138, 0.1)',
                  borderRadius: 2,
                }}
              >
                <CardContent>
                  <Typography variant="h6" sx={{ fontWeight: 600, mb: 3 }}>
                    {translate('platform_settings.current_config')}
                  </Typography>
                  <Stack spacing={2}>
                    <Box>
                      <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                        {translate('platform_settings.default_user_limit')}
                      </Typography>
                      <Typography variant="h6" sx={{ fontWeight: 700, color: '#1e3a8a' }}>
                        {quotas.maxUsers.toLocaleString()} {translate('platform_settings.users')}
                      </Typography>
                    </Box>
                    <Box>
                      <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                        {translate('platform_settings.base_monthly_fee')}
                      </Typography>
                      <Typography variant="h6" sx={{ fontWeight: 700, color: '#1e3a8a' }}>
                        ${pricing.baseFee}
                      </Typography>
                    </Box>
                    <Box>
                      <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                        {translate('platform_settings.max_providers')}
                      </Typography>
                      <Typography variant="h6" sx={{ fontWeight: 700, color: '#1e3a8a' }}>
                        {systemSettings.maxProviders}
                      </Typography>
                    </Box>
                    <Box>
                      <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                        {translate('platform_settings.registration')}
                      </Typography>
                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                        {systemSettings.enableRegistration
                          ? translate('platform_settings.open')
                          : translate('platform_settings.closed')}
                        {systemSettings.requireApproval &&
                          ` (${translate('platform_settings.approval_required')})`}
                      </Typography>
                    </Box>
                  </Stack>
                </CardContent>
              </Card>

              <Alert severity="warning">
                <AlertTitle>{translate('platform_settings.caution')}</AlertTitle>
                {translate('platform_settings.caution_message')}
              </Alert>

              <Alert severity="info">
                <AlertTitle>{translate('platform_settings.best_practices')}</AlertTitle>
                {translate('platform_settings.best_practices_message')}
              </Alert>
            </Stack>
          </Grid>
        </Grid>
      )}

      {currentTab === 1 && <PricingPlans />}
    </Container>
  );
};

PlatformSettings.displayName = 'PlatformSettings';
