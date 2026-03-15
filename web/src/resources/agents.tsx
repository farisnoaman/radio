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
    SimpleShowLayout,
    useTranslate,
    useLocale,
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
    const translate = useTranslate();
    if (!record) return null;
    return (
        <Tooltip title={translate('resources.agents.actions.show')}>
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

const AgentFilter = (props: any) => {
    const translate = useTranslate();
    return (
        <Filter {...props}>
            <TextInput source="q" label={translate('ra.action.search')} alwaysOn />
            <TextInput source="username" label={translate('resources.agents.fields.username')} />
            <TextInput source="realname" label={translate('resources.agents.fields.realname')} />
            <TextInput source="email" label={translate('resources.agents.fields.email')} />
        </Filter>
    );
};

const TopupButton = () => {
    const record = useRecordContext();
    const [open, setOpen] = useState(false);
    const [amount, setAmount] = useState('');
    const [remark, setRemark] = useState('');
    const notify = useNotify();
    const refresh = useRefresh();

    const handleOpen = () => setOpen(true);
    const handleClose = () => setOpen(false);

    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

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
            notify('resources.agents.dialogs.topup_success', { type: 'success' });
            setOpen(false);
            refresh();
        } catch (error) {
            notify('resources.agents.dialogs.topup_failed', { type: 'error' });
        }
    };

    return (
        <>
            <Tooltip title={translate('resources.agents.actions.topup')}>
                <MuiButton 
                    size="small" 
                    variant="contained"
                    color="success"
                    onClick={(e) => { e.stopPropagation(); handleOpen(); }} 
                    startIcon={<AddCardIcon />}
                    sx={{ textTransform: 'none', minWidth: 0, px: 1, ml: isRTL ? 0 : 0.5, mr: isRTL ? 0.5 : 0 }}
                >
                    {translate('resources.agents.actions.topup')}
                </MuiButton>
            </Tooltip>
            <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth dir={isRTL ? 'rtl' : 'ltr'}>
                <DialogTitle>{translate('resources.agents.dialogs.topup_title')}</DialogTitle>
                <DialogContent>
                    <MuiTextField
                        autoFocus
                        margin="dense"
                        label={translate('resources.agents.fields.amount')}
                        type="number"
                        fullWidth
                        placeholder="0"
                        value={amount}
                        onChange={(e) => setAmount(e.target.value)}
                        dir={isRTL ? 'rtl' : 'ltr'}
                        slotProps={{
                            input: { sx: { textAlign: isRTL ? 'right' : 'left' } },
                            inputLabel: { sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }
                        }}
                    />
                    <MuiTextField
                        margin="dense"
                        label={translate('resources.agents.fields.remark')}
                        fullWidth
                        value={remark}
                        onChange={(e) => setRemark(e.target.value)}
                        dir={isRTL ? 'rtl' : 'ltr'}
                        slotProps={{
                            input: { sx: { textAlign: isRTL ? 'right' : 'left' } },
                            inputLabel: { sx: { transformOrigin: isRTL ? 'top right' : 'top left', left: isRTL ? 'auto' : 0, right: isRTL ? 24 : 'auto' } }
                        }}
                    />
                </DialogContent>
                <DialogActions sx={{ justifyContent: isRTL ? 'flex-start' : 'flex-end' }}>
                    <MuiButton onClick={handleClose}>{translate('resources.agents.actions.cancel')}</MuiButton>
                    <MuiButton onClick={handleSubmit} variant="contained">{translate('resources.agents.actions.submit')}</MuiButton>
                </DialogActions>
            </Dialog>
        </>
    );
};

