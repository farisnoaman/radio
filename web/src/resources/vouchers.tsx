import React, { useState } from 'react';
import {
    List,
    Datagrid,
    TextField,
    DateField,
    Create,
    SimpleForm,
    TextInput,
    NumberInput,
    SelectInput,
    ReferenceInput,
    required,
    ListProps,
    CreateProps,
    ReferenceField,
    Button,
    useNotify,
    useRefresh,
    useDataProvider,
    useRecordContext,
    DateTimeInput,
    useGetOne,
} from 'react-admin';
import { Box, Dialog, DialogTitle, DialogContent, DialogActions, /* TextField as MuiTextField, */ Button as MuiButton } from '@mui/material';
import RedeemIcon from '@mui/icons-material/Redeem';
import VisibilityIcon from '@mui/icons-material/Visibility';
import DownloadIcon from '@mui/icons-material/Download';
import ToggleOnIcon from '@mui/icons-material/ToggleOn';
import ToggleOffIcon from '@mui/icons-material/ToggleOff';
import DeleteIcon from '@mui/icons-material/Delete';
import RestoreIcon from '@mui/icons-material/Restore';
import CurrencyExchangeIcon from '@mui/icons-material/CurrencyExchange';
import { Link } from 'react-router-dom';

import { httpClient } from '../utils/apiClient';

// --- Voucher Batch ---

const BatchActions = () => {
    const record = useRecordContext();
    const notify = useNotify();
    const refresh = useRefresh();
    const dataProvider = useDataProvider();

    if (!record) return null;

    const handleAction = async (action: string) => {
        try {
            await httpClient(`/voucher-batches/${record.id}/${action}`, { method: 'POST' });
            notify(`Batch ${action}ed successfully`, { type: 'success' });
            refresh();
        } catch (error: any) {
            const msg = error?.json?.msg || error?.message || `Failed to ${action} batch`;
            notify(msg, { type: 'error' });
        }
    };

    const handleDelete = async () => {
        if (window.confirm('Are you sure you want to delete this batch and all its vouchers?')) {
            try {
                await dataProvider.delete('voucher-batches', { id: record.id });
                notify('Batch deleted successfully', { type: 'success' });
                refresh();
            } catch (error) {
                notify('Failed to delete batch', { type: 'error' });
            }
        }
    };

    const handleDownload = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`/api/v1/voucher-batches/${record.id}/export`, {
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });

            if (!response.ok) {
                throw new Error('Export failed');
            }

            // Try to extract filename from Content-Disposition header
            const contentDisposition = response.headers.get('Content-Disposition');
            let filename = `voucher_batch_${record.id}.csv`;
            if (contentDisposition) {
                const filenameMatch = contentDisposition.match(/filename="?([^"]+)"?/);
                if (filenameMatch && filenameMatch.length === 2) {
                    filename = filenameMatch[1];
                }
            }

            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
            notify('Export successful', { type: 'success' });
        } catch (error) {
            notify('Failed to export batch', { type: 'error' });
        }
    };

    const userStr = localStorage.getItem('user');
    const user = userStr ? JSON.parse(userStr) : null;
    const isAdmin = user && user.level !== 'agent';

    return (
        <Box display="flex">
            <Button
                label="Voucher List"
                size="small"
                component={Link}
                to={`/vouchers?filter=${JSON.stringify({ batch_id: record.id })}`}
                onClick={(e) => e.stopPropagation()}
            >
                <VisibilityIcon />
            </Button>
            {!record.is_deleted && (
                <>
                    <Button label="Download" size="small" onClick={handleDownload}>
                        <DownloadIcon />
                    </Button>
                    <Button label="Activate" size="small" onClick={() => handleAction('activate')} color="primary">
                        <ToggleOnIcon />
                    </Button>
                    <Button label="Deactivate" size="small" onClick={() => handleAction('deactivate')} color="warning">
                        <ToggleOffIcon />
                    </Button>
                    <Button label="Delete" size="small" onClick={handleDelete} color="error">
                        <DeleteIcon />
                    </Button>
                </>
            )}
            {record.is_deleted && isAdmin && (
                <>
                    <Button label="Restore" size="small" onClick={() => handleAction('restore')} color="primary">
                        <RestoreIcon />
                    </Button>
                    {record.agent_id && record.agent_id !== "0" && (
                        <Button label="Refund Unused" size="small" onClick={() => handleAction('refund')} color="success">
                            <CurrencyExchangeIcon />
                        </Button>
                    )}
                </>
            )}
        </Box>
    );
};

