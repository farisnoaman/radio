import { useState } from 'react';
import {
    List,
    Datagrid,
    TextField,
    EmailField,
    Create,
    Edit,
    SimpleForm,
    TextInput,
    PasswordInput,
    SelectInput,
    Toolbar,
    SaveButton,
    DeleteButton,
    required,
    useRecordContext,
    ListProps,
    CreateProps,
    EditProps,
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
import VisibilityIcon from '@mui/icons-material/Visibility';
import EditIcon from '@mui/icons-material/Edit';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import AddCardIcon from '@mui/icons-material/AddCard';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import { useNavigate } from 'react-router-dom';
import { httpClient } from '../utils/apiClient';
import {
    FormSection,
    FieldGrid,
    FieldGridItem,
    formLayoutSx
} from '../components';

const ShowIconButton = () => {
    const record = useRecordContext();

    return (
        <Tooltip title="Show">
            <IconButton
                component={Link}
                to={`/agents/${record?.id}/show`}
                size="small"
                sx={{ color: 'primary.main' }}
            >
                <VisibilityIcon fontSize="small" />
            </IconButton>
        </Tooltip>
    );
};

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
            <Tooltip title="Topup">
                <MuiButton 
                    size="small" 
                    variant="contained"
                    color="success"
                    onClick={(e) => { e.stopPropagation(); handleOpen(); }} 
                    startIcon={<AddCardIcon />}
                    sx={{ textTransform: 'none', minWidth: 0, px: 1 }}
                >
                    Topup
                </MuiButton>
            </Tooltip>
            <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
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

const EditButton = () => {
    const record = useRecordContext();
    const navigate = useNavigate();

    const handleClick = (e: React.MouseEvent) => {
        e.stopPropagation();
        if (record) {
            navigate(`/agents/${record.id}`);
        }
    };

    return (
        <Tooltip title="Edit">
            <MuiButton 
                size="small" 
                variant="outlined"
                color="info"
                onClick={handleClick}
                startIcon={<EditIcon />}
                sx={{ textTransform: 'none', minWidth: 0, px: 1 }}
            >
                Edit
            </MuiButton>
        </Tooltip>
    );
};

const HierarchyButton = () => {
    const record = useRecordContext();
    const [open, setOpen] = useState(false);
    const [parentId, setParentId] = useState<string>('');
    const [commissionRate, setCommissionRate] = useState<string>('5');
    const [territory, setTerritory] = useState<string>('');
    const [availableAgents, setAvailableAgents] = useState<any[]>([]);
    const [loadingAgents, setLoadingAgents] = useState(false);
    const notify = useNotify();
    const refresh = useRefresh();

    const handleOpen = async () => {
        if (!record) return;
        setOpen(true);
        setLoadingAgents(true);
        try {
            // Fetch current hierarchy
            const hierarchyRes = await httpClient(`/agents/${record.id}/hierarchy`, { method: 'GET' });
            const hierarchy = hierarchyRes.json as any;
            if (hierarchy && hierarchy.parent_id) {
                setParentId(String(hierarchy.parent_id));
            }
            if (hierarchy && hierarchy.commission_rate) {
                setCommissionRate(String(hierarchy.commission_rate * 100));
            }
            if (hierarchy && hierarchy.territory) {
                setTerritory(hierarchy.territory);
            }

            // Fetch available parent agents
            const agentsRes = await httpClient(`/agents?filter={"level":"agent"}&perPage=1000`, { method: 'GET' });
            const agentsData = agentsRes.json as any;
            const filtered = (agentsData.data || []).filter((a: any) => a.id !== record.id);
            setAvailableAgents(filtered);
        } catch (error) {
            console.error('Failed to load hierarchy data:', error);
        } finally {
            setLoadingAgents(false);
        }
    };
    const handleClose = () => setOpen(false);

    const handleSubmit = async () => {
        if (!record) return;
        try {
            const parentIdValue = parentId && parentId !== '' ? parentId : null;
            await httpClient(`/agents/${record.id}/assign-parent`, {
                method: 'POST',
                body: JSON.stringify({
                    parent_id: parentIdValue,
                    commission_rate: parseFloat(commissionRate) / 100,
                    territory: territory,
                }),
            });
            notify('Hierarchy updated successfully', { type: 'success' });
            handleClose();
            refresh();
        } catch (error: any) {
            notify(error.message || 'Failed to update hierarchy', { type: 'error' });
        }
    };

    return (
        <>
            <Tooltip title="Hierarchy">
                <MuiButton 
                    size="small" 
                    variant="contained"
                    color="primary"
                    onClick={(e) => { e.stopPropagation(); handleOpen(); }}
                    startIcon={<AccountTreeIcon />}
                    sx={{ textTransform: 'none', minWidth: 0, px: 1 }}
                >
                    Hierarchy
                </MuiButton>
            </Tooltip>
            <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
                <DialogTitle>Manage Agent Hierarchy</DialogTitle>
                <DialogContent>
                    <MuiTextField
                        select
                        margin="dense"
                        label="Parent Agent"
                        fullWidth
                        value={parentId}
                        onChange={(e) => setParentId(e.target.value)}
                        SelectProps={{ native: true }}
                        disabled={loadingAgents}
                    >
                        <option value="">Root Agent (No Parent)</option>
                        {availableAgents.map((agent) => (
                            <option key={String(agent.id)} value={String(agent.id)}>
                                {agent.realname || agent.username} ({String(agent.id).slice(-8)})
                            </option>
                        ))}
                    </MuiTextField>
                    <MuiTextField
                        margin="dense"
                        label="Commission Rate (%)"
                        type="number"
                        fullWidth
                        value={commissionRate}
                        onChange={(e) => setCommissionRate(e.target.value)}
                        inputProps={{ min: 0, max: 100, step: 0.01 }}
                    />
                    <MuiTextField
                        margin="dense"
                        label="Territory"
                        fullWidth
                        value={territory}
                        onChange={(e) => setTerritory(e.target.value)}
                        placeholder="e.g., North Region, City Center"
                    />
                </DialogContent>
                <DialogActions>
                    <MuiButton onClick={handleClose}>Cancel</MuiButton>
                    <MuiButton onClick={handleSubmit} variant="contained">Save</MuiButton>
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
                <Datagrid rowClick={false}>
                    <TextField source="id" />
                    <TextField source="username" />
                    <TextField source="realname" />
                    <EmailField source="mobile" />
                    <FunctionField
                        label="Balance"
                        render={(record: any) => (record.balance || 0).toFixed(2)}
                    />
                    <TextField source="status" />
                    <ShowIconButton />
                    <EditButton />
                    <TopupButton />
                    <HierarchyButton />
                </Datagrid>
            )}
        </List>
    );
};

