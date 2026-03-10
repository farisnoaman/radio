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
import VisibilityIcon from '@mui/icons-material/Visibility';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
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
        <Box 
            display="grid" 
            gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)' }} 
            gap={3} 
            p={3} 
            sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}
        >
            {data.map(record => (
                <RecordContextProvider value={record} key={record.id}>
                    <Card 
                        elevation={2} 
                        sx={{ 
                            borderRadius: 3, 
                            transition: 'all 0.3s ease',
                            '&:hover': { 
                                transform: 'translateY(-4px)',
                                boxShadow: 6 
                            }
                        }}
                    >
                        <CardContent sx={{ pb: 2, pt: 2.5, px: 2.5 }}>
                            {/* Header: Avatar, Name, Status */}
                            <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2.5}>
                                <Box display="flex" alignItems="center" gap={1.5}>
                                    <Avatar 
                                        sx={{ 
                                            bgcolor: 'primary.main', 
                                            width: 48, 
                                            height: 48, 
                                            fontSize: '1.1rem', 
                                            fontWeight: 'bold',
                                            boxShadow: 2
                                        }}
                                    >
                                        {record.username?.charAt(0).toUpperCase()}
                                    </Avatar>
                                    <Box minWidth={0}>
                                        <Typography 
                                            variant="h6" 
                                            component="div" 
                                            sx={{ 
                                                fontWeight: 700, 
                                                lineHeight: 1.2,
                                                fontSize: '1rem',
                                                overflow: 'hidden', 
                                                textOverflow: 'ellipsis', 
                                                whiteSpace: 'nowrap',
                                                maxWidth: 150
                                            }}
                                        >
                                            {record.username}
                                        </Typography>
                                        <Typography variant="caption" color="text.secondary">
                                            ID: #{String(record.id).slice(-8)}
                                        </Typography>
                                    </Box>
                                </Box>
                                <Chip 
                                    label={record.status === 'enabled' ? 'Active' : 'Inactive'} 
                                    size="medium" 
                                    color={record.status === 'enabled' ? 'success' : 'default'}
                                    variant="filled" 
                                    sx={{ 
                                        fontWeight: 'bold',
                                        fontSize: '0.7rem',
                                        height: 28
                                    }}
                                />
                            </Box>
                            
                            {/* Agent Info Section */}
                            <Box 
                                sx={{ 
                                    bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.03)', 
                                    p: 2, 
                                    borderRadius: 2, 
                                    mb: 2 
                                }}
                            >
                                <Box display="flex" alignItems="center" gap={1} mb={1.5}>
                                    <Typography variant="caption" color="text.secondary" sx={{ minWidth: 50 }}>
                                        Name:
                                    </Typography>
                                    <Typography variant="body2" fontWeight="bold">
                                        {record.realname || 'N/A'}
                                    </Typography>
                                </Box>
                                <Box display="flex" alignItems="center" gap={1} mb={1.5}>
                                    <Typography variant="caption" color="text.secondary" sx={{ minWidth: 50 }}>
                                        Contact:
                                    </Typography>
                                    <Typography variant="body2">
                                        {record.mobile || record.email || 'N/A'}
                                    </Typography>
                                </Box>
                                <Box display="flex" alignItems="center" gap={1}>
                                    <Typography variant="caption" color="text.secondary" sx={{ minWidth: 50 }}>
                                        Balance:
                                    </Typography>
                                    <Typography 
                                        variant="body1" 
                                        fontWeight="bold" 
                                        color="primary.main"
                                    >
                                        ${(record.balance || 0).toFixed(2)}
                                    </Typography>
                                </Box>
                            </Box>

                            {/* Statistics Section */}
                            <Box 
                                sx={{ 
                                    bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(33, 150, 243, 0.15)' : 'rgba(33, 150, 243, 0.08)', 
                                    p: 2, 
                                    borderRadius: 2,
                                    mb: 2
                                }}
                            >
                                <Typography 
                                    variant="caption" 
                                    color="primary.main" 
                                    sx={{ 
                                        display: 'block', 
                                        mb: 1.5,
                                        fontWeight: 600,
                                        letterSpacing: 0.5
                                    }}
                                >
                                    STATISTICS
                                </Typography>
                                <Box display="flex" justifyContent="space-between" gap={2}>
                                    <Box textAlign="center" flex={1}>
                                        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5 }}>
                                            Level
                                        </Typography>
                                        <Typography variant="body1" fontWeight="bold" sx={{ fontSize: '1.1rem' }}>
                                            {record.level !== undefined ? record.level : '-'}
                                        </Typography>
                                    </Box>
                                    <Box textAlign="center" flex={1}>
                                        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5 }}>
                                            Tier
                                        </Typography>
                                        <Typography variant="body1" fontWeight="bold" sx={{ fontSize: '1.1rem' }}>
                                            {record.level === 0 ? 'Root' : record.level === 1 ? 'Level 1' : `L${record.level || 0}`}
                                        </Typography>
                                    </Box>
                                    <Box textAlign="center" flex={1}>
                                        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5 }}>
                                            Status
                                        </Typography>
                                        <Typography 
                                            variant="body1" 
                                            fontWeight="bold" 
                                            color={record.status === 'enabled' ? 'success.main' : 'error.main'}
                                            sx={{ fontSize: '1.1rem' }}
                                        >
                                            {record.status === 'enabled' ? 'Active' : 'Off'}
                                        </Typography>
                                    </Box>
                                </Box>
                            </Box>
                        </CardContent>
                        
                        {/* Actions */}
                        <CardActions sx={{ 
                            justifyContent: 'flex-end', 
                            borderTop: theme => `1px solid ${theme.palette.divider}`, 
                            px: 2, 
                            py: 1,
                            gap: 0.5,
                            bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(0,0,0,0.2)' : 'rgba(0,0,0,0.02)'
                        }}>
                            <Tooltip title="View Details">
                                <IconButton 
                                    component={Link} 
                                    to={`/agents/${record.id}/show`}
                                    size="small"
                                    sx={{ 
                                        bgcolor: 'primary.main',
                                        color: 'white',
                                        '&:hover': { bgcolor: 'primary.dark' },
                                        width: 36,
                                        height: 36
                                    }}
                                >
                                    <VisibilityIcon sx={{ fontSize: 20 }} />
                                </IconButton>
                            </Tooltip>
                            <TopupButton />
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
