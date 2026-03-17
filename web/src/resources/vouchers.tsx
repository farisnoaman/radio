import React, { useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';
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
import SettingsIcon from '@mui/icons-material/Settings';
const LinkWrapper = React.forwardRef<HTMLAnchorElement, any>((props, ref) => (
    <RouterLink ref={ref} {...props} />
));

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
    SortButton,
    useTranslate,
    useLocale
} from 'react-admin';
import { useFormContext, useWatch } from 'react-hook-form';

import { httpClient } from '../utils/apiClient';

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
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

    if (!record) return null;

    const handleAction = async (action: string) => {
        try {
            await httpClient(`/voucher-batches/${record.id}/${action}`, { method: 'POST' });
            const actionKey = action === 'activate' ? 'activated' : action === 'deactivate' ? 'deactivated' : action === 'restore' ? 'restored' : 'refunded';
            notify(translate(`resources.voucher-batches.notifications.${actionKey}`), { type: 'success' });
            refresh();
        } catch (error: any) {
            const msg = error?.json?.msg || error?.message || `Failed to ${action} batch`;
            notify(msg, { type: 'error' });
        }
    };

    const handleDelete = async () => {
        if (window.confirm(translate('resources.voucher-batches.notifications.delete_confirm'))) {
            try {
                await dataProvider.delete('voucher-batches', { id: record.id });
                notify(translate('resources.voucher-batches.notifications.deleted'), { type: 'success' });
                refresh();
            } catch (error) {
                notify(translate('resources.voucher-batches.notifications.delete_error'), { type: 'error' });
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
            notify(translate('resources.voucher-batches.notifications.export_success'), { type: 'success' });
        } catch (error) {
            notify(translate('resources.voucher-batches.notifications.export_error'), { type: 'error' });
        }
    };

    const userStr = localStorage.getItem('user');
    const user = userStr ? JSON.parse(userStr) : null;
    const isAdmin = user && user.level !== 'agent';


    const [transferOpen, setTransferOpen] = useState(false);

    return (
        <Box display="flex" gap={0.5} flexWrap="wrap" justifyContent="flex-start">
            <MuiButton
                size="small"
                variant="outlined"
                component={LinkWrapper}
                to={`/vouchers?filter=${JSON.stringify({ batch_id: record.id })}`}
                onClick={(e: any) => e.stopPropagation()}
                startIcon={!isRTL ? <VisibilityIcon /> : undefined}
                endIcon={isRTL ? <VisibilityIcon /> : undefined}
                sx={{ minWidth: 80, justifyContent: 'space-between' }}
            >
                {translate('resources.voucher-batches.actions.view_list')}
            </MuiButton>
            {!record.is_deleted && (
                <>
                    <MuiButton size="small" variant="outlined" onClick={handleDownload} startIcon={!isRTL ? <DownloadIcon /> : undefined} endIcon={isRTL ? <DownloadIcon /> : undefined} sx={{ minWidth: 80, justifyContent: 'space-between' }}>
                        {translate('resources.voucher-batches.actions.download')}
                    </MuiButton>
                    <MuiButton size="small" variant="outlined" component={LinkWrapper} to={`/voucher-printing?batch=${record.id}`} onClick={(e: any) => e.stopPropagation()} startIcon={!isRTL ? <PrintIcon /> : undefined} endIcon={isRTL ? <PrintIcon /> : undefined} sx={{ minWidth: 65, justifyContent: 'space-between' }}>
                        {translate('resources.voucher-batches.actions.print')}
                    </MuiButton>
                    {record.activated_at ? (
                        <MuiButton size="small" variant="outlined" onClick={() => handleAction('deactivate')} color="warning" startIcon={!isRTL ? <ToggleOffIcon /> : undefined} endIcon={isRTL ? <ToggleOffIcon /> : undefined} sx={{ minWidth: 85, justifyContent: 'space-between' }}>
                            {translate('resources.voucher-batches.actions.deactivate')}
                        </MuiButton>
                    ) : (
                        <MuiButton size="small" variant="outlined" onClick={() => handleAction('activate')} color="primary" startIcon={!isRTL ? <ToggleOnIcon /> : undefined} endIcon={isRTL ? <ToggleOnIcon /> : undefined} sx={{ minWidth: 75, justifyContent: 'space-between' }}>
                            {translate('resources.voucher-batches.actions.activate')}
                        </MuiButton>
                    )}
                    <MuiButton size="small" variant="outlined" onClick={() => setTransferOpen(true)} color="secondary" startIcon={!isRTL ? <SwapHorizIcon /> : undefined} endIcon={isRTL ? <SwapHorizIcon /> : undefined} sx={{ minWidth: 70, justifyContent: 'space-between' }}>
                        {translate('resources.voucher-batches.actions.transfer')}
                    </MuiButton>
                    <MuiButton size="small" variant="outlined" onClick={handleDelete} color="error" startIcon={!isRTL ? <DeleteIcon /> : undefined} endIcon={isRTL ? <DeleteIcon /> : undefined} sx={{ minWidth: 60, justifyContent: 'space-between' }}>
                        {translate('resources.voucher-batches.actions.delete')}
                    </MuiButton>
                </>
            )}
            {record.is_deleted && isAdmin && (
                <>
                    <MuiButton size="small" variant="outlined" onClick={() => handleAction('restore')} color="primary" startIcon={!isRTL ? <RestoreIcon /> : undefined} endIcon={isRTL ? <RestoreIcon /> : undefined} sx={{ minWidth: 75, justifyContent: 'space-between' }}>
                        {translate('resources.voucher-batches.actions.restore')}
                    </MuiButton>
                    {record.agent_id && record.agent_id !== "0" && (
                        <MuiButton size="small" variant="outlined" onClick={() => handleAction('refund')} color="success" startIcon={!isRTL ? <CurrencyExchangeIcon /> : undefined} endIcon={isRTL ? <CurrencyExchangeIcon /> : undefined} sx={{ minWidth: 100, justifyContent: 'space-between' }}>
                            {translate('resources.voucher-batches.actions.refund')}
                        </MuiButton>
                    )}
                </>
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
    const translate = useTranslate();
    if (!record) return null;
    if (record.is_deleted) {
        return <Chip label={translate('resources.vouchers.status.deleted')} color="error" size="small" variant="outlined" />;
    }
    if (!record.activated_at) {
        return <Chip label={translate('resources.vouchers.status.inactive')} color="default" size="small" variant="outlined" />;
    }
    return <Chip label={translate('resources.vouchers.status.active')} color="success" size="small" variant="filled" />;
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


const VoucherBatchInputs = () => {
    const { setValue, control } = useFormContext();
    const [balance, setBalance] = useState<number | null>(null);
    const [user, setUser] = useState<any>(null);
    const notify = useNotify();
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

    const textInputProps = { style: { textAlign: isRTL ? 'right' : 'left', direction: isRTL ? 'rtl' : 'ltr' } } as const;
    const numInputProps = { style: { textAlign: isRTL ? 'right' : 'left', direction: isRTL ? 'rtl' : 'ltr' } } as const;

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
        <Box sx={{ width: '100%' }}>
            <Box mb={2}>
                <TextInput 
                    source="name" 
                    validate={[required()]} 
                    fullWidth 
                    size="small"
                    label={translate('pages.voucher.create.name')} 
                    inputProps={textInputProps} 
                    placeholder={translate('pages.voucher.create.name_placeholder')}
                    InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                />
            </Box>

            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: user?.level !== 'agent' ? '1fr 1fr' : '1fr' }, gap: 2, mb: 2, direction: isRTL ? 'rtl' : 'ltr' }}>
                <Box>
                    <ReferenceInput source="product_id" reference="products">
                        <SelectInput 
                            optionText="name" 
                            validate={[required()]} 
                            fullWidth 
                            size="small"
                            label={translate('pages.voucher.create.product')} 
                            InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                        />
                    </ReferenceInput>
                </Box>

                {user?.level !== 'agent' && (
                    <Box>
                        <ReferenceInput source="agent_id" reference="agents">
                            <SelectInput 
                                optionText="realname" 
                                helperText={translate('pages.voucher.create.agent_placeholder')} 
                                fullWidth 
                                size="small"
                                label={translate('pages.voucher.create.agent')} 
                                InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                            />
                        </ReferenceInput>
                    </Box>
                )}
            </Box>

            {user?.level === 'agent' && balance !== null && (
                <Box mb={2} p={1.5} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.03)', borderRadius: 2, border: theme => `1px solid ${theme.palette.divider}`, direction: isRTL ? 'rtl' : 'ltr' }}>
                    <Typography variant="body2" component="div">
                        {translate('pages.voucher.create.wallet.balance')}: <strong>{balance.toFixed(2)}</strong>
                        {product && (
                            <span> | {translate('pages.voucher.create.wallet.cost')}: <strong>{effectivePrice}</strong> | {translate('pages.voucher.create.wallet.max')}: <strong>{maxAffordable}</strong></span>
                        )}
                    </Typography>
                </Box>
            )}

            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, mb: 2, direction: isRTL ? 'rtl' : 'ltr' }}>
                <Box>
                    <NumberInput
                        source="count"
                        validate={[required()]}
                        min={1}
                        max={maxAffordable || 10000}
                        fullWidth
                        size="small"
                        label={translate('pages.voucher.create.count')}
                        helperText={maxAffordable !== null ? `${translate('pages.voucher.create.wallet.max')}: ${maxAffordable}` : translate('pages.voucher.create.count_helper')}
                        inputProps={numInputProps}
                        placeholder="1"
                        InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                    />
                </Box>
                <Box>
                    <TextInput 
                        source="prefix" 
                        fullWidth 
                        size="small"
                        label={translate('pages.voucher.create.prefix')} 
                        placeholder={translate('pages.voucher.create.prefix_placeholder')} 
                        inputProps={textInputProps} 
                        InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                    />
                </Box>
            </Box>

            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, mb: 2, direction: isRTL ? 'rtl' : 'ltr' }}>
                <Box>
                    <NumberInput 
                        source="length" 
                        placeholder="10" 
                        min={6} 
                        max={20} 
                        fullWidth 
                        size="small"
                        label={translate('pages.voucher.create.code_length')} 
                        inputProps={numInputProps} 
                        InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                    />
                </Box>
                <Box>
                    <SelectInput 
                        source="type" 
                        label={translate('pages.voucher.create.code_type')} 
                        choices={[
                            { id: 'mixed', name: translate('pages.voucher.create.code_type_mixed') },
                            { id: 'number', name: translate('pages.voucher.create.code_type_number') },
                            { id: 'alpha', name: translate('pages.voucher.create.code_type_alpha') },
                        ]} 
                        defaultValue="mixed" 
                        fullWidth 
                        size="small"
                        InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                    />
                </Box>
            </Box>

            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, mb: 2, direction: isRTL ? 'rtl' : 'ltr' }}>
                <Box>
                    <DateTimeInput 
                        source="expire_time" 
                        fullWidth 
                        size="small"
                        label={translate('pages.voucher.create.expiry')} 
                        helperText={translate('pages.voucher.create.expiry_helper')} 
                        InputLabelProps={{ shrink: true, sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                    />
                </Box>
                <Box>
                    <TextInput 
                        source="remark" 
                        multiline 
                        fullWidth 
                        size="small"
                        label={translate('pages.voucher.create.remark')} 
                        placeholder={translate('pages.voucher.create.remark_placeholder')} 
                        inputProps={textInputProps} 
                        InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                    />
                </Box>
            </Box>

            <Box mt={3} p={2} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.01)', borderRadius: 2, border: theme => `1px solid ${theme.palette.divider}` }}>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 2, display: 'flex', alignItems: 'center', gap: 1, color: 'primary.main', direction: isRTL ? 'rtl' : 'ltr' }}>
                    <SettingsIcon fontSize="small" />
                    {translate('pages.voucher.create.advanced')}
                </Typography>

                <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, direction: isRTL ? 'rtl' : 'ltr' }}>
                    <Box>
                        <BooleanInput 
                            source="generate_pin" 
                            label={translate('pages.voucher.create.generate_pin')} 
                            defaultValue={false} 
                            sx={{ '& .MuiFormControlLabel-root': { ml: isRTL ? 0 : undefined, mr: isRTL ? '-11px' : undefined, flexFlow: isRTL ? 'row-reverse' : 'row' } }} 
                        />
                    </Box>
                    <Box>
                        <SelectInput 
                            source="expiration_type" 
                            label={translate('pages.voucher.create.expiration_type')} 
                            choices={[
                                { id: 'fixed', name: translate('pages.voucher.create.expiration_fixed') },
                                { id: 'first_use', name: translate('pages.voucher.create.expiration_first_use') },
                            ]} 
                            defaultValue="fixed" 
                            fullWidth 
                            size="small"
                            InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                        />
                    </Box>
                </Box>

                <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, direction: isRTL ? 'rtl' : 'ltr' }}>
                    <Box>
                        {useWatch({ control, name: 'generate_pin' }) && (
                            <NumberInput 
                                source="pin_length" 
                                label={translate('pages.voucher.create.pin_length')} 
                                placeholder="4" 
                                min={4} 
                                max={8} 
                                fullWidth 
                                size="small"
                                inputProps={numInputProps} 
                                InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                            />
                        )}
                    </Box>
                    <Box>
                        {useWatch({ control, name: 'expiration_type' }) === 'first_use' && (
                            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, direction: isRTL ? 'rtl' : 'ltr' }}>
                                <NumberInput 
                                    source="validity_value_virtual" 
                                    label={translate('pages.voucher.create.validity')} 
                                    placeholder="30" 
                                    min={1} 
                                    fullWidth 
                                    size="small"
                                    inputProps={numInputProps} 
                                    InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                                />
                                <SelectInput 
                                    source="validity_unit_virtual" 
                                    label={translate('pages.voucher.create.validity_unit')} 
                                    choices={[
                                        { id: 'minutes', name: translate('common.minutes') },
                                        { id: 'hours', name: translate('common.hours') },
                                        { id: 'days', name: translate('common.days') },
                                    ]} 
                                    defaultValue="days" 
                                    fullWidth 
                                    size="small"
                                    InputLabelProps={{ sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }}
                                />
                            </Box>)}
                    </Box>
                </Box>
            </Box>
        </Box>
    );
};

