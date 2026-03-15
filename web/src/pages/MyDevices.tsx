import { useEffect, useState } from 'react';
import { 
    Box, Card, CardContent, Typography, Stack, Button, 
    Table, TableBody, TableCell, TableContainer, TableHead, TableRow, 
    Paper, Divider, useMediaQuery, useTheme, alpha 
} from '@mui/material';
import Grid from '@mui/material/GridLegacy';
import { useTranslate, useLocale, useNotify } from 'react-admin';
import DevicesIcon from '@mui/icons-material/Devices';
import PowerSettingsNewIcon from '@mui/icons-material/PowerSettingsNew';
import VpnKeyIcon from '@mui/icons-material/VpnKey';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import LanguageIcon from '@mui/icons-material/Language';
import RouterIcon from '@mui/icons-material/Router';
import AccessTimeIcon from '@mui/icons-material/AccessTime';

const MyDevices = () => {
    const translate = useTranslate();
    const notify = useNotify();
    const locale = useLocale();
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const isDark = theme.palette.mode === 'dark';
    const isRTL = locale === 'ar';
    const [sessions, setSessions] = useState<any[]>([]);
    const [usage, setUsage] = useState<any>(null);
    const [loading, setLoading] = useState(false);

    const fetchData = async () => {
        try {
            const token = localStorage.getItem('token');
            const [sessionsRes, usageRes] = await Promise.all([
                fetch('/api/v1/portal/sessions', { headers: { 'Authorization': `Bearer ${token}` } }),
                fetch('/api/v1/portal/usage', { headers: { 'Authorization': `Bearer ${token}` } })
            ]);
            
            const sessionsData = await sessionsRes.json();
            const usageData = await usageRes.json();

            if (sessionsData.code === 0) setSessions(sessionsData.data || []);
            if (usageData.code === 0) setUsage(usageData.data);
        } catch (error) {
            console.error('Failed to fetch devices data:', error);
            notify('common.network_error', { type: 'error' });
        }
    };

    useEffect(() => {
        fetchData();
    }, []);

    const handleDisconnect = async (sessionId: string) => {
        setLoading(true);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`/api/v1/portal/sessions/${sessionId}`, {
                method: 'DELETE',
                headers: { 'Authorization': `Bearer ${token}` }
            });
            const result = await response.json();
            if (result.code === 0) {
                notify('portal.disconnect_success', { type: 'success' });
                fetchData();
            } else {
                notify(result.msg, { type: 'error' });
            }
        } catch (error) {
            notify('common.network_error', { type: 'error' });
        } finally {
            setLoading(false);
        }
    };

    const handleUnbind = async () => {
        setLoading(true);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('/api/v1/portal/unbind-mac', {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${token}` }
            });
            const result = await response.json();
            if (result.code === 0) {
                notify('portal.unbind_success', { type: 'success' });
                fetchData();
            } else {
                notify(result.msg, { type: 'error' });
            }
        } catch (error) {
            notify('common.network_error', { type: 'error' });
        } finally {
            setLoading(false);
        }
    };

    const SessionCard = ({ session }: { session: any }) => (
        <Card sx={{ 
            borderRadius: 4, 
            mb: 2, 
            border: '1px solid', 
            borderColor: 'divider',
            bgcolor: 'background.paper',
            boxShadow: isDark ? 'none' : '0 2px 4px rgba(0,0,0,0.05)',
            transition: 'transform 0.2s',
            '&:active': { transform: 'scale(0.98)' }
        }}>
            <CardContent>
                <Stack spacing={2}>
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                        <Stack direction="row" spacing={1} alignItems="center">
                            <Box sx={{ p: 1, borderRadius: 2, bgcolor: alpha(theme.palette.primary.main, 0.1) }}>
                                <LanguageIcon color="primary" fontSize="small" />
                            </Box>
                            <Typography variant="subtitle1" fontWeight={700}>
                                {session.framed_ip}
                            </Typography>
                        </Stack>
                        <Button 
                            variant="outlined" 
                            color="error" 
                            size="small"
                            onClick={() => handleDisconnect(session.acct_session_id)}
                            disabled={loading}
                            sx={{ borderRadius: 2, fontWeight: 700, minWidth: 'auto', px: 2 }}
                        >
                            {translate('portal.disconnect')}
                        </Button>
                    </Stack>
                    
                    <Box sx={{ p: 2, bgcolor: alpha(theme.palette.action.hover, 0.4), borderRadius: 2 }}>
                        <Stack direction="row" spacing={2}>
                            <Box sx={{ flex: 1 }}>
                                <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 600, display: 'block', mb: 0.5 }}>
                                    {translate('common.mac_address')}
                                </Typography>
                                <Stack direction="row" spacing={1} alignItems="center">
                                    <RouterIcon sx={{ fontSize: 16, color: 'text.secondary' }} />
                                    <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 600 }}>
                                        {session.calling_station_id}
                                    </Typography>
                                </Stack>
                            </Box>
                            <Box sx={{ minWidth: 100 }}>
                                <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 600, display: 'block', mb: 0.5 }}>
                                    {translate('common.duration')}
                                </Typography>
                                <Stack direction="row" spacing={1} alignItems="center">
                                    <AccessTimeIcon sx={{ fontSize: 16, color: 'text.secondary' }} />
                                    <Typography variant="body2" fontWeight={600}>
                                        {Math.floor(session.acct_session_time / 60)}m
                                    </Typography>
                                </Stack>
                            </Box>
                        </Stack>
                    </Box>
                </Stack>
            </CardContent>
        </Card>
    );

    return (
        <Box sx={{ 
            mt: { xs: 1, md: 2 }, 
            direction: isRTL ? 'rtl' : 'ltr',
            p: { xs: 1, md: 0 } 
        }}>
            <Paper 
                sx={{ 
                    p: { xs: 2.5, md: 4 }, 
                    mb: { xs: 2, md: 4 }, 
                    borderRadius: { xs: 4, md: 6 }, 
                    background: isDark 
                        ? 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)' 
                        : 'linear-gradient(135deg, #1e40af 0%, #3b82f6 100%)',
                    color: 'white',
                    boxShadow: isDark 
                        ? '0 10px 25px -5px rgba(0, 0, 0, 0.5)' 
                        : '0 10px 25px -5px rgba(30, 64, 175, 0.4)',
                    border: isDark ? '1px solid rgba(255,255,255,0.05)' : 'none'
                }}
            >
                <Stack direction="row" spacing={2} alignItems="center">
                    <Box sx={{ p: 1.5, borderRadius: 3, bgcolor: 'rgba(255,255,255,0.1)', backdropFilter: 'blur(5px)' }}>
                        <DevicesIcon sx={{ fontSize: { xs: 24, md: 32 } }} />
                    </Box>
                    <Box>
                        <Typography variant={isMobile ? "h5" : "h4"} sx={{ fontWeight: 800 }}>
                            {translate('portal.my_devices')}
                        </Typography>
                        <Typography variant="body2" sx={{ opacity: 0.8 }}>
                            {translate('portal.usage_description')}
                        </Typography>
                    </Box>
                </Stack>
            </Paper>

            <Grid container spacing={isMobile ? 2 : 3}>
                <Grid item xs={12} md={8}>
                    {isMobile ? (
                        <Box>
                            <Typography variant="h6" fontWeight={800} sx={{ mb: 2, px: 1 }}>
                                {translate('portal.active_sessions')} ({sessions.length})
                            </Typography>
                            {sessions.length === 0 ? (
                                <Paper sx={{ p: 4, textAlign: 'center', borderRadius: 4, border: '1px dashed', borderColor: 'divider' }}>
                                    <Typography color="text.secondary">{translate('portal.no_sessions')}</Typography>
                                </Paper>
                            ) : (
                                sessions.map(session => <SessionCard key={session.id} session={session} />)
                            )}
                        </Box>
                    ) : (
                        <Paper 
                            sx={{ 
                                borderRadius: 6, 
                                overflow: 'hidden',
                                border: '1px solid',
                                borderColor: 'divider',
                                bgcolor: 'background.paper',
                                boxShadow: isDark ? 'none' : '0 4px 6px -1px rgba(0,0,0,0.05)'
                            }}
                        >
                            <Box sx={{ p: 3, borderBottom: '1px solid', borderColor: 'divider', bgcolor: alpha(theme.palette.action.hover, 0.4) }}>
                                <Typography variant="h6" fontWeight={700}>
                                    {translate('portal.active_sessions')}
                                </Typography>
                            </Box>
                            
                            <TableContainer>
                                <Table>
                                    <TableHead>
                                        <TableRow sx={{ bgcolor: alpha(theme.palette.background.default, 0.5) }}>
                                            <TableCell sx={{ fontWeight: 700 }}>{translate('common.ip_address')}</TableCell>
                                            <TableCell sx={{ fontWeight: 700 }}>{translate('common.mac_address')}</TableCell>
                                            <TableCell sx={{ fontWeight: 700 }}>{translate('common.duration')}</TableCell>
                                            <TableCell align="right" sx={{ fontWeight: 700 }}>{translate('resources.vouchers.fields.action')}</TableCell>
                                        </TableRow>
                                    </TableHead>
                                    <TableBody>
                                        {sessions.length === 0 ? (
                                            <TableRow>
                                                <TableCell colSpan={4} align="center" sx={{ py: 8 }}>
                                                    <Typography color="text.secondary">
                                                        {translate('portal.no_sessions')}
                                                    </Typography>
                                                </TableCell>
                                            </TableRow>
                                        ) : (
                                            sessions.map((session) => (
                                                <TableRow key={session.id} hover sx={{ '&:last-child td, &:last-child th': { border: 0 } }}>
                                                    <TableCell sx={{ fontWeight: 500 }}>{session.framed_ip}</TableCell>
                                                    <TableCell sx={{ fontFamily: 'monospace', color: 'text.secondary' }}>{session.calling_station_id}</TableCell>
                                                    <TableCell>{Math.floor(session.acct_session_time / 60)}m</TableCell>
                                                    <TableCell align="right">
                                                        <Button 
                                                            variant="text" 
                                                            color="error" 
                                                            size="small"
                                                            startIcon={<PowerSettingsNewIcon />}
                                                            onClick={() => handleDisconnect(session.acct_session_id)}
                                                            disabled={loading}
                                                            sx={{ borderRadius: 2, fontWeight: 700 }}
                                                        >
                                                            {translate('portal.disconnect')}
                                                        </Button>
                                                    </TableCell>
                                                </TableRow>
                                            ))
                                        )}
                                    </TableBody>
                                </Table>
                            </TableContainer>
                        </Paper>
                    )}
                </Grid>

                <Grid item xs={12} md={4}>
                    <Card sx={{ 
                        borderRadius: { xs: 4, md: 6 }, 
                        mb: 3, 
                        border: '1px solid', 
                        borderColor: 'divider', 
                        bgcolor: 'background.paper',
                        boxShadow: isDark ? 'none' : '0 4px 6px -1px rgba(0,0,0,0.05)' 
                    }}>
                        <CardContent sx={{ p: { xs: 2, md: 3 } }}>
                            <Stack direction="row" spacing={2} alignItems="center" sx={{ mb: 2 }}>
                                <Box sx={{ p: 1, borderRadius: 2, bgcolor: alpha(theme.palette.error.main, 0.1) }}>
                                    <VpnKeyIcon color="error" />
                                </Box>
                                <Typography variant="h6" fontWeight={700}>
                                    {translate('portal.unbind_mac')}
                                </Typography>
                            </Stack>
                            <Typography variant="body2" color="text.secondary" sx={{ mb: 3, lineHeight: 1.6 }}>
                                {translate('portal.unbind_mac_description')}
                            </Typography>
                            <Divider sx={{ mb: 3 }} />
                            <Stack spacing={2}>
                                <Box sx={{ p: 2, bgcolor: alpha(theme.palette.action.hover, 0.4), borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
                                    <Typography variant="caption" color="text.secondary" sx={{ fontWeight: 600, display: 'block', mb: 0.5 }}>
                                        {translate('portal.current_mac')}
                                    </Typography>
                                    <Typography variant="body1" fontWeight={700} sx={{ fontFamily: 'monospace' }}>
                                        {usage?.mac_addr || '-'}
                                    </Typography>
                                </Box>
                                <Button 
                                    variant="contained" 
                                    color="error" 
                                    fullWidth
                                    onClick={handleUnbind}
                                    disabled={loading || !usage?.mac_addr}
                                    sx={{ 
                                        borderRadius: 3, 
                                        py: 1.5, 
                                        fontWeight: 700,
                                        boxShadow: isDark ? 'none' : '0 4px 12px rgba(211, 47, 47, 0.2)'
                                    }}
                                >
                                    {translate('portal.unbind_mac')}
                                </Button>
                            </Stack>
                        </CardContent>
                    </Card>

                    <Paper 
                        sx={{ 
                            p: 3, 
                            borderRadius: { xs: 4, md: 6 }, 
                            bgcolor: alpha(theme.palette.primary.main, 0.05), 
                            border: '1px solid', 
                            borderColor: alpha(theme.palette.primary.main, 0.1),
                            display: 'flex',
                            alignItems: 'center',
                            gap: 2
                        }}
                    >
                        <Box sx={{ p: 1, borderRadius: 2, bgcolor: 'background.paper', display: 'flex', boxShadow: '0 2px 4px rgba(0,0,0,0.05)', border: '1px solid', borderColor: 'divider' }}>
                            <InfoOutlinedIcon color="primary" />
                        </Box>
                        <Typography variant="body2" color="text.primary" fontWeight={500}>
                            {translate('portal.active_num')}: <strong>{usage?.bind_mac || translate('resources.products.units.unlimited')}</strong>
                        </Typography>
                    </Paper>
                </Grid>
            </Grid>
        </Box>
    );
};

export default MyDevices;
