import AccountBalanceWalletOutlinedIcon from '@mui/icons-material/AccountBalanceWalletOutlined';
import ReceiptLongOutlinedIcon from '@mui/icons-material/ReceiptLongOutlined';
import ConfirmationNumberOutlinedIcon from '@mui/icons-material/ConfirmationNumberOutlined';
import InventoryOutlinedIcon from '@mui/icons-material/InventoryOutlined';
import HistoryOutlinedIcon from '@mui/icons-material/HistoryOutlined';
import {
    Box,
    Card,
    CardContent,
    Typography,
    Stack,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Chip,
    LinearProgress,
} from '@mui/material';
import { alpha, useTheme } from '@mui/material/styles';
import { Grid } from '@mui/material';
import { useState, useEffect } from 'react';
import { useGetIdentity, useTranslate } from 'react-admin';
import { httpClient } from '../utils/apiClient';

interface AgentStats {
    balance: number;
    total_batches: number;
    total_vouchers: number;
    used_vouchers: number;
    unused_vouchers: number;
    recent_transactions: WalletLog[];
    recent_batches: BatchStats[];
}

interface BatchStats {
    id: string;
    name: string;
    total_vouchers: number;
    used_vouchers: number;
    unused_vouchers: number;
    created_at: string;
}

interface WalletLog {
    id: string;
    type: string;
    amount: number;
    balance: number;
    reference_id: string;
    remark: string;
    created_at: string;
}

