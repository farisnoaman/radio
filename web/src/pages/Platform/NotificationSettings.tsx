import { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Switch,
  FormControlLabel,
  TextField,
  Button,
  Stack,
  CircularProgress,
} from '@mui/material';
import { useNotify, useTranslate } from 'react-admin';
import { useApiQuery } from '../../hooks/useApiQuery';
import { useMutation } from '@tanstack/react-query';
import { apiRequest } from '../../utils/apiClient';

interface NotificationPreferences {
  alert_percentages: string;
  alert_percentages_enabled: boolean;
  max_users_threshold: number;
  max_data_bytes_threshold: number;
  absolute_alerts_enabled: boolean;
  anomaly_detection_enabled: boolean;
  anomaly_threshold_percent: number;
  email_enabled: boolean;
  sms_enabled: boolean;
}

export default function NotificationSettings() {
  const translate = useTranslate();
  const notify = useNotify();

  const [preferences, setPreferences] = useState<NotificationPreferences>({
    alert_percentages: '70,85,100',
    alert_percentages_enabled: true,
    max_users_threshold: 0,
    max_data_bytes_threshold: 0,
    absolute_alerts_enabled: false,
    anomaly_detection_enabled: false,
    anomaly_threshold_percent: 50,
    email_enabled: true,
    sms_enabled: false,
  });

  const { data, isLoading, refetch } = useApiQuery<NotificationPreferences>({
    path: '/api/v1/reporting/notifications/preferences',
    queryKey: ['reporting', 'notification-preferences'],
    enabled: true,
  });

  useEffect(() => {
    if (data) {
      setPreferences(data);
    }
  }, [data]);

  const mutation = useMutation({
    mutationFn: (body: NotificationPreferences) =>
      apiRequest<{ message: string }>('/api/v1/reporting/notifications/preferences', {
        method: 'PUT',
        body: JSON.stringify(body),
      }),
    onSuccess: () => {
      notify(translate('common.save_success'), { type: 'success' });
      refetch();
    },
    onError: (error: Error) => {
      notify(error.message || translate('common.save_failed'), { type: 'error' });
    },
  });

  const handleSave = () => {
    mutation.mutate(preferences);
  };

  const updatePref = <K extends keyof NotificationPreferences>(
    key: K,
    value: NotificationPreferences[K]
  ) => {
    setPreferences((prev) => ({ ...prev, [key]: value }));
  };

  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3, maxWidth: 800 }}>
      <Typography variant="h5" sx={{ fontWeight: 'bold', mb: 3 }}>
        Notification Settings
      </Typography>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>Percentage-Based Alerts</Typography>
          <FormControlLabel
            control={
              <Switch
                checked={preferences.alert_percentages_enabled}
                onChange={(e) => updatePref('alert_percentages_enabled', e.target.checked)}
              />
            }
            label="Enable percentage alerts"
          />
          <TextField
            label="Alert Percentages (comma-separated)"
            value={preferences.alert_percentages}
            onChange={(e) => updatePref('alert_percentages', e.target.value)}
            disabled={!preferences.alert_percentages_enabled}
            fullWidth
            sx={{ mt: 2 }}
            placeholder="70,85,100"
          />
        </CardContent>
      </Card>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>Absolute Thresholds</Typography>
          <FormControlLabel
            control={
              <Switch
                checked={preferences.absolute_alerts_enabled}
                onChange={(e) => updatePref('absolute_alerts_enabled', e.target.checked)}
              />
            }
            label="Enable absolute threshold alerts"
          />
          <TextField
            label="Max Users Threshold"
            type="number"
            value={preferences.max_users_threshold}
            onChange={(e) => updatePref('max_users_threshold', parseInt(e.target.value) || 0)}
            disabled={!preferences.absolute_alerts_enabled}
            fullWidth
            sx={{ mt: 2 }}
          />
        </CardContent>
      </Card>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>Anomaly Detection</Typography>
          <FormControlLabel
            control={
              <Switch
                checked={preferences.anomaly_detection_enabled}
                onChange={(e) => updatePref('anomaly_detection_enabled', e.target.checked)}
              />
            }
            label="Enable anomaly detection"
          />
          <TextField
            label="Anomaly Threshold (%)"
            type="number"
            value={preferences.anomaly_threshold_percent}
            onChange={(e) => updatePref('anomaly_threshold_percent', parseInt(e.target.value) || 50)}
            disabled={!preferences.anomaly_detection_enabled}
            fullWidth
            sx={{ mt: 2 }}
            inputProps={{ min: 1, max: 100 }}
          />
        </CardContent>
      </Card>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>Notification Channels</Typography>
          <Stack direction="row" spacing={4}>
            <FormControlLabel
              control={
                <Switch
                  checked={preferences.email_enabled}
                  onChange={(e) => updatePref('email_enabled', e.target.checked)}
                />
              }
              label="Email"
            />
            <FormControlLabel
              control={
                <Switch
                  checked={preferences.sms_enabled}
                  onChange={(e) => updatePref('sms_enabled', e.target.checked)}
                />
              }
              label="SMS"
            />
          </Stack>
        </CardContent>
      </Card>

      <Button
        variant="contained"
        onClick={handleSave}
        disabled={mutation.isPending}
        startIcon={mutation.isPending ? <CircularProgress size={20} /> : null}
      >
        Save Settings
      </Button>
    </Box>
  );
}
