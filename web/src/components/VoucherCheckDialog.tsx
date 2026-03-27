import React, { useState } from 'react';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button,
    TextField,
    Box,
    Typography,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Paper,
    Divider,
    IconButton,
    InputAdornment,
    CircularProgress,
    Alert,
    Chip,
    Tooltip
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import CloseIcon from '@mui/icons-material/Close';
import HistoryIcon from '@mui/icons-material/History';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import { useTranslate, useNotify } from 'react-admin';
import { apiRequest } from '../utils/apiClient';

interface Session {
    id: string;
    acct_session_id: string;
    acct_session_time: number;
    acct_input_total: number;
    acct_output_total: number;
    acct_start_time: string;
    acct_stop_time: string;
}

interface VoucherSummary {
    username: string;
    time_quota: number;
    used_time: number;
    remaining_time: number;
    data_quota: number;
    used_data: number;
    remaining_data: number;
    status: string;
    idle_timeout: number;
    session_timeout: number;
    sessions: Session[];
}

interface VoucherCheckDialogProps {
    open: boolean;
    onClose: () => void;
}

const VoucherCheckDialog: React.FC<VoucherCheckDialogProps> = ({ open, onClose }) => {
    const translate = useTranslate();
    const notify = useNotify();
    const [code, setCode] = useState('');
    const [loading, setLoading] = useState(false);
    const [data, setData] = useState<VoucherSummary | null>(null);
    const [error, setError] = useState<string | null>(null);

    const handleCheck = async () => {
        if (!code.trim()) return;
        setLoading(true);
        setError(null);
        setData(null);
        try {
            const result = await apiRequest<VoucherSummary>(`/vouchers/check?code=${encodeURIComponent(code.trim())}`);
            setData(result);
        } catch (err: any) {
            setError(err.message || 'Failed to check voucher');
        } finally {
            setLoading(false);
        }
    };

    const formatDuration = (seconds: number) => {
        const h = Math.floor(seconds / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        const s = seconds % 60;
        return `${h}h ${m}m ${s}s`;
    };

    const formatBytes = (bytes: number) => {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const handleCopy = () => {
        if (!data) return;
        const report = `
${translate('resources.vouchers.diagnostic.title')}: ${data.username}
-----------------------------------
${translate('resources.radius/users.fields.status')}: ${data.status}
${translate('resources.vouchers.diagnostic.time_usage')}:
  ${translate('resources.vouchers.diagnostic.quota_limit')}: ${formatDuration(data.time_quota)}
  ${translate('resources.vouchers.diagnostic.consumed')}: ${formatDuration(data.used_time)}
  ${translate('resources.vouchers.diagnostic.remaining')}: ${formatDuration(data.remaining_time)}

${translate('resources.vouchers.diagnostic.data_usage')}:
  ${translate('resources.vouchers.diagnostic.quota_limit')}: ${data.data_quota} MB
  ${translate('resources.vouchers.diagnostic.consumed')}: ${formatBytes(data.used_data)}
  ${translate('resources.vouchers.diagnostic.remaining')}: ${formatBytes(data.remaining_data)}

${translate('resources.vouchers.diagnostic.idle_timeout')}: ${data.idle_timeout > 0 ? data.idle_timeout + ' s' : translate('resources.products.units.unlimited')}
${translate('resources.vouchers.diagnostic.session_timeout')}: ${data.session_timeout > 0 ? data.session_timeout + ' s' : translate('resources.products.units.unlimited')}

${translate('resources.vouchers.diagnostic.sessions')}:
${data.sessions.map(s => `- ${new Date(s.acct_start_time).toLocaleString()}: ${formatDuration(s.acct_session_time)} (${formatBytes(s.acct_input_total + s.acct_output_total)})`).join('\n')}
        `.trim();

        navigator.clipboard.writeText(report);
        notify('resources.vouchers.diagnostic.copied', { type: 'success' });
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth sx={{ '& .MuiDialog-paper': { borderRadius: 3 } }}>
            <DialogTitle sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', pb: 1 }}>
                <Box display="flex" alignItems="center" gap={1}>
                    <HistoryIcon color="primary" />
                    <Typography variant="h6" fontWeight={700}>{translate('resources.vouchers.diagnostic.title')}</Typography>
                </Box>
                <IconButton onClick={onClose} size="small"><CloseIcon /></IconButton>
            </DialogTitle>
            <DialogContent dividers>
                <Box mb={3} display="flex" gap={1}>
                    <TextField
                        fullWidth
                        size="small"
                        label={translate('resources.vouchers.diagnostic.placeholder')}
                        variant="outlined"
                        value={code}
                        onChange={(e) => setCode(e.target.value)}
                        onKeyPress={(e) => e.key === 'Enter' && handleCheck()}
                        InputProps={{
                            startAdornment: (
                                <InputAdornment position="start">
                                    <SearchIcon color="action" />
                                </InputAdornment>
                            ),
                        }}
                    />
                    <Button
                        variant="contained"
                        onClick={handleCheck}
                        disabled={loading || !code.trim()}
                        startIcon={loading ? <CircularProgress size={20} color="inherit" /> : <CheckCircleIcon />}
                        sx={{ borderRadius: 2, px: 3 }}
                    >
                        {translate('resources.vouchers.diagnostic.check')}
                    </Button>
                </Box>

                {error && <Alert severity="error" sx={{ mb: 2, borderRadius: 2 }}>{error}</Alert>}

                {data && (
                    <Box>
                        <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                            <Box>
                                <Typography variant="h5" fontWeight={800} color="primary">{data.username}</Typography>
                                <Chip
                                    label={translate(`resources.vouchers.status.${data.status}`)}
                                    color={data.status === 'unused' ? 'success' : data.status === 'used' ? 'error' : 'default'}
                                    size="small"
                                    sx={{ mt: 1, fontWeight: 'bold' }}
                                />
                            </Box>
                            <Tooltip title={translate('resources.vouchers.diagnostic.copy_report')}>
                                <Button
                                    size="small"
                                    variant="outlined"
                                    startIcon={<ContentCopyIcon />}
                                    onClick={handleCopy}
                                    sx={{ borderRadius: 2 }}
                                >
                                    {translate('resources.vouchers.diagnostic.copy_report')}
                                </Button>
                            </Tooltip>
                        </Box>

                        <Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr' }} gap={2} mb={3}>
                            <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: 'rgba(0,0,0,0.02)' }}>
                                <Typography variant="caption" color="text.secondary" fontWeight={700} gutterBottom display="block">{translate('resources.vouchers.diagnostic.time_usage')}</Typography>
                                <Box display="flex" justifyContent="space-between" mb={1}>
                                    <Typography variant="body2">{translate('resources.vouchers.diagnostic.quota_limit')}:</Typography>
                                    <Typography variant="body2" fontWeight={600}>{formatDuration(data.time_quota)}</Typography>
                                </Box>
                                <Box display="flex" justifyContent="space-between" mb={1}>
                                    <Typography variant="body2">{translate('resources.vouchers.diagnostic.consumed')}:</Typography>
                                    <Typography variant="body2" fontWeight={600} color="error.main">{formatDuration(data.used_time)}</Typography>
                                </Box>
                                <Divider sx={{ my: 1 }} />
                                <Box display="flex" justifyContent="space-between">
                                    <Typography variant="body1" fontWeight={700}>{translate('resources.vouchers.diagnostic.remaining')}:</Typography>
                                    <Typography variant="body1" fontWeight={700} color={data.remaining_time > 0 ? "success.main" : "error.main"}>
                                        {formatDuration(data.remaining_time)}
                                    </Typography>
                                </Box>
                            </Paper>

                            <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: 'rgba(0,0,0,0.02)' }}>
                                <Typography variant="caption" color="text.secondary" fontWeight={700} gutterBottom display="block">{translate('resources.vouchers.diagnostic.data_usage')}</Typography>
                                <Box display="flex" justifyContent="space-between" mb={1}>
                                    <Typography variant="body2">{translate('resources.vouchers.diagnostic.quota_limit')}:</Typography>
                                    <Typography variant="body2" fontWeight={600}>{data.data_quota} MB</Typography>
                                </Box>
                                <Box display="flex" justifyContent="space-between" mb={1}>
                                    <Typography variant="body2">{translate('resources.vouchers.diagnostic.consumed')}:</Typography>
                                    <Typography variant="body2" fontWeight={600} color="error.main">{formatBytes(data.used_data)}</Typography>
                                </Box>
                                <Divider sx={{ my: 1 }} />
                                <Box display="flex" justifyContent="space-between">
                                    <Typography variant="body1" fontWeight={700}>{translate('resources.vouchers.diagnostic.remaining')}:</Typography>
                                    <Typography variant="body1" fontWeight={700} color={data.remaining_data > 0 ? "success.main" : "error.main"}>
                                        {formatBytes(data.remaining_data)}
                                    </Typography>
                                </Box>
                            </Paper>
                        </Box>

                        <Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr' }} gap={2} mb={3}>
                            <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: 'rgba(25, 118, 210, 0.04)', border: '1px solid rgba(25, 118, 210, 0.2)' }}>
                                <Typography variant="caption" color="primary.main" fontWeight={700} gutterBottom display="block">{translate('resources.vouchers.diagnostic.idle_timeout')}</Typography>
                                <Typography variant="body1" fontWeight={700}>
                                    {data.idle_timeout > 0 ? `${data.idle_timeout} ${translate('resources.products.units.seconds', { _: 'Seconds' })}` : translate('resources.products.units.unlimited', { _: 'Unlimited' })}
                                </Typography>
                            </Paper>

                            <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: 'rgba(156, 39, 176, 0.04)', border: '1px solid rgba(156, 39, 176, 0.2)' }}>
                                <Typography variant="caption" color="secondary.main" fontWeight={700} gutterBottom display="block">{translate('resources.vouchers.diagnostic.session_timeout')}</Typography>
                                <Typography variant="body1" fontWeight={700}>
                                    {data.session_timeout > 0 ? `${data.session_timeout} ${translate('resources.products.units.seconds', { _: 'Seconds' })}` : translate('resources.products.units.unlimited', { _: 'Unlimited' })}
                                </Typography>
                            </Paper>
                        </Box>

                        <Typography variant="subtitle2" fontWeight={700} mb={1} sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <HistoryIcon fontSize="small" /> {translate('resources.vouchers.diagnostic.sessions')} ({data.sessions.length})
                        </Typography>
                        <TableContainer component={Paper} variant="outlined" sx={{ borderRadius: 2, maxHeight: 300 }}>
                            <Table size="small" stickyHeader>
                                <TableHead>
                                    <TableRow>
                                        <TableCell sx={{ fontWeight: 700 }}>{translate('resources.radius/accounting.fields.acct_start_time')}</TableCell>
                                        <TableCell sx={{ fontWeight: 700 }}>{translate('resources.radius/accounting.fields.acct_session_time')}</TableCell>
                                        <TableCell sx={{ fontWeight: 700 }}>{translate('resources.radius/accounting.fields.total_traffic')}</TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {data.sessions.length > 0 ? data.sessions.map((session) => (
                                        <TableRow key={session.acct_session_id}>
                                            <TableCell>{new Date(session.acct_start_time).toLocaleString()}</TableCell>
                                            <TableCell>{formatDuration(session.acct_session_time)}</TableCell>
                                            <TableCell>{formatBytes(session.acct_input_total + session.acct_output_total)}</TableCell>
                                        </TableRow>
                                    )) : (
                                        <TableRow>
                                            <TableCell colSpan={3} align="center">{translate('resources.vouchers.diagnostic.no_sessions')}</TableCell>
                                        </TableRow>
                                    )}
                                </TableBody>
                            </Table>
                        </TableContainer>
                    </Box>
                )}

                {!data && !loading && !error && (
                    <Box py={8} textAlign="center">
                        <HistoryIcon sx={{ fontSize: 60, color: 'action.disabled', mb: 2 }} />
                        <Typography color="text.secondary">{translate('resources.vouchers.diagnostic.empty_hint')}</Typography>
                    </Box>
                )}
            </DialogContent>
            <DialogActions sx={{ p: 2 }}>
                <Button onClick={onClose}>{translate('ra.action.cancel')}</Button>
            </DialogActions>
        </Dialog>
    );
};

export default VoucherCheckDialog;