import { Chip } from '@mui/material';

const StatusField = () => {
    const record = useRecordContext();
    if (!record) return null;
    if (record.is_deleted) {
        return <Chip label="Deleted" color="error" size="small" variant="outlined" />;
    }
    return <Chip label="Active" color="success" size="small" variant="outlined" />;
};

export const VoucherBatchList = (props: ListProps) => (
    <List {...props} sort={{ field: 'id', order: 'DESC' }}>
        <Datagrid>
            <TextField source="id" />
            <TextField source="name" />
            <ReferenceField source="product_id" reference="products">
                <TextField source="name" />
            </ReferenceField>
            <ReferenceField source="agent_id" reference="agents" emptyText="System">
                <TextField source="realname" />
            </ReferenceField>
            <TextField source="count" />
            <StatusField />
            <DateField source="expire_time" showTime label="Expiry Time" />
            <DateField source="created_at" showTime />
            <BatchActions />
        </Datagrid>
    </List>
);

import { useWatch, useFormContext } from 'react-hook-form';

const VoucherBatchInputs = () => {
    const { setValue, control } = useFormContext();
    const [balance, setBalance] = useState<number | null>(null);
    const [user, setUser] = useState<any>(null);
    const notify = useNotify();

    const productId = useWatch({ control, name: 'product_id' });
    const { data: product } = useGetOne('products', { id: productId }, { enabled: !!productId });

    React.useEffect(() => {
        const userStr = localStorage.getItem('user');
        if (userStr) {
            const u = JSON.parse(userStr);
            setUser(u);
            if (u.level === 'agent') {
                httpClient(`/agents/${u.id}/wallet`)
                    .then(({ json }) => {
                        setBalance(json.data.balance);
                    })
                    .catch(() => {
                        notify('Failed to fetch wallet balance', { type: 'warning' });
                    });
            }
        }
    }, [notify]);

    React.useEffect(() => {
        if (user?.level === 'agent' && product && balance !== null) {
            const price = product.cost_price || product.price || 0;
            if (price > 0) {
                const max = Math.floor(balance / price);
                setValue('count', max);
                notify(`Max affordable vouchers: ${max}`, { type: 'info' });
            }
        }
    }, [product, balance, user, setValue, notify]);

    const effectivePrice = product ? (product.cost_price || product.price || 0) : 0;
    const maxAffordable = (user?.level === 'agent' && effectivePrice > 0 && balance !== null)
        ? Math.floor(balance / effectivePrice)
        : null;

    return (
        <>
            <TextInput source="name" validate={[required()]} fullWidth />
            <ReferenceInput source="product_id" reference="products">
                <SelectInput optionText="name" validate={[required()]} />
            </ReferenceInput>

            {user?.level !== 'agent' && (
                <ReferenceInput source="agent_id" reference="agents">
                    <SelectInput optionText="realname" helperText="Optional: Charge to agent wallet" />
                </ReferenceInput>
            )}

            {user?.level === 'agent' && balance !== null && (
                <Box mb={2} p={1} bgcolor="background.default" borderRadius={1} border="1px solid #e0e0e0">
                    Available Balance: <strong>{balance.toFixed(2)}</strong>
                    {product && (
                        <span> | Agent Cost: <strong>{effectivePrice}</strong> | Max Affordable: <strong>{maxAffordable}</strong></span>
                    )}
                </Box>
            )}

            <Box display={{ xs: 'block', sm: 'flex', width: '100%' }}>
                <Box flex={1} mr={{ xs: 0, sm: '0.5em' }}>
                    <NumberInput
                        source="count"
                        validate={[required()]}
                        min={1}
                        max={maxAffordable || 10000}
                        fullWidth
                        helperText={maxAffordable !== null ? `Max: ${maxAffordable}` : ''}
                    />
                </Box>
                <Box flex={1} ml={{ xs: 0, sm: '0.5em' }}>
                    <TextInput source="prefix" fullWidth />
                </Box>
            </Box>
            <Box display={{ xs: 'block', sm: 'flex', width: '100%' }}>
                <Box flex={1} mr={{ xs: 0, sm: '0.5em' }}>
                    <NumberInput source="length" defaultValue={10} min={6} max={20} fullWidth label="Code Length" />
                </Box>
                <Box flex={1} ml={{ xs: 0, sm: '0.5em' }}>
                    <SelectInput source="type" choices={[
                        { id: 'mixed', name: 'Mixed (A-Z, 0-9)' },
                        { id: 'number', name: 'Numbers Only' },
                        { id: 'alpha', name: 'Letters Only' },
                    ]} defaultValue="mixed" fullWidth />
                </Box>
            </Box>
            <Box display={{ xs: 'block', sm: 'flex', width: '100%' }}>
                <Box flex={1} mr={{ xs: 0, sm: '0.5em' }}>
                    <DateTimeInput source="expire_time" fullWidth label="Voucher Batch Expiry" helperText="Vouchers will not be redeemable after this date" />
                </Box>
                <Box flex={1} ml={{ xs: 0, sm: '0.5em' }}>
                    <TextInput source="remark" multiline fullWidth />
                </Box>
            </Box>
        </>
    );
};

