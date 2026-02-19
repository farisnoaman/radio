import { useState } from 'react';
import { useNotify, useTranslate } from 'react-admin';
import {
    Box,
    Button,
    Typography,
    Alert,
    Switch,
    FormControlLabel,
    TextField,
    Accordion,
    AccordionSummary,
    AccordionDetails,
    CircularProgress,
    useTheme,
    useMediaQuery,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ConstructionIcon from '@mui/icons-material/Construction';
import DeleteSweepIcon from '@mui/icons-material/DeleteSweep';
import SaveIcon from '@mui/icons-material/Save';
import { useApiQuery } from '../../hooks/useApiQuery';
import { httpClient } from '../../utils/apiClient';

export const MaintenanceSettings = () => {
    const notify = useNotify();
    const translate = useTranslate();
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

    const [archiving, setArchiving] = useState(false);
    const [toggling, setToggling] = useState(false);
    const [drain, setDrain] = useState(true);
    const [days, setDays] = useState(30);
    const [expanded, setExpanded] = useState<string | false>('maintenance');

    const { data: status, refetch, isLoading } = useApiQuery<{ active: boolean }>({
        path: '/system/maintenance',
        queryKey: ['system', 'maintenance'],
    });

    const handleAccordionChange = (panel: string) => (_: React.SyntheticEvent, isExpanded: boolean) => {
        setExpanded(isExpanded ? panel : false);
    };

    const handleToggleMaintenance = async () => {
        setToggling(true);
        try {
            const action = status?.active ? 'disable' : 'enable';
            const path = `/system/maintenance/${action}${action === 'enable' ? `?drain=${drain}` : ''}`;
            await httpClient(path, { method: 'POST' });
            notify(`Maintenance mode ${action}d successfully`, { type: 'success' });
            refetch();
        } catch (error: any) {
            notify(error?.body?.message || 'Failed to toggle maintenance mode', { type: 'error' });
        } finally {
            setToggling(false);
        }
    };

    const handleArchiveLogs = async () => {
        if (!window.confirm(`Are you sure you want to archive and compress logs older than ${days} days?`)) return;
        setArchiving(true);
        try {
            await httpClient(`/system/logs/archive?days=${days}`, { method: 'POST' });
            notify('Logs archived and compressed successfully', { type: 'success' });
        } catch (error: any) {
            notify(error?.body?.message || 'Log archival failed', { type: 'error' });
        } finally {
            setArchiving(false);
        }
    };

    if (isLoading) {
        return (
            <Box sx={{ p: 3, display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
                <CircularProgress />
            </Box>
        );
    }

    return (
        <Box sx={{ p: { xs: 2, sm: 3 } }}>
            {/* Page Title */}
            <Box sx={{ mb: 3 }}>
                <Typography variant={isMobile ? 'h5' : 'h4'} gutterBottom>
                    {translate('pages.maintenance.title', { _: 'System Maintenance' })}
                </Typography>
                <Typography variant="body1" color="textSecondary">
                    {translate('pages.maintenance.subtitle', { _: 'Manage system availability and log data lifecycle.' })}
                </Typography>
            </Box>

            {/* Info Alert */}
            <Alert severity="info" sx={{ mb: 3 }}>
                {translate('pages.maintenance.info_message', { _: 'Maintenance operations affect system availability. Use with caution.' })}
            </Alert>

            {/* Maintenance Mode Section */}
            <Accordion
                expanded={expanded === 'maintenance'}
                onChange={handleAccordionChange('maintenance')}
                sx={{
                    mb: 2,
                    border: status?.active ? '2px solid #ff9800' : 'none',
                }}
            >
                <AccordionSummary
                    expandIcon={<ExpandMoreIcon />}
                    sx={{
                        backgroundColor: status?.active
                            ? (theme.palette.mode === 'dark' ? 'rgba(255, 152, 0, 0.2)' : '#fff3e0')
                            : (theme.palette.mode === 'dark' ? 'rgba(25, 118, 210, 0.15)' : '#e3f2fd'),
                        '&:hover': {
                            backgroundColor: status?.active
                                ? (theme.palette.mode === 'dark' ? 'rgba(255, 152, 0, 0.3)' : '#ffe0b2')
                                : (theme.palette.mode === 'dark' ? 'rgba(25, 118, 210, 0.25)' : '#bbdefb')
                        }
                    }}
                >
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: { xs: 1, sm: 2 } }}>
                        <Box sx={{ color: status?.active ? '#ff9800' : '#1976d2' }}>
                            <ConstructionIcon />
                        </Box>
                        <Box>
                            <Typography variant="h6" sx={{ color: status?.active ? '#ff9800' : '#1976d2', fontSize: isMobile ? '1rem' : '1.25rem' }}>
                                {translate('pages.maintenance.maintenance_mode.title', { _: 'Maintenance Mode' })}
                            </Typography>
                            <Typography variant="body2" color="textSecondary" sx={{ display: { xs: 'none', sm: 'block' } }}>
                                {translate('pages.maintenance.maintenance_mode.description', { _: 'Block non-admin users and drain active sessions.' })}
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
                            {status?.active ? (
                                <Alert severity="warning" sx={{ mb: 2 }}>
                                    {translate('pages.maintenance.maintenance_mode.active_warning', { _: 'System is currently in MAINTENANCE MODE. Only administrators can access the API.' })}
                                </Alert>
                            ) : (
                                <Alert severity="info" sx={{ mb: 2 }}>
                                    {translate('pages.maintenance.maintenance_mode.normal_info', { _: 'System is running normally.' })}
                                </Alert>
                            )}

                            <Box sx={{ mb: 2 }}>
                                <FormControlLabel
                                    control={
                                        <Switch
                                            checked={status?.active || false}
                                            onChange={handleToggleMaintenance}
                                            disabled={toggling}
                                            color="warning"
                                        />
                                    }
                                    label={status?.active
                                        ? translate('pages.maintenance.maintenance_mode.status_active', { _: 'Active' })
                                        : translate('pages.maintenance.maintenance_mode.status_inactive', { _: 'Inactive' })
                                    }
                                />
                            </Box>

                            {!status?.active && (
                                <FormControlLabel
                                    control={
                                        <Switch
                                            checked={drain}
                                            onChange={(e) => setDrain(e.target.checked)}
                                        />
                                    }
                                    label={translate('pages.maintenance.maintenance_mode.drain_label', { _: 'Drain sessions (Disconnect online users upon enabling)' })}
                                />
                            )}

                            <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mt: 2 }}>
                                {translate('pages.maintenance.maintenance_mode.help', { _: 'Enabling maintenance mode will prevent new user logins and optionally disconnect all current users.' })}
                            </Typography>
                        </Box>
                    </Box>
                </AccordionDetails>
            </Accordion>

            {/* Log Archival Section */}
            <Accordion
                expanded={expanded === 'logs'}
                onChange={handleAccordionChange('logs')}
                sx={{ mb: 2 }}
            >
                <AccordionSummary
                    expandIcon={<ExpandMoreIcon />}
                    sx={{
                        backgroundColor: theme.palette.mode === 'dark' ? 'rgba(156, 39, 176, 0.15)' : '#f3e5f5',
                        '&:hover': {
                            backgroundColor: theme.palette.mode === 'dark' ? 'rgba(156, 39, 176, 0.25)' : '#e1bee7'
                        }
                    }}
                >
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: { xs: 1, sm: 2 } }}>
                        <Box sx={{ color: '#9c27b0' }}>
                            <DeleteSweepIcon />
                        </Box>
                        <Box>
                            <Typography variant="h6" sx={{ color: '#9c27b0', fontSize: isMobile ? '1rem' : '1.25rem' }}>
                                {translate('pages.maintenance.log_archival.title', { _: 'System Log Archival' })}
                            </Typography>
                            <Typography variant="body2" color="textSecondary" sx={{ display: { xs: 'none', sm: 'block' } }}>
                                {translate('pages.maintenance.log_archival.description', { _: 'Archive and compress old logs to save disk space.' })}
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
                            <Box sx={{
                                display: 'flex',
                                flexDirection: isMobile ? 'column' : 'row',
                                gap: 2,
                                alignItems: isMobile ? 'stretch' : 'flex-start',
                                mb: 2
                            }}>
                                <TextField
                                    label={translate('pages.maintenance.log_archival.retention_days', { _: 'Retention Days' })}
                                    type="number"
                                    value={days}
                                    onChange={(e) => setDays(Number(e.target.value))}
                                    sx={{ width: isMobile ? '100%' : 150 }}
                                    helperText={isMobile ? translate('pages.maintenance.log_archival.retention_help', { _: 'Logs older than this will be compressed.' }) : undefined}
                                />
                                <Button
                                    variant="contained"
                                    color="secondary"
                                    startIcon={archiving ? <CircularProgress size={20} color="inherit" /> : <SaveIcon />}
                                    onClick={handleArchiveLogs}
                                    disabled={archiving}
                                    sx={{ height: isMobile ? undefined : 56 }}
                                    fullWidth={isMobile}
                                >
                                    {archiving
                                        ? translate('pages.maintenance.log_archival.archiving', { _: 'Archiving...' })
                                        : translate('pages.maintenance.log_archival.archive_now', { _: 'Archive Now' })
                                    }
                                </Button>
                            </Box>
                            {!isMobile && (
                                <Typography variant="caption" color="textSecondary">
                                    {translate('pages.maintenance.log_archival.help', { _: 'Compressed logs are stored in the system log directory as .csv.gz files. This process is also triggered automatically on a daily schedule.' })}
                                </Typography>
                            )}
                        </Box>
                    </Box>
                </AccordionDetails>
            </Accordion>
        </Box>
    );
};
