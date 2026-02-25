import { useState, useCallback } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  TextField,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  IconButton,
  CircularProgress,
  Alert,
  Stack,
  useTheme,
  alpha,
  Tooltip,
  LinearProgress,
} from '@mui/material';
import {
  Search as SearchIcon,
  Add as AddIcon,
  Devices as DeviceIcon,
  Wifi as WifiIcon,
} from '@mui/icons-material';
import { useNotify } from 'react-admin';

interface DiscoveredDevice {
  ip: string;
  port: number;
  is_router_os: boolean;
  identity?: string;
  board_name?: string;
  version?: string;
  model?: string;
  serial?: string;
  timestamp?: string;
  error?: string;
}

interface ScanResult {
  cidr: string;
  duration: number;
  found_count: number;
  total_hosts: number;
  results: DiscoveredDevice[];
}

const NetworkDiscovery = () => {
  const theme = useTheme();
  void theme; // satisfies TypeScript - theme is used in sx callbacks
  const notify = useNotify();
  
  const [ipRange, setIpRange] = useState('192.168.1.0/24');
  const [scanning, setScanning] = useState(false);
  const [, setProgress] = useState(0);
  const [scanResult, setScanResult] = useState<ScanResult | null>(null);
  const [addingDevices, setAddingDevices] = useState<Set<string>>(new Set());

  const handleScan = useCallback(async () => {
    if (!ipRange.trim()) {
      notify('Please enter an IP range', { type: 'warning' });
      return;
    }

    setScanning(true);
    setProgress(0);
    setScanResult(null);

    try {
      const response = await fetch('/api/v1/network/discovery/scan', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ip_range: ipRange,
          ports: [8728, 8729],
          timeout: 2,
          workers: 10,
        }),
      });

      const data = await response.json();
      
      if (response.ok && data.code === 0) {
        setScanResult(data.data);
        notify(`Found ${data.data.found_count} MikroTik devices`, { type: 'success' });
      } else {
        notify(data.message || 'Scan failed', { type: 'error' });
      }
    } catch (error) {
      notify('Failed to start scan: ' + (error as Error).message, { type: 'error' });
    } finally {
      setScanning(false);
      setProgress(100);
    }
  }, [ipRange, notify]);

  const handleAddDevice = useCallback(async (device: DiscoveredDevice) => {
    const secret = prompt('Enter RADIUS secret for this device:', 'mikrotik');
    if (!secret) return;

    setAddingDevices(prev => new Set(prev).add(device.ip));

    try {
      const response = await fetch('/api/v1/network/discovery', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ip: device.ip,
          secret: secret,
          name: device.identity || device.model || `Mikrotik-${device.ip}`,
          model: device.model,
          tags: 'discovered',
        }),
      });

      const data = await response.json();

      if (response.ok && data.code === 0) {
        notify('Device added to NAS successfully', { type: 'success' });
      } else {
        notify(data.message || 'Failed to add device', { type: 'error' });
      }
    } catch (error) {
      notify('Failed to add device: ' + (error as Error).message, { type: 'error' });
    } finally {
      setAddingDevices(prev => {
        const next = new Set(prev);
        next.delete(device.ip);
        return next;
      });
    }
  }, [notify]);

  const handleAddAll = useCallback(async () => {
    if (!scanResult) return;

    const devices = scanResult.results.filter(r => r.is_router_os);
    if (devices.length === 0) return;

    const secret = prompt('Enter RADIUS secret for all devices:', 'mikrotik');
    if (!secret) return;

    setAddingDevices(new Set(devices.map(d => d.ip)));

    try {
      const response = await fetch('/api/v1/network/discovery/bulk', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(
          devices.map(d => ({
            ip: d.ip,
            secret: secret,
            name: d.identity || d.model || `Mikrotik-${d.ip}`,
            model: d.model,
            tags: 'discovered',
          }))
        ),
      });

      const data = await response.json();

      if (response.ok && data.code === 0) {
        notify(`Added ${data.data.added_count} devices to NAS`, { type: 'success' });
      } else {
        notify(data.message || 'Failed to add devices', { type: 'error' });
      }
    } catch (error) {
      notify('Failed to add devices: ' + (error as Error).message, { type: 'error' });
    } finally {
      setAddingDevices(new Set());
    }
  }, [scanResult, notify]);

  const mikrotikDevices = scanResult?.results.filter(r => r.is_router_os) || [];
  const otherDevices = scanResult?.results.filter(r => !r.is_router_os) || [];

  return (
    <Box sx={{ p: 3 }}>
      {/* Header */}
      <Box sx={{ mb: 3 }}>
        <Typography variant="h5" sx={{ fontWeight: 600, mb: 1 }}>
          Network Discovery
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Scan your network to discover MikroTik RouterOS devices
        </Typography>
      </Box>

      {/* Scan Form */}
      <Card
        elevation={0}
        sx={{
          mb: 3,
          borderRadius: 2,
          border: theme => `1px solid ${theme.palette.divider}`,
        }}
      >
        <CardContent>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} alignItems="flex-end">
            <TextField
              label="IP Range (CIDR)"
              value={ipRange}
              onChange={e => setIpRange(e.target.value)}
              placeholder="192.168.1.0/24"
              size="small"
              sx={{ flexGrow: 1, maxWidth: 400 }}
              disabled={scanning}
              helperText="Enter IP range in CIDR notation (e.g., 192.168.1.0/24)"
            />
            <Button
              variant="contained"
              startIcon={scanning ? <CircularProgress size={20} color="inherit" /> : <SearchIcon />}
              onClick={handleScan}
              disabled={scanning || !ipRange.trim()}
              sx={{ minWidth: 150 }}
            >
              {scanning ? 'Scanning...' : 'Start Scan'}
            </Button>
          </Stack>

          {scanning && (
            <Box sx={{ mt: 2 }}>
              <LinearProgress />
              <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                Scanning network for MikroTik devices on ports 8728, 8729...
              </Typography>
            </Box>
          )}
        </CardContent>
      </Card>

      {/* Results */}
      {scanResult && (
        <>
          {/* Summary */}
          <Card
            elevation={0}
            sx={{
              mb: 3,
              borderRadius: 2,
              border: theme => `1px solid ${theme.palette.divider}`,
              bgcolor: theme => alpha(theme.palette.success.main, 0.1),
            }}
          >
            <CardContent sx={{ py: 2 }}>
              <Stack direction="row" spacing={3} alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography variant="h4" sx={{ fontWeight: 700, color: 'success.main' }}>
                    {scanResult.found_count}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    MikroTik Devices Found
                  </Typography>
                </Box>
                <Box>
                  <Typography variant="h4" sx={{ fontWeight: 700 }}>
                    {scanResult.total_hosts}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Hosts Scanned
                  </Typography>
                </Box>
                <Box>
                  <Typography variant="h4" sx={{ fontWeight: 700 }}>
                    {scanResult.duration.toFixed(1)}s
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Scan Duration
                  </Typography>
                </Box>
                {mikrotikDevices.length > 0 && (
                  <Button
                    variant="contained"
                    color="success"
                    startIcon={<AddIcon />}
                    onClick={handleAddAll}
                    disabled={addingDevices.size > 0}
                  >
                    Add All to NAS
                  </Button>
                )}
              </Stack>
            </CardContent>
          </Card>

          {/* MikroTik Devices */}
          {mikrotikDevices.length > 0 && (
            <Card
              elevation={0}
              sx={{
                mb: 3,
                borderRadius: 2,
                border: theme => `1px solid ${theme.palette.divider}`,
              }}
            >
              <Box
                sx={{
                  px: 2,
                  py: 1.5,
                  bgcolor: theme => alpha(theme.palette.primary.main, 0.05),
                  borderBottom: theme => `1px solid ${theme.palette.divider}`,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 1,
                }}
              >
                <DeviceIcon sx={{ color: 'primary.main' }} />
                <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                  MikroTik Devices ({mikrotikDevices.length})
                </Typography>
              </Box>
              <TableContainer>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>IP Address</TableCell>
                      <TableCell>Port</TableCell>
                      <TableCell>Identity</TableCell>
                      <TableCell>Model</TableCell>
                      <TableCell>Version</TableCell>
                      <TableCell>Serial</TableCell>
                      <TableCell align="right">Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {mikrotikDevices.map((device, index) => (
                      <TableRow key={index} hover>
                        <TableCell>
                          <Typography variant="body2" sx={{ fontWeight: 600, fontFamily: 'monospace' }}>
                            {device.ip}
                          </Typography>
                        </TableCell>
                        <TableCell>
                          <Chip label={device.port} size="small" variant="outlined" />
                        </TableCell>
                        <TableCell>{device.identity || '-'}</TableCell>
                        <TableCell>{device.model || device.board_name || '-'}</TableCell>
                        <TableCell>{device.version || '-'}</TableCell>
                        <TableCell>{device.serial || '-'}</TableCell>
                        <TableCell align="right">
                          <Tooltip title="Add to NAS">
                            <IconButton
                              size="small"
                              onClick={() => handleAddDevice(device)}
                              disabled={addingDevices.has(device.ip)}
                            >
                              {addingDevices.has(device.ip) ? (
                                <CircularProgress size={20} />
                              ) : (
                                <AddIcon />
                              )}
                            </IconButton>
                          </Tooltip>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </Card>
          )}

          {/* Other Devices (found but not MikroTik) */}
          {otherDevices.length > 0 && (
            <Alert severity="info" sx={{ mb: 3 }}>
              Found {otherDevices.length} hosts but they are not MikroTik RouterOS devices.
              Only devices responding on RouterOS API ports (8728/8729) are shown above.
            </Alert>
          )}

          {/* Empty State */}
          {scanResult.total_hosts > 0 && scanResult.found_count === 0 && (
            <Alert severity="warning">
              No MikroTik devices found in the specified IP range. Make sure:
              <ul style={{ margin: '8px 0 0 0', paddingLeft: '20px' }}>
                <li>MikroTik devices are powered on and accessible</li>
                <li>RouterOS API service is enabled on the device</li>
                <li>Firewall allows access on ports 8728 (HTTP) or 8729 (HTTPS)</li>
              </ul>
            </Alert>
          )}
        </>
      )}

      {/* Initial State */}
      {!scanResult && !scanning && (
        <Card
          elevation={0}
          sx={{
            p: 6,
            borderRadius: 2,
            border: theme => `1px dashed ${theme.palette.divider}`,
            textAlign: 'center',
          }}
        >
          <WifiIcon sx={{ fontSize: 60, color: 'text.disabled', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            Ready to Scan
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Enter an IP range above and click "Start Scan" to discover MikroTik devices on your network.
          </Typography>
          <Typography variant="caption" color="text.disabled">
            Supported: MikroTik RouterOS devices with API enabled on port 8728 or 8729
          </Typography>
        </Card>
      )}
    </Box>
  );
};

export default NetworkDiscovery;
