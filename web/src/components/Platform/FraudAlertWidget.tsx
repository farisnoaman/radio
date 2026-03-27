import {
  Card,
  CardContent,
  Typography,
  Box,
  Alert,
  Skeleton,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { useApiQuery } from '../../hooks/useApiQuery';

interface FraudEvent {
  id: number;
  ip_address: string;
  event_type: string;
  details: string;
  created_at: string;
}

export const FraudAlertWidget = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const { data: fraudLogs, isLoading } = useApiQuery<FraudEvent[]>({
    path: '/api/v1/reporting/fraud',
    queryKey: ['reporting', 'fraud'],
    enabled: true,
  });

  if (isLoading) {
    return (
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>Fraud Detection</Typography>
          <Box>
            <Skeleton variant="text" />
            <Skeleton variant="text" />
          </Box>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent sx={{ py: isMobile ? 1.5 : 2 }}>
        <Typography
          variant="h6"
          gutterBottom
          sx={{ fontSize: isMobile ? '1rem' : undefined }}
        >
          Fraud Detection
        </Typography>
        {!fraudLogs || fraudLogs.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            No fraud events detected
          </Typography>
        ) : (
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
            {fraudLogs.slice(0, 5).map((log) => (
              <Alert
                severity="warning"
                key={log.id}
                sx={{ fontSize: isMobile ? '0.75rem' : undefined }}
              >
                <Typography
                  variant="body2"
                  sx={{ fontSize: isMobile ? '0.75rem' : undefined }}
                >
                  <strong>{log.event_type}</strong>
                  {!isMobile && ` — IP: ${log.ip_address}`}
                </Typography>
                <Typography
                  variant="caption"
                  sx={{ fontSize: isMobile ? '0.65rem' : undefined }}
                >
                  {new Date(log.created_at).toLocaleString()}
                </Typography>
              </Alert>
            ))}
          </Box>
        )}
      </CardContent>
    </Card>
  );
};
