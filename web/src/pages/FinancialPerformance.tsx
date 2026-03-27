import { useState, useEffect } from 'react';
import {
    Box,
    Card,
    CardContent,
    Typography,
    Stack,
    Tab,
    Tabs,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    TextField,
    Button,
    LinearProgress,
    Chip,
    Divider,
} from '@mui/material';
import { Grid } from '@mui/material';
import { useTheme, alpha } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import DownloadIcon from '@mui/icons-material/Download';
import AccountBalanceWalletOutlinedIcon from '@mui/icons-material/AccountBalanceWalletOutlined';
import ReceiptLongOutlinedIcon from '@mui/icons-material/ReceiptLongOutlined';
import PeopleAltOutlinedIcon from '@mui/icons-material/PeopleAltOutlined';
import VerifiedUserOutlinedIcon from '@mui/icons-material/VerifiedUserOutlined';
import InventoryOutlinedIcon from '@mui/icons-material/InventoryOutlined';
import MonetizationOnOutlinedIcon from '@mui/icons-material/MonetizationOnOutlined';
import HistoryOutlinedIcon from '@mui/icons-material/HistoryOutlined';
import { httpClient } from '../utils/apiClient';
import { Title, useTranslate, useLocale } from 'react-admin';

interface FinancialReport {
    overview: {
        total_batches: number;
        total_agents: number;
        total_balance: number;
    };
    agent_summary: {
        total_agents: number;
        total_batches: number;
        total_vouchers: number;
        total_cost: number;
        used_cost: number;
        unused_cost: number;
    };
    agents: AgentPerformance[];
    admin: {
        total_batches: number;
        total_vouchers: number;
        used_vouchers: number;
        unused_vouchers: number;
        total_cost: number;
        used_cost: number;
        unused_cost: number;
        batches: AdminBatchDetail[];
    };
    date_range: {
        start: string | null;
        end: string | null;
    };
}

interface AgentPerformance {
    id: string;
    name: string;
    username: string;
    balance: number;
    total_vouchers: number;
    used_vouchers: number;
    unused_vouchers: number;
    total_sales: number;
}

interface AdminBatchDetail {
    id: string;
    name: string;
    product_name: string;
    count: number;
    used_vouchers: number;
    unused_vouchers: number;
    total_cost: number;
    created_at: string;
}

