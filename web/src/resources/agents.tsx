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
    useListContext,
    RecordContextProvider,
    Link,
    Filter,
    Show,
    SimpleShowLayout
} from 'react-admin';
import { Box, Card, CardContent, CardActions, Typography, useMediaQuery, Theme, Avatar, Chip } from '@mui/material';
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

const AgentFilter = (props: any) => (
    <Filter {...props}>
        <TextInput source="q" label="Search" alwaysOn />
        <TextInput source="username" label="Username" />
        <TextInput source="realname" label="Name" />
        <TextInput source="email" label="Email" />
    </Filter>
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


const AgentGrid = () => {
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
                                <Box display="flex" alignItems="center" gap={1.5}>
                                    <Avatar sx={{ bgcolor: 'secondary.main', width: 40, height: 40, fontWeight: 'bold' }}>
                                        {record.username?.charAt(0).toUpperCase()}
                                    </Avatar>
                                    <Box>
                                        <Typography variant="subtitle1" component="div" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
                                            <TextField source="username" />
                                        </Typography>
                                        <Typography variant="caption" color="text.secondary">
                                            ID: {record.id}
                                        </Typography>
                                    </Box>
                                </Box>
                                <Chip 
                                    label={record.status === 'enabled' ? 'Active' : 'Disabled'} 
                                    size="small" 
                                    color={record.status === 'enabled' ? 'success' : 'default'}
                                    variant="outlined" 
                                />
                            </Box>
                            
                            <Box sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.03)', p: 1.5, borderRadius: 2, mb: 2 }}>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Name:</span>
                                    <strong style={{ textAlign: 'right' }}><TextField source="realname" /></strong>
                                </Typography>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Contact:</span>
                                    <strong style={{ textAlign: 'right' }}><EmailField source="mobile" /></strong>
                                </Typography>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                    <span style={{ color: 'text.secondary' }}>Wallet:</span>
                                    <Typography variant="subtitle2" sx={{ fontWeight: 'bold', color: 'primary.main', fontSize: '1.1rem' }}>
                                        <FunctionField render={(r:any) => `$${(r.balance||0).toFixed(2)}`} />
                                    </Typography>
                                </Typography>
                            </Box>
                            {/* Agent Stats Card Section */}
                            <Box sx={{ mt: 2, p: 1.5, bgcolor: 'rgba(33, 150, 243, 0.1)', borderRadius: 2 }}>
                                <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
                                    AGENT STATISTICS
                                </Typography>
                                <Box display="flex" justifyContent="space-between" gap={1}>
                                    <Box textAlign="center" flex={1}>
                                        <Typography variant="caption" color="text.secondary">Level</Typography>
                                        <Typography variant="body2" fontWeight="bold">
                                            {record.level !== undefined ? record.level : 'N/A'}
                                        </Typography>
                                    </Box>
                                    <Box textAlign="center" flex={1}>
                                        <Typography variant="caption" color="text.secondary">Status</Typography>
                                        <Typography variant="body2" fontWeight="bold" color={record.status === 'enabled' ? 'success.main' : 'error.main'}>
                                            {record.status}
                                        </Typography>
                                    </Box>
                                    <Box textAlign="center" flex={1}>
                                        <Typography variant="caption" color="text.secondary">Tier</Typography>
                                        <Typography variant="body2" fontWeight="bold">
                                            {record.level === 0 ? 'Root' : record.level === 1 ? 'Level 1' : 'Level ' + (record.level || 0)}
                                        </Typography>
                                    </Box>
                                </Box>
                            </Box>
                        </CardContent>
                        <CardActions sx={{ justifyContent: 'flex-end', borderTop: theme => `1px solid ${theme.palette.divider}`, px: 2, py: 1.5 }}>
                            <TopupButton />
                            <Button 
                                component={Link} 
                                to={`/agents/${record.id}/show`}
                                label="Hierarchy" 
                                sx={{ textTransform: 'none' }}
                            />
                        </CardActions>
                    </Card>
                </RecordContextProvider>
            ))}
        </Box>
    );
};
export const AgentList = (props: ListProps) => {
    const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
    return (
        <List {...props} actions={<AgentListActions />} filters={<AgentFilter />}>
            {isSmall ? (
                <AgentGrid />
            ) : (
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
            )}
        </List>
    );
};

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

import { AgentHierarchyTree } from './AgentHierarchyTree';

export const AgentShow = (props: any) => (
    <Show {...props}>
        <SimpleShowLayout>
            <Box sx={{ mb: 2 }}>
                <Typography variant="h5" gutterBottom>
                    Agent Details
                </Typography>
            </Box>
            <TextField source="id" label="ID" />
            <TextField source="username" label="Username" />
            <TextField source="realname" label="Name" />
            <EmailField source="email" label="Email" />
            <TextField source="mobile" label="Mobile" />
            <TextField source="status" label="Status" />
            <FunctionField
                label="Wallet Balance"
                render={(record: any) => `$${(record.balance || 0).toFixed(2)}`}
            />
            <AgentHierarchyTree />
        </SimpleShowLayout>
    </Show>
);
