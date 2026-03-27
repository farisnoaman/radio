import DevicesOutlinedIcon from '@mui/icons-material/DevicesOutlined';
import ThermostatIcon from '@mui/icons-material/Thermostat';
import ElectricBoltIcon from '@mui/icons-material/ElectricBolt';
import SignalCellularAltIcon from '@mui/icons-material/SignalCellularAlt';
import AirIcon from '@mui/icons-material/Air';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import { 
  Box, 
  Card, 
  CardContent, 
  Chip, 
  CircularProgress, 
  Grid, 
  Stack, 
  Typography 
} from '@mui/material';
import { alpha, Theme, useTheme } from '@mui/material/styles';
import { useTranslate } from 'react-admin';
import { useApiQuery } from '../hooks/useApiQuery';

interface EnvironmentMetric {
  id: number;
  nas_id: number;
  nas_name: string;
  metric_type: string;
  value: number;
  unit: string;
  severity: string;
  collected_at: string;
}

interface EnvironmentAlert {
  id: number;
  nas_id: number;
  metric_type: string;
  alert_value: number;
  threshold_value: number;
  threshold_type: string;
  severity: string;
  status: string;
  fired_at: string;
}

interface NASDevice {
  id: string | number;
  name: string;
  ipaddr: string;
  status: string;
}

interface EnvOverview {
  total_devices: number;
  online_devices: number;
  warning_alerts: number;
  critical_alerts: number;
}

const getSeverityColor = (severity: string, theme: Theme) => {
  switch (severity) {
    case 'critical': return theme.palette.error.main;
    case 'warning': return theme.palette.warning.main;
    default: return theme.palette.success.main;
  }
};

