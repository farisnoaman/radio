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
    CircularProgress,
    FormControl,
    InputLabel,
    Select,
    MenuItem
} from '@mui/material';
import SaveIcon from '@mui/icons-material/Save';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import RestoreIcon from '@mui/icons-material/Restore';
import { useApiQuery } from '../../hooks/useApiQuery';
import { httpClient } from '../../utils/apiClient';

interface BackupConfig {
    provider: string; // "local", "gdrive", "s3"
    retention_days: number;
    schedule: string; // cron expression
    gdrive_client_id?: string;
    gdrive_client_secret?: string;
    gdrive_folder_id?: string;
}

export const BackupSettings = () => {
    const notify = useNotify();
    const [saving, setSaving] = useState(false);
    const [testing, setTesting] = useState(false);
    const [config, setConfig] = useState<BackupConfig>({
        provider: 'local',
        retention_days: 7,
        schedule: '0 0 2 * * *'
    });

    const { data: remoteConfig, isLoading: configLoading } = useApiQuery<BackupConfig>({
        path: '/admin/backup/config',
        queryKey: ['backup', 'config'],
    });

    React.useEffect(() => {
        if (remoteConfig) {
            setConfig(remoteConfig);
        }
    }, [remoteConfig]);

    const handleSave = async () => {
        setSaving(true);
        try {
            await httpClient('/admin/backup/config', {
                method: 'POST',
                body: JSON.stringify(config),
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
            await httpClient('/admin/backup/test', { method: 'POST' });
            notify('Test backup upload initiated successfully', { type: 'success' });
        } catch (error: any) {
            notify(error?.body?.msg || 'Test backup failed', { type: 'error' });
        } finally {
            setTesting(false);
        }
    };

    if (configLoading) return <CircularProgress />;

    return (
        <Box sx={{ p: 3, maxWidth: 800 }}>
            <Box sx={{ mb: 3 }}>
                <Typography variant="h4" gutterBottom display="flex" alignItems="center" gap={1}>
                    <RestoreIcon fontSize="large" color="primary" />
                    Backup & Disaster Recovery
                </Typography>
                <Typography variant="body1" color="textSecondary">
                    Configure automated backups to local storage or cloud providers.
                </Typography>
            </Box>

            <Stack spacing={3}>
                <Card>
                    <CardHeader title="General Settings" />
                    <CardContent>
                        <Stack spacing={2}>
                            <FormControl fullWidth>
                                <InputLabel>Backup Provider</InputLabel>
                                <Select
                                    value={config.provider}
                                    label="Backup Provider"
                                    onChange={(e) => setConfig({ ...config, provider: e.target.value })}
                                >
                                    <MenuItem value="local">Local Storage</MenuItem>
                                    <MenuItem value="gdrive">Google Drive</MenuItem>
                                </Select>
                            </FormControl>

                            <TextField
                                label="Retention Days"
                                type="number"
                                value={config.retention_days}
                                onChange={(e) => setConfig({ ...config, retention_days: Number(e.target.value) })}
                                fullWidth
                                helperText="How many days to keep backup files."
                            />

                            <TextField
                                label="Schedule (Cron)"
                                value={config.schedule}
                                onChange={(e) => setConfig({ ...config, schedule: e.target.value })}
                                fullWidth
                                helperText="Cron expression for automated backups (e.g. 0 0 2 * * *)"
                            />
                        </Stack>
                    </CardContent>
                </Card>

                {config.provider === 'gdrive' && (
                    <Card>
                        <CardHeader title="Google Drive Configuration" />
                        <CardContent>
                            <Alert severity="warning" sx={{ mb: 2 }}>
                                You must configure OAuth2 credentials in Google Cloud Console.
                            </Alert>
                            <Stack spacing={2}>
                                <TextField
                                    label="Client ID"
                                    value={config.gdrive_client_id || ''}
                                    onChange={(e) => setConfig({ ...config, gdrive_client_id: e.target.value })}
                                    fullWidth
                                />
                                <TextField
                                    label="Client Secret"
                                    type="password"
                                    value={config.gdrive_client_secret || ''}
                                    onChange={(e) => setConfig({ ...config, gdrive_client_secret: e.target.value })}
                                    fullWidth
                                />
                                <TextField
                                    label="Folder ID (Optional)"
                                    value={config.gdrive_folder_id || ''}
                                    onChange={(e) => setConfig({ ...config, gdrive_folder_id: e.target.value })}
                                    fullWidth
                                    helperText="ID of the Google Drive folder to upload to."
                                />
                            </Stack>
                        </CardContent>
                    </Card>
                )}

                <Box sx={{ display: 'flex', gap: 2 }}>
                    <Button
                        variant="contained"
                        startIcon={<SaveIcon />}
                        onClick={handleSave}
                        disabled={saving}
                    >
                        {saving ? 'Saving...' : 'Save Configuration'}
                    </Button>

                    <Button
                        variant="outlined"
                        color="secondary"
                        startIcon={<CloudUploadIcon />}
                        onClick={handleTestUpload}
                        disabled={saving || testing}
                    >
                        {testing ? 'Uploading...' : 'Test Cloud Upload'}
                    </Button>
                </Box>
            </Stack>
        </Box>
    );
};
