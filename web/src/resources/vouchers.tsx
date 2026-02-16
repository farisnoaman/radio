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
} from 'react-admin';
import { Box, Dialog, DialogTitle, DialogContent, DialogActions, TextField as MuiTextField, Button as MuiButton } from '@mui/material';
import RedeemIcon from '@mui/icons-material/Redeem';

// --- Voucher Batch ---

export const VoucherBatchList = (props: ListProps) => (
    <List {...props} sort={{ field: 'id', order: 'DESC' }}>
        <Datagrid rowClick="show">
            <TextField source="id" />
            <TextField source="name" />
            <ReferenceField source="product_id" reference="products">
                <TextField source="name" />
            </ReferenceField>
            <ReferenceField source="agent_id" reference="agents" emptyText="System">
                <TextField source="realname" />
            </ReferenceField>
            <TextField source="count" />
            <TextField source="prefix" />
            <DateField source="expire_time" showTime label="Expiry Time" />
            <DateField source="created_at" showTime />
        </Datagrid>
    </List>
);

export const VoucherBatchCreate = (props: CreateProps) => (
    <Create {...props}>
        <SimpleForm>
            <TextInput source="name" validate={[required()]} fullWidth />
            <ReferenceInput source="product_id" reference="products">
                <SelectInput optionText="name" validate={[required()]} />
            </ReferenceInput>
            <ReferenceInput source="agent_id" reference="agents">
                <SelectInput optionText="realname" helperText="Optional: Charge to agent wallet" />
            </ReferenceInput>
            <Box display={{ xs: 'block', sm: 'flex', width: '100%' }}>
                <Box flex={1} mr={{ xs: 0, sm: '0.5em' }}>
                    <NumberInput source="count" validate={[required()]} min={1} max={10000} fullWidth />
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
