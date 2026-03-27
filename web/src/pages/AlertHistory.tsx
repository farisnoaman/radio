import { useTranslate } from 'react-admin';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  CircularProgress,
} from '@mui/material';
import {
  History as HistoryIcon,
  Email as EmailIcon,
  Sms as SmsIcon,
} from '@mui/icons-material';
import { useApiQuery } from '../hooks/useApiQuery';

interface UsageAlert {
  id: number;
  user_id: number;
  threshold: number;
  alert_type: string;
  sent_at: string | null;
  created_at: string;
}

export default function AlertHistory() {
  const translate = useTranslate();

  const { data: alerts, isLoading } = useApiQuery<UsageAlert[]>(
    ['portal', 'alerts', 'history'],
    '/api/v1/portal/alerts/history',
    { enabled: true }
  );

  const getThresholdColor = (threshold: number) => {
    if (threshold >= 100) return 'error';
    if (threshold >= 90) return 'warning';
    return 'info';
  };

  return (
    <Box sx={{ p: 3, maxWidth: 1000, mx: 'auto' }}>
      <Box sx={{ mb: 3 }}>
        <Typography variant="h5" sx={{ fontWeight: 700, mb: 1 }}>
          <HistoryIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
          {translate('portal.alert_history')}
        </Typography>
        <Typography variant="body2" sx={{ color: 'text.secondary' }}>
          {translate('portal.alert_history_desc')}
        </Typography>
      </Box>

      <Card>
        <CardContent>
          {(isLoading && !alerts) ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
              <CircularProgress size={20} />
            </Box>
          ) : alerts && alerts.length > 0 ? (
            <TableContainer component={Paper} variant="outlined">
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>{translate('portal.threshold')}</TableCell>
                    <TableCell>{translate('portal.type')}</TableCell>
                    <TableCell>{translate('portal.alert_sent_at')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {alerts.map((alert) => (
                    <TableRow key={alert.id} hover>
                      <TableCell>
                        <Chip
                          label={`${alert.threshold}%`}
                          color={getThresholdColor(alert.threshold)}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          {alert.alert_type === 'email' ? (
                            <EmailIcon fontSize="small" color="action" />
                          ) : (
                            <SmsIcon fontSize="small" color="action" />
                          )}
                          {alert.alert_type === 'email'
                            ? translate('portal.alert_type_email')
                            : translate('portal.alert_type_sms')}
                        </Box>
                      </TableCell>
                      <TableCell>
                        {alert.sent_at
                          ? new Date(alert.sent_at).toLocaleString()
                          : new Date(alert.created_at).toLocaleString()}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          ) : (
            <Box sx={{ textAlign: 'center', py: 4 }}>
              <Typography variant="body1" sx={{ color: 'text.secondary' }}>
                {translate('portal.no_alerts')}
              </Typography>
            </Box>
          )}
        </CardContent>
      </Card>
    </Box>
  );
}
