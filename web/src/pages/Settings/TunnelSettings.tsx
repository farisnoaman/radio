import React, { useState } from 'react';
import { useNotify, useTranslate } from 'react-admin';
import {
    Box,
    TextField,
    Button,
    Typography,
    Alert,
    Chip,
    CircularProgress,
    Accordion,
    AccordionSummary,
    AccordionDetails,
    useTheme,
    useMediaQuery,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import SaveIcon from '@mui/icons-material/Save';
import RestartAltIcon from '@mui/icons-material/RestartAlt';
import CloudQueueIcon from '@mui/icons-material/CloudQueue';
import InfoIcon from '@mui/icons-material/Info';
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
    const translate = useTranslate();
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

    const [saving, setSaving] = useState(false);
    const [restarting, setRestarting] = useState(false);
    const [token, setToken] = useState('');
    const [expanded, setExpanded] = useState<string | false>('status');

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

    const handleAccordionChange = (panel: string) => (_: React.SyntheticEvent, isExpanded: boolean) => {
        setExpanded(isExpanded ? panel : false);
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            await httpClient('/system/tunnel/config', {
                method: 'POST',
                body: JSON.stringify({ tunnel_type: 'cloudflare', token }),
            });
            notify('Tunnel configuration saved', { type: 'success' });
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

    if (configLoading) {
        return (
            <Box sx={{ p: 3, display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
                <CircularProgress />
            </Box>
        );
    }

    const isConnected = status?.status === 'connected';

    return (
        <Box sx={{ p: { xs: 2, sm: 3 } }}>
            {/* Page Title */}
            <Box sx={{ mb: 3 }}>
                <Typography variant={isMobile ? 'h5' : 'h4'} gutterBottom>
                    {translate('pages.tunnel.title', { _: 'Tunnel Management' })}
                </Typography>
                <Typography variant="body1" color="textSecondary">
                    {translate('pages.tunnel.subtitle', { _: 'Configure remote access via Cloudflare Tunnel.' })}
                </Typography>
            </Box>

            {/* Info Alert */}
            <Alert severity="info" sx={{ mb: 3 }}>
                {translate('pages.tunnel.info_message', { _: 'Tunnel provides secure remote access without opening ports. Requires a Cloudflare Zero Trust account.' })}
            </Alert>

            {/* Status Section */}
            <Accordion
                expanded={expanded === 'status'}
                onChange={handleAccordionChange('status')}
                sx={{ mb: 2 }}
            >
                <AccordionSummary
                    expandIcon={<ExpandMoreIcon />}
                    sx={{
                        backgroundColor: isConnected
                            ? (theme.palette.mode === 'dark' ? 'rgba(46, 125, 50, 0.2)' : '#e8f5e9')
                            : (theme.palette.mode === 'dark' ? 'rgba(211, 47, 47, 0.2)' : '#ffebee'),
                        '&:hover': {
                            backgroundColor: isConnected
                                ? (theme.palette.mode === 'dark' ? 'rgba(46, 125, 50, 0.3)' : '#c8e6c9')
                                : (theme.palette.mode === 'dark' ? 'rgba(211, 47, 47, 0.3)' : '#ffcdd2')
                        }
                    }}
                >
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: { xs: 1, sm: 2 } }}>
                        <Box sx={{ color: isConnected ? '#2e7d32' : '#d32f2f' }}>
                            <CloudQueueIcon />
                        </Box>
                        <Box>
                            <Typography variant="h6" sx={{ color: isConnected ? '#2e7d32' : '#d32f2f', fontSize: isMobile ? '1rem' : '1.25rem' }}>
                                {translate('pages.tunnel.status.title', { _: 'Current Status' })}
                            </Typography>
                            <Typography variant="body2" color="textSecondary" sx={{ display: { xs: 'none', sm: 'block' } }}>
                                {translate('pages.tunnel.status.description', { _: 'Real-time tunnel connection status and details.' })}
                            </Typography>
                        </Box>
                    </Box>
                </AccordionSummary>
                <AccordionDetails>
                    <Box sx={{
                        display: 'grid',
                        gridTemplateColumns: { xs: '1fr', md: 'repeat(auto-fit, minmax(400px, 1fr))' },
                        gap: 3
                    }}>
                        <Box>
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                                <Box
                                    sx={{
                                        width: 16,
                                        height: 16,
                                        borderRadius: '50%',
                                        bgcolor: isConnected ? 'success.main' : 'error.main',
                                        animation: isConnected ? 'pulse 2s infinite' : 'none',
                                        '@keyframes pulse': {
                                            '0%': { opacity: 1 },
                                            '50%': { opacity: 0.5 },
                                            '100%': { opacity: 1 },
                                        },
                                    }}
                                />
                                <Typography variant="h6">
                                    {status?.status?.toUpperCase() || 'UNKNOWN'}
                                </Typography>
                            </Box>
                            <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                                {status?.region && (
                                    <Chip
                                        label={`${translate('pages.tunnel.status.region', { _: 'Region' })}: ${status.region}`}
                                        size="small"
                                        color="primary"
                                        variant="outlined"
                                    />
                                )}
                                {status?.protocol && (
                                    <Chip
                                        label={`${translate('pages.tunnel.status.protocol', { _: 'Protocol' })}: ${status.protocol}`}
                                        size="small"
                                        variant="outlined"
                                    />
                                )}
                                {status?.uptime && (
                                    <Chip
                                        label={`${translate('pages.tunnel.status.uptime', { _: 'Uptime' })}: ${Math.floor(status.uptime / 60)}m`}
                                        size="small"
                                        variant="outlined"
                                    />
                                )}
                            </Box>
                        </Box>
                    </Box>
                </AccordionDetails>
            </Accordion>

            {/* Configuration Section */}
            <Accordion
                expanded={expanded === 'config'}
                onChange={handleAccordionChange('config')}
                sx={{ mb: 2 }}
            >
                <AccordionSummary
                    expandIcon={<ExpandMoreIcon />}
                    sx={{
                        backgroundColor: theme.palette.mode === 'dark' ? 'rgba(25, 118, 210, 0.15)' : '#e3f2fd',
                        '&:hover': {
                            backgroundColor: theme.palette.mode === 'dark' ? 'rgba(25, 118, 210, 0.25)' : '#bbdefb'
                        }
                    }}
                >
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: { xs: 1, sm: 2 } }}>
                        <Box sx={{ color: '#1976d2' }}>
                            <InfoIcon />
                        </Box>
                        <Box>
                            <Typography variant="h6" sx={{ color: '#1976d2', fontSize: isMobile ? '1rem' : '1.25rem' }}>
                                {translate('pages.tunnel.config.title', { _: 'Configuration' })}
                            </Typography>
                            <Typography variant="body2" color="textSecondary" sx={{ display: { xs: 'none', sm: 'block' } }}>
                                {translate('pages.tunnel.config.description', { _: 'Set up your Cloudflare Tunnel token for remote access.' })}
                            </Typography>
                        </Box>
                    </Box>
                </AccordionSummary>
                <AccordionDetails>
                    <Box sx={{
                        display: 'grid',
                        gridTemplateColumns: { xs: '1fr', md: 'repeat(auto-fit, minmax(400px, 1fr))' },
                        gap: 3
                    }}>
                        <Box>
                            <Alert severity="info" sx={{ mb: 3 }}>
                                {translate('pages.tunnel.config.token_info', { _: 'You need a Cloudflare Tunnel Token to enable remote access. If you don\'t have one, create a tunnel in Cloudflare Zero Trust dashboard.' })}
                            </Alert>

                            <TextField
                                label={translate('pages.tunnel.config.token_label', { _: 'Cloudflare Tunnel Token' })}
                                value={token}
                                onChange={(e) => setToken(e.target.value)}
                                fullWidth
                                type="password"
                                helperText={translate('pages.tunnel.config.token_help', { _: 'Paste your base64 encoded tunnel token here.' })}
                                sx={{ mb: 3 }}
                            />

                            <Box sx={{
                                display: 'flex',
                                flexDirection: isMobile ? 'column' : 'row',
                                gap: 2
                            }}>
                                <Button
                                    variant="contained"
                                    startIcon={saving ? <CircularProgress size={20} color="inherit" /> : <SaveIcon />}
                                    onClick={handleSave}
                                    disabled={saving || restarting}
                                    fullWidth={isMobile}
                                >
                                    {saving
                                        ? translate('pages.tunnel.config.saving', { _: 'Saving...' })
                                        : translate('pages.tunnel.config.save', { _: 'Save Configuration' })
                                    }
                                </Button>

                                <Button
                                    variant="outlined"
                                    color="warning"
                                    startIcon={restarting ? <CircularProgress size={20} color="inherit" /> : <RestartAltIcon />}
                                    onClick={handleRestart}
                                    disabled={saving || restarting}
                                    fullWidth={isMobile}
                                >
                                    {restarting
                                        ? translate('pages.tunnel.config.restarting', { _: 'Restarting...' })
                                        : translate('pages.tunnel.config.restart', { _: 'Restart Tunnel Service' })
                                    }
                                </Button>
                            </Box>
                        </Box>
                    </Box>
                </AccordionDetails>
            </Accordion>
        </Box>
    );
};
