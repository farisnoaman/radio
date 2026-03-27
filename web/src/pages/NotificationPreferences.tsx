import { useState, useEffect } from 'react';
import { useNotify, useTranslate } from 'react-admin';
import {
  Card,
  CardContent,
  Typography,
  Box,
  Switch,
  FormControlLabel,
  Button,
  Stack,
  Divider,
  CircularProgress,
  Checkbox,
  FormGroup,
  Paper,
  Chip,
} from '@mui/material';
import {
  Email as EmailIcon,
  Sms as SmsIcon,
  NotificationsActive as AlertIcon,
} from '@mui/icons-material';
import { useApiQuery } from '../hooks/useApiQuery';
import { useApiMutation } from '../hooks/useApiMutation';

interface NotificationPreferences {
  id?: number;
  email_enabled: boolean;
  sms_enabled: boolean;
  email_thresholds: string;
  sms_thresholds: string;
  daily_summary_enabled: boolean;
}

const THRESHOLD_OPTIONS = [70, 80, 85, 90, 95, 100];

export default function NotificationPreferences() {
  const translate = useTranslate();
  const notify = useNotify();
  const [loading, setLoading] = useState(false);
  const [preferences, setPreferences] = useState<NotificationPreferences>({
    email_enabled: true,
    sms_enabled: false,
    email_thresholds: '80,90,100',
    sms_thresholds: '100',
    daily_summary_enabled: false,
  });

  const { data, isLoading, refetch } = useApiQuery<NotificationPreferences>(
    ['portal', 'preferences', 'notifications'],
    '/api/v1/portal/preferences/notifications',
    { enabled: true }
  );

  useEffect(() => {
    if (data) {
      setPreferences({
        ...data,
      });
    }
  }, [data]);

  const mutation = useApiMutation<
    NotificationPreferences,
    NotificationPreferences
  >(
    'PUT',
    '/api/v1/portal/preferences/notifications',
    {
      onSuccess: () => {
        notify(translate('common.save_success'), { type: 'success' });
        refetch();
      },
      onError: (error: Error) => {
        notify(error.message || translate('common.save_failed'), { type: 'error' });
      },
    }
  );

  const handleSave = async () => {
    setLoading(true);
    try {
      await mutation.mutateAsync(preferences);
    } finally {
      setLoading(false);
    }
  };

  const handleThresholdToggle = (type: 'email' | 'sms', value: number) => () => {
    const key = type === 'email' ? 'email_thresholds' : 'sms_thresholds';
    const currentThresholds = preferences[key]
      .split(',')
      .map((t) => parseInt(t.trim()))
      .filter((n) => !isNaN(n));

    const newThresholds = currentThresholds.includes(value)
      ? currentThresholds.filter((t) => t !== value)
      : [...currentThresholds, value].sort((a, b) => a - b);

    setPreferences({
      ...preferences,
      [key]: newThresholds.join(','),
    });
  };

  const getThresholdsArray = (thresholds: string): number[] => {
    return thresholds
      .split(',')
      .map((t) => parseInt(t.trim()))
      .filter((n) => !isNaN(n));
  };

  if (isLoading && !data) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress size={20} />
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3, maxWidth: 800, mx: 'auto' }}>
      <Box sx={{ mb: 3 }}>
        <Typography variant="h5" sx={{ fontWeight: 700, mb: 1 }}>
          <AlertIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
          {translate('portal.notification_preferences')}
        </Typography>
        <Typography variant="body2" sx={{ color: 'text.secondary' }}>
          {translate('portal.notification_preferences_desc')}
        </Typography>
      </Box>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Stack spacing={3}>
            <Box>
              <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 2 }}>
                <EmailIcon color="primary" />
                <Typography variant="h6">
                  {translate('portal.email_alerts')}
                </Typography>
              </Stack>
              <Typography variant="body2" sx={{ color: 'text.secondary', mb: 2 }}>
                {translate('portal.email_alerts_desc')}
              </Typography>
              <FormControlLabel
                control={
                  <Switch
                    checked={preferences.email_enabled}
                    onChange={(e) =>
                      setPreferences({
                        ...preferences,
                        email_enabled: e.target.checked,
                      })
                    }
                  />
                }
                label={translate('portal.enable_email_alerts')}
              />
            </Box>

            <Divider />

            <Box>
              <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 2 }}>
                <SmsIcon color="primary" />
                <Typography variant="h6">
                  {translate('portal.sms_alerts')}
                </Typography>
              </Stack>
              <Typography variant="body2" sx={{ color: 'text.secondary', mb: 2 }}>
                {translate('portal.sms_alerts_desc')}
              </Typography>
              <FormControlLabel
                control={
                  <Switch
                    checked={preferences.sms_enabled}
                    onChange={(e) =>
                      setPreferences({
                        ...preferences,
                        sms_enabled: e.target.checked,
                      })
                    }
                  />
                }
                label={translate('portal.enable_sms_alerts')}
              />
            </Box>

            <Divider />

            <Box>
              <Typography variant="h6" sx={{ mb: 1 }}>
                {translate('portal.alert_thresholds')}
              </Typography>
              <Typography variant="body2" sx={{ color: 'text.secondary', mb: 2 }}>
                {translate('portal.alert_thresholds_desc')}
              </Typography>

              <Paper sx={{ p: 2, bgcolor: 'background.default' }}>
                <Typography variant="subtitle2" sx={{ mb: 1 }}>
                  {translate('portal.email_alerts')}:
                </Typography>
                <FormGroup row>
                  {THRESHOLD_OPTIONS.map((threshold) => (
                    <FormControlLabel
                      key={`email-${threshold}`}
                      control={
                        <Checkbox
                          checked={getThresholdsArray(preferences.email_thresholds).includes(
                            threshold
                          )}
                          onChange={handleThresholdToggle('email', threshold)}
                          disabled={!preferences.email_enabled}
                        />
                      }
                      label={<Chip label={`${threshold}%`} size="small" />}
                    />
                  ))}
                </FormGroup>

                <Typography variant="subtitle2" sx={{ mt: 2, mb: 1 }}>
                  {translate('portal.sms_alerts')}:
                </Typography>
                <FormGroup row>
                  {THRESHOLD_OPTIONS.map((threshold) => (
                    <FormControlLabel
                      key={`sms-${threshold}`}
                      control={
                        <Checkbox
                          checked={getThresholdsArray(preferences.sms_thresholds).includes(
                            threshold
                          )}
                          onChange={handleThresholdToggle('sms', threshold)}
                          disabled={!preferences.sms_enabled}
                        />
                      }
                      label={<Chip label={`${threshold}%`} size="small" />}
                    />
                  ))}
                </FormGroup>
              </Paper>
            </Box>

            <Divider />

            <FormControlLabel
              control={
                <Switch
                  checked={preferences.daily_summary_enabled}
                  onChange={(e) =>
                    setPreferences({
                      ...preferences,
                      daily_summary_enabled: e.target.checked,
                    })
                  }
                />
              }
              label={translate('portal.daily_summary')}
            />
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('portal.daily_summary_desc')}
            </Typography>
          </Stack>

          <Box sx={{ mt: 3 }}>
            <Button
              variant="contained"
              size="large"
              onClick={handleSave}
              disabled={loading}
            >
              {loading ? <CircularProgress size={24} /> : translate('common.save')}
            </Button>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
}
