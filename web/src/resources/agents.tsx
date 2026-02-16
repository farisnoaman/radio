import React, { useState } from 'react';
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
    useDataProvider,
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
    const dataProvider = useDataProvider();

    const handleOpen = () => setOpen(true);
    const handleClose = () => setOpen(false);

    const handleSubmit = async () => {
        if (!record) return;
        try {
            await dataProvider.post(`agents/${record.id}/topup`, {
                amount: parseFloat(amount),
                remark,
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

// Custom field to fetch and display wallet balance
const WalletBalanceField = () => {
    const record = useRecordContext();
    const [balance, setBalance] = useState<number | null>(null);
    const dataProvider = useDataProvider();

    React.useEffect(() => {
        if (record && record.id) {
            dataProvider.getOne('agents', { id: record.id + '/wallet' })
                .then(({ data }) => setBalance(data.balance))
                .catch(() => setBalance(0));
        }
    }, [record, dataProvider]);

    if (balance === null) return <span>...</span>;
    return <span>{balance.toFixed(2)}</span>;
};


export const AgentList = (props: ListProps) => (
    <List {...props} actions={<AgentListActions />}>
        <Datagrid>
            <TextField source="id" />
            <TextField source="username" />
            <TextField source="realname" />
            <EmailField source="mobile" />
            {/* Display Wallet Balance via custom component or ensure API returns it */}
            <FunctionField label="Balance" render={() => <WalletBalanceField />} />
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