const MetricCard = ({ title, value, unit, severity, icon }: { 
  title: string; 
  value?: number; 
  unit: string; 
  severity: string;
  icon: React.ReactNode;
}) => {
  const theme = useTheme();
  const color = getSeverityColor(severity, theme);
  
  return (
    <Card sx={{ 
      borderRadius: 3,
      border: 1,
      borderColor: alpha(color, 0.3),
      background: alpha(color, 0.05)
    }}>
      <CardContent>
        <Stack direction="row" alignItems="center" spacing={2}>
          <Box sx={{ 
            width: 48, 
            height: 48, 
            borderRadius: 2, 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            backgroundColor: alpha(color, 0.15),
            color: color
          }}>
            {icon}
          </Box>
          <Box>
            <Typography variant="body2" color="text.secondary">
              {title}
            </Typography>
            <Stack direction="row" alignItems="baseline" spacing={1}>
              <Typography variant="h5" fontWeight={600}>
                {value !== undefined ? value.toFixed(1) : '--'}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {unit}
              </Typography>
            </Stack>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
};

const DeviceRow = ({ device, metrics, alerts }: {
  device: NASDevice;
  metrics: EnvironmentMetric[];
  alerts: EnvironmentAlert[];
}) => {
  const theme = useTheme();
  
  const tempMetric = metrics.find(m => m.metric_type === 'temperature');
  const powerMetric = metrics.find(m => m.metric_type === 'power');
  const voltageMetric = metrics.find(m => m.metric_type === 'voltage');
  const fanMetric = metrics.find(m => m.metric_type === 'fan_speed');
  const signalMetric = metrics.find(m => m.metric_type === 'signal_strength');
  
  const hasWarning = alerts.some(a => a.severity === 'warning');
  const hasCritical = alerts.some(a => a.severity === 'critical');
  
  const statusColor = hasCritical ? theme.palette.error.main : hasWarning ? theme.palette.warning.main : theme.palette.success.main;

  return (
    <Card sx={{ borderRadius: 3, mb: 2 }}>
      <CardContent>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Stack direction="row" alignItems="center" spacing={2}>
            <Box sx={{ 
              width: 40, 
              height: 40, 
              borderRadius: 2, 
              display: 'flex', 
              alignItems: 'center', 
              justifyContent: 'center',
              backgroundColor: alpha(statusColor, 0.15),
              color: statusColor
            }}>
              <DevicesOutlinedIcon />
            </Box>
            <Box>
              <Typography variant="subtitle1" fontWeight={600}>
                {device.name}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {device.ipaddr}
              </Typography>
            </Box>
          </Stack>
          
          <Stack direction="row" spacing={1}>
            {hasCritical && <Chip icon={<ErrorIcon />} label="Critical" color="error" size="small" />}
            {hasWarning && <Chip icon={<WarningAmberIcon />} label="Warning" color="warning" size="small" />}
            {!hasCritical && !hasWarning && <Chip icon={<CheckCircleIcon />} label="Healthy" color="success" size="small" />}
          </Stack>
        </Stack>
        
        <Grid container spacing={2} sx={{ mt: 2 }}>
          <Grid size={{ xs: 6, sm: 4, md: 2.4 }}>
            <MetricCard 
              title="Temperature" 
              value={tempMetric?.value} 
              unit="°C" 
              severity={tempMetric?.severity || 'normal'}
              icon={<ThermostatIcon />}
            />
          </Grid>
          <Grid size={{ xs: 6, sm: 4, md: 2.4 }}>
            <MetricCard 
              title="Power" 
              value={powerMetric?.value} 
              unit="W" 
              severity={powerMetric?.severity || 'normal'}
              icon={<ElectricBoltIcon />}
            />
          </Grid>
          <Grid size={{ xs: 6, sm: 4, md: 2.4 }}>
            <MetricCard 
              title="Voltage" 
              value={voltageMetric?.value} 
              unit="V" 
              severity={voltageMetric?.severity || 'normal'}
              icon={<ElectricBoltIcon />}
            />
          </Grid>
          <Grid size={{ xs: 6, sm: 4, md: 2.4 }}>
            <MetricCard 
              title="Fan Speed" 
              value={fanMetric?.value} 
              unit="RPM" 
              severity={fanMetric?.severity || 'normal'}
              icon={<AirIcon />}
            />
          </Grid>
          <Grid size={{ xs: 6, sm: 4, md: 2.4 }}>
            <MetricCard 
              title="Signal" 
              value={signalMetric?.value} 
              unit="dBm" 
              severity={signalMetric?.severity || 'normal'}
              icon={<SignalCellularAltIcon />}
            />
          </Grid>
        </Grid>
        
        {tempMetric?.collected_at && (
          <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
            Last updated: {new Date(tempMetric.collected_at).toLocaleString()}
          </Typography>
        )}
      </CardContent>
    </Card>
  );
};

const EnvironmentMonitoring = () => {
  const theme = useTheme();
  const translate = useTranslate();
  
  const { data: overview, isLoading: overviewLoading } = useApiQuery<EnvOverview>({
    path: '/dashboard/env-health',
    queryKey: ['env-monitoring', 'overview'],
    staleTime: 60 * 1000,
    refetchInterval: 60 * 1000,
  });

  const { data: nasDevices } = useApiQuery<NASDevice[]>({
    path: '/network/nas',
    queryKey: ['network', 'nas'],
    staleTime: 300 * 1000,
  });

  // Fetch metrics for each device
  const { data: metricsData, isLoading: metricsLoading } = useApiQuery<{ metrics: EnvironmentMetric[] }>({
    path: '/network/nas/metrics',
    queryKey: ['env-monitoring', 'metrics'],
    staleTime: 60 * 1000,
    refetchInterval: 60 * 1000,
  });

  const { data: alertsData } = useApiQuery<{ alerts: EnvironmentAlert[] }>({
    path: '/network/nas/alerts',
    queryKey: ['env-monitoring', 'alerts'],
    staleTime: 60 * 1000,
    refetchInterval: 60 * 1000,
  });

  const metrics = metricsData?.metrics || [];
  const alerts = alertsData?.alerts || [];
  
  // Count critical severity metrics
  const criticalMetrics = metrics.filter(m => m.severity === 'critical').length;

  if (overviewLoading || metricsLoading) {
    return (
      <Box sx={{ p: 3 }}>
        <CircularProgress />
      </Box>
    );
  }

  const devicesWithMetrics = (nasDevices || []).map(device => {
    const deviceId = typeof device.id === 'string' ? parseInt(device.id, 10) : device.id;
    const deviceMetrics = metrics.filter(m => m.nas_id === deviceId);
    const deviceAlerts = alerts.filter(a => a.nas_id === deviceId);
    return { device, metrics: deviceMetrics, alerts: deviceAlerts };
  });

  return (
    <Box sx={{ p: 3 }}>
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={3}>
        <Box>
          <Typography variant="h4" fontWeight={700}>
            {translate('dashboard.env_monitoring') || 'Environment Monitoring'}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {translate('dashboard.env_monitoring_desc') || 'Monitor device temperature, power, voltage, and more'}
          </Typography>
        </Box>
      </Stack>

      {/* Overview Stats */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid size={{ xs: 6, md: 3 }}>
          <Card sx={{ borderRadius: 3, background: alpha(theme.palette.primary.main, 0.1) }}>
            <CardContent sx={{ textAlign: 'center', py: 3 }}>
              <Typography variant="h3" fontWeight={700} sx={{ mb: 1 }}>
                {overview?.total_devices || 0}
              </Typography>
              <Stack direction="row" alignItems="center" justifyContent="center" spacing={1}>
                <DevicesOutlinedIcon sx={{ fontSize: 20, color: theme.palette.primary.main }} />
                <Typography variant="body2" color="text.secondary">
                  {translate('dashboard.total_devices') || 'Total Devices'}
                </Typography>
              </Stack>
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Card sx={{ borderRadius: 3, background: alpha('#34d399', 0.1) }}>
            <CardContent sx={{ textAlign: 'center', py: 3 }}>
              <Typography variant="h3" fontWeight={700} sx={{ mb: 1, color: '#34d399' }}>
                {overview?.online_devices || 0}
              </Typography>
              <Stack direction="row" alignItems="center" justifyContent="center" spacing={1}>
                <CheckCircleIcon sx={{ fontSize: 20, color: '#34d399' }} />
                <Typography variant="body2" color="text.secondary">
                  {translate('dashboard.online_devices') || 'Online'}
                </Typography>
              </Stack>
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Card sx={{ borderRadius: 3, background: alpha(theme.palette.warning.main, 0.1) }}>
            <CardContent sx={{ textAlign: 'center', py: 3 }}>
              <Typography variant="h3" fontWeight={700} sx={{ mb: 1, color: theme.palette.warning.main }}>
                {overview?.warning_alerts || 0}
              </Typography>
              <Stack direction="row" alignItems="center" justifyContent="center" spacing={1}>
                <WarningAmberIcon sx={{ fontSize: 20, color: theme.palette.warning.main }} />
                <Typography variant="body2" color="text.secondary">
                  {translate('dashboard.warning_alerts') || 'Warnings'}
                </Typography>
              </Stack>
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Card sx={{ borderRadius: 3, background: alpha(theme.palette.error.main, 0.1) }}>
            <CardContent sx={{ textAlign: 'center', py: 3 }}>
              <Typography variant="h3" fontWeight={700} sx={{ mb: 1, color: theme.palette.error.main }}>
                {criticalMetrics > 0 ? criticalMetrics : (overview?.critical_alerts || 0)}
              </Typography>
              <Stack direction="row" alignItems="center" justifyContent="center" spacing={1}>
                <ErrorIcon sx={{ fontSize: 20, color: theme.palette.error.main }} />
                <Typography variant="body2" color="text.secondary">
                  {translate('dashboard.critical_alerts') || 'Critical'}
                </Typography>
              </Stack>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Device List */}
      <Typography variant="h6" fontWeight={600} mb={2}>
        {translate('dashboard.devices') || 'Devices'}
      </Typography>
      
      {!nasDevices || nasDevices.length === 0 ? (
        <Card sx={{ borderRadius: 3, p: 4, textAlign: 'center' }}>
          <DevicesOutlinedIcon sx={{ fontSize: 48, color: 'text.disabled', mb: 2 }} />
          <Typography color="text.secondary">
            {translate('dashboard.no_nas_devices') || 'No NAS devices configured'}
          </Typography>
          <Typography variant="body2" color="text.disabled">
            {translate('dashboard.no_nas_devices_desc') || 'Add NAS devices in Device Management to monitor environment'}
          </Typography>
        </Card>
      ) : devicesWithMetrics.length === 0 ? (
        <Card sx={{ borderRadius: 3, p: 4, textAlign: 'center' }}>
          <DevicesOutlinedIcon sx={{ fontSize: 48, color: 'text.disabled', mb: 2 }} />
          <Typography color="text.secondary">
            {translate('dashboard.no_metrics_collected') || 'Waiting for metrics'}
          </Typography>
          <Typography variant="body2" color="text.disabled">
            {translate('dashboard.no_metrics_desc') || 'Environment metrics will appear here once collected from your devices'}
          </Typography>
        </Card>
      ) : (
        devicesWithMetrics.map(({ device, metrics: deviceMetrics, alerts: deviceAlerts }) => (
          <DeviceRow 
            key={device.id} 
            device={device} 
            metrics={deviceMetrics} 
            alerts={deviceAlerts} 
          />
        ))
      )}
    </Box>
  );
};

export default EnvironmentMonitoring;