const FinancialPerformance = () => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
    const translate = useTranslate();
    const locale = useLocale();
    const isRtl = locale === 'ar';
    const [tabValue, setTabValue] = useState(0);
    const [loading, setLoading] = useState(true);
    const [report, setReport] = useState<FinancialReport | null>(null);
    const [startDate, setStartDate] = useState('');
    const [endDate, setEndDate] = useState('');

    const fetchData = async () => {
        setLoading(true);
        try {
            let url = '/financial/report';
            const params = new URLSearchParams();
            if (startDate) params.append('start_date', new Date(startDate).toISOString());
            if (endDate) params.append('end_date', new Date(endDate).toISOString());

            if (params.toString()) {
                url += `?${params.toString()}`;
            }

            const { json } = await httpClient(url);
            setReport(json.data);
        } catch (error) {
            console.error('Failed to fetch financial report', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, [startDate, endDate]);

    const handleTabChange = (_: React.SyntheticEvent, newValue: number) => {
        setTabValue(newValue);
    };

    const handleExport = () => {
        if (!report) return;

        let csvContent = "data:text/csv;charset=utf-8,";

        if (tabValue === 1) { // Agent Performance
            csvContent += "ID,Name,Username,Balance,Total Vouchers,Used (Sold),Unused,Total Sales\n";
            (report.agents || []).forEach(agent => {
                csvContent += `${agent.id},${agent.name},${agent.username},${agent.balance.toFixed(2)},${agent.total_vouchers},${agent.used_vouchers},${agent.unused_vouchers},${(agent.used_vouchers * 0).toFixed(2)}\n`;
            });
            const encodedUri = encodeURI(csvContent);
            const link = document.createElement("a");
            link.setAttribute("href", encodedUri);
            link.setAttribute("download", `agent_performance_${new Date().toISOString().slice(0, 10)}.csv`);
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        } else {
            // General Export (Overview + Admin)
            csvContent += "Category,Metric,Value\n";
            csvContent += `Overview,Total Batches,${report.overview.total_batches}\n`;
            csvContent += `Overview,Total Agents,${report.overview.total_agents}\n`;
            csvContent += `Overview,Total Balance,${report.overview.total_balance}\n`;
            csvContent += `Admin,Total Batches,${report.admin.total_batches}\n`;
            csvContent += `Admin,Total Vouchers,${report.admin.total_vouchers}\n`;
            csvContent += `Admin,Used Vouchers,${report.admin.used_vouchers}\n`;
            csvContent += `Admin,Unused Vouchers,${report.admin.unused_vouchers}\n`;
            csvContent += `Admin,Total Cost,${report.admin.total_cost}\n`;
            csvContent += `Admin,Used Cost,${report.admin.used_cost}\n`;
            csvContent += `Admin,Unused Cost,${report.admin.unused_cost}\n`;

            const encodedUri = encodeURI(csvContent);
            const link = document.createElement("a");
            link.setAttribute("href", encodedUri);
            link.setAttribute("download", `financial_overview_${new Date().toISOString().slice(0, 10)}.csv`);
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        }
    };

    const StatCard = ({ title, value, icon, color, isCurrency = false }: any) => (
        <Card sx={{ height: '100%', borderRadius: 3, transition: 'transform 0.2s', '&:hover': { transform: 'translateY(-4px)' } }}>
            <CardContent>
                <Stack direction="row" justifyContent="space-between" alignItems="center">
                    <Box>
                        <Typography color="textSecondary" variant="subtitle2" gutterBottom sx={{ fontWeight: 600, textTransform: 'uppercase', letterSpacing: 0.5 }}>
                            {title}
                        </Typography>
                        <Typography variant="h4" fontWeight="bold" sx={{ color: color }}>
                            {isCurrency ? `$${Number(value).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}` : value}
                        </Typography>
                    </Box>
                    <Box
                        sx={{
                            backgroundColor: alpha(color, 0.1),
                            color: color,
                            borderRadius: '50%',
                            p: 1.5,
                            display: 'flex',
                        }}
                    >
                        {icon}
                    </Box>
                </Stack>
            </CardContent>
        </Card>
    );

    if (loading && !report) return <LinearProgress />;
    if (!report) return <Typography>Error loading report</Typography>;

    return (
        <Box sx={{ p: { xs: 1.5, sm: 2, md: 3 }, direction: isRtl ? 'rtl' : 'ltr' }}>
            <Title title={translate('pages.financial.title')} />

            {/* Header & Controls */}
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={3} mb={3} mt={{ xs: 5, md: 0 }} justifyContent="space-between" alignItems={{ xs: 'stretch', md: 'center' }}>
                <Typography variant="h5" fontWeight="bold" sx={{ display: { xs: 'none', md: 'block' } }}>
                    {translate('pages.financial.title')}
                </Typography>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} alignItems={{ xs: 'stretch', sm: 'center' }} sx={{ width: { xs: '100%', md: 'auto' } }}>
                    <TextField
                        type="date"
                        label={translate('pages.financial.controls.start_date')}
                        size="small"
                        InputLabelProps={{
                            shrink: true,
                            sx: isRtl ? {
                                transformOrigin: 'top right',
                                left: 'unset',
                                right: '1.75rem',
                            } : {}
                        }}
                        sx={{
                            minWidth: { sm: 160 },
                            '& legend': isRtl ? { textAlign: 'right' } : {}
                        }}
                        value={startDate}
                        onChange={(e) => setStartDate(e.target.value)}
                        fullWidth={isMobile}
                    />
                    <TextField
                        type="date"
                        label={translate('pages.financial.controls.end_date')}
                        size="small"
                        InputLabelProps={{
                            shrink: true,
                            sx: isRtl ? {
                                transformOrigin: 'top right',
                                left: 'unset',
                                right: '1.75rem',
                            } : {}
                        }}
                        sx={{
                            minWidth: { sm: 160 },
                            '& legend': isRtl ? { textAlign: 'right' } : {}
                        }}
                        value={endDate}
                        onChange={(e) => setEndDate(e.target.value)}
                        fullWidth={isMobile}
                    />
                    <Button 
                        variant="contained" 
                        startIcon={!isRtl && <DownloadIcon />}
                        href="" 
                        onClick={handleExport} 
                        sx={{ 
                            borderRadius: 2, 
                            whiteSpace: 'nowrap', 
                            py: { xs: 1, sm: 0.8 },
                            height: 40,
                            minWidth: 140
                        }} 
                        fullWidth={isMobile}
                    >
                        {isRtl && <DownloadIcon sx={{ mr: 0, ml: 1 }} />}
                        {translate('pages.financial.controls.export_csv')}
                    </Button>
                </Stack>
            </Stack>

            <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
                <Tabs value={tabValue} onChange={handleTabChange} variant="fullWidth" sx={{ '& .MuiTab-root': { fontSize: { xs: '0.75rem', sm: '0.875rem' }, minWidth: 0, px: 1 } }}>
                    <Tab label={translate('pages.financial.tabs.overview')} />
                    <Tab label={isMobile ? translate('pages.financial.agent.agents') : translate('pages.financial.tabs.agent_performance')} />
                    <Tab label={isMobile ? translate('pages.financial.agent.batches') : translate('pages.financial.tabs.admin_performance')} />
                </Tabs>
            </Box>

            {/* Overview Tab */}
            {tabValue === 0 && (
                <Grid container spacing={3}>
                    <Grid size={{ xs: 12, md: 4 }}>
                        <StatCard
                            title={translate('pages.financial.overview.total_agents')}
                            value={report.overview.total_agents}
                            icon={<PeopleAltOutlinedIcon fontSize="medium" />}
                            color={theme.palette.primary.main}
                        />
                    </Grid>
                    <Grid size={{ xs: 12, md: 4 }}>
                        <StatCard
                            title={translate('pages.financial.overview.total_batches_system')}
                            value={report.overview.total_batches}
                            icon={<ReceiptLongOutlinedIcon fontSize="medium" />}
                            color={theme.palette.secondary.main}
                        />
                    </Grid>
                    <Grid size={{ xs: 12, md: 4 }}>
                        <StatCard
                            title={translate('pages.financial.overview.total_agent_balance')}
                            value={report.overview.total_balance}
                            icon={<AccountBalanceWalletOutlinedIcon fontSize="medium" />}
                            color="#10b981"
                            isCurrency={true}
                        />
                    </Grid>
                </Grid>
            )}

            {/* Agent Performance Tab */}
            {tabValue === 1 && (
                <Stack spacing={4}>
                    <Box>
                        <Typography variant="h6" gutterBottom sx={{ fontWeight: 600, display: 'flex', alignItems: 'center', mb: 2 }}>
                            <PeopleAltOutlinedIcon color="primary" sx={{ mr: isRtl ? 0 : 1, ml: isRtl ? 1 : 0 }} />
                            {translate('pages.financial.agent.network_summary')}
                        </Typography>
                        <Grid container spacing={2}>
                            <Grid size={{ xs: 12, md: 4, lg: 2 }}>
                                <StatCard
                                    title={translate('pages.financial.agent.agents')}
                                    value={report.agent_summary.total_agents}
                                    icon={<PeopleAltOutlinedIcon fontSize="small" />}
                                    color={theme.palette.primary.main}
                                />
                            </Grid>
                            <Grid size={{ xs: 12, md: 4, lg: 2 }}>
                                <StatCard
                                    title={translate('pages.financial.agent.batches')}
                                    value={report.agent_summary.total_batches}
                                    icon={<ReceiptLongOutlinedIcon fontSize="small" />}
                                    color={theme.palette.info.main}
                                />
                            </Grid>
                            <Grid size={{ xs: 12, md: 4, lg: 2 }}>
                                <StatCard
                                    title={translate('pages.financial.agent.vouchers')}
                                    value={report.agent_summary.total_vouchers}
                                    icon={<VerifiedUserOutlinedIcon fontSize="small" />}
                                    color={theme.palette.secondary.main}
                                />
                            </Grid>
                            <Grid size={{ xs: 12, md: 4, lg: 2 }}>
                                <StatCard
                                    title={translate('pages.financial.agent.total_value')}
                                    value={report.agent_summary.total_cost}
                                    icon={<MonetizationOnOutlinedIcon fontSize="small" />}
                                    color={theme.palette.success.main}
                                    isCurrency={true}
                                />
                            </Grid>
                            <Grid size={{ xs: 12, md: 4, lg: 2 }}>
                                <StatCard
                                    title={translate('pages.financial.agent.sold_value')}
                                    value={report.agent_summary.used_cost}
                                    icon={<InventoryOutlinedIcon fontSize="small" />}
                                    color={theme.palette.success.main}
                                    isCurrency={true}
                                />
                            </Grid>
                            <Grid size={{ xs: 12, md: 4, lg: 2 }}>
                                <StatCard
                                    title={translate('pages.financial.agent.unused_value')}
                                    value={report.agent_summary.unused_cost}
                                    icon={<HistoryOutlinedIcon fontSize="small" />}
                                    color={theme.palette.warning.main}
                                    isCurrency={true}
                                />
                            </Grid>
                        </Grid>
                    </Box>

                    <Card sx={{ borderRadius: 3 }}>
                        <Typography variant="h6" sx={{ p: 2, fontWeight: 600 }}>{translate('pages.financial.agent.detailed_performance')}</Typography>
                        <Divider />
                        {isMobile ? (
                            <Box sx={{ p: 2 }}>
                                {(report.agents || []).map((agent) => (
                                    <Card variant="outlined" key={agent.id} sx={{ mb: 2, borderRadius: 2 }}>
                                        <CardContent sx={{ pb: 2 }}>
                                            <Stack direction="row" justifyContent="space-between" mb={1}>
                                                <Typography variant="subtitle1" fontWeight="bold">{agent.name}</Typography>
                                                <Typography variant="subtitle1" fontWeight="bold" color="primary.main">${agent.balance.toFixed(2)}</Typography>
                                            </Stack>
                                            <Typography variant="body2" color="text.secondary" mb={2}>@{agent.username}</Typography>
                                            <Grid container spacing={1}>
                                                <Grid size={{ xs: 4 }}>
                                                    <Typography variant="caption" color="text.secondary">{translate('pages.financial.agent.card.total')}</Typography>
                                                    <Typography variant="body2" fontWeight="bold">{agent.total_vouchers}</Typography>
                                                </Grid>
                                                <Grid size={{ xs: 4 }}>
                                                    <Typography variant="caption" color="text.secondary">{translate('pages.financial.agent.card.sold')}</Typography>
                                                    <Box><Chip label={agent.used_vouchers} color="success" size="small" variant="outlined" sx={{ fontWeight: 600, height: 20 }} /></Box>
                                                </Grid>
                                                <Grid size={{ xs: 4 }}>
                                                    <Typography variant="caption" color="text.secondary">{translate('pages.financial.agent.card.unused')}</Typography>
                                                    <Box><Chip label={agent.unused_vouchers} color="warning" size="small" variant="outlined" sx={{ fontWeight: 600, height: 20 }} /></Box>
                                                </Grid>
                                            </Grid>
                                        </CardContent>
                                    </Card>
                                ))}
                                {(report.agents || []).length === 0 && (
                                    <Typography color="text.secondary" textAlign="center" py={3}>{translate('pages.financial.agent.no_agents')}</Typography>
                                )}
                            </Box>
                        ) : (
                            <Table>
                                <TableHead>
                                    <TableRow sx={{ backgroundColor: alpha(theme.palette.primary.main, 0.05) }}>
                                        <TableCell align={isRtl ? "right" : "left"}>{translate('pages.financial.agent.table.agent_name')}</TableCell>
                                        <TableCell align={isRtl ? "right" : "left"}>{translate('pages.financial.agent.table.username')}</TableCell>
                                        <TableCell align={isRtl ? "left" : "right"}>{translate('pages.financial.agent.table.wallet_balance')}</TableCell>
                                        <TableCell align={isRtl ? "left" : "right"}>{translate('pages.financial.agent.table.total_vouchers')}</TableCell>
                                        <TableCell align={isRtl ? "left" : "right"}>{translate('pages.financial.agent.table.sold')}</TableCell>
                                        <TableCell align={isRtl ? "left" : "right"}>{translate('pages.financial.agent.table.unused')}</TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {(report.agents || []).map((agent) => (
                                        <TableRow key={agent.id} hover>
                                            <TableCell align={isRtl ? "right" : "left"} sx={{ fontWeight: 'bold' }}>{agent.name}</TableCell>
                                            <TableCell align={isRtl ? "right" : "left"}>{agent.username}</TableCell>
                                            <TableCell align={isRtl ? "left" : "right"} sx={{ fontWeight: 'bold', color: theme.palette.primary.main }}>
                                                ${agent.balance.toFixed(2)}
                                            </TableCell>
                                            <TableCell align={isRtl ? "left" : "right"}>{agent.total_vouchers}</TableCell>
                                            <TableCell align={isRtl ? "left" : "right"}>
                                                <Chip label={agent.used_vouchers} color="success" size="small" variant="outlined" sx={{ fontWeight: 600 }} />
                                            </TableCell>
                                            <TableCell align={isRtl ? "left" : "right"}>
                                                <Chip label={agent.unused_vouchers} color="warning" size="small" variant="outlined" sx={{ fontWeight: 600 }} />
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                    {(report.agents || []).length === 0 && (
                                        <TableRow>
                                            <TableCell colSpan={6} align="center" sx={{ py: 3, color: 'text.secondary' }}>{translate('pages.financial.agent.no_agents')}</TableCell>
                                        </TableRow>
                                    )}
                                </TableBody>
                            </Table>
                        )}
                    </Card>
                </Stack>
            )}

            {/* Admin Performance Tab */}
            {tabValue === 2 && (
                <Stack spacing={4}>
                    <Box>
                        <Typography variant="h6" gutterBottom sx={{ fontWeight: 600, display: 'flex', alignItems: 'center', mb: 2 }}>
                            <MonetizationOnOutlinedIcon color="primary" sx={{ mr: isRtl ? 0 : 1, ml: isRtl ? 1 : 0 }} />
                            {translate('pages.financial.admin.financial_metrics')}
                        </Typography>
                        <Grid container spacing={3}>
                            <Grid size={{ xs: 12, md: 4 }}>
                                <StatCard
                                    title={translate('pages.financial.agent.total_value')}
                                    value={report.admin.total_cost}
                                    icon={<MonetizationOnOutlinedIcon fontSize="medium" />}
                                    color={theme.palette.primary.main}
                                    isCurrency={true}
                                />
                            </Grid>
                            <Grid size={{ xs: 12, md: 4 }}>
                                <StatCard
                                    title={translate('pages.financial.admin.value_of_sold')}
                                    value={report.admin.used_cost}
                                    icon={<InventoryOutlinedIcon fontSize="medium" />}
                                    color={theme.palette.success.main}
                                    isCurrency={true}
                                />
                            </Grid>
                            <Grid size={{ xs: 12, md: 4 }}>
                                <StatCard
                                    title={translate('pages.financial.admin.value_of_unused')}
                                    value={report.admin.unused_cost}
                                    icon={<HistoryOutlinedIcon fontSize="medium" />}
                                    color={theme.palette.warning.main}
                                    isCurrency={true}
                                />
                            </Grid>
                        </Grid>
                    </Box>

                    <Box>
                        <Typography variant="h6" gutterBottom sx={{ fontWeight: 600, display: 'flex', alignItems: 'center', mb: 2 }}>
                            <HistoryOutlinedIcon color="primary" sx={{ mr: isRtl ? 0 : 1, ml: isRtl ? 1 : 0 }} />
                            {translate('pages.financial.admin.batch_details')}
                        </Typography>
                        <Card sx={{ borderRadius: 3 }}>
                            {isMobile ? (
                                <Box sx={{ p: 2 }}>
                                    {(report.admin.batches || []).map((batch) => (
                                        <Card variant="outlined" key={batch.id} sx={{ mb: 2, borderRadius: 2 }}>
                                            <CardContent sx={{ pb: 2 }}>
                                                <Stack direction="row" justifyContent="space-between" mb={1}>
                                                    <Typography variant="subtitle1" fontWeight="bold">{batch.name}</Typography>
                                                    <Typography variant="subtitle1" fontWeight="bold" color="primary.main">${batch.total_cost.toFixed(2)}</Typography>
                                                </Stack>
                                                <Box mb={2}>
                                                    <Chip label={batch.product_name} size="small" sx={{ borderRadius: 1 }} />
                                                </Box>
                                                <Grid container spacing={1}>
                                                    <Grid size={{ xs: 4 }}>
                                                        <Typography variant="caption" color="text.secondary">{translate('pages.financial.agent.card.total')}</Typography>
                                                        <Typography variant="body2" fontWeight="bold">{batch.count}</Typography>
                                                    </Grid>
                                                    <Grid size={{ xs: 4 }}>
                                                        <Typography variant="caption" color="text.secondary">{translate('pages.financial.agent.card.sold')}</Typography>
                                                        <Typography variant="body2" color="success.main" fontWeight={600}>{batch.used_vouchers}</Typography>
                                                    </Grid>
                                                    <Grid size={{ xs: 4 }}>
                                                        <Typography variant="caption" color="text.secondary">{translate('pages.financial.agent.card.unused')}</Typography>
                                                        <Typography variant="body2" color="warning.main" fontWeight={600}>{batch.unused_vouchers}</Typography>
                                                    </Grid>
                                                </Grid>
                                                <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 2, textAlign: 'right' }}>
                                                    {new Date(batch.created_at).toLocaleString()}
                                                </Typography>
                                            </CardContent>
                                        </Card>
                                    ))}
                                    {(report.admin.batches || []).length === 0 && (
                                        <Typography color="text.secondary" textAlign="center" py={3}>{translate('pages.financial.admin.no_batches')}</Typography>
                                    )}
                                </Box>
                            ) : (
                                <Table>
                                    <TableHead>
                                        <TableRow sx={{ backgroundColor: alpha(theme.palette.primary.main, 0.05) }}>
                                            <TableCell align={isRtl ? "right" : "left"}>{translate('pages.financial.admin.table.batch_name')}</TableCell>
                                            <TableCell align={isRtl ? "right" : "left"}>{translate('pages.financial.admin.table.product')}</TableCell>
                                            <TableCell align={isRtl ? "left" : "right"}>{translate('pages.financial.admin.table.count')}</TableCell>
                                            <TableCell align={isRtl ? "left" : "right"}>{translate('pages.financial.admin.table.sold')}</TableCell>
                                            <TableCell align={isRtl ? "left" : "right"}>{translate('pages.financial.admin.table.unused')}</TableCell>
                                            <TableCell align={isRtl ? "left" : "right"}>{translate('pages.financial.admin.table.total_value')}</TableCell>
                                            <TableCell align={isRtl ? "left" : "right"}>{translate('pages.financial.admin.table.generated_at')}</TableCell>
                                        </TableRow>
                                    </TableHead>
                                    <TableBody>
                                        {(report.admin.batches || []).map((batch) => (
                                            <TableRow key={batch.id} hover>
                                                <TableCell align={isRtl ? "right" : "left"} sx={{ fontWeight: 'bold' }}>{batch.name}</TableCell>
                                                <TableCell align={isRtl ? "right" : "left"}>
                                                    <Chip label={batch.product_name} size="small" sx={{ borderRadius: 1 }} />
                                                </TableCell>
                                                <TableCell align={isRtl ? "left" : "right"}>{batch.count}</TableCell>
                                                <TableCell align={isRtl ? "left" : "right"}>
                                                    <Typography variant="body2" color="success.main" fontWeight={600}>{batch.used_vouchers}</Typography>
                                                </TableCell>
                                                <TableCell align={isRtl ? "left" : "right"}>
                                                    <Typography variant="body2" color="warning.main" fontWeight={600}>{batch.unused_vouchers}</Typography>
                                                </TableCell>
                                                <TableCell align={isRtl ? "left" : "right"} sx={{ fontWeight: 'bold' }}>
                                                    ${batch.total_cost.toFixed(2)}
                                                </TableCell>
                                                <TableCell align={isRtl ? "left" : "right"} sx={{ color: 'text.secondary', fontSize: '0.875rem' }}>
                                                    {new Date(batch.created_at).toLocaleString()}
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                        {(report.admin.batches || []).length === 0 && (
                                            <TableRow>
                                                <TableCell colSpan={7} align="center" sx={{ py: 3, color: 'text.secondary' }}>{translate('pages.financial.admin.no_batches')}</TableCell>
                                            </TableRow>
                                        )}
                                    </TableBody>
                                </Table>
                            )}
                        </Card>
                    </Box>
                </Stack>
            )}
        </Box>
    );
};

export default FinancialPerformance;
