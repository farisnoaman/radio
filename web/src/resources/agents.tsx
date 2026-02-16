import { useState } from 'react';
import {
    List,
    Datagrid,
    TextField,
    EmailField,
    Create,
    SimpleForm,
    TextInput,
    PasswordInput,
    required,
    useRecordContext,
    ListProps,
    CreateProps,
    Button,
    useNotify,
    useRefresh,
    TopToolbar,
    CreateButton,
    FunctionField,
} from 'react-admin';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    TextField as MuiTextField,
    Button as MuiButton,
} from '@mui/material';
import AddCardIcon from '@mui/icons-material/AddCard';
import AttachMoneyIcon from '@mui/icons-material/AttachMoney';
import { httpClient } from '../utils/apiClient';

const AgentListActions = () => (
    <TopToolbar>
        <CreateButton />
    </TopToolbar>
);

const TopupButton = () => {
    const record = useRecordContext();
    const [open, setOpen] = useState(false);
    const [amount, setAmount] = useState('');
    const [remark, setRemark] = useState('');
    const notify = useNotify();
    const refresh = useRefresh();

    const handleOpen = () => setOpen(true);
    const handleClose = () => setOpen(false);

    const handleSubmit = async () => {
        if (!record) return;
        try {
            await httpClient(`/agents/${record.id}/topup`, {
                method: 'POST',
                body: JSON.stringify({
                    amount: parseFloat(amount),
                    remark,
                }),
            });
            notify('Topup successful', { type: 'success' });
            setOpen(false);
            refresh();
        } catch (error) {
            notify('Topup failed', { type: 'error' });
        }
    };

    return (
        <>
            <Button label="Topup" onClick={handleOpen} startIcon={<AddCardIcon />}>
                <AttachMoneyIcon />
            </Button>
            <Dialog open={open} onClose={handleClose}>
                <DialogTitle>Topup Agent Wallet</DialogTitle>
                <DialogContent>
                    <MuiTextField
                        autoFocus
                        margin="dense"
                        label="Amount"
                        type="number"
                        fullWidth
                        value={amount}
                        onChange={(e) => setAmount(e.target.value)}
                    />
                    <MuiTextField
                        margin="dense"
                        label="Remark"
                        fullWidth
                        value={remark}
                        onChange={(e) => setRemark(e.target.value)}
                    />
                </DialogContent>
                <DialogActions>
                    <MuiButton onClick={handleClose}>Cancel</MuiButton>
                    <MuiButton onClick={handleSubmit}>Submit</MuiButton>
                </DialogActions>
            </Dialog>
        </>
    );
};

export const AgentList = (props: ListProps) => (
    <List {...props} actions={<AgentListActions />}>
        <Datagrid>
            <TextField source="id" />
            <TextField source="username" />
            <TextField source="realname" />
            <EmailField source="mobile" />
            <FunctionField
                label="Balance"
                render={(record: any) => (record.balance || 0).toFixed(2)}
            />
            <TextField source="status" />
            <TopupButton />
        </Datagrid>
    </List>
);

export const AgentCreate = (props: CreateProps) => (
    <Create {...props}>
        <SimpleForm>
            <TextInput source="username" validate={[required()]} />
            <PasswordInput source="password" validate={[required()]} />
            <TextInput source="realname" validate={[required()]} />
            <TextInput source="mobile" />
            <TextInput source="email" />
            <TextInput source="remark" multiline />
            {/* Hidden field to enforce level='agent' */}
            <TextInput source="level" defaultValue="agent" style={{ display: 'none' }} />
            <TextInput source="status" defaultValue="enabled" style={{ display: 'none' }} />
        </SimpleForm>
    </Create>
);
