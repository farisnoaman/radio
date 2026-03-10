import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useMediaQuery, Theme, Box, Card, CardContent, CardActions, TextField as MuiTextField, Button as MuiButton, IconButton, InputAdornment, Dialog, DialogTitle, DialogContent, DialogActions, Typography, Chip } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import ClearIcon from '@mui/icons-material/Clear';
import RedeemIcon from '@mui/icons-material/Redeem';
import VisibilityIcon from '@mui/icons-material/Visibility';
import DownloadIcon from '@mui/icons-material/Download';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import ToggleOnIcon from '@mui/icons-material/ToggleOn';
import ToggleOffIcon from '@mui/icons-material/ToggleOff';
import DeleteIcon from '@mui/icons-material/Delete';
import RestoreIcon from '@mui/icons-material/Restore';
import CurrencyExchangeIcon from '@mui/icons-material/CurrencyExchange';
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
    FunctionField,
    ReferenceField,
    Button,
    useNotify,
    useRefresh,
    useDataProvider,
    useRecordContext,
    DateTimeInput,
    useGetOne,
    BooleanInput,
    RecordContextProvider,
    useListContext,
    TopToolbar,
    ExportButton,
    SortButton
} from 'react-admin';

import { httpClient } from '../utils/apiClient';
import VoucherPrintDialog from '../components/VoucherPrintDialog';
import VoucherTransferDialog from '../components/VoucherTransferDialog';
import { ServerPagination } from '../components/datagrid/ServerPagination';
import PrintIcon from '@mui/icons-material/Print';
import SwapHorizIcon from '@mui/icons-material/SwapHoriz';

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

    const [printOpen, setPrintOpen] = useState(false);
    const [transferOpen, setTransferOpen] = useState(false);
    const { data: product } = useGetOne('products', { id: record.product_id });

    return (
        <Box display="flex">
            <Button
                label="Voucher List"
                size="small"
                component={Link}
                to={`/vouchers?filter=${JSON.stringify({ batch_id: record.id })}`}
                onClick={(e: any) => e.stopPropagation()}
            >
                <VisibilityIcon />
            </Button>
            {!record.is_deleted && (
                <>
                    <Button label="Download" size="small" onClick={handleDownload}>
                        <DownloadIcon />
                    </Button>
                    <Button label="Print" size="small" onClick={() => setPrintOpen(true)}>
                        <PrintIcon />
                    </Button>
                    {record.activated_at ? (
                        <Button label="Deactivate" size="small" onClick={() => handleAction('deactivate')} color="warning">
                            <ToggleOffIcon />
                        </Button>
                    ) : (
                        <Button label="Activate" size="small" onClick={() => handleAction('activate')} color="primary">
                            <ToggleOnIcon />
                        </Button>
                    )}
                    <Button label="Transfer" size="small" onClick={() => setTransferOpen(true)} color="secondary">
                        <SwapHorizIcon />
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
            {printOpen && (
                <VoucherPrintDialog
                    open={printOpen}
                    onClose={() => setPrintOpen(false)}
                    batchId={record.id}
                    batchName={record.name}
                    productName={product ? product.name : ''}
                    productColor={product ? product.color : '#000000'}
                    productValidity={product ? product.validity_seconds : 0}
                />
            )}
            {transferOpen && (
                <VoucherTransferDialog
                    open={transferOpen}
                    onClose={() => setTransferOpen(false)}
                    batchId={record.id}
                    batchName={record.name}
                />
            )}
        </Box>
    );

};

const StatusField = () => {
    const record = useRecordContext();
    if (!record) return null;
    if (record.is_deleted) {
        return <Chip label="Deleted" color="error" size="small" variant="outlined" />;
    }
    if (!record.activated_at) {
        return <Chip label="Inactive" color="default" size="small" variant="outlined" />;
    }
    return <Chip label="Active" color="success" size="small" variant="filled" />;
};


const VoucherBatchGrid = () => {
    const { data, isLoading } = useListContext();
    if (isLoading || !data) return null;
    return (
        <Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)' }} gap={2} p={2} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}>
            {data.map(record => (
                <RecordContextProvider value={record} key={record.id}>
                    <Card
                        elevation={0}
                        sx={{
                            borderRadius: 3,
                            border: theme => `1px solid ${theme.palette.divider}`,
                            transition: 'box-shadow 0.2s',
                            '&:hover': { boxShadow: 4 }
                        }}
                    >
                        <CardContent sx={{ pb: 1 }}>
                            <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                                <Box>
                                    <Typography variant="subtitle1" component="div" sx={{ fontWeight: 700, lineHeight: 1.2, mb: 0.5 }}>
                                        <TextField source="name" />
                                    </Typography>
                                    <Typography variant="caption" color="text.secondary">
                                        BATCH ID: {record.id}
                                    </Typography>
                                </Box>
                                <StatusField />
                            </Box>

                            <Box sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.03)', p: 1.5, borderRadius: 2, mb: 2 }}>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Product:</span>
                                    <strong style={{ textAlign: 'right' }}><ReferenceField source="product_id" reference="products"><TextField source="name" /></ReferenceField></strong>
                                </Typography>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Count:</span>
                                    <strong style={{ textAlign: 'right', fontSize: '1.1em' }}><TextField source="count" /></strong>
                                </Typography>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Agent:</span>
                                    <strong style={{ textAlign: 'right' }}><ReferenceField source="agent_id" reference="agents" emptyText="System"><TextField source="realname" /></ReferenceField></strong>
                                </Typography>
                                <Typography variant="caption" sx={{ display: 'flex', justifyContent: 'space-between', color: 'error.main' }}>
                                    <span>Expires:</span>
                                    <DateField source="expire_time" showTime />
                                </Typography>
                            </Box>
                        </CardContent>
                        <CardActions sx={{ justifyContent: 'flex-start', borderTop: theme => `1px solid ${theme.palette.divider}`, px: 2, py: 1.5, flexWrap: 'wrap', gap: 1 }}>
                            <BatchActions />
                        </CardActions>
                    </Card>
                </RecordContextProvider>
            ))}
        </Box>
    );
};
export const VoucherBatchList = (props: ListProps) => {
    const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
    return (
        <List {...props} sort={{ field: 'id', order: 'DESC' }}>
            {isSmall ? (
                <VoucherBatchGrid />
            ) : (
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
            )}
        </List>
    );
};

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

            <Box mt={2}>
                <Typography variant="h6" gutterBottom>Advanced Options</Typography>
                <Box display={{ xs: 'block', sm: 'flex', width: '100%' }}>
                    <Box flex={1} mr={{ xs: 0, sm: '0.5em' }}>
                        <BooleanInput source="generate_pin" label="Generate PIN for Vouchers" defaultValue={false} />
                    </Box>
                    <Box flex={1} ml={{ xs: 0, sm: '0.5em' }}>
                        <SelectInput source="expiration_type" choices={[
                            { id: 'fixed', name: 'Fixed (From creation)' },
                            { id: 'first_use', name: 'First-Use (From activation)' },
                        ]} defaultValue="fixed" fullWidth />
                    </Box>
                </Box>

                <Box display={{ xs: 'block', sm: 'flex', width: '100%' }}>
                    <Box flex={1} mr={{ xs: 0, sm: '0.5em' }}>
                        {useWatch({ control, name: 'generate_pin' }) && (
                            <NumberInput source="pin_length" label="PIN Length" defaultValue={4} min={4} max={8} fullWidth />
                        )}
                    </Box>
                    <Box flex={1} ml={{ xs: 0, sm: '0.5em' }}>
                        {useWatch({ control, name: 'expiration_type' }) === 'first_use' && (
                            <NumberInput source="validity_days" label="Validity Days" defaultValue={30} min={1} fullWidth />
                        )}
                    </Box>
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