import { useGetList } from 'react-admin';

export const VoucherBatchCreate = (props: CreateProps) => {
    const translate = useTranslate();
    const { data: latestBatches, isLoading } = useGetList('voucher-batches', {
        pagination: { page: 1, perPage: 1 },
        sort: { field: 'id', order: 'DESC' }
    });

    if (isLoading) return null;

    const nextId = (latestBatches && latestBatches.length > 0) ? latestBatches[0].id + 1 : 1;
    const defaultName = `${translate('pages.voucher.batch.default_name_prefix')}${nextId}`;

    return (
        <Create {...props} record={{ name: defaultName }}>
            <SimpleForm>
                <VoucherBatchInputs />
            </SimpleForm>
        </Create>
    );
};

// --- Voucher ---

const RedeemButton = () => {
    const record = useRecordContext();
    const [open, setOpen] = useState(false);
    const notify = useNotify();
    const refresh = useRefresh();
    const dataProvider = useDataProvider();
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

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
            notify(translate('resources.vouchers.notifications.redeemed'), { type: 'success' });
            setOpen(false);
            refresh();
        } catch (error) {
            notify(translate('resources.vouchers.notifications.redeem_error'), { type: 'error' });
        }
    };

    return (
        <>
            <MuiButton onClick={handleOpen} size="small" startIcon={!isRTL ? <RedeemIcon /> : undefined} endIcon={isRTL ? <RedeemIcon /> : undefined}>
                {translate('resources.vouchers.actions.redeem')}
            </MuiButton>
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
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

    if (!record || (record.status !== 'used' && record.status !== 'expired')) return null;

    const handleOpen = (e: any) => {
        e.stopPropagation();
        setOpen(true);
    };

    return (
        <>
            <MuiButton onClick={handleOpen} size="small" color="secondary" startIcon={!isRTL ? <UpdateIcon /> : undefined} endIcon={isRTL ? <UpdateIcon /> : undefined}>
                {translate('resources.vouchers.actions.extend')}
            </MuiButton>
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
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

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
            <MuiButton
                component={LinkWrapper}
                to="/voucher-batches"
                size="small"
                startIcon={!isRTL ? <ArrowBackIcon /> : undefined}
                endIcon={isRTL ? <ArrowBackIcon /> : undefined}
            >
                {translate('resources.voucher-batches.name')}
            </MuiButton>
            {isMobile && (
                <>
                    <MuiButton
                        variant="outlined"
                        color="primary"
                        size="small"
                        onClick={() => setDialogOpen(true)}
                        startIcon={!isRTL ? <SearchIcon /> : undefined}
                        endIcon={isRTL ? <SearchIcon /> : undefined}
                    >
                        {filterValues?.sn ? `${translate('common.search')}: ${filterValues.sn}` : translate('common.search')}
                    </MuiButton>
                    <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} fullWidth maxWidth="sm">
                        <DialogTitle>{translate('pages.voucher.search.title')}</DialogTitle>
                        <DialogContent>
                            <Box display="flex" flexDirection="column" gap={2} pt={1}>
                                <MuiTextField
                                    fullWidth
                                    label={translate('pages.voucher.search.placeholder')}
                                    value={searchInput}
                                    onChange={(e) => setSearchInput(e.target.value)}
                                    onKeyPress={(e: any) => e.key === 'Enter' && handleSearch()}
                                    placeholder={translate('pages.voucher.search.example')}
                                    autoFocus
                                />
                                <Box display="flex" gap={1} justifyContent="flex-end">
                                    <MuiButton onClick={handleClear} disabled={!filterValues?.sn}>
                                        {translate('pages.voucher.search.clear')}
                                    </MuiButton>
                                    <MuiButton variant="contained" onClick={handleSearch}>
                                        {translate('common.search')}
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
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

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
                label={translate('pages.voucher.search.placeholder_short')}
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                onKeyPress={(e: any) => e.key === 'Enter' && handleSearch()}
                placeholder={translate('pages.voucher.search.example')}
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
            <MuiButton variant="contained" size="small" onClick={handleSearch} startIcon={!isRTL ? <SearchIcon /> : undefined} endIcon={isRTL ? <SearchIcon /> : undefined}>
                {translate('common.search')}
            </MuiButton>
            {filterValues?.sn && (
                <MuiButton size="small" onClick={handleClear}>
                    {translate('pages.voucher.search.clear')}
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
                    <DialogTitle>{translate('pages.voucher.search.title')}</DialogTitle>
                    <DialogContent>
                        <Box display="flex" flexDirection="column" gap={2} pt={1}>
                            <MuiTextField
                                fullWidth
                                label={translate('pages.voucher.search.placeholder')}
                                value={searchInput}
                                onChange={(e) => setSearchInput(e.target.value)}
                                onKeyPress={(e: any) => e.key === 'Enter' && handleSearch()}
                                placeholder={translate('pages.voucher.search.example')}
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

