import React, { useState, useEffect } from 'react';
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
    FormControl,
    InputLabel,
    Select,
    MenuItem,
    Chip,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ConstructionIcon from '@mui/icons-material/Construction';
import DeleteSweepIcon from '@mui/icons-material/DeleteSweep';
import SaveIcon from '@mui/icons-material/Save';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import RestoreIcon from '@mui/icons-material/Restore';
import DownloadIcon from '@mui/icons-material/Download';
import StorageIcon from '@mui/icons-material/Storage';
import CloudIcon from '@mui/icons-material/Cloud';
import HistoryIcon from '@mui/icons-material/History';
import SettingsIcon from '@mui/icons-material/Settings';
import { useApiQuery } from '../hooks/useApiQuery';
import { httpClient } from '../utils/apiClient';

interface BackupConfig {
    provider: string;
    retention_days: number;
    schedule: string;
    gdrive_client_id?: string;
    gdrive_client_secret?: string;
    gdrive_folder_id?: string;
}

interface BackupItem {
    id: string;
    file_name: string;
    size: number;
    created_at: string;
    restored_at?: string;
    type: string;
}

export const SystemSettings = () => {
    const notify = useNotify();
    const translate = useTranslate();
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

    const [archiving, setArchiving] = useState(false);
    const [toggling, setToggling] = useState(false);
    const [drain, setDrain] = useState(true);
    const [days, setDays] = useState(30);

    const [saving, setSaving] = useState(false);
    const [testing, setTesting] = useState(false);
    const [restoring, setRestoring] = useState(false);
    const [config, setConfig] = useState<BackupConfig>({
        provider: 'local',
        retention_days: 7,
        schedule: '0 0 2 * * *'
    });

    const [expanded, setExpanded] = useState<string | false>('maintenance');

    const { data: status, refetch: refetchMaintenance, isLoading: maintenanceLoading } = useApiQuery<{ active: boolean }>({
        path: '/system/maintenance',
        queryKey: ['system', 'maintenance'],
    });

    const { data: remoteConfig, isLoading: configLoading } = useApiQuery<BackupConfig>({
        path: '/system/settings?type=backup',
        queryKey: ['backup', 'config'],
    });

    const { data: backups, refetch: refetchBackups } = useApiQuery<BackupItem[]>({
        path: '/system/backup',
        queryKey: ['backup', 'list'],
    });

    useEffect(() => {
        if (remoteConfig && (remoteConfig as any).data) {
            const data = (remoteConfig as any).data;
            if (Array.isArray(data)) {
                const configSetting = data.find((s: any) => s.name === 'config');
                if (configSetting?.value) {
                    try {
                        setConfig(JSON.parse(configSetting.value));
                    } catch (e) {
                        console.error("Failed to parse backup config", e);
                    }
                }
            }
        }
    }, [remoteConfig]);

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
            refetchMaintenance();
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

    const handleSaveBackup = async () => {
        setSaving(true);
        try {
            await httpClient('/system/settings', {
                method: 'POST',
                body: JSON.stringify({ type: 'backup', name: 'config', value: JSON.stringify(config) }),
            });
            notify('Backup configuration saved', { type: 'success' });
        } catch (error: any) {
            notify(error?.body?.msg || 'Failed to save configuration', { type: 'error' });
        } finally {
            setSaving(false);
        }
    };

    const handleTestUpload = async () => {
        setTesting(true);
        try {
            await httpClient('/system/backup', { method: 'POST' });
            notify('Backup created successfully', { type: 'success' });
            refetchBackups();
        } catch (error: any) {
            notify(error?.body?.msg || 'Backup failed', { type: 'error' });
        } finally {
            setTesting(false);
        }
    };

    const handleDownload = async (id: string, filename?: string) => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`/api/v1/system/backup/${id}/download`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });

            if (!response.ok) {
                throw new Error('Download failed');
            }

            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename || `backup-${id}.db`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        } catch (error: any) {
            notify(error?.message || 'Download failed', { type: 'error' });
        }
    };

    const handleRestore = async (id: string) => {
        if (!window.confirm(translate('pages.backup.backups.restore_confirm', { _: 'Are you sure you want to restore this backup? This will overwrite the current database!' }))) return;
        setRestoring(true);
        try {
            await httpClient(`/system/backup/${id}/restore`, { method: 'POST' });
            notify('Database restored successfully', { type: 'success' });
            refetchBackups();
        } catch (error: any) {
            notify(error?.body?.msg || 'Restore failed', { type: 'error' });
        } finally {
            setRestoring(false);
        }
    };

    if (maintenanceLoading || configLoading) {
        return (
            <Box sx={{ p: 3, display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
                <CircularProgress />
            </Box>
        );
    }

    const formatSize = (bytes: number) => {
        if (bytes < 1024) return `${bytes} B`;
        if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
        return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
    };

    const formatDateTime = (dateStr: string) => {
        try {
            const date = new Date(dateStr);
            return date.toLocaleString();
        } catch {
            return dateStr;
        }
    };

    return (
        <Box sx={{ p: { xs: 2, sm: 3 } }}>
            <Box sx={{ mb: 3 }}>
                <Typography variant={isMobile ? 'h5' : 'h4'} gutterBottom>
                    {translate('pages.maintenance.title', { _: 'System Maintenance & Backup' })}
                </Typography>
                <Typography variant="body1" color="textSecondary">
                    {translate('pages.maintenance.subtitle', { _: 'Manage system availability, log data lifecycle, and backup configuration.' })}
                </Typography>
            </Box>

            <Alert severity="info" sx={{ mb: 3 }}>
                {translate('pages.maintenance.info_message', { _: 'Maintenance operations affect system availability. Regular backups protect your data.' })}
            </Alert>

            <Box sx={{
                mb: 3,
                display: 'flex',
                flexDirection: isMobile ? 'column' : 'row',
                gap: 2
            }}>
                <Button
                    variant="contained"
                    startIcon={saving ? <CircularProgress size={20} color="inherit" /> : <SaveIcon />}
                    onClick={handleSaveBackup}
                    disabled={saving || testing}
                    fullWidth={isMobile}
                >
                    {saving
                        ? translate('pages.backup.actions.saving', { _: 'Saving...' })
                        : translate('pages.backup.actions.save', { _: 'Save Configuration' })
                    }
                </Button>

                <Button
                    variant="outlined"
                    color="secondary"
                    startIcon={testing ? <CircularProgress size={20} color="inherit" /> : <CloudUploadIcon />}
                    onClick={handleTestUpload}
                    disabled={saving || testing}
                    fullWidth={isMobile}
                >
                    {testing
                        ? translate('pages.backup.actions.creating', { _: 'Creating Backup...' })
                        : translate('pages.backup.actions.backup_now', { _: 'Backup Now' })
                    }
                </Button>
            </Box>

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

            <Accordion
                expanded={expanded === 'general'}
                onChange={handleAccordionChange('general')}
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
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                        <Box sx={{ color: '#1976d2' }}>
                            <SettingsIcon />
                        </Box>
                        <Box>
                            <Typography variant="h6" sx={{ color: '#1976d2', fontSize: isMobile ? '1rem' : '1.25rem' }}>
                                {translate('pages.backup.general.title', { _: 'Backup Settings' })}
                            </Typography>
                            <Typography variant="body2" color="textSecondary">
                                {translate('pages.backup.general.description', { _: 'Configure backup provider and schedule.' })}
                            </Typography>
                        </Box>
                    </Box>
                </AccordionSummary>
                <AccordionDetails>
                    <Box sx={{
                        display: 'grid',
                        gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' },
                        gap: 3
                    }}>
                        <Box>
                            <FormControl fullWidth sx={{ mb: 2 }}>
                                <InputLabel>{translate('pages.backup.general.provider_label', { _: 'Backup Provider' })}</InputLabel>
                                <Select
                                    value={config.provider}
                                    label={translate('pages.backup.general.provider_label', { _: 'Backup Provider' })}
                                    onChange={(e) => setConfig({ ...config, provider: e.target.value })}
                                >
                                    <MenuItem value="local">
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                            <StorageIcon fontSize="small" />
                                            {translate('pages.backup.general.provider_local', { _: 'Local Storage' })}
                                        </Box>
                                    </MenuItem>
                                    <MenuItem value="gdrive">
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                            <CloudIcon fontSize="small" />
                                            {translate('pages.backup.general.provider_gdrive', { _: 'Google Drive' })}
                                        </Box>
                                    </MenuItem>
                                </Select>
                            </FormControl>

                            <TextField
                                label={translate('pages.backup.general.retention_label', { _: 'Retention Days' })}
                                type="number"
                                value={config.retention_days}
                                onChange={(e) => setConfig({ ...config, retention_days: Number(e.target.value) })}
                                fullWidth
                                helperText={translate('pages.backup.general.retention_help', { _: 'How many days to keep backup files.' })}
                                sx={{ mb: 2 }}
                            />

                            <TextField
                                label={translate('pages.backup.general.schedule_label', { _: 'Schedule (Cron)' })}
                                value={config.schedule}
                                onChange={(e) => setConfig({ ...config, schedule: e.target.value })}
                                fullWidth
                                helperText={translate('pages.backup.general.schedule_help', { _: 'Cron expression for automated backups (e.g. 0 0 2 * * *)' })}
                            />
                        </Box>
                    </Box>
                </AccordionDetails>
            </Accordion>

            {config.provider === 'gdrive' && (
                <Accordion
                    expanded={expanded === 'gdrive'}
                    onChange={handleAccordionChange('gdrive')}
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
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                            <Box sx={{ color: '#9c27b0' }}>
                                <CloudIcon />
                            </Box>
                            <Box>
                                <Typography variant="h6" sx={{ color: '#9c27b0', fontSize: isMobile ? '1rem' : '1.25rem' }}>
                                    {translate('pages.backup.gdrive.title', { _: 'Google Drive Configuration' })}
                                </Typography>
                                <Typography variant="body2" color="textSecondary">
                                    {translate('pages.backup.gdrive.description', { _: 'Configure OAuth2 credentials for Google Drive.' })}
                                </Typography>
                            </Box>
                        </Box>
                    </AccordionSummary>
                    <AccordionDetails>
                        <Box sx={{
                            display: 'grid',
                            gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' },
                            gap: 3
                        }}>
                            <Box>
                                <Alert severity="warning" sx={{ mb: 3 }}>
                                    {translate('pages.backup.gdrive.warning', { _: 'You must configure OAuth2 credentials in Google Cloud Console.' })}
                                </Alert>

                                <TextField
                                    label={translate('pages.backup.gdrive.client_id', { _: 'Client ID' })}
                                    value={config.gdrive_client_id || ''}
                                    onChange={(e) => setConfig({ ...config, gdrive_client_id: e.target.value })}
                                    fullWidth
                                    sx={{ mb: 2 }}
                                />

                                <TextField
                                    label={translate('pages.backup.gdrive.client_secret', { _: 'Client Secret' })}
                                    type="password"
                                    value={config.gdrive_client_secret || ''}
                                    onChange={(e) => setConfig({ ...config, gdrive_client_secret: e.target.value })}
                                    fullWidth
                                    sx={{ mb: 2 }}
                                />

                                <TextField
                                    label={translate('pages.backup.gdrive.folder_id', { _: 'Folder ID (Optional)' })}
                                    value={config.gdrive_folder_id || ''}
                                    onChange={(e) => setConfig({ ...config, gdrive_folder_id: e.target.value })}
                                    fullWidth
                                    helperText={translate('pages.backup.gdrive.folder_id_help', { _: 'ID of the Google Drive folder to upload to.' })}
                                />
                            </Box>
                        </Box>
                    </AccordionDetails>
                </Accordion>
            )}

            <Accordion
                expanded={expanded === 'backups'}
                onChange={handleAccordionChange('backups')}
                sx={{ mb: 2 }}
            >
                <AccordionSummary
                    expandIcon={<ExpandMoreIcon />}
                    sx={{
                        backgroundColor: theme.palette.mode === 'dark' ? 'rgba(46, 125, 50, 0.15)' : '#e8f5e9',
                        '&:hover': {
                            backgroundColor: theme.palette.mode === 'dark' ? 'rgba(46, 125, 50, 0.25)' : '#c8e6c9'
                        }
                    }}
                >
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                        <Box sx={{ color: '#2e7d32' }}>
                            <HistoryIcon />
                        </Box>
                        <Box>
                            <Typography variant="h6" sx={{ color: '#2e7d32', fontSize: isMobile ? '1rem' : '1.25rem' }}>
                                {translate('pages.backup.backups.title', { _: 'Available Backups' })}
                            </Typography>
                            <Typography variant="body2" color="textSecondary">
                                {translate('pages.backup.backups.description', { _: 'Download or restore previous backups.' })}
                                {backups && backups.length > 0 && (
                                    <Chip
                                        label={`${backups.length} ${translate('pages.backup.backups.count', { _: 'backups' })}`}
                                        size="small"
                                        sx={{ ml: 1 }}
                                    />
                                )}
                            </Typography>
                        </Box>
                    </Box>
                </AccordionSummary>
                <AccordionDetails>
                    <Box sx={{
                        display: 'grid',
                        gridTemplateColumns: '1fr',
                        gap: 2
                    }}>
                        {backups && backups.length > 0 ? (
                            backups.map((item) => (
                                <Box
                                    key={item.id}
                                    sx={{
                                        display: 'flex',
                                        flexDirection: isMobile ? 'column' : 'row',
                                        justifyContent: 'space-between',
                                        alignItems: isMobile ? 'flex-start' : 'center',
                                        p: 2,
                                        borderRadius: 1,
                                        backgroundColor: theme.palette.mode === 'dark'
                                            ? 'rgba(255, 255, 255, 0.05)'
                                            : 'rgba(0, 0, 0, 0.02)',
                                        border: '1px solid',
                                        borderColor: theme.palette.mode === 'dark'
                                            ? 'rgba(255, 255, 255, 0.1)'
                                            : 'rgba(0, 0, 0, 0.08)',
                                        '&:hover': {
                                            backgroundColor: theme.palette.mode === 'dark'
                                                ? 'rgba(255, 255, 255, 0.08)'
                                                : 'rgba(0, 0, 0, 0.04)',
                                        }
                                    }}
                                >
                                    <Box sx={{ mb: isMobile ? 1.5 : 0, flex: 1 }}>
                                        <Typography variant="subtitle2" sx={{ fontWeight: 600, wordBreak: 'break-all' }}>
                                            {item.file_name}
                                        </Typography>
                                        <Box sx={{
                                            display: 'flex',
                                            flexDirection: isMobile ? 'column' : 'row',
                                            gap: isMobile ? 0.5 : 2,
                                            mt: 0.5
                                        }}>
                                            <Typography variant="caption" color="textSecondary">
                                                {translate('pages.backup.backups.created', { _: 'Created' })}: {formatDateTime(item.created_at)}
                                            </Typography>
                                            <Typography variant="caption" color="textSecondary">
                                                {translate('pages.backup.backups.size', { _: 'Size' })}: {formatSize(item.size)}
                                            </Typography>
                                        </Box>
                                        <Box sx={{ mt: 0.5 }}>
                                            <Typography variant="caption" color={item.restored_at ? 'success.main' : 'textSecondary'}>
                                                {translate('pages.backup.backups.restored', { _: 'Restored' })}: {item.restored_at ? formatDateTime(item.restored_at) : '-'}
                                            </Typography>
                                        </Box>
                                    </Box>
                                    <Box sx={{
                                        display: 'flex',
                                        flexDirection: isMobile ? 'row' : 'row',
                                        gap: 1,
                                        width: isMobile ? '100%' : 'auto'
                                    }}>
                                        <Button
                                            size="small"
                                            variant="outlined"
                                            startIcon={<DownloadIcon />}
                                            onClick={() => handleDownload(item.id, item.file_name)}
                                            fullWidth={isMobile}
                                        >
                                            {translate('pages.backup.backups.download', { _: 'Download' })}
                                        </Button>
                                        <Button
                                            size="small"
                                            variant="outlined"
                                            color="warning"
                                            startIcon={restoring ? <CircularProgress size={16} color="inherit" /> : <RestoreIcon />}
                                            onClick={() => handleRestore(item.id)}
                                            disabled={restoring}
                                            fullWidth={isMobile}
                                        >
                                            {translate('pages.backup.backups.restore', { _: 'Restore' })}
                                        </Button>
                                    </Box>
                                </Box>
                            ))
                        ) : (
                            <Box sx={{
                                textAlign: 'center',
                                py: 4,
                                color: 'text.secondary'
                            }}>
                                <StorageIcon sx={{ fontSize: 48, opacity: 0.3, mb: 1 }} />
                                <Typography variant="body1">
                                    {translate('pages.backup.backups.empty', { _: 'No backups found. Click "Backup Now" to create your first backup.' })}
                                </Typography>
                            </Box>
                        )}
                    </Box>
                </AccordionDetails>
            </Accordion>
        </Box>
    );
};