import VoucherExtensionDialog from '../components/VoucherExtensionDialog';
import UpdateIcon from '@mui/icons-material/Update';

const ExtendButton = () => {
    const record = useRecordContext();
    const [open, setOpen] = useState(false);

    if (!record || (record.status !== 'used' && record.status !== 'expired')) return null;

    const handleOpen = (e: any) => {
        e.stopPropagation();
        setOpen(true);
    };

    return (
        <>
            <Button label="Extend" onClick={handleOpen} size="small" color="secondary">
                <UpdateIcon />
            </Button>
            {open && (
                <VoucherExtensionDialog
                    open={open}
                    onClose={() => setOpen(false)}
                    voucherCode={record.code}
                    currentExpiry={record.expire_time}
                />
            )}
        </>
    );
};


const VoucherGrid = () => {
    const { data, isLoading } = useListContext();
    if (isLoading || !data) return null;
    return (
        <Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)', lg: 'repeat(4, 1fr)' }} gap={1} p={1} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}>
            {data.map(record => (
                <RecordContextProvider value={record} key={record.id}>
                    <Card
                        elevation={0}
                        sx={{
                            borderRadius: 2,
                            border: theme => `1px solid ${theme.palette.divider}`,
                            transition: 'box-shadow 0.2s',
                            '&:hover': { boxShadow: 2 },
                            position: 'relative',
                            overflow: 'hidden'
                        }}
                    >
                        <Box sx={{ position: 'absolute', left: 0, top: 0, bottom: 0, width: 3, bgcolor: record.status === 'unused' ? 'success.main' : record.status === 'used' ? 'error.main' : 'warning.main' }} />

                        <CardContent sx={{ pb: 1, pl: 2, pt: 1.5 }}>
                            <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                                <Typography variant="body2" component="div" sx={{ fontFamily: 'monospace', fontWeight: 600, letterSpacing: 0.5 }}>
                                    <TextField source="code" />
                                </Typography>
                                <Chip
                                    label={record.status.toUpperCase()}
                                    size="small"
                                    color={record.status === 'unused' ? 'success' : record.status === 'used' ? 'error' : 'default'}
                                    variant={record.status === 'unused' ? 'filled' : 'outlined'}
                                    sx={{ fontWeight: 'bold', fontSize: '0.65rem', height: 20 }}
                                />
                            </Box>

                            <Box sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.03)', p: 1, borderRadius: 1, mb: 1 }}>
                                <Box display="flex" justifyContent="space-between" mb={0.5}>
                                    <Typography variant="caption" sx={{ color: 'text.secondary' }}>Batch:</Typography>
                                    <Typography variant="caption" noWrap sx={{ maxWidth: 80 }}>
                                        <ReferenceField source="batch_id" reference="voucher-batches"><TextField source="name" /></ReferenceField>
                                    </Typography>
                                </Box>
                                <Box display="flex" justifyContent="space-between" mb={0.5}>
                                    <Typography variant="caption" sx={{ color: 'text.secondary' }}>Price:</Typography>
                                    <Typography variant="caption" sx={{ color: 'success.main', fontWeight: 600 }}>$<TextField source="price" /></Typography>
                                </Box>
                                <Box display="flex" justifyContent="space-between">
                                    <Typography variant="caption" sx={{ color: 'text.secondary' }}>Exp:</Typography>
                                    <Typography variant="caption"><DateField source="expire_time" /></Typography>
                                </Box>
                            </Box>
                        </CardContent>
                        <CardActions sx={{ justifyContent: 'flex-end', borderTop: theme => `1px solid ${theme.palette.divider}`, px: 1, py: 0.5 }}>
                            <RedeemButton />
                            <ExtendButton />
                        </CardActions>
                    </Card>
                </RecordContextProvider>
            ))}
        </Box>
    );
};
const VoucherListActions = () => {
    // VoucherListActions - toolbar for voucher list page

    const isMobile = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
    const { filterValues, setFilters, displayedFilters } = useListContext();
    const [searchInput, setSearchInput] = useState(filterValues?.sn || '');
    const [dialogOpen, setDialogOpen] = useState(false);

    const handleSearch = () => {
        if (searchInput.trim() === '') {
            const newFilters = { ...filterValues };
            delete newFilters.sn;
            setFilters(newFilters, displayedFilters);
        } else {
            setFilters({ ...filterValues, sn: searchInput.trim() }, displayedFilters);
        }
        setDialogOpen(false);
    };

    const handleClear = () => {
        setSearchInput('');
        const newFilters = { ...filterValues };
        delete newFilters.sn;
        setFilters(newFilters, displayedFilters);
        setDialogOpen(false);
    };

    return (
        <TopToolbar sx={{ flexWrap: 'nowrap', gap: 1, overflowX: 'auto' }}>
            <Button
                label="Batches"
                component={Link}
                to="/voucher-batches"
                size="small"
            >
                <ArrowBackIcon />
            </Button>
            {isMobile && (
                <>
                    <MuiButton
                        variant="outlined"
                        color="primary"
                        size="small"
                        onClick={() => setDialogOpen(true)}
                        startIcon={<SearchIcon />}
                    >
                        {filterValues?.sn ? `Search: ${filterValues.sn}` : 'Search'}
                    </MuiButton>
                    <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} fullWidth maxWidth="sm">
                        <DialogTitle>Search Vouchers</DialogTitle>
                        <DialogContent>
                            <Box display="flex" flexDirection="column" gap={2} pt={1}>
                                <MuiTextField
                                    fullWidth
                                    label="Serial Number, ID, or Code"
                                    value={searchInput}
                                    onChange={(e) => setSearchInput(e.target.value)}
                                    onKeyPress={(e: any) => e.key === 'Enter' && handleSearch()}
                                    placeholder="e.g. 3-150, 123, or ABC123"
                                    autoFocus
                                />
                                <Box display="flex" gap={1} justifyContent="flex-end">
                                    <MuiButton onClick={handleClear} disabled={!filterValues?.sn}>
                                        Clear
                                    </MuiButton>
                                    <MuiButton variant="contained" onClick={handleSearch}>
                                        Search
                                    </MuiButton>
                                </Box>
                            </Box>
                        </DialogContent>
                    </Dialog>
                </>
            )}
            <SortButton fields={['id', 'created_at', 'expire_time', 'status']} />
            <ExportButton />
        </TopToolbar>
    );
};
const VoucherSearchFilters = () => {
    const { filterValues, setFilters, displayedFilters } = useListContext();
    const [searchInput, setSearchInput] = useState(filterValues?.sn || '');
    const isMobile = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
    const [dialogOpen, setDialogOpen] = useState(false);

    if (isMobile) return null;

    const handleSearch = () => {
        if (searchInput.trim() === '') {
            const newFilters = { ...filterValues };
            delete newFilters.sn;
            setFilters(newFilters, displayedFilters);
        } else {
            setFilters({ ...filterValues, sn: searchInput.trim() }, displayedFilters);
        }
        setDialogOpen(false);
    };

    const handleClear = () => {
        setSearchInput('');
        const newFilters = { ...filterValues };
        delete newFilters.sn;
        setFilters(newFilters, displayedFilters);
        setDialogOpen(false);
    };

    const searchContent = (
        <Box display="flex" gap={1} alignItems="center" flex={isMobile ? 1 : 'none'}>
            <MuiTextField
                size="small"
                label="SN / ID / Code"
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                onKeyPress={(e: any) => e.key === 'Enter' && handleSearch()}
                placeholder="e.g. 3-150 or ABC123"
                sx={{ minWidth: isMobile ? 120 : 200 }}
                InputProps={{
                    endAdornment: searchInput && (
                        <InputAdornment position="end">
                            <IconButton size="small" onClick={() => setSearchInput('')}>
                                <ClearIcon fontSize="small" />
                            </IconButton>
                        </InputAdornment>
                    ),
                }}
            />
            <MuiButton variant="contained" size="small" onClick={handleSearch}>
                Search
            </MuiButton>
            {filterValues?.sn && (
                <MuiButton size="small" onClick={handleClear}>
                    Clear
                </MuiButton>
            )}
        </Box>
    );

    if (isMobile) {
        return (
            <>
                <IconButton onClick={() => setDialogOpen(true)} color="primary">
                    <SearchIcon />
                </IconButton>
                <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} fullWidth maxWidth="sm">
                    <DialogTitle>Search Vouchers</DialogTitle>
                    <DialogContent>
                        <Box display="flex" flexDirection="column" gap={2} pt={1}>
                            <MuiTextField
                                fullWidth
                                label="Serial Number, ID, or Code"
                                value={searchInput}
                                onChange={(e) => setSearchInput(e.target.value)}
                                onKeyPress={(e: any) => e.key === 'Enter' && handleSearch()}
                                placeholder="e.g. 3-150, 123, or ABC123"
                                autoFocus
                            />
                            <Box display="flex" gap={1} justifyContent="flex-end">
                                <MuiButton onClick={handleClear} disabled={!filterValues?.sn}>
                                    Clear
                                </MuiButton>
                                <MuiButton variant="contained" onClick={handleSearch}>
                                    Search
                                </MuiButton>
                            </Box>
                        </Box>
                    </DialogContent>
                </Dialog>
            </>
        );
    }

    return (
        <Card sx={{ mb: 2 }}>
            <CardContent sx={{ py: 1.5 }}>
                {searchContent}
            </CardContent>
        </Card>
    );
};
export const VoucherList = (props: ListProps) => {
    const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
    return (
        <List {...props} sort={{ field: 'id', order: 'DESC' }} perPage={50} pagination={<ServerPagination />} actions={<VoucherListActions />} filters={<VoucherSearchFilters />}>
            {isSmall ? (
                <VoucherGrid />
            ) : (
                <Datagrid>
                    <TextField source="id" />
                    <TextField source="code" />
                    <TextField source="status" />
                    <ReferenceField source="batch_id" reference="voucher-batches">
                        <TextField source="name" />
                    </ReferenceField>
                    <TextField source="price" />
                    <FunctionField label="PIN" render={(record: any) => record.require_pin ? (record.pin_view ? record.pin : '****') : 'N/A'} />
                    <RedeemButton />
                    <ExtendButton />
                    <DateField source="expire_time" showTime />
                    <DateField source="created_at" showTime />
                </Datagrid>
            )}
        </List>
    );
};

