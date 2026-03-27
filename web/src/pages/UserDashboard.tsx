import { useEffect, useState } from 'react';
import { Box, Card, CardContent, Typography, Stack, LinearProgress, Divider, Paper, alpha, useTheme } from '@mui/material';
import { Grid } from '@mui/material';
import { useTranslate, useGetIdentity, useLocale } from 'react-admin';
import ReceiptLongOutlinedIcon from '@mui/icons-material/ReceiptLongOutlined';
import AccountCircleOutlinedIcon from '@mui/icons-material/AccountCircleOutlined';
import SpeedIcon from '@mui/icons-material/Speed';
import TimerIcon from '@mui/icons-material/Timer';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import DevicesIcon from '@mui/icons-material/Devices';
import DateRangeIcon from '@mui/icons-material/DateRange';

const UserDashboard = () => {
    const translate = useTranslate();
    const { data: identity } = useGetIdentity();
    const locale = useLocale();
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';
    const isRTL = locale === 'ar';
    const [usage, setUsage] = useState<any>(null);

    useEffect(() => {
        const fetchUsage = async () => {
            try {
                const token = localStorage.getItem('token');
                const response = await fetch('/api/v1/portal/usage', {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });
                const result = await response.json();
                if (result.data) {
                    setUsage(result.data);
                }
            } catch (error) {
                console.error('Failed to fetch usage:', error);
            }
        };

        fetchUsage();
        const interval = setInterval(fetchUsage, 30000);

        return () => clearInterval(interval);
    }, []);

    const formatData = (bytes: number) => {
        if (!bytes) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const formatTime = (seconds: number) => {
        if (!seconds) return '0s';
        const h = Math.floor(seconds / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        return `${h}h ${m}m`;
    };

    const dataPercent = usage?.data_quota > 0 ? Math.min(100, (usage.data_used / (usage.data_quota * 1024 * 1024)) * 100) : 0;

    return (
        <Box sx={{ mt: 2, direction: isRTL ? 'rtl' : 'ltr' }}>
            {/* Welcome Hero Section with Gradient */}
            <Paper 
                sx={{ 
                    p: { xs: 3, md: 4 }, 
                    mb: 4, 
                    borderRadius: 6, 
                    background: isDark 
                        ? 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)' 
                        : 'linear-gradient(135deg, #1e3a8a 0%, #3b82f6 100%)',
                    color: 'white',
                    boxShadow: isDark 
                        ? '0 10px 25px -5px rgba(0, 0, 0, 0.5)' 
                        : '0 10px 25px -5px rgba(59, 130, 246, 0.4)',
                    position: 'relative',
                    overflow: 'hidden',
                    border: isDark ? '1px solid rgba(255,255,255,0.05)' : 'none'
                }}
            >
                {/* Decorative background circle */}
                <Box sx={{ 
                    position: 'absolute', 
                    top: -50, 
                    right: -50, 
                    width: 200, 
                    height: 200, 
                    borderRadius: '50%', 
                    background: 'rgba(255,255,255,0.1)',
                    display: { xs: 'none', md: 'block' }
                }} />

                <Typography variant="h4" sx={{ fontWeight: 800, mb: 1 }}>
                    {translate('auth.welcome')}, {identity?.fullName || identity?.username}
                </Typography>
                <Typography variant="h6" sx={{ opacity: 0.9, fontWeight: 400 }}>
                    {translate('portal.usage_stats')}
                </Typography>
            </Paper>
            
            <Grid container spacing={3}>
                <Grid size={{ xs: 12, md: 8 }}>
                    <Paper 
                        sx={{ 
                            p: 3, 
                            borderRadius: 6, 
                            mb: 3, 
                            backdropFilter: 'blur(10px)',
                            border: '1px solid',
                            borderColor: 'divider',
                            background: (theme) => alpha(theme.palette.background.paper, isDark ? 0.6 : 0.8),
                            boxShadow: isDark ? 'none' : '0 4px 6px -1px rgba(0,0,0,0.1)'
                        }}
                    >
                        <Stack direction="row" spacing={1.5} alignItems="center" sx={{ mb: 4 }}>
                            <Box sx={{ p: 1, borderRadius: 2, bgcolor: alpha(theme.palette.primary.main, 0.1) }}>
                                <SpeedIcon color="primary" />
                            </Box>
                            <Typography variant="h6" fontWeight={700}>
                                {translate('portal.usage_stats')}
                            </Typography>
                        </Stack>
                        
                        <Grid container spacing={4}>
                            <Grid size={{ xs: 12, sm: 6 }}>
                                <Typography variant="subtitle2" color="text.secondary" gutterBottom sx={{ fontWeight: 600 }}>
                                    {translate('portal.data_used')}
                                </Typography>
                                <Typography variant="h3" fontWeight={800} color="primary.main" sx={{ mb: 2 }}>
                                    {formatData(usage?.data_used)}
                                </Typography>
                                <Box sx={{ mt: 2, mb: 1 }}>
                                    <LinearProgress 
                                        variant="determinate" 
                                        value={dataPercent} 
                                        sx={{ 
                                            height: 10, 
                                            borderRadius: 5, 
                                            bgcolor: isDark ? alpha(theme.palette.divider, 0.1) : 'action.hover',
                                            '& .MuiLinearProgress-bar': {
                                                borderRadius: 5,
                                                background: isDark 
                                                    ? `linear-gradient(90deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.light} 100%)`
                                                    : 'linear-gradient(90deg, #3b82f6 0%, #60a5fa 100%)'
                                            }
                                        }}
                                    />
                                </Box>
                                <Typography variant="body2" color="text.secondary" fontWeight={500}>
                                    {translate('portal.remaining')}: <Box component="span" sx={{ color: 'text.primary', fontWeight: 700 }}>{usage?.data_quota > 0 ? formatData((usage.data_quota * 1024 * 1024) - usage.data_used) : translate('resources.products.units.unlimited')}</Box>
                                </Typography>
                            </Grid>
                            
                            <Grid size={{ xs: 12, sm: 6 }}>
                                <Typography variant="subtitle2" color="text.secondary" gutterBottom sx={{ fontWeight: 600 }}>
                                    {translate('portal.time_quota')}
                                </Typography>
                                <Typography variant="h3" fontWeight={800} color="secondary.main" sx={{ mb: 2 }}>
                                    {usage?.time_quota > 0 ? formatTime(usage?.time_quota - usage?.time_used) : translate('resources.products.units.unlimited')}
                                </Typography>
                                <Box sx={{ mt: 2, mb: 1 }}>
                                    <LinearProgress
                                        variant="determinate"
                                        value={usage?.time_quota > 0 ? Math.min(100, ((usage?.time_used || 0) / usage?.time_quota) * 100) : 0}
                                        sx={{
                                            height: 10,
                                            borderRadius: 5,
                                            bgcolor: isDark ? alpha(theme.palette.divider, 0.1) : 'action.hover',
                                            '& .MuiLinearProgress-bar': {
                                                borderRadius: 5,
                                                background: isDark
                                                    ? `linear-gradient(90deg, ${theme.palette.secondary.main} 0%, ${theme.palette.secondary.light} 100%)`
                                                    : 'linear-gradient(90deg, #6366f1 0%, #9c27b0 100%)'
                                            }
                                        }}
                                    />
                                </Box>
                                <Typography variant="body2" color="text.secondary" fontWeight={500}>
                                    {translate('portal.remaining')}: <Box component="span" sx={{ color: 'text.primary', fontWeight: 700 }}>
                                        {usage?.time_quota > 0
                                            ? `${formatTime(usage?.time_quota - usage?.time_used)} / ${formatTime(usage?.time_quota)}`
                                            : translate('resources.products.units.unlimited')
                                        }
                                    </Box>
                                </Typography>
                            </Grid>
                        </Grid>
                    </Paper>

                    <Grid container spacing={3}>
                        <Grid size={{ xs: 12, sm: 6 }}>
                            <Card sx={{ 
                                borderRadius: 6, 
                                bgcolor: isDark ? alpha(theme.palette.primary.main, 0.1) : theme.palette.primary.main, 
                                color: isDark ? theme.palette.primary.main : theme.palette.primary.contrastText, 
                                boxShadow: isDark ? 'none' : '0 8px 16px rgba(37, 99, 235, 0.2)',
                                border: isDark ? `1px solid ${alpha(theme.palette.primary.main, 0.2)}` : 'none'
                            }}>
                                <CardContent sx={{ p: 3 }}>
                                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                                        <Box>
                                            <Typography variant="subtitle2" sx={{ opacity: 0.8, fontWeight: 600 }}>
                                                {translate('portal.monthly_fee')}
                                            </Typography>
                                            <Typography variant="h4" fontWeight={800}>
                                                ${usage?.monthly_fee || 0}
                                            </Typography>
                                        </Box>
                                        <Paper sx={{ p: 1.5, borderRadius: 3, bgcolor: isDark ? alpha(theme.palette.primary.main, 0.1) : 'rgba(255,255,255,0.15)', backdropFilter: 'blur(5px)', border: 'none' }}>
                                            <ReceiptLongOutlinedIcon sx={{ fontSize: 32, color: isDark ? theme.palette.primary.main : 'white' }} />
                                        </Paper>
                                    </Stack>
                                </CardContent>
                            </Card>
                        </Grid>
                        <Grid size={{ xs: 12, sm: 6 }}>
                            <Card sx={{ 
                                borderRadius: 6, 
                                bgcolor: isDark ? alpha(theme.palette.secondary.main, 0.1) : theme.palette.secondary.main, 
                                color: isDark ? theme.palette.secondary.main : theme.palette.secondary.contrastText, 
                                boxShadow: isDark ? 'none' : '0 8px 16px rgba(139, 92, 246, 0.2)',
                                border: isDark ? `1px solid ${alpha(theme.palette.secondary.main, 0.2)}` : 'none'
                            }}>
                                <CardContent sx={{ p: 3 }}>
                                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                                        <Box>
                                            <Typography variant="subtitle2" sx={{ opacity: 0.8, fontWeight: 600 }}>
                                                {translate('portal.next_bill')}
                                            </Typography>
                                            <Typography variant="h5" fontWeight={800}>
                                                {usage?.next_bill_date ? new Date(usage.next_bill_date).toLocaleDateString(locale, { year: 'numeric', month: 'long', day: 'numeric' }) : '-'}
                                            </Typography>
                                        </Box>
                                        <Paper sx={{ p: 1.5, borderRadius: 3, bgcolor: isDark ? alpha(theme.palette.secondary.main, 0.1) : 'rgba(255,255,255,0.15)', backdropFilter: 'blur(5px)', border: 'none' }}>
                                            <DateRangeIcon sx={{ fontSize: 32, color: isDark ? theme.palette.secondary.main : 'white' }} />
                                        </Paper>
                                    </Stack>
                                </CardContent>
                            </Card>
                        </Grid>
                    </Grid>
                </Grid>

                <Grid size={{ xs: 12, md: 4 }}>
                    <Card sx={{ 
                        borderRadius: 6, 
                        mb: 3, 
                        border: '1px solid', 
                        borderColor: 'divider', 
                        bgcolor: 'background.paper',
                        boxShadow: isDark ? 'none' : '0 4px 6px -1px rgba(0,0,0,0.05)' 
                    }}>
                        <CardContent sx={{ p: 3 }}>
                            <Stack direction="row" spacing={2} alignItems="center" sx={{ mb: 3 }}>
                                <Box sx={{ p: 1, borderRadius: 2, bgcolor: alpha(theme.palette.primary.main, 0.1) }}>
                                    <AccountCircleOutlinedIcon color="primary" />
                                </Box>
                                <Typography variant="h6" fontWeight={700}>
                                    {translate('appbar.account_settings')}
                                </Typography>
                            </Stack>
                            <Divider sx={{ mb: 3 }} />
                            <Stack spacing={2.5}>
                                <Box>
                                    <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 600, display: 'block', mb: 0.5 }}>
                                        {translate('auth.username')}
                                    </Typography>
                                    <Typography variant="body1" fontWeight={700}>{usage?.username}</Typography>
                                </Box>
                                <Box>
                                    <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 600, display: 'block', mb: 0.5 }}>
                                        {translate('portal.current_mac')}
                                    </Typography>
                                    <Typography variant="body1" fontWeight={700} sx={{ fontFamily: 'monospace', bgcolor: alpha(theme.palette.action.hover, 0.5), px: 1, py: 0.5, borderRadius: 1.5, display: 'inline-block' }}>
                                        {usage?.mac_addr || '-'}
                                    </Typography>
                                </Box>
                                <Box>
                                    <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 600, display: 'block', mb: 0.5 }}>
                                        {translate('portal.active_num')}
                                    </Typography>
                                    <Typography variant="body1" fontWeight={700}>{usage?.bind_mac || translate('resources.products.units.unlimited')}</Typography>
                                </Box>
                            </Stack>
                        </CardContent>
                    </Card>

                    <Paper 
                        sx={{ 
                            p: 3, 
                            borderRadius: 6, 
                            bgcolor: alpha(theme.palette.primary.main, 0.05), 
                            border: '2px dashed', 
                            borderColor: alpha(theme.palette.primary.main, 0.2),
                            textAlign: 'center'
                        }}
                    >
                         <Stack direction="column" spacing={2} alignItems="center">
                            <Box sx={{ p: 1.5, borderRadius: '50%', bgcolor: 'background.paper', display: 'flex', boxShadow: '0 4px 10px rgba(0,0,0,0.05)', border: '1px solid', borderColor: 'divider' }}>
                                <DevicesIcon color="primary" fontSize="large" />
                            </Box>
                            <Typography variant="subtitle1" fontWeight={700}>
                                {translate('portal.my_devices')}
                            </Typography>
                            <Typography variant="body2" color="text.secondary">
                                {translate('portal.usage_description')}
                            </Typography>
                         </Stack>
                    </Paper>
                </Grid>
            </Grid>
        </Box>
    );
};

export default UserDashboard;