export const VoucherBatchCreate = (props: CreateProps) => (
    <Create {...props}>
        <SimpleForm>
            <VoucherBatchInputs />
        </SimpleForm>
    </Create>
);

// --- Voucher ---

const RedeemButton = () => {
    const record = useRecordContext();
    const [open, setOpen] = useState(false);
    const notify = useNotify();
    const refresh = useRefresh();
    const dataProvider = useDataProvider();

    if (!record || record.status !== 'unused') return null;

    const handleOpen = (e: any) => {
        e.stopPropagation();
        setOpen(true);
    };

    const handleClose = (e: any) => {
        e.stopPropagation();
        setOpen(false);
    };

    const handleRedeem = async (e: any) => {
        e.stopPropagation();
        try {
            await dataProvider.post('vouchers/redeem', {
                code: record.code,
            });
            notify('Voucher redeemed successfully', { type: 'success' });
            setOpen(false);
            refresh();
        } catch (error) {
            notify('Redemption failed', { type: 'error' });
        }
    };

    return (
        <>
            <Button label="Redeem" onClick={handleOpen} size="small">
                <RedeemIcon />
            </Button>
            <Dialog open={open} onClose={handleClose} onClick={(e) => e.stopPropagation()}>
                <DialogTitle>Redeem Voucher</DialogTitle>
                <DialogContent>
                    Are you sure you want to activate voucher <b>{record.code}</b>?
                    <br />
                    This will create a new Radius User with this code as username and password.
                </DialogContent>
                <DialogActions>
                    <MuiButton onClick={handleClose}>Cancel</MuiButton>
                    <MuiButton onClick={handleRedeem} color="primary" variant="contained">
                        Confirm & Activate
                    </MuiButton>
                </DialogActions>
            </Dialog>
        </>
    );
};

export const VoucherList = (props: ListProps) => (
    <List {...props} sort={{ field: 'id', order: 'DESC' }}>
        <Datagrid>
            <TextField source="id" />
            <TextField source="code" />
            <TextField source="status" />
            <ReferenceField source="batch_id" reference="voucher-batches">
                <TextField source="name" />
            </ReferenceField>
            <TextField source="price" />
            <RedeemButton />
            <DateField source="expire_time" showTime />
            <DateField source="created_at" showTime />
        </Datagrid>
    </List>
);
