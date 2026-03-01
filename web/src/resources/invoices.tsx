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
    ListButton
} from 'react-admin';
import { Box, Card, CardContent, Typography, Chip, Stack } from '@mui/material';
import PaymentsIcon from '@mui/icons-material/Payments';
import ReceiptIcon from '@mui/icons-material/Receipt';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { httpClient } from '../utils/apiClient';

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

export const InvoiceList = () => (
    <List sort={{ field: 'id', order: 'DESC' }}>
        <Datagrid rowClick="show">
            <TextField source="id" />
            <TextField source="username" />
            <NumberField source="amount" />
            <TextField source="currency" />
            <StatusChip label="Status" />
            <DateField source="issue_date" />
            <DateField source="due_date" />
            <PayButton />
        </Datagrid>
    </List>
);

const InvoiceShowActions = () => (
    <TopToolbar>
        <ListButton icon={<ArrowBackIcon />} />
    </TopToolbar>
);

export const InvoiceShow = () => (
    <Show actions={<InvoiceShowActions />}>
        <SimpleShowLayout>
            <Box sx={{ p: 2 }}>
                <Stack spacing={3}>
                    <Card elevation={3} sx={{ borderRadius: 2 }}>
                        <CardContent>
                            <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                                <Box display="flex" alignItems="center" gap={1}>
                                    <ReceiptIcon color="primary" />
                                    <Typography variant="h6">Invoice Detail</Typography>
                                </Box>
                                <StatusChip />
                            </Box>

                            <Box display="grid" gridTemplateColumns="1fr 1fr" gap={2}>
                                <Box>
                                    <Typography variant="caption" color="textSecondary">Invoice ID</Typography>
                                    <Typography variant="body1" fontWeight={600}><TextField source="id" /></Typography>
                                </Box>
                                <Box>
                                    <Typography variant="caption" color="textSecondary">Username</Typography>
                                    <Typography variant="body1"><TextField source="username" /></Typography>
                                </Box>
                                <Box>
                                    <Typography variant="caption" color="textSecondary">Amount</Typography>
                                    <Typography variant="body1" color="primary.main" fontWeight={700}>
                                        <NumberField source="amount" options={{ style: 'currency', currency: 'USD' }} />
                                    </Typography>
                                </Box>
                                <Box>
                                    <Typography variant="caption" color="textSecondary">Issue Date</Typography>
                                    <Typography variant="body1"><DateField source="issue_date" /></Typography>
                                </Box>
                                <Box>
                                    <Typography variant="caption" color="textSecondary">Due Date</Typography>
                                    <Typography variant="body1"><DateField source="due_date" /></Typography>
                                </Box>
                                <Box>
                                    <Typography variant="caption" color="textSecondary">Billing Period</Typography>
                                    <Typography variant="body2">
                                        <DateField source="billing_period_start" /> - <DateField source="billing_period_end" />
                                    </Typography>
                                </Box>
                            </Box>

                            <Box mt={4} display="flex" justifyContent="flex-end">
                                <PayButton />
                            </Box>
                        </CardContent>
                    </Card>

                    <Card elevation={1} sx={{ bgcolor: 'grey.50', borderRadius: 2 }}>
                        <CardContent>
                            <Typography variant="subtitle2" gutterBottom>Remark</Typography>
                            <Typography variant="body2" color="textSecondary">
                                <TextField source="remark" emptyText="No remarks" />
                            </Typography>
                        </CardContent>
                    </Card>
                </Stack>
            </Box>
        </SimpleShowLayout>
    </Show>
);