const EditButton = () => {
    const record = useRecordContext();
    const navigate = useNavigate();
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

    const handleClick = (e: React.MouseEvent) => {
        e.stopPropagation();
        if (record) {
            navigate(`/agents/${record.id}`);
        }
    };

    return (
        <Tooltip title={translate('resources.agents.actions.edit')}>
            <MuiButton 
                size="small" 
                variant="outlined"
                color="info"
                onClick={handleClick}
                startIcon={<EditIcon />}
                sx={{ textTransform: 'none', minWidth: 0, px: 1, ml: isRTL ? 0 : 0.5, mr: isRTL ? 0.5 : 0 }}
            >
                {translate('ra.action.edit')}
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

    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

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
            notify('resources.agents.hierarchy.success', { type: 'success' });
            handleClose();
            refresh();
        } catch (error: any) {
            notify(error.message || 'resources.agents.hierarchy.error', { type: 'error' });
        }
    };

    return (
        <>
            <Tooltip title={translate('resources.agents.actions.hierarchy')}>
                <MuiButton 
                    size="small" 
                    variant="contained"
                    color="primary"
                    onClick={(e) => { e.stopPropagation(); handleOpen(); }}
                    startIcon={<AccountTreeIcon />}
                    sx={{ textTransform: 'none', minWidth: 0, px: 1, ml: isRTL ? 0 : 0.5, mr: isRTL ? 0.5 : 0 }}
                >
                    {translate('resources.agents.actions.hierarchy')}
                </MuiButton>
            </Tooltip>
            <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth dir={isRTL ? 'rtl' : 'ltr'}>
                <DialogTitle>{translate('resources.agents.hierarchy.manage')}</DialogTitle>
                <DialogContent>
                    <MuiTextField
                        select
                        margin="dense"
                        label={translate('resources.agents.fields.parent')}
                        fullWidth
                        value={parentId}
                        onChange={(e) => setParentId(e.target.value)}
                        SelectProps={{ native: true }}
                        disabled={loadingAgents}
                        dir={isRTL ? 'rtl' : 'ltr'}
                        sx={{ textAlign: isRTL ? 'right' : 'left' }}
                    >
                        <option value="">{translate('resources.agents.hierarchy.no_parent')}</option>
                        {availableAgents.map((agent) => (
                            <option key={String(agent.id)} value={String(agent.id)}>
                                {agent.realname || agent.username} ({String(agent.id).slice(-8)})
                            </option>
                        ))}
                    </MuiTextField>
                    <MuiTextField
                        margin="dense"
                        label={translate('resources.agents.fields.commission_rate')}
                        type="number"
                        fullWidth
                        value={commissionRate}
                        onChange={(e) => setCommissionRate(e.target.value)}
                        inputProps={{ min: 0, max: 100, step: 0.01 }}
                        dir={isRTL ? 'rtl' : 'ltr'}
                        sx={{ textAlign: isRTL ? 'right' : 'left' }}
                    />
                    <MuiTextField
                        margin="dense"
                        label={translate('resources.agents.fields.territory')}
                        fullWidth
                        value={territory}
                        onChange={(e) => setTerritory(e.target.value)}
                        placeholder={translate('resources.agents.hierarchy.territory_placeholder')}
                        dir={isRTL ? 'rtl' : 'ltr'}
                        sx={{ textAlign: isRTL ? 'right' : 'left' }}
                    />
                </DialogContent>
                <DialogActions sx={{ justifyContent: isRTL ? 'flex-start' : 'flex-end' }}>
                    <MuiButton onClick={handleClose}>{translate('resources.agents.actions.cancel')}</MuiButton>
                    <MuiButton onClick={handleSubmit} variant="contained">{translate('resources.agents.actions.save')}</MuiButton>
                </DialogActions>
            </Dialog>
        </>
    );
};


