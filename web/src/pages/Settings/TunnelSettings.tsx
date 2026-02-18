import React, { useState } from 'react';
import { useNotify } from 'react-admin';
import {
    Box,
    Card,
    CardContent,
    CardHeader,
    TextField,
    Button,
    Typography,
    Stack,
    Alert,
    Chip,
    CircularProgress
} from '@mui/material';
import SaveIcon from '@mui/icons-material/Save';
import RestartAltIcon from '@mui/icons-material/RestartAlt';
import CloudQueueIcon from '@mui/icons-material/CloudQueue';
import { useApiQuery } from '../../hooks/useApiQuery';
import { httpClient } from '../../utils/apiClient';

interface TunnelStatus {
    status: string; // "connected", "disconnected", "error"
    protocol: string;
    region: string;
    uptime: number; // seconds
}

interface TunnelConfig {
    tunnel_type: string;
    token: string;
}

export const TunnelSettings = () => {
    const notify = useNotify();
    const [saving, setSaving] = useState(false);
    const [restarting, setRestarting] = useState(false);
    const [token, setToken] = useState('');

    const { data: config, isLoading: configLoading } = useApiQuery<TunnelConfig>({
        path: '/system/tunnel/config',
        queryKey: ['tunnel', 'config'],
    });

    React.useEffect(() => {
        if (config?.token) {
            setToken(config.token);
        }
    }, [config]);

    const { data: status, refetch: refetchStatus } = useApiQuery<TunnelStatus>({
        path: '/system/tunnel/status',
        queryKey: ['tunnel', 'status'],
        refetchInterval: 5000,
    });

    const handleSave = async () => {
        setSaving(true);
        try {
            await httpClient('/system/tunnel/config', {
                method: 'POST',
                body: JSON.stringify({ tunnel_type: 'cloudflare', token }),
            });
            notify('Tunnel configuration saved', { type: 'success' });
            // Optionally restart after save
        } catch (error: any) {
            notify(error?.body?.msg || 'Failed to save configuration', { type: 'error' });
        } finally {
            setSaving(false);
        }
    };

    const handleRestart = async () => {
        setRestarting(true);
        try {
            await httpClient('/system/tunnel/restart', { method: 'POST' });
            notify('Tunnel service restart initiated', { type: 'success' });
            refetchStatus();
        } catch (error: any) {
            notify(error?.body?.msg || 'Failed to restart tunnel', { type: 'error' });
        } finally {
            setRestarting(false);
        }
    };

    if (configLoading) return <CircularProgress />;

    return (
        <Box sx={{ p: 3, maxWidth: 800 }}>
            <Box sx={{ mb: 3 }}>
                <Typography variant="h4" gutterBottom display="flex" alignItems="center" gap={1}>
                    <CloudQueueIcon fontSize="large" color="primary" />
                    Tunnel Management
                </Typography>
                <Typography variant="body1" color="textSecondary">
                    Configure remote access via Cloudflare Tunnel.
                </Typography>
            </Box>

            <Stack spacing={3}>
                {/* Status Card */}
                <Card>
                    <CardHeader title="Current Status" />
                    <CardContent>
                        <Stack direction="row" spacing={2} alignItems="center">
                            <Box sx={{
                                width: 12,
                                height: 12,
                                borderRadius: '50%',
                                bgcolor: status?.status === 'connected' ? 'success.main' : 'error.main'
                            }} />
                            <Typography variant="h6">
                                {status?.status?.toUpperCase() || 'UNKNOWN'}
                            </Typography>
                            {status?.region && <Chip label={`Region: ${status.region}`} size="small" />}
                            {status?.protocol && <Chip label={`Protocol: ${status.protocol}`} size="small" variant="outlined" />}
                        </Stack>
                    </CardContent>
                </Card>

                {/* Config Card */}
                <Card>
                    <CardHeader title="Configuration" />
                    <CardContent>
                        <Alert severity="info" sx={{ mb: 3 }}>
                            You need a Cloudflare Tunnel Token to enable remote access.
                            If you don't have one, create a tunnel in Cloudflare Zero Trust dashboard.
                        </Alert>

                        <TextField
                            label="Cloudflare Tunnel Token"
                            value={token}
                            onChange={(e) => setToken(e.target.value)}
                            fullWidth
                            type="password"
                            helperText="Paste your base64 encoded tunnel token here."
                            sx={{ mb: 3 }}
                        />

                        <Stack direction="row" spacing={2}>
                            <Button
                                variant="contained"
                                startIcon={<SaveIcon />}
                                onClick={handleSave}
                                disabled={saving || restarting}
                            >
                                {saving ? 'Saving...' : 'Save Configuration'}
                            </Button>

                            <Button
                                variant="outlined"
                                color="warning"
                                startIcon={<RestartAltIcon />}
                                onClick={handleRestart}
                                disabled={saving || restarting}
                            >
                                {restarting ? 'Restarting...' : 'Restart Tunnel Service'}
                            </Button>
                        </Stack>
                    </CardContent>
                </Card>
            </Stack>
        </Box>
    );
};