export const AgentCreate = (props: CreateProps) => (
    <Create {...props}>
        <SimpleForm sx={formLayoutSx}>
            <FormSection
                title="Basic Information"
                description="Agent account credentials and basic details"
            >
                <FieldGrid columns={{ xs: 1, sm: 2 }}>
                    <FieldGridItem>
                        <TextInput
                            source="username"
                            label="Username"
                            validate={[required()]}
                            fullWidth
                            size="small"
                            helperText="3-30 characters, letters, numbers and underscores"
                        />
                    </FieldGridItem>
                    <FieldGridItem>
                        <PasswordInput
                            source="password"
                            label="Password"
                            validate={[required()]}
                            fullWidth
                            size="small"
                        />
                    </FieldGridItem>
                    <FieldGridItem>
                        <TextInput
                            source="realname"
                            label="Real Name"
                            validate={[required()]}
                            fullWidth
                            size="small"
                        />
                    </FieldGridItem>
                </FieldGrid>
            </FormSection>

            <FormSection
                title="Contact Information"
                description="Agent contact details"
            >
                <FieldGrid columns={{ xs: 1, sm: 2 }}>
                    <FieldGridItem>
                        <TextInput
                            source="mobile"
                            label="Mobile"
                            fullWidth
                            size="small"
                            helperText="China mobile number"
                        />
                    </FieldGridItem>
                    <FieldGridItem>
                        <TextInput
                            source="email"
                            label="Email"
                            type="email"
                            fullWidth
                            size="small"
                            helperText="For system notifications"
                        />
                    </FieldGridItem>
                </FieldGrid>
            </FormSection>

            <FormSection
                title="Additional Information"
                description="Other optional details"
            >
                <FieldGrid columns={{ xs: 1, sm: 2 }}>
                    <FieldGridItem span={{ xs: 1, sm: 2 }}>
                        <TextInput
                            source="remark"
                            label="Remark"
                            multiline
                            fullWidth
                            size="small"
                        />
                    </FieldGridItem>
                </FieldGrid>
            </FormSection>
            <TextInput source="level" defaultValue="agent" style={{ display: 'none' }} />
            <TextInput source="status" defaultValue="enabled" style={{ display: 'none' }} />
        </SimpleForm>
    </Create>
);

const AgentFormToolbar = (props: any) => (
    <Toolbar {...props}>
        <SaveButton />
        <DeleteButton mutationMode="pessimistic" />
    </Toolbar>
);

export const AgentEdit = (props: EditProps) => {
    return (
        <Edit {...props}>
            <SimpleForm toolbar={<AgentFormToolbar />} sx={formLayoutSx}>
                <FormSection
                    title="Basic Information"
                    description="Agent account credentials and basic details"
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <TextInput
                                source="id"
                                label="ID"
                                disabled
                                fullWidth
                                size="small"
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <TextInput
                                source="username"
                                label="Username"
                                validate={[required()]}
                                fullWidth
                                size="small"
                                helperText="3-30 characters, letters, numbers and underscores"
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <PasswordInput
                                source="password"
                                label="New Password"
                                fullWidth
                                size="small"
                                helperText="Leave blank to keep current password"
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <TextInput
                                source="realname"
                                label="Real Name"
                                validate={[required()]}
                                fullWidth
                                size="small"
                            />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title="Contact Information"
                    description="Agent contact details"
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <TextInput
                                source="mobile"
                                label="Mobile"
                                fullWidth
                                size="small"
                                helperText="China mobile number"
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <TextInput
                                source="email"
                                label="Email"
                                type="email"
                                fullWidth
                                size="small"
                                helperText="For system notifications"
                            />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title="Status"
                    description="Agent account status"
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <SelectInput
                                source="status"
                                label="Status"
                                choices={[
                                    { id: 'enabled', name: 'Enabled' },
                                    { id: 'disabled', name: 'Disabled' }
                                ]}
                                fullWidth
                                size="small"
                            />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title="Additional Information"
                    description="Other optional details"
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem span={{ xs: 1, sm: 2 }}>
                            <TextInput
                                source="remark"
                                label="Remark"
                                multiline
                                fullWidth
                                size="small"
                            />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>
            </SimpleForm>
        </Edit>
    );
};

import { AgentHierarchyTree } from './AgentHierarchyTree';
import { AgentHierarchyForm } from './AgentHierarchyForm';

const AgentHierarchySection = () => {
    const record = useRecordContext();
    return record ? <AgentHierarchyForm agentId={String(record.id)} /> : null;
};

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
                render={(record: any) => `${(record.balance || 0).toFixed(2)}`}
            />
            <AgentHierarchyTree />
            <AgentHierarchySection />
        </SimpleShowLayout>
    </Show>
);
