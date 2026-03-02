import React from 'react';
import {
    List,
    Datagrid,
    TextField,
    DateField,
    NumberField,
    Show,
    SimpleShowLayout,
    Button,
    useNotify,
    useRefresh,
    useRecordContext,
    TopToolbar,
    ListButton,
    FunctionField
} from 'react-admin';
import { Box, Card, CardContent, Typography, Chip, Stack, alpha } from '@mui/material';
import PaymentsIcon from '@mui/icons-material/Payments';
import ReceiptIcon from '@mui/icons-material/Receipt';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import PrintIcon from '@mui/icons-material/Print';
import NoteIcon from '@mui/icons-material/Note';
import WhatsAppIcon from '@mui/icons-material/WhatsApp';
import { httpClient } from '../utils/apiClient';

const printStyles = `
    @media print {
        @page {
            margin: 8mm;
            size: A4 portrait;
        }
        html, body {
            height: 100% !important;
            overflow: hidden !important;
            background: white !important;
            margin: 0 !important;
            padding: 0 !important;
            font-size: 10pt !important;
        }
        #root {
            height: 100% !important;
            overflow: hidden !important;
        }
        /* Hide sidebar, header, actions */
        .MuiDrawer-root,
        .MuiDrawer-paper,
        .RaSidebar-root,
        .RaSidebar-fixed,
        header.MuiAppBar-root,
        .RaAppBar-root,
        .no-print,
        .RaShow-actions {
            display: none !important;
            visibility: hidden !important;
        }
        /* Reset layout */
        .RaLayout-appFrame,
        .RaLayout-main,
        .RaLayout-content {
            margin: 0 !important;
            padding: 0 !important;
            width: 100% !important;
            height: 100% !important;
            overflow: hidden !important;
        }
        /* Content fills page */
        .invoice-content {
            margin: 0 !important;
            padding: 0 !important;
            width: 100% !important;
            height: 100% !important;
            max-width: 100% !important;
            overflow: hidden !important;
            display: flex !important;
            flex-direction: column !important;
        }
        /* Compact cards */
        .MuiCard-root {
            page-break-inside: avoid;
            margin-bottom: 6px !important;
        }
        .MuiCardContent-root {
            padding: 8px 12px !important;
        }
        /* Reduce typography sizes */
        .MuiTypography-h5 {
            font-size: 16pt !important;
        }
        .MuiTypography-h6 {
            font-size: 12pt !important;
        }
        .MuiTypography-body1 {
            font-size: 10pt !important;
        }
        .MuiTypography-body2 {
            font-size: 9pt !important;
        }
        .MuiTypography-caption {
            font-size: 8pt !important;
        }
        /* Reduce stack spacing */
        .MuiStack-root {
            gap: 6px !important;
        }
    }
`;

const PrintStyles = () => <style>{printStyles}</style>;

const StatusChip = ({ label }: { label?: string }) => {
    const record = useRecordContext();
    if (!record) return null;

    let color: "success" | "error" | "warning" | "default" = "default";
    const status = record.status || 'unpaid';
    switch (status) {
        case 'paid': color = 'success'; break;
        case 'unpaid': color = 'warning'; break;
        case 'overdue': color = 'error'; break;
    }

    return (
        <Chip
            label={label || status.toUpperCase()}
            color={color}
            size="small"
            sx={{ fontWeight: 600 }}
        />
    );
};

const PayButton = () => {
    const record = useRecordContext();
    const notify = useNotify();
    const refresh = useRefresh();

    if (!record || record.status !== 'unpaid') return null;

    const handlePay = async (e: React.MouseEvent) => {
        e.stopPropagation();
        try {
            await httpClient(`/radius/invoices/${record.id}/pay`, { method: 'POST' });
            notify('resources.radius/invoices.notifications.paid', { type: 'success' });
            refresh();
        } catch (error) {
            notify('resources.radius/invoices.notifications.pay_error', { type: 'error' });
        }
    };

    return (
        <Button
            label="resources.invoices.actions.pay"
            onClick={handlePay}
            variant="contained"
            color="primary"
            startIcon={<PaymentsIcon />}
        />
    );
};

