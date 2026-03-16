import { useState, useCallback } from 'react';
import {
  Box,
  Typography,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
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
  Stack,
  LinearProgress,
  Tooltip,
} from '@mui/material';
import {
  Search as SearchIcon,
  Add as AddIcon,
  Devices as DeviceIcon,
} from '@mui/icons-material';
import { useNotify, useRefresh, useTranslate, useLocale } from 'react-admin';
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
}

interface ScanResult {
  cidr: string;
  duration: number;
  found_count: number;
  total_hosts: number;
  results: DiscoveredDevice[];
}

interface ScanNetworkModalProps {
  open: boolean;
  onClose: () => void;
}

export const ScanNetworkModal = ({ open, onClose }: ScanNetworkModalProps) => {
  const notify = useNotify();
  const refresh = useRefresh();
  const translate = useTranslate();
  const locale = useLocale();
  const isRTL = locale === 'ar';

  const [ipRange, setIpRange] = useState('192.168.1.0/24');
  const [username, setUsername] = useState('admin');
  const [password, setPassword] = useState('');
  const [scanning, setScanning] = useState(false);
  const [scanResult, setScanResult] = useState<ScanResult | null>(null);
  const [addingDevices, setAddingDevices] = useState<Set<string>>(new Set());

  const handleScan = useCallback(async () => {
    if (!ipRange.trim()) {
      notify(translate('pages.scan_modal.enter_ip_range', { _: 'Please enter an IP range' }), { type: 'warning' });
      return;
    }

    setScanning(true);
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
        notify(translate('pages.scan_modal.found_devices', { count: result.found_count, _: `Found ${result.found_count} MikroTik devices` }), { type: 'success' });
      }
    } catch (error: any) {
      console.error('Scan error:', error);
      if (error.name === 'AbortError') {
        notify(translate('pages.scan_modal.scan_timeout', { _: 'Scan timed out' }), { type: 'error' });
      } else {
        notify(error.message || translate('pages.scan_modal.scan_failed', { _: 'Scan failed' }), { type: 'error' });
      }
    } finally {
      setScanning(false);
    }
  }, [ipRange, username, password, notify]);

  const handleAddDevice = useCallback(async (device: DiscoveredDevice) => {
    const secret = prompt(translate('pages.scan_modal.enter_secret', { _: 'Enter RADIUS secret for this device:' }), translate('pages.scan_modal.default_secret', { _: 'mikrotik' }));
    if (!secret) return;

    setAddingDevices(prev => new Set(prev).add(device.ip));

    try {
      await apiRequest('/network/discovery', {
        method: 'POST',
        body: JSON.stringify({
          ip: device.ip,
          secret: secret,
          name: device.identity || device.model || `Mikrotik-${device.ip}`,
          model: device.model,
          tags: 'discovered',
        }),
      });
      notify(translate('pages.scan_modal.device_added', { _: 'Device added to NAS successfully' }), { type: 'success' });
      refresh();
    } catch (error: any) {
      notify(error.message || translate('pages.scan_modal.scan_failed', { _: 'Failed to add device' }), { type: 'error' });
    } finally {
      setAddingDevices(prev => {
        const next = new Set(prev);
        next.delete(device.ip);
        return next;
      });
    }
  }, [notify, refresh]);

  const handleAddAll = useCallback(async () => {
    if (!scanResult) return;

    const devices = scanResult.results.filter(r => r.is_router_os);
    if (devices.length === 0) return;

    const secret = prompt(translate('pages.scan_modal.enter_secret_all', { _: 'Enter RADIUS secret for all devices:' }), translate('pages.scan_modal.default_secret', { _: 'mikrotik' }));
    if (!secret) return;

    setAddingDevices(new Set(devices.map(d => d.ip)));

    try {
      const result = await apiRequest<any>('/network/discovery/bulk', {
        method: 'POST',
        body: JSON.stringify(devices.map(d => ({
          ip: d.ip,
          secret: secret,
          name: d.identity || d.model || `Mikrotik-${d.ip}`,
          model: d.model,
          tags: 'discovered',
        }))),
      });
      notify(translate('pages.scan_modal.devices_added', { count: result.added_count, _: `Added ${result.added_count} devices to NAS` }), { type: 'success' });
      refresh();
    } catch (error: any) {
      notify(error.message || 'Failed to add devices', { type: 'error' });
    } finally {
      setAddingDevices(new Set());
    }
  }, [scanResult, notify, refresh]);

  const mikrotikDevices = scanResult?.results.filter(r => r.is_router_os) || [];

  const handleClose = () => {
    setScanResult(null);
    setIpRange('192.168.1.0/24');
    setUsername('admin');
    setPassword('');
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth dir={isRTL ? 'rtl' : 'ltr'}>
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <DeviceIcon color="primary" />
        {translate('pages.scan_modal.title', { _: 'Network Discovery' })}
      </DialogTitle>
      <DialogContent>
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ mb: 3, mt: 1 }} dir={isRTL ? 'rtl' : 'ltr'}>
          <TextField
            label={translate('pages.scan_modal.ip_range', { _: 'IP Range (CIDR)' })}
            value={ipRange}
            onChange={e => setIpRange(e.target.value)}
            placeholder={translate('pages.scan_modal.ip_range_placeholder', { _: '192.168.1.0/24' })}
            size="small"
            sx={{ flexGrow: 1 }}
            disabled={scanning}
          />
          <TextField
            label={translate('pages.scan_modal.username', { _: 'Username' })}
            value={username}
            onChange={e => setUsername(e.target.value)}
            size="small"
            sx={{ width: 120 }}
            disabled={scanning}
          />
          <TextField
            label={translate('pages.scan_modal.password', { _: 'Password' })}
            value={password}
            onChange={e => setPassword(e.target.value)}
            type="password"
            size="small"
            sx={{ width: 120 }}
            disabled={scanning}
          />
          <Button
            variant="contained"
            startIcon={scanning ? <CircularProgress size={20} color="inherit" /> : <SearchIcon />}
            onClick={handleScan}
            disabled={scanning || !ipRange.trim()}
          >
            {scanning ? translate('pages.scan_modal.scanning', { _: 'Scanning...' }) : translate('pages.scan_modal.scan', { _: 'Scan' })}
          </Button>
        </Stack>

        {scanning && (
          <Box sx={{ mb: 2 }}>
            <LinearProgress />
            <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
              {translate('pages.scan_modal.scanning_network', { _: 'Scanning network for MikroTik devices...' })}
            </Typography>
          </Box>
        )}

        {scanResult && mikrotikDevices.length > 0 && (
          <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="body2">
              {translate('pages.scan_modal.found_devices', { count: mikrotikDevices.length, _: 'Found <strong>{count}</strong> MikroTik devices' })}
            </Typography>
            <Button
              variant="contained"
              size="small"
              startIcon={<AddIcon />}
              onClick={handleAddAll}
              disabled={addingDevices.size > 0}
            >
              {translate('pages.scan_modal.add_all', { _: 'Add All' })}
            </Button>
          </Box>
        )}

        {mikrotikDevices.length > 0 && (
          <TableContainer sx={{ maxHeight: 400 }}>
            <Table size="small" stickyHeader>
              <TableHead>
                <TableRow>
                  <TableCell>{translate('pages.scan_modal.ip_address', { _: 'IP Address' })}</TableCell>
                  <TableCell>{translate('pages.scan_modal.port', { _: 'Port' })}</TableCell>
                  <TableCell>{translate('pages.scan_modal.identity', { _: 'Identity' })}</TableCell>
                  <TableCell>{translate('pages.scan_modal.model', { _: 'Model' })}</TableCell>
                  <TableCell>{translate('pages.scan_modal.version', { _: 'Version' })}</TableCell>
                  <TableCell align="right">{translate('pages.scan_modal.actions', { _: 'Actions' })}</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {mikrotikDevices.map((device) => (
                  <TableRow key={device.ip} hover>
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
                    <TableCell align="right">
                      <Tooltip title={translate('pages.scan_modal.add_to_nas', { _: 'Add to NAS' })}>
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
        )}

        {scanResult && scanResult.found_count === 0 && (
          <Typography color="text.secondary" sx={{ textAlign: 'center', py: 4 }}>
            {translate('pages.scan_modal.no_devices_found', { _: 'No MikroTik devices found in the specified IP range.' })}
          </Typography>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>{translate('pages.scan_modal.close', { _: 'Close' })}</Button>
      </DialogActions>
    </Dialog>
  );
};