const AgentDashboard = () => {
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';
    const translate = useTranslate();
    const { data: identity, isLoading: identityLoading } = useGetIdentity();
    const [stats, setStats] = useState<AgentStats | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        if (identity && identity.id) {
            httpClient(`/agents/${identity.id}/stats`)
                .then(({ json }) => {
                    setStats(json.data);
                    setLoading(false);
                })
                .catch(() => {
                    setLoading(false);
                });
        }
    }, [identity]);

    if (loading || identityLoading) {
        return <LinearProgress sx={{ mt: 2 }} />;
    }

    if (!stats) {
        return <Typography>{translate('agent_dashboard.load_error')}</Typography>;
    }

    const numberFormatter = new Intl.NumberFormat();
    const dateFormatter = new Intl.DateTimeFormat(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
    });

    const statCards = [
        {
            title: translate('agent_dashboard.current_balance'),
            value: `${stats.balance.toFixed(2)}`,
            icon: <AccountBalanceWalletOutlinedIcon fontSize="large" />,
            accent: theme.palette.primary.main,
        },
        {
            title: translate('agent_dashboard.voucher_batches'),
            value: numberFormatter.format(stats.total_batches),
            icon: <ReceiptLongOutlinedIcon fontSize="large" />,
            accent: '#34d399',
        },
        {
            title: translate('agent_dashboard.total_vouchers'),
            value: numberFormatter.format(stats.total_vouchers),
            icon: <ConfirmationNumberOutlinedIcon fontSize="large" />,
            accent: '#f97316',
        },
        {
            title: translate('agent_dashboard.sold_redeemed'),
            value: numberFormatter.format(stats.used_vouchers),
            icon: <InventoryOutlinedIcon fontSize="large" />,
            accent: '#8b5cf6',
        },
    ];

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
            <Card
                sx={{
                    borderRadius: 4,
                    overflow: 'hidden',
                    background: isDark
                        ? 'linear-gradient(135deg, #1e293b, #334155)'
                        : 'linear-gradient(135deg, #eef2ff, #fdf2f8)',
                    border: `1px solid ${isDark ? 'rgba(148, 163, 184, 0.1)' : 'rgba(255, 255, 255, 0.6)'}`,
                }}
            >
                <CardContent>
                    <Stack
                        direction={{ xs: 'column', md: 'row' }}
                        spacing={3}
                        alignItems="center"
                        justifyContent="space-between"
                    >
                        <Box>
                            <Chip label={translate('agent_dashboard.title')} color="primary" sx={{ mb: 2, fontWeight: 600 }} />
                            <Typography variant="h5" sx={{ fontWeight: 700, mb: 1 }}>
                                {translate('agent_dashboard.welcome', { name: identity?.fullName })}
                            </Typography>
                            <Typography variant="body1" sx={{ color: 'text.secondary', maxWidth: 520 }}>
                                {translate('agent_dashboard.subtitle')}
                            </Typography>
                        </Box>
                        <Box sx={{ textAlign: 'center', minWidth: 200 }}>
                            <Typography variant="subtitle2" color="text.secondary">{translate('agent_dashboard.available_balance')}</Typography>
                            <Typography variant="h3" sx={{ fontWeight: 700, color: theme.palette.primary.main }}>
                                {stats.balance.toFixed(2)}
                            </Typography>
                        </Box>
                    </Stack>
                </CardContent>
            </Card>

            <Grid container spacing={3}>
                {statCards.map((card) => (
                    <Grid size={{ xs: 12, sm: 6, lg: 3 }} key={card.title}>
                        <Card sx={{ height: '100%', borderRadius: 4 }}>
                            <CardContent>
                                <Stack direction="row" justifyContent="space-between" alignItems="center">
                                    <Box>
                                        <Typography variant="subtitle2" color="text.secondary">
                                            {card.title}
                                        </Typography>
                                        <Typography variant="h4" sx={{ fontWeight: 700, my: 1 }}>
                                            {card.value}
                                        </Typography>
                                    </Box>
                                    <Box
                                        sx={{
                                            width: 48,
                                            height: 48,
                                            borderRadius: 2,
                                            display: 'grid',
                                            placeItems: 'center',
                                            backgroundColor: alpha(card.accent, 0.15),
                                            color: card.accent,
                                        }}
                                    >
                                        {card.icon}
                                    </Box>
                                </Stack>
                            </CardContent>
                        </Card>
                    </Grid>
                ))}
            </Grid>

            <Card sx={{ borderRadius: 4 }}>
                <CardContent>
                    <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 3 }}>
                        <InventoryOutlinedIcon color="primary" />
                        <Typography variant="h6" sx={{ fontWeight: 700 }}>
                            {translate('agent_dashboard.recent_batches')}
                        </Typography>
                    </Stack>
                    <Table>
                        <TableHead>
                            <TableRow>
                                <TableCell>{translate('agent_dashboard.date')}</TableCell>
                                <TableCell>{translate('agent_dashboard.batch_name')}</TableCell>
                                <TableCell align="right">{translate('agent_dashboard.total')}</TableCell>
                                <TableCell align="right">{translate('agent_dashboard.used_sold')}</TableCell>
                                <TableCell align="right">{translate('agent_dashboard.unused_remaining')}</TableCell>
                                <TableCell align="right">{translate('agent_dashboard.usage_rate')}</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {(stats.recent_batches || []).map((batch) => {
                                const usageRate = batch.total_vouchers > 0
                                    ? (batch.used_vouchers / batch.total_vouchers) * 100
                                    : 0;
                                return (
                                    <TableRow key={batch.id}>
                                        <TableCell>{dateFormatter.format(new Date(batch.created_at))}</TableCell>
                                        <TableCell sx={{ fontWeight: 600 }}>{batch.name}</TableCell>
                                        <TableCell align="right">{batch.total_vouchers}</TableCell>
                                        <TableCell align="right">
                                            <Chip label={batch.used_vouchers} size="small" color="success" />
                                        </TableCell>
                                        <TableCell align="right">
                                            <Chip label={batch.unused_vouchers} size="small" variant="outlined" />
                                        </TableCell>
                                        <TableCell align="right">
                                            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: 1 }}>
                                                <Typography variant="body2">{usageRate.toFixed(0)}%</Typography>
                                                <LinearProgress
                                                    variant="determinate"
                                                    value={usageRate}
                                                    sx={{ width: 60, height: 6, borderRadius: 3 }}
                                                />
                                            </Box>
                                        </TableCell>
                                    </TableRow>
                                );
                            })}
                            {(!stats.recent_batches || stats.recent_batches.length === 0) && (
                                <TableRow>
                                    <TableCell colSpan={6} align="center">{translate('agent_dashboard.no_batches')}</TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            <Card sx={{ borderRadius: 4 }}>
                <CardContent>
                    <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 3 }}>
                        <HistoryOutlinedIcon color="primary" />
                        <Typography variant="h6" sx={{ fontWeight: 700 }}>
                            {translate('agent_dashboard.recent_activities')}
                        </Typography>
                    </Stack>
                    <Table>
                        <TableHead>
                            <TableRow>
                                <TableCell>{translate('agent_dashboard.date')}</TableCell>
                                <TableCell>{translate('agent_dashboard.type')}</TableCell>
                                <TableCell>{translate('agent_dashboard.reference')}</TableCell>
                                <TableCell>{translate('agent_dashboard.remark')}</TableCell>
                                <TableCell align="right">{translate('agent_dashboard.amount')}</TableCell>
                                <TableCell align="right">{translate('agent_dashboard.after_balance')}</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {(stats.recent_transactions || []).map((tr) => (
                                <TableRow key={tr.id}>
                                    <TableCell>{dateFormatter.format(new Date(tr.created_at))}</TableCell>
                                    <TableCell>
                                        <Chip
                                            label={tr.type.toUpperCase()}
                                            size="small"
                                            color={tr.type === 'deposit' ? 'success' : 'primary'}
                                            variant="outlined"
                                        />
                                    </TableCell>
                                    <TableCell>
                                        <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>{tr.reference_id}</Typography>
                                    </TableCell>
                                    <TableCell>{tr.remark}</TableCell>
                                    <TableCell align="right" sx={{ color: tr.amount < 0 ? 'error.main' : 'success.main', fontWeight: 600 }}>
                                        {tr.amount > 0 ? `+${tr.amount.toFixed(2)}` : tr.amount.toFixed(2)}
                                    </TableCell>
                                    <TableCell align="right">{tr.balance.toFixed(2)}</TableCell>
                                </TableRow>
                            ))}
                            {stats.recent_transactions.length === 0 && (
                                <TableRow>
                                    <TableCell colSpan={6} align="center">{translate('agent_dashboard.no_transactions')}</TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </Box>
    );
};

export default AgentDashboard;
