import { useState } from 'react';
import { useNotify } from 'react-admin';
import {
    Box,
    Card,
    CardContent,
    CardHeader,
    Button,
    Typography,
    Stack,
    Alert,
    Switch,
    FormControlLabel,
    Divider,
    TextField,
} from '@mui/material';
import ConstructionIcon from '@mui/icons-material/Construction';
import DeleteSweepIcon from '@mui/icons-material/DeleteSweep';
import { useApiQuery } from '../../hooks/useApiQuery';
import { httpClient } from '../../utils/apiClient';

export const MaintenanceSettings = () => {
    const notify = useNotify();
    const [archiving, setArchiving] = useState(false);
    const [toggling, setToggling] = useState(false);
    const [drain, setDrain] = useState(true);
    const [days, setDays] = useState(30);

    const { data: status, refetch } = useApiQuery<{ active: boolean }>({
        path: '/system/maintenance',
        queryKey: ['system', 'maintenance'],
    });

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

    return (
        <Box sx={{ p: 3, maxWidth: 800 }}>
            <Box sx={{ mb: 3 }}>
                <Typography variant="h4" gutterBottom display="flex" alignItems="center" gap={1}>
                    <ConstructionIcon fontSize="large" color="primary" />
                    System Maintenance
                </Typography>
                <Typography variant="body1" color="textSecondary">
                    Manage system availability and log data lifecycle.
                </Typography>
            </Box>

            <Stack spacing={3}>
                <Card sx={{ border: status?.active ? '2px solid orange' : 'none' }}>
                    <CardHeader
                        title="Maintenance Mode"
                        subheader="Block non-admin users and drain active sessions."
                    />
                    <Divider />
                    <CardContent>
                        <Stack spacing={2}>
                            {status?.active ? (
                                <Alert severity="warning">
                                    System is currently in MAINTENANCE MODE. Only administrators can access the API.
                                </Alert>
                            ) : (
                                <Alert severity="info">
                                    System is running normally.
                                </Alert>
                            )}

                            <Box>
                                <FormControlLabel
                                    control={
                                        <Switch
                                            checked={status?.active || false}
                                            onChange={handleToggleMaintenance}
                                            disabled={toggling}
                                            color="warning"
                                        />
                                    }
                                    label={status?.active ? "Active" : "Inactive"}
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
                                    label="Drain sessions (Disconnect online users upon enabling)"
                                />
                            )}

                            <Typography variant="caption" color="textSecondary">
                                Enabling maintenance mode will prevent new user logins and optionally disconnect all current users.
                            </Typography>
                        </Stack>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader
                        title="System Log Archival"
                        subheader="Archive and compress old logs to save disk space."
                    />
                    <Divider />
                    <CardContent>
                        <Stack spacing={3} direction="row" alignItems="center">
                            <TextField
                                label="Retention Days"
                                type="number"
                                value={days}
                                onChange={(e) => setDays(Number(e.target.value))}
                                sx={{ width: 150 }}
                                helperText="Logs older than this will be compressed."
                            />
                            <Button
                                variant="contained"
                                color="secondary"
                                startIcon={<DeleteSweepIcon />}
                                onClick={handleArchiveLogs}
                                disabled={archiving}
                                sx={{ height: 56 }}
                            >
                                {archiving ? 'Archiving...' : 'Archive Now'}
                            </Button>
                        </Stack>
                        <Box sx={{ mt: 2 }}>
                            <Typography variant="caption" color="textSecondary">
                                Compressed logs are stored in the system log directory as .csv.gz files.
                                This process is also triggered automatically on a daily schedule.
                            </Typography>
                        </Box>
                    </CardContent>
                </Card>
            </Stack>
        </Box>
    );
};
