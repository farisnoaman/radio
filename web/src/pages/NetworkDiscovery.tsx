import { useState, useCallback, useEffect } from 'react';
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
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
} from '@mui/material';
import {
  Search as SearchIcon,
  Add as AddIcon,
  Devices as DeviceIcon,
  Wifi as WifiIcon,
  ContentCopy as CopyIcon,
  CheckCircle as CheckIcon,
} from '@mui/icons-material';
import { useNotify } from 'react-admin';
import { apiRequest } from '../utils/apiClient';

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

interface SetupInfo {
  device: DiscoveredDevice;
  secret: string;
  radiusServerIP: string;
}

const CodeBlock = ({ code }: { code: string }) => {
  const [copied, setCopied] = useState(false);
  const handleCopy = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <Box
      sx={{
        position: 'relative',
        bgcolor: '#1e1e1e',
        borderRadius: 1,
        p: 2,
        mb: 1,
        fontFamily: 'monospace',
        fontSize: '0.8rem',
        overflowX: 'auto',
      }}
    >
      <Typography
        component="pre"
        sx={{ color: '#d4d4d4', m: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}
      >
        {code}
      </Typography>
      <IconButton
        size="small"
        onClick={handleCopy}
        sx={{ position: 'absolute', top: 4, right: 4, color: copied ? 'success.main' : 'grey.400' }}
      >
        {copied ? <CheckIcon fontSize="small" /> : <CopyIcon fontSize="small" />}
      </IconButton>
    </Box>
  );
};