const WhatsAppShareButton = ({ variant = "outlined" }: { variant?: "outlined" | "contained" | "text" }) => {
    const record = useRecordContext();
    if (!record) return null;

    const handleShare = () => {
        // Format invoice details for WhatsApp message
        const message = encodeURIComponent(
            `*Invoice #${record.id}*\n\n` +
            `User: ${record.username}\n` +
            `Amount: $${Number(record.amount).toFixed(2)}\n` +
            `Base Fee: $${Number(record.base_amount || 0).toFixed(2)}\n` +
            `Usage: ${Number(record.usage_gb || 0).toFixed(2)} GB\n` +
            `Status: ${record.status?.toUpperCase()}\n` +
            `Due Date: ${new Date(record.due_date).toLocaleDateString()}\n\n` +
            `_Please make payment before the due date._`
        );

        // Open WhatsApp with pre-filled message
        const whatsappUrl = `https://wa.me/?text=${message}`;
        window.open(whatsappUrl, '_blank');
    };

    return (
        <Button
            label="WhatsApp"
            onClick={handleShare}
            variant={variant}
            color="success"
            startIcon={<WhatsAppIcon />}
        />
    );
};

const WhatsAppListButton = () => {
    const record = useRecordContext();
    if (!record) return null;

    const handleShare = (e: React.MouseEvent) => {
        e.stopPropagation();
        const message = encodeURIComponent(
            `*Invoice #${record.id}*\n\n` +
            `User: ${record.username}\n` +
            `Amount: $${Number(record.amount).toFixed(2)}\n` +
            `Status: ${record.status?.toUpperCase()}\n` +
            `Due: ${new Date(record.due_date).toLocaleDateString()}\n\n` +
            `_Please make payment before the due date._`
        );
        window.open(`https://wa.me/?text=${message}`, '_blank');
    };

    return (
        <Button
            label=""
            onClick={handleShare}
            color="success"
            startIcon={<WhatsAppIcon />}
        />
    );
};

export const InvoiceList = () => (
    <List sort={{ field: 'id', order: 'DESC' }}>
        <Datagrid rowClick="show">
            <TextField source="id" />
            <TextField source="username" />
            <NumberField source="usage_gb" options={{ maximumFractionDigits: 2 }} />
            <NumberField source="amount" options={{ style: 'currency', currency: 'USD' }} />
            <StatusChip label="Status" />
            <DateField source="issue_date" />
            <PayButton />
            <WhatsAppListButton />
        </Datagrid>
    </List>
);

const InvoiceShowActions = () => (
    <TopToolbar>
        <ListButton icon={<ArrowBackIcon />} />
        <Button
            label="Print"
            onClick={() => window.print()}
            variant="text"
            color="primary"
            startIcon={<PrintIcon />}
        />
        <WhatsAppShareButton variant="text" />
    </TopToolbar>
);

