import React, { useState, useEffect } from 'react';
import {
    Card, CardContent, Typography, Button, Grid,
    Box, LinearProgress, Chip
} from '@mui/material';
import {
    Refresh, Backup, Speed, Wifi
} from '@mui/icons-material';
import { useDataProvider, useNotify } from 'react-admin';

export const DeviceManagement = () => {
    const dataProvider = useDataProvider();
    const notify = useNotify();
    const [loading, setLoading] = useState(false);
    const [devices, setDevices] = useState([]);

    useEffect(() => {
        loadDevices();
    }, []);

    const loadDevices = async () => {
        setLoading(true);
        try {
            const { data } = await dataProvider.getList('network/nas', {
                pagination: { page: 1, perPage: 50 },
                sort: { field: 'name', order: 'ASC' },
            });
            setDevices(data);
        } catch (error) {
            notify('Error loading devices', { type: 'error' });
        } finally {
            setLoading(false);
        }
    };

    const handleBackup = async (deviceId: number) => {
        try {
            await dataProvider.create(`network/nas/${deviceId}/backup`, {});
            notify('Backup started', { type: 'success' });
        } catch (error) {
            notify('Backup failed', { type: 'error' });
        }
    };

    const handleSpeedTest = async (deviceId: number) => {
        try {
            await dataProvider.create(`network/nas/${deviceId}/speedtest`, {});
            notify('Speed test started', { type: 'success' });
        } catch (error) {
            notify('Speed test failed', { type: 'error' });
        }
    };

    if (loading) return <LinearProgress />;

    return (
        <Box p={3}>
            <Typography variant="h4" gutterBottom>
                Device Management
            </Typography>

            <Grid container spacing={3}>
                {devices.map((device: any) => (
                    <Grid item xs={12} md={6} lg={4} key={device.id}>
                        <Card>
                            <CardContent>
                                <Typography variant="h6" gutterBottom>
                                    {device.name}
                                </Typography>
                                <Typography variant="body2" color="textSecondary">
                                    IP: {device.ipaddr}
                                </Typography>
                                <Typography variant="body2" color="textSecondary">
                                    Vendor: {device.vendor_code}
                                </Typography>
                                <Box mt={2}>
                                    <Chip
                                        label={device.status}
                                        color={device.status === 'enabled' ? 'success' : 'default'}
                                        size="small"
                                    />
                                </Box>
                                <Box mt={2} display="flex" gap={1}>
                                    <Button
                                        size="small"
                                        startIcon={<Backup />}
                                        onClick={() => handleBackup(device.id)}
                                    >
                                        Backup
                                    </Button>
                                    <Button
                                        size="small"
                                        startIcon={<Speed />}
                                        onClick={() => handleSpeedTest(device.id)}
                                    >
                                        Speed Test
                                    </Button>
                                    <Button
                                        size="small"
                                        startIcon={<Wifi />}
                                        href={`#/network/nas/${device.id}/neighbors`}
                                    >
                                        Neighbors
                                    </Button>
                                </Box>
                            </CardContent>
                        </Card>
                    </Grid>
                ))}
            </Grid>
        </Box>
    );
};