const NetworkDiscovery = () => {
  const theme = useTheme();
  void theme;
  const notify = useNotify();

  const [ipRange, setIpRange] = useState('192.168.1.0/24');
  const [username, setUsername] = useState('admin');
  const [password, setPassword] = useState('');
  const [scanning, setScanning] = useState(false);
  const [, setProgress] = useState(0);
  const [scanResult, setScanResult] = useState<ScanResult | null>(null);
  const [addingDevices, setAddingDevices] = useState<Set<string>>(new Set());
  const [setupInfo, setSetupInfo] = useState<SetupInfo | null>(null);
  const [radiusServerIP, setRadiusServerIP] = useState('');

  // Load the configured RADIUS server IP
  useEffect(() => {
    apiRequest<any>('/system/settings?type=radius&name=ServerIP&perPage=100&page=1')
      .then((data: any) => {
        const items: any[] = Array.isArray(data) ? data : (data?.data ?? []);
        const entry = items.find((s: any) => s.type === 'radius' && s.name === 'ServerIP');
        if (entry?.value) setRadiusServerIP(entry.value);
      })
      .catch(() => {/*ignore*/ });
  }, []);

  const handleScan = useCallback(async () => {
    if (!ipRange.trim()) {
      notify('Please enter an IP range', { type: 'warning' });
      return;
    }

    setScanning(true);
    setProgress(0);
    setScanResult(null);

    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 300000);

      const result = await apiRequest<any>('/network/discovery/scan', {
        method: 'POST',
        signal: controller.signal,
        body: JSON.stringify({
          ip_range: ipRange,
          ports: [8728, 8729],
          timeout: 5,
          workers: 20,
          username: username || 'admin',
          password: password,
        }),
      });

      clearTimeout(timeoutId);

      if (result) {
        setScanResult(result);
        notify(`Found ${result.found_count} MikroTik devices`, { type: 'success' });
      }
    } catch (error: any) {
      if (error.name === 'AbortError') {
        notify('Scan timed out', { type: 'error' });
      } else {
        notify(error.message || 'Scan failed', { type: 'error' });
      }
    } finally {
      setScanning(false);
      setProgress(100);
    }
  }, [ipRange, username, password, notify]);

  const handleAddDevice = useCallback(async (device: DiscoveredDevice) => {
    const secret = prompt('Enter RADIUS shared secret for this device:', 'mikrotik');
    if (!secret) return;

    setAddingDevices(prev => new Set(prev).add(device.ip));

    try {
      await apiRequest('/network/discovery', {
        method: 'POST',
        body: JSON.stringify({
          ip: device.ip,
          secret,
          name: device.identity || device.model || `Mikrotik-${device.ip}`,
          model: device.model,
          tags: 'discovered',
        }),
      });
      notify('Device added to NAS successfully', { type: 'success' });
      setSetupInfo({ device, secret, radiusServerIP });
    } catch (error: any) {
      notify(error.message || 'Failed to add device', { type: 'error' });
    } finally {
      setAddingDevices(prev => {
        const next = new Set(prev);
        next.delete(device.ip);
        return next;
      });
    }
  }, [notify, radiusServerIP]);

  const handleAddAll = useCallback(async () => {
    if (!scanResult) return;

    const devices = scanResult.results.filter(r => r.is_router_os);
    if (devices.length === 0) return;

    const secret = prompt('Enter RADIUS shared secret for all devices:', 'mikrotik');
    if (!secret) return;

    setAddingDevices(new Set(devices.map(d => d.ip)));

    try {
      const result = await apiRequest<any>('/network/discovery/bulk', {
        method: 'POST',
        body: JSON.stringify(devices.map(d => ({
          ip: d.ip,
          secret,
          name: d.identity || d.model || `Mikrotik-${d.ip}`,
          model: d.model,
          tags: 'discovered',
        }))),
      });
      notify(`Added ${result.added_count} devices to NAS`, { type: 'success' });
      // Show setup for first device as example
      if (devices.length > 0) {
        setSetupInfo({ device: devices[0], secret, radiusServerIP });
      }
    } catch (error: any) {
      notify(error.message || 'Failed to add devices', { type: 'error' });
    } finally {
      setAddingDevices(new Set());
    }
  }, [scanResult, notify, radiusServerIP]);

  const mikrotikDevices = scanResult?.results.filter(r => r.is_router_os) || [];
  const otherDevices = scanResult?.results.filter(r => !r.is_router_os) || [];

  const radiusIP = setupInfo?.radiusServerIP || radiusServerIP || '<RADIUS_SERVER_IP>';
  const secret = setupInfo?.secret || 'your-secret';

  return (
    <Box sx={{ p: 3 }}>
      {/* Header */}
      <Box sx={{ mb: 3 }}>
        <Typography variant="h5" sx={{ fontWeight: 600, mb: 1 }}>
          Network Discovery
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Scan your network to discover MikroTik RouterOS devices and add them to RADIUS
        </Typography>
        {!radiusServerIP && (
          <Alert severity="warning" sx={{ mt: 1 }}>
            RADIUS Server IP not configured. Go to <strong>Settings → System Config</strong> and set the{' '}
            <strong>radius.ServerIP</strong> to this server's IP address so discovered routers know where to send authentication requests.
          </Alert>
        )}
        {radiusServerIP && (
          <Alert severity="success" sx={{ mt: 1 }}>
            RADIUS Server IP: <strong>{radiusServerIP}</strong>. Discovered routers will be configured to point here.
          </Alert>
        )}
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
              sx={{ flexGrow: 1, maxWidth: 300 }}
              disabled={scanning}
              helperText="e.g., 192.168.1.0/24"
            />
            <TextField
              label="RouterOS Username"
              value={username}
              onChange={e => setUsername(e.target.value)}
              placeholder="admin"
              size="small"
              sx={{ width: 150 }}
              disabled={scanning}
            />
            <TextField
              label="RouterOS Password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              placeholder="password"
              type="password"
              size="small"
              sx={{ width: 150 }}
              disabled={scanning}
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
            <Button
              variant="outlined"
              onClick={async () => {
                try {
                  const testIp = prompt('Enter IP to test connection:', '10.0.0.1');
                  if (!testIp) return;
                  const result = await apiRequest<any>(`/network/discovery/test?ip=${testIp}&port=8728`);
                  notify(result?.message || 'Connection OK', { type: 'success' });
                } catch (error: any) {
                  notify(error.message || 'Connection failed', { type: 'error' });
                }
              }}
            >
              Test IP
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
                          <Tooltip title="Add to NAS & get setup guide">
                            <IconButton
                              size="small"
                              onClick={() => handleAddDevice(device)}
                              disabled={addingDevices.has(device.ip)}
                              color="primary"
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

          {/* Other Devices */}
          {otherDevices.length > 0 && (
            <Alert severity="info" sx={{ mb: 3 }}>
              Found {otherDevices.length} hosts but they are not MikroTik RouterOS devices.
              Only devices responding on RouterOS API ports (8728/8729) are shown above.
            </Alert>
          )}

          {/* Empty State */}
          {scanResult.total_hosts > 0 && scanResult.found_count === 0 && (
            <Alert severity="warning">
              No MikroTik devices found in the specified range. Make sure:
              <ul style={{ margin: '8px 0 0 0', paddingLeft: '20px' }}>
                <li>MikroTik devices are powered on and accessible</li>
                <li>RouterOS API service is enabled (IP → Services → api: enabled)</li>
                <li>Firewall allows access on ports 8728 or 8729</li>
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
            Supported: MikroTik RouterOS with API enabled on port 8728 or 8729
          </Typography>
        </Card>
      )}

      {/* Setup Instructions Dialog */}
      <Dialog
        open={!!setupInfo}
        onClose={() => setSetupInfo(null)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <CheckIcon color="success" />
          Device Added — Configure Mikrotik RADIUS
        </DialogTitle>
        <DialogContent>
          <DialogContentText component="div">
            <Alert severity="success" sx={{ mb: 2 }}>
              <strong>{setupInfo?.device.identity || setupInfo?.device.ip}</strong> has been added to NAS.
              Now configure RADIUS on the Mikrotik router using the commands below.
            </Alert>

            <Typography variant="subtitle1" fontWeight={600} gutterBottom>
              1. Add RADIUS Server (via Terminal / SSH)
            </Typography>
            <CodeBlock code={`/radius add address=${radiusIP} secret=${secret} service=hotspot,ppp authentication-port=1812 accounting-port=1813 comment="ToughRADIUS"`} />

            <Typography variant="subtitle1" fontWeight={600} gutterBottom sx={{ mt: 2 }}>
              2. Enable RADIUS for Hotspot
            </Typography>
            <CodeBlock code={`/ip hotspot profile set [find default=yes] use-radius=yes`} />

            <Typography variant="subtitle1" fontWeight={600} gutterBottom sx={{ mt: 2 }}>
              3. Enable RADIUS for PPPoE
            </Typography>
            <CodeBlock code={`/ppp aaa set use-radius=yes`} />

            <Typography variant="subtitle1" fontWeight={600} gutterBottom sx={{ mt: 2 }}>
              4. Verify RADIUS Connection
            </Typography>
            <CodeBlock code={`/radius print\n/radius monitor [find]`} />

            <Alert severity="info" sx={{ mt: 2 }}>
              <strong>Tip:</strong> Make sure this server (<strong>{radiusIP || 'your RADIUS server IP'}</strong>) allows UDP traffic on ports 1812 and 1813 from the Mikrotik router's IP (<strong>{setupInfo?.device.ip}</strong>).
              {!setupInfo?.radiusServerIP && (
                <> Configure the RADIUS Server IP in <strong>Settings → System Config → RADIUS Server IP</strong> to avoid using a placeholder here.</>
              )}
            </Alert>
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSetupInfo(null)} variant="contained">
            Done
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default NetworkDiscovery;