export const InvoiceShow = () => {
    return (
        <Show actions={<InvoiceShowActions />}>
            <PrintStyles />
            <SimpleShowLayout>
                <Box sx={{ p: 0.5 }} className="invoice-content">
                    <Stack spacing={0.5} sx={{ height: '100%' }}>
                        {/* Header Summary Card - Compact */}
                        <Card elevation={2} sx={{
                            borderRadius: 1,
                            background: 'linear-gradient(135deg, #1e3a8a 0%, #3b82f6 100%)',
                            color: 'white',
                            flexShrink: 0
                        }}>
                            <CardContent sx={{ p: 1, py: 1.5 }}>
                                <Box display="flex" justifyContent="space-between" alignItems="center">
                                    <Box>
                                        <Typography variant="h6" fontWeight={700} sx={{ fontSize: '1.1rem' }}>
                                            INVOICE
                                        </Typography>
                                        <Typography variant="body2" sx={{ opacity: 0.9, fontSize: '0.85rem' }}>
                                            User: <TextField source="username" sx={{ fontWeight: 600, color: 'inherit' }} />
                                        </Typography>
                                    </Box>
                                    <Box textAlign="right">
                                        <StatusChip />
                                        <Typography variant="h6" fontWeight={800} sx={{ mt: 0.5, fontSize: '1.25rem' }}>
                                            <NumberField source="amount" options={{ style: 'currency', currency: 'USD' }} />
                                        </Typography>
                                    </Box>
                                </Box>
                            </CardContent>
                        </Card>

                        <Box display="grid" gridTemplateColumns={{ xs: '1fr', md: '2fr 1fr' }} gap={0.5} sx={{ flex: 1, minHeight: 0 }}>
                            <Stack spacing={0.5}>
                                {/* Billing Details Card - Compact */}
                                <Card elevation={1} sx={{ borderRadius: 1 }}>
                                    <CardContent sx={{ p: 1, py: 0.75 }}>
                                        <Typography variant="subtitle1" fontWeight={600} sx={{ fontSize: '0.95rem', mb: 0.5, display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                            <ReceiptIcon color="primary" sx={{ fontSize: 18 }} /> Billing Breakdown
                                        </Typography>
                                        <Box sx={{ fontSize: '0.85rem' }}>
                                            <Box display="flex" justifyContent="space-between" py={0.25} borderBottom="1px solid #f0f0f0">
                                                <Typography color="textSecondary" sx={{ fontSize: '0.85rem' }}>Base Monthly Fee</Typography>
                                                <Typography fontWeight={600} sx={{ fontSize: '0.85rem' }}><NumberField source="base_amount" options={{ style: 'currency', currency: 'USD' }} /></Typography>
                                            </Box>
                                            <Box display="flex" justifyContent="space-between" py={0.25} borderBottom="1px solid #f0f0f0">
                                                <Typography color="textSecondary" sx={{ fontSize: '0.85rem' }}>Data Usage</Typography>
                                                <Typography fontWeight={600} sx={{ fontSize: '0.85rem' }}><NumberField source="usage_gb" options={{ maximumFractionDigits: 2 }} /> GB</Typography>
                                            </Box>
                                            <Box display="flex" justifyContent="space-between" py={0.25} borderBottom="1px solid #f0f0f0">
                                                <Typography color="textSecondary" sx={{ fontSize: '0.85rem' }}>Price per GB</Typography>
                                                <Typography fontWeight={600} sx={{ fontSize: '0.85rem' }}><NumberField source="price_per_gb" options={{ style: 'currency', currency: 'USD' }} /></Typography>
                                            </Box>
                                            <Box display="flex" justifyContent="space-between" py={0.25} borderBottom="1px solid #f0f0f0" bgcolor="rgba(59, 130, 246, 0.04)" px={0.5}>
                                                <Typography fontWeight={600} color="primary" sx={{ fontSize: '0.85rem' }}>Usage Charge</Typography>
                                                <FunctionField render={(record) => {
                                                    const consumption = (record.usage_gb || 0) * (record.price_per_gb || 0);
                                                    return <Typography fontWeight={700} color="primary" sx={{ fontSize: '0.85rem' }}>{consumption.toLocaleString('en-US', { style: 'currency', currency: 'USD' })}</Typography>;
                                                }} />
                                            </Box>
                                            <Box display="flex" justifyContent="space-between" py={0.5}>
                                                <Typography fontWeight={700} sx={{ fontSize: '0.95rem' }}>Total</Typography>
                                                <Typography fontWeight={800} color="primary.main" sx={{ fontSize: '0.95rem' }}>
                                                    <NumberField source="amount" options={{ style: 'currency', currency: 'USD' }} />
                                                </Typography>
                                            </Box>
                                        </Box>
                                    </CardContent>
                                </Card>

                                {/* Usage Statistics - Compact */}
                                <Card elevation={1} sx={{ borderRadius: 1 }}>
                                    <CardContent sx={{ p: 1, py: 0.75 }}>
                                        <Typography variant="subtitle1" fontWeight={600} sx={{ fontSize: '0.95rem', mb: 0.5, display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                            <PaymentsIcon color="primary" sx={{ fontSize: 18 }} /> Usage Stats
                                        </Typography>
                                        <Box display="grid" gridTemplateColumns="1fr 1fr" gap={0.5}>
                                            <Box px={0.5} py={0.25} bgcolor="#f8fafc" borderRadius={1}>
                                                <Typography variant="caption" color="textSecondary" display="block" sx={{ fontSize: '0.75rem' }}>Sessions</Typography>
                                                <Typography fontWeight={700} sx={{ fontSize: '0.9rem' }}><NumberField source="session_count" /></Typography>
                                            </Box>
                                            <Box px={0.5} py={0.25} bgcolor="#f8fafc" borderRadius={1}>
                                                <Typography variant="caption" color="textSecondary" display="block" sx={{ fontSize: '0.75rem' }}>Data Used</Typography>
                                                <Typography fontWeight={700} sx={{ fontSize: '0.9rem' }}><NumberField source="usage_gb" options={{ maximumFractionDigits: 2 }} /> GB</Typography>
                                            </Box>
                                        </Box>
                                    </CardContent>
                                </Card>
                            </Stack>

                            <Stack spacing={0.5}>
                                {/* Dates and References - Compact */}
                                <Card elevation={1} sx={{ borderRadius: 1 }}>
                                    <CardContent sx={{ p: 1, py: 0.75 }}>
                                        <Typography variant="body2" color="textSecondary" sx={{ fontSize: '0.8rem' }}>Invoice Ref</Typography>
                                        <Typography fontWeight={600} sx={{ fontSize: '0.95rem' }} gutterBottom>#<TextField source="id" /></Typography>

                                        <Box mt={0.25}>
                                            <Typography variant="caption" color="textSecondary" display="block" sx={{ fontSize: '0.75rem' }}>Issue Date</Typography>
                                            <Typography sx={{ fontSize: '0.85rem' }} fontWeight={500}><DateField source="issue_date" showTime /></Typography>
                                        </Box>

                                        <Box mt={0.25} p={0.5} borderRadius={1} bgcolor={alpha('#ef4444', 0.05)} border={`1px solid ${alpha('#ef4444', 0.1)}`}>
                                            <Typography variant="caption" color="error" display="block" fontWeight={600} sx={{ fontSize: '0.75rem' }}>Due Date</Typography>
                                            <Typography sx={{ fontSize: '0.85rem' }} fontWeight={700} color="error.main"><DateField source="due_date" /></Typography>
                                        </Box>

                                        <Box mt={0.25}>
                                            <Typography variant="caption" color="textSecondary" display="block" sx={{ fontSize: '0.75rem' }}>Billing Period</Typography>
                                            <Typography sx={{ fontSize: '0.85rem' }} fontWeight={500}>
                                                <DateField source="billing_period_start" /> - <DateField source="billing_period_end" />
                                            </Typography>
                                        </Box>

                                        {useRecordContext()?.paid_at && (
                                            <Box mt={0.25} p={0.5} borderRadius={1} bgcolor={alpha('#10b981', 0.05)} border={`1px solid ${alpha('#10b981', 0.1)}`}>
                                                <Typography variant="caption" color="success.main" display="block" fontWeight={600} sx={{ fontSize: '0.75rem' }}>Paid On</Typography>
                                                <Typography sx={{ fontSize: '0.85rem' }} fontWeight={700} color="success.main"><DateField source="paid_at" showTime /></Typography>
                                            </Box>
                                        )}
                                    </CardContent>
                                </Card>

                                {/* Actions Card - No Print */}
                                <Card elevation={1} sx={{ borderRadius: 1, bgcolor: '#f1f5f9' }} className="no-print">
                                    <CardContent sx={{ p: 1 }}>
                                        <Typography variant="subtitle2" gutterBottom sx={{ fontSize: '0.85rem' }}>Actions</Typography>
                                        <Box display="flex" flexDirection="column" gap={0.5}>
                                            <PayButton />
                                            <Button
                                                label="Print"
                                                onClick={() => window.print()}
                                                variant="outlined"
                                                color="primary"
                                                startIcon={<PrintIcon />}
                                            />
                                            <WhatsAppShareButton />
                                        </Box>
                                    </CardContent>
                                </Card>
                            </Stack>
                        </Box>

                        {/* Remark Section - No Print */}
                        <Card elevation={0} sx={{ borderRadius: 1, bgcolor: '#f8fafc', border: '1px dashed #cbd5e1', flexShrink: 0 }} className="no-print">
                            <CardContent sx={{ p: 1, py: 0.75 }}>
                                <Typography variant="subtitle2" color="textSecondary" sx={{ fontSize: '0.8rem' }}>
                                    <NoteIcon sx={{ fontSize: 14, verticalAlign: 'middle', mr: 0.5 }} /> Remarks
                                </Typography>
                                <Typography sx={{ fontSize: '0.8rem' }}>
                                    <TextField source="remark" emptyText="No remarks" />
                                </Typography>
                            </CardContent>
                        </Card>
                    </Stack>
                </Box>
            </SimpleShowLayout>
        </Show>
    );
};
