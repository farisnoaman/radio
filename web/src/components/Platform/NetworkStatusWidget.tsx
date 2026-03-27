import { Card, CardContent, Typography, Box, Chip, useMediaQuery, useTheme } from '@mui/material';
import { Grid } from '@mui/material';
import { useApiQuery } from '../../hooks/useApiQuery';
import LanIcon from '@mui/icons-material/Lan';
import DnsIcon from '@mui/icons-material/Dns';

interface NetworkMetrics {
  nodes: { active: number; total: number };
  servers: { active: number; total: number };
}

const DeviceStatus = ({
  active,
  total,
  label,
  icon,
}: {
  active: number;
  total: number;
  label: string;
  icon: React.ReactNode;
}) => {
  const percent = total > 0 ? (active / total) * 100 : 0;
  const color = percent >= 80 ? 'success' : percent >= 50 ? 'warning' : 'error';

  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
          {icon}
          <Typography variant="body1" sx={{ ml: 1, fontWeight: 600 }}>
            {label}
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Typography variant="h5" sx={{ fontWeight: 'bold' }}>
            {active}/{total}
          </Typography>
          <Chip label={`${percent.toFixed(0)}%`} color={color} size="small" />
        </Box>
        <Typography variant="caption" color="text.secondary">
          {active} active of {total} total
        </Typography>
      </CardContent>
    </Card>
  );
};

export const NetworkStatusWidget = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const { data, isLoading } = useApiQuery<NetworkMetrics>({
    path: '/api/v1/reporting/network-status',
    queryKey: ['reporting', 'network-status'],
    enabled: true,
  });

  if (isLoading) {
    return (
      <Card>
        <CardContent>
          <Typography variant="h6" sx={{ mb: 2 }}>
            Network Status (Real-time)
          </Typography>
          <Grid container spacing={2}>
            <Grid size={{ xs: 12, sm: 6 }}>
              <Card sx={{ bgcolor: 'action.hover' }}>
                <CardContent>
                  <Box sx={{ height: 80 }} />
                </CardContent>
              </Card>
            </Grid>
            <Grid size={{ xs: 12, sm: 6 }}>
              <Card sx={{ bgcolor: 'action.hover' }}>
                <CardContent>
                  <Box sx={{ height: 80 }} />
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent>
        <Typography
          variant="h6"
          sx={{ mb: isMobile ? 1.5 : 2, fontSize: isMobile ? '1rem' : undefined }}
        >
          Network Status (Real-time)
        </Typography>
        <Grid container spacing={isMobile ? 1.5 : 2}>
          <Grid size={{ xs: 12, sm: 6 }}>
            <DeviceStatus
              active={data?.nodes?.active ?? 0}
              total={data?.nodes?.total ?? 0}
              label="Nodes"
              icon={<LanIcon color="primary" />}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6 }}>
            <DeviceStatus
              active={data?.servers?.active ?? 0}
              total={data?.servers?.total ?? 0}
              label="Servers"
              icon={<DnsIcon color="primary" />}
            />
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  );
};