const AgentGrid = () => {
    const { data, isLoading } = useListContext();
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

    if (isLoading || !data) return null;

    return (
        <Box
            display="grid"
            gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)' }}
            gap={3}
            p={3}
            dir={isRTL ? 'rtl' : 'ltr'}
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
                                    label={record.status === 'enabled' ? translate('resources.agents.status.active') : translate('resources.agents.status.inactive')}
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
                                    <Typography variant="caption" color="text.secondary" sx={{ minWidth: 60 }}>
                                        {translate('resources.agents.fields.realname')}:
                                    </Typography>
                                    <Typography variant="body2" fontWeight="bold">
                                        {record.realname || 'N/A'}
                                    </Typography>
                                </Box>
                                <Box display="flex" alignItems="center" gap={1} mb={1.5}>
                                    <Typography variant="caption" color="text.secondary" sx={{ minWidth: 60 }}>
                                        {translate('resources.agents.fields.contact')}:
                                    </Typography>
                                    <Typography variant="body2">
                                        {record.mobile || record.email || 'N/A'}
                                    </Typography>
                                </Box>
                                <Box display="flex" alignItems="center" gap={1}>
                                    <Typography variant="caption" color="text.secondary" sx={{ minWidth: 60 }}>
                                        {translate('resources.agents.fields.balance')}:
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
                                        letterSpacing: 0.5,
                                        textAlign: isRTL ? 'right' : 'left'
                                    }}
                                >
                                    {translate('resources.agents.sections.statistics')}
                                </Typography>
                                <Box display="flex" justifyContent="space-between" gap={2}>
                                    <Box textAlign="center" flex={1}>
                                        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5 }}>
                                            {translate('resources.agents.fields.level')}
                                        </Typography>
                                        <Typography variant="body1" fontWeight="bold" sx={{ fontSize: '1.1rem' }}>
                                            {record.level !== undefined ? record.level : '-'}
                                        </Typography>
                                    </Box>
                                    <Box textAlign="center" flex={1}>
                                        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5 }}>
                                            {translate('resources.agents.fields.tier')}
                                        </Typography>
                                        <Typography variant="body1" fontWeight="bold" sx={{ fontSize: '1.1rem' }}>
                                            {record.level === 0 ? translate('resources.agents.hierarchy.no_parent') : record.level === 1 ? 'Level 1' : `L${record.level || 0}`}
                                        </Typography>
                                    </Box>
                                    <Box textAlign="center" flex={1}>
                                        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5 }}>
                                            {translate('resources.agents.fields.status')}
                                        </Typography>
                                        <Typography
                                            variant="body1"
                                            fontWeight="bold"
                                            color={record.status === 'enabled' ? 'success.main' : 'error.main'}
                                            sx={{ fontSize: '1.1rem' }}
                                        >
                                            {record.status === 'enabled' ? translate('resources.agents.status.active') : translate('resources.agents.status.off')}
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
                            <Tooltip title={translate('resources.agents.actions.show')}>
                                <IconButton
                                    component={Link}
                                    to={`/agents/${record.id}/show`}
                                    size="small"
                                    sx={{
                                        bgcolor: 'primary.main',
                                        color: 'white',
                                        '&:hover': { bgcolor: 'primary.dark' },
                                        width: 36,
                                        height: 36,
                                        ml: isRTL ? 0.5 : 0,
                                        mr: isRTL ? 0 : 0.5
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
    const translate = useTranslate();
    return (
        <List {...props} actions={<AgentListActions />} filters={<AgentFilter />}>
            {isSmall ? (
                <AgentGrid />
            ) : (
                <Datagrid rowClick={false}>
                    <TextField source="id" label={translate('resources.agents.fields.id')} />
                    <TextField source="username" label={translate('resources.agents.fields.username')} />
                    <TextField source="realname" label={translate('resources.agents.fields.realname')} />
                    <EmailField source="mobile" label={translate('resources.agents.fields.mobile')} />
                    <FunctionField
                        label={translate('resources.agents.fields.balance')}
                        render={(record: any) => (record.balance || 0).toFixed(2)}
                    />
                    <FunctionField
                        source="status"
                        label={translate('resources.agents.fields.status')}
                        render={(record: any) => (
                            <Chip
                                label={record.status === 'enabled' ? translate('resources.agents.status.enabled', { _: 'Enabled' }) : translate('resources.agents.status.disabled', { _: 'Disabled' })}
                                size="small"
                                color={record.status === 'enabled' ? 'success' : 'default'}
                                variant={record.status === 'enabled' ? 'filled' : 'outlined'}
                                sx={{ fontWeight: 600, fontSize: '0.75rem' }}
                            />
                        )}
                    />
                    <ShowIconButton />
                    <EditButton />
                    <TopupButton />
                    <HierarchyButton />
                </Datagrid>
            )}
        </List>
    );
};

export const AgentCreate = (props: CreateProps) => {
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

    const inputLabelProps = {
        sx: {
            transformOrigin: isRTL ? 'top right' : 'top left',
            left: isRTL ? 'auto' : 0,
            right: isRTL ? 24 : 'auto',
        }
    };

    const textInputProps = {
        style: { textAlign: (isRTL ? 'right' : 'left') as any },
        dir: isRTL ? 'rtl' : 'ltr'
    };

    return (
        <Create {...props}>
            <SimpleForm sx={{ ...formLayoutSx, direction: isRTL ? 'rtl' : 'ltr' }}>
                <FormSection
                    title={translate('resources.agents.sections.basic')}
                    description={translate('resources.agents.sections.basic_desc')}
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <TextInput
                                source="username"
                                label={translate('resources.agents.fields.username')}
                                validate={[required()]}
                                fullWidth
                                size="small"
                                inputProps={textInputProps}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <PasswordInput
                                source="password"
                                label={translate('resources.agents.fields.password')}
                                validate={[required()]}
                                fullWidth
                                size="small"
                                inputProps={textInputProps}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <TextInput
                                source="realname"
                                label={translate('resources.agents.fields.realname')}
                                validate={[required()]}
                                fullWidth
                                size="small"
                                inputProps={textInputProps}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title={translate('resources.agents.sections.contact')}
                    description={translate('resources.agents.sections.contact_desc')}
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <TextInput
                                source="mobile"
                                label={translate('resources.agents.fields.mobile')}
                                fullWidth
                                size="small"
                                inputProps={textInputProps}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <TextInput
                                source="email"
                                label={translate('resources.agents.fields.email')}
                                type="email"
                                fullWidth
                                size="small"
                                inputProps={textInputProps}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title={translate('resources.agents.sections.additional')}
                    description={translate('resources.agents.sections.additional_desc')}
                >
                    <TextInput
                        source="remark"
                        label={translate('resources.agents.fields.remark')}
                        multiline
                        rows={3}
                        fullWidth
                        size="small"
                        inputProps={textInputProps}
                        InputLabelProps={inputLabelProps}
                    />
                </FormSection>
                <TextInput source="level" defaultValue="agent" style={{ display: 'none' }} />
                <TextInput source="status" defaultValue="enabled" style={{ display: 'none' }} />
            </SimpleForm>
        </Create>
    );
};

const AgentFormToolbar = (props: any) => (
    <Toolbar {...props}>
        <SaveButton />
        <DeleteButton mutationMode="pessimistic" />
    </Toolbar>
);

export const AgentEdit = (props: EditProps) => {
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

    const inputLabelProps = {
        sx: {
            transformOrigin: isRTL ? 'top right' : 'top left',
            left: isRTL ? 'auto' : 0,
            right: isRTL ? 24 : 'auto',
        }
    };

    const textInputProps = {
        style: { textAlign: (isRTL ? 'right' : 'left') as any },
        dir: isRTL ? 'rtl' : 'ltr'
    };

    return (
        <Edit {...props}>
            <SimpleForm toolbar={<AgentFormToolbar />} sx={{ ...formLayoutSx, direction: isRTL ? 'rtl' : 'ltr' }}>
                <FormSection
                    title={translate('resources.agents.sections.basic')}
                    description={translate('resources.agents.sections.basic_desc')}
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <TextInput
                                source="id"
                                label={translate('resources.agents.fields.id')}
                                disabled
                                fullWidth
                                size="small"
                                inputProps={textInputProps}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <TextInput
                                source="username"
                                label={translate('resources.agents.fields.username')}
                                validate={[required()]}
                                fullWidth
                                size="small"
                                inputProps={textInputProps}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <PasswordInput
                                source="password"
                                label={translate('resources.agents.fields.password')}
                                fullWidth
                                size="small"
                                inputProps={textInputProps}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <SelectInput
                                source="status"
                                label={translate('resources.agents.fields.status')}
                                choices={[
                                    { id: 'enabled', name: translate('resources.agents.status.enabled') },
                                    { id: 'disabled', name: translate('resources.agents.status.disabled') }
                                ]}
                                fullWidth
                                size="small"
                                dir={isRTL ? 'rtl' : 'ltr'}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <TextInput
                                source="realname"
                                label={translate('resources.agents.fields.realname')}
                                validate={[required()]}
                                fullWidth
                                size="small"
                                inputProps={textInputProps}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title={translate('resources.agents.sections.contact')}
                    description={translate('resources.agents.sections.contact_desc')}
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <TextInput
                                source="mobile"
                                label={translate('resources.agents.fields.mobile')}
                                fullWidth
                                size="small"
                                inputProps={textInputProps}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                        <FieldGridItem>
                            <TextInput
                                source="email"
                                label={translate('resources.agents.fields.email')}
                                type="email"
                                fullWidth
                                size="small"
                                inputProps={textInputProps}
                                InputLabelProps={inputLabelProps}
                            />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title={translate('resources.agents.sections.additional')}
                    description={translate('resources.agents.sections.additional_desc')}
                >
                    <TextInput
                        source="remark"
                        label={translate('resources.agents.fields.remark')}
                        multiline
                        rows={3}
                        fullWidth
                        size="small"
                        inputProps={textInputProps}
                        InputLabelProps={inputLabelProps}
                    />
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

export const AgentShow = (props: any) => {
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

    return (
        <Show {...props}>
            <SimpleShowLayout sx={{ direction: isRTL ? 'rtl' : 'ltr' }}>
                <Box sx={{ mb: 2 }}>
                    <Typography variant="h5" gutterBottom>
                        {translate('resources.agents.sections.details')}
                    </Typography>
                </Box>
                <TextField source="id" label={translate('resources.agents.fields.id')} />
                <TextField source="username" label={translate('resources.agents.fields.username')} />
                <TextField source="realname" label={translate('resources.agents.fields.realname')} />
                <EmailField source="email" label={translate('resources.agents.fields.email')} />
                <TextField source="mobile" label={translate('resources.agents.fields.mobile')} />
                <TextField source="status" label={translate('resources.agents.fields.status')} />
                <FunctionField
                    label={translate('resources.agents.fields.balance')}
                    render={(record: any) => `${(record.balance || 0).toFixed(2)}`}
                />
                <AgentHierarchyTree />
                <AgentHierarchySection />
            </SimpleShowLayout>
        </Show>
    );
};
