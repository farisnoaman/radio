import { useState } from 'react';
import { 
    Box, Card, CardContent, Typography, Stack, TextField, Button, 
    alpha, CircularProgress, useTheme, useMediaQuery
} from '@mui/material';
import Grid from '@mui/material/GridLegacy';
import { useTranslate, useLocale, useNotify } from 'react-admin';
import RedeemIcon from '@mui/icons-material/Redeem';

const VoucherRedeem = () => {
    const translate = useTranslate();
    const notify = useNotify();
    const locale = useLocale();
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const isRTL = locale === 'ar';
    const [code, setCode] = useState('');
    const [loading, setLoading] = useState(false);

    const handleRedeem = async () => {
        if (!code) {
            notify('portal.voucher_code', { type: 'warning' });
            return;
        }

        setLoading(true);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('/api/v1/portal/vouchers/redeem', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ code })
            });
            const result = await response.json();
            if (result.code === 0) {
                notify('portal.redeem_success', { type: 'success' });
                setCode('');
            } else {
                notify(result.msg || 'portal.redeem_error', { type: 'error' });
            }
        } catch (error) {
            notify('common.network_error', { type: 'error' });
        } finally {
            setLoading(false);
        }
    };

    return (
        <Box sx={{ 
            mt: { xs: 2, md: 4 }, 
            direction: isRTL ? 'rtl' : 'ltr', 
            display: 'flex', 
            flexDirection: 'column', 
            alignItems: 'center',
            maxWidth: 1200,
            mx: 'auto',
            p: 2
        }}>
            {/* Hero Section */}
            <Typography variant="h3" sx={{ 
                fontWeight: 900, 
                mb: 1, 
                textAlign: 'center', 
                background: isDark 
                    ? `linear-gradient(90deg, ${theme.palette.primary.light} 0%, ${theme.palette.secondary.light} 100%)`
                    : 'linear-gradient(90deg, #3b82f6 0%, #8b5cf6 100%)', 
                WebkitBackgroundClip: 'text', 
                WebkitTextFillColor: 'transparent',
                textShadow: isDark ? '0 0 30px rgba(59, 130, 246, 0.3)' : 'none'
            }}>
                {translate('portal.redeem_voucher')}
            </Typography>
            <Typography variant="h6" color="text.secondary" sx={{ mb: { xs: 4, md: 6 }, fontWeight: 400, textAlign: 'center', maxWidth: 600 }}>
                {translate('portal.redeem_instruction')}
            </Typography>

            <Grid container spacing={isMobile ? 3 : 6} justifyContent="center" alignItems="stretch">
                <Grid item xs={12} md={6}>
                    <Card sx={{ 
                        borderRadius: 8, 
                        p: { xs: 1, md: 3 }, 
                        height: '100%',
                        boxShadow: isDark ? 'none' : '0 20px 50px rgba(0,0,0,0.1)',
                        border: '1px solid',
                        borderColor: 'divider',
                        backdropFilter: 'blur(20px)',
                        background: (theme) => alpha(theme.palette.background.paper, isDark ? 0.6 : 0.8)
                    }}>
                        <CardContent sx={{ p: { xs: 2, md: 4 } }}>
                            <Box sx={{ mb: 4, textAlign: 'center' }}>
                                <Box sx={{ 
                                    display: 'inline-flex', 
                                    p: 2, 
                                    borderRadius: 4, 
                                    bgcolor: alpha(theme.palette.primary.main, 0.1),
                                    mb: 2
                                }}>
                                    <RedeemIcon color="primary" sx={{ fontSize: 40 }} />
                                </Box>
                                <Typography variant="h5" fontWeight={800} gutterBottom>
                                    {translate('portal.voucher_code')}
                                </Typography>
                            </Box>

                            <Stack spacing={3}>
                                <TextField
                                    fullWidth
                                    variant="outlined"
                                    placeholder={translate('portal.enter_code')}
                                    value={code}
                                    onChange={(e) => setCode(e.target.value)}
                                    disabled={loading}
                                    sx={{
                                        '& .MuiOutlinedInput-root': {
                                            borderRadius: 4,
                                            height: 64,
                                            fontSize: '1.2rem',
                                            fontWeight: 700,
                                            fontFamily: 'monospace',
                                            textAlign: 'center',
                                            bgcolor: alpha(theme.palette.action.hover, 0.5),
                                            '& fieldset': { borderColor: 'divider' },
                                            '&:hover fieldset': { borderColor: 'primary.main' }
                                        }
                                    }}
                                />
                                <Button
                                    fullWidth
                                    variant="contained"
                                    size="large"
                                    onClick={handleRedeem}
                                    disabled={loading || !code}
                                    sx={{
                                        borderRadius: 4,
                                        height: 64,
                                        fontSize: '1.1rem',
                                        fontWeight: 800,
                                        background: isDark
                                            ? `linear-gradient(90deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.dark} 100%)`
                                            : 'linear-gradient(90deg, #3b82f6 0%, #2563eb 100%)',
                                        boxShadow: isDark ? 'none' : '0 8px 20px rgba(37, 99, 235, 0.3)',
                                        '&:hover': {
                                            background: isDark
                                                ? `linear-gradient(90deg, ${theme.palette.primary.dark} 0%, ${theme.palette.primary.main} 100%)`
                                                : 'linear-gradient(90deg, #2563eb 0%, #1d4ed8 100%)'
                                        }
                                    }}
                                >
                                    {loading ? <CircularProgress size={24} color="inherit" /> : translate('portal.redeem_voucher')}
                                </Button>
                            </Stack>
                        </CardContent>
                    </Card>
                </Grid>

                <Grid item xs={12} md={6}>
                    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center', p: { xs: 1, md: 4 } }}>
                         <Typography variant="h5" fontWeight={800} sx={{ mb: 4 }}>
                            {translate('portal.how_it_works')}
                        </Typography>
                        
                        <Stack spacing={4}>
                            <Box sx={{ display: 'flex', gap: 2 }}>
                                <Box sx={{ minWidth: 48, height: 48, borderRadius: 3, bgcolor: alpha(theme.palette.primary.main, 0.1), display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 800, color: 'primary.main' }}>
                                    1
                                </Box>
                                <Box>
                                    <Typography variant="subtitle1" fontWeight={700} gutterBottom>
                                        {translate('common.create')}
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary">
                                        {translate('portal.how_it_works_step1')}
                                    </Typography>
                                </Box>
                            </Box>

                            <Box sx={{ display: 'flex', gap: 2 }}>
                                <Box sx={{ minWidth: 48, height: 48, borderRadius: 3, bgcolor: alpha(theme.palette.secondary.main, 0.1), display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 800, color: 'secondary.main' }}>
                                    2
                                </Box>
                                <Box>
                                    <Typography variant="subtitle1" fontWeight={700} gutterBottom>
                                        {translate('portal.voucher_code')}
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary">
                                        {translate('portal.how_it_works_step2')}
                                    </Typography>
                                </Box>
                            </Box>

                            <Box sx={{ display: 'flex', gap: 2 }}>
                                <Box sx={{ minWidth: 48, height: 48, borderRadius: 3, bgcolor: alpha(theme.palette.success.main, 0.1), display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 800, color: 'success.main' }}>
                                    3
                                </Box>
                                <Box>
                                    <Typography variant="subtitle1" fontWeight={700} gutterBottom>
                                        {translate('portal.redeem_success')}
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary">
                                        {translate('portal.how_it_works_step3')}
                                    </Typography>
                                </Box>
                            </Box>
                        </Stack>
                    </Box>
                </Grid>
            </Grid>
        </Box>
    );
};

export default VoucherRedeem;
