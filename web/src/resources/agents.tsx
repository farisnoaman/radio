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
    Filter,
    Show,
    useTranslate,
    useLocale,
} from 'react-admin';
import { Link as RouterLink } from 'react-router-dom';
import React from 'react';
import { Box, Card, CardContent, CardActions, Typography, useMediaQuery, Theme, Avatar, Chip, Stack } from '@mui/material';
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
    formLayoutSx,
    DetailItem,
    DetailSectionCard,
    EmptyValue,
} from '../components';
import PersonIcon from '@mui/icons-material/Person';
import PhoneAndroidIcon from '@mui/icons-material/PhoneAndroid';
import AccountBalanceWalletIcon from '@mui/icons-material/AccountBalanceWallet';
import BadgeIcon from '@mui/icons-material/Badge';
import NotesIcon from '@mui/icons-material/Notes';
import CalendarTodayIcon from '@mui/icons-material/CalendarToday';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import CancelIcon from '@mui/icons-material/Cancel';
import { alpha } from '@mui/material/styles';

const AgentLink = React.forwardRef<HTMLAnchorElement, any>((props, ref) => (
    <RouterLink ref={ref} {...(props as any)} />
));

const ShowIconButton = () => {
    const record = useRecordContext();
    const translate = useTranslate();
    if (!record) return null;
    return (
        <Tooltip title={translate('resources.agents.actions.show')}>
            <IconButton
                component={AgentLink}
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
            dir={isRTL ? 'rtl' : 'ltr'}
            sx={{
                display: 'grid',
                gridTemplateColumns: {
                    xs: '1fr',
                    sm: '1fr 1fr',
                    md: 'repeat(3,1fr)'
                },
                gap: { xs: 0.75, sm: 1.25 },
                px: { xs: 1, sm: 1 },
                py: 1
            }}
        >
            {data.map(record => (
                <RecordContextProvider value={record} key={record.id}>
                    <Card
                        elevation={1}
                        sx={{
                            borderRadius: 2,
                            transition: 'all .2s ease',
                            '&:hover': {
                                transform: 'translateY(-2px)',
                                boxShadow: 4
                            }
                        }}
                    >
                        <CardContent
                            sx={{
                                p: 1.75,
                                '&:last-child': { pb: 1.75 }
                            }}
                        >
                            {/* Header */}
                            <Box
                                display="flex"
                                justifyContent="space-between"
                                alignItems="center"
                                mb={1.5}
                            >
                                <Box display="flex" alignItems="center" gap={1}>
                                    <Avatar
                                        sx={{
                                            width: 38,
                                            height: 38,
                                            bgcolor: 'primary.main',
                                            fontSize: 14,
                                            fontWeight: 700
                                        }}
                                    >
                                        {record.username?.charAt(0)?.toUpperCase()}
                                    </Avatar>

                                    <Box>
                                        <Typography
                                            variant="subtitle2"
                                            sx={{ fontWeight: 700 }}
                                        >
                                            {record.username}
                                        </Typography>

                                        <Typography
                                            variant="caption"
                                            color="text.secondary"
                                        >
                                            #{String(record.id).slice(-6)}
                                        </Typography>
                                    </Box>
                                </Box>

                                <Chip
                                    size="small"
                                    label={
                                        record.status === 'enabled'
                                            ? translate('resources.agents.status.active')
                                            : translate('resources.agents.status.inactive')
                                    }
                                    color={
                                        record.status === 'enabled'
                                            ? 'success'
                                            : 'default'
                                    }
                                />
                            </Box>

                            {/* Agent Info */}
                            <Box
                                sx={{
                                    bgcolor: theme =>
                                        theme.palette.mode === 'dark'
                                            ? 'rgba(255,255,255,0.05)'
                                            : 'rgba(0,0,0,0.03)',
                                    p: 1.25,
                                    borderRadius: 1.5,
                                    mb: 1.25
                                }}
                            >
                                <Typography variant="body2">
                                    <strong>
                                        {translate('resources.agents.fields.realname')}:
                                    </strong>{' '}
                                    {record.realname || 'N/A'}
                                </Typography>

                                <Typography variant="body2">
                                    <strong>
                                        {translate('resources.agents.fields.contact')}:
                                    </strong>{' '}
                                    {record.mobile || record.email || 'N/A'}
                                </Typography>

                                <Typography
                                    variant="body2"
                                    sx={{
                                        fontWeight: 700,
                                        color: 'primary.main'
                                    }}
                                >
                                    {translate('resources.agents.fields.balance')}:
                                    ${(record.balance || 0).toFixed(2)}
                                </Typography>
                            </Box>
                        </CardContent>

                        {/* Actions */}
                        <CardActions
                            sx={{
                                px: 1,
                                py: 0.75,
                                justifyContent: 'flex-end',
                                borderTop: theme => `1px solid ${theme.palette.divider}`
                            }}
                        >
                            <Tooltip title={translate('resources.agents.actions.show')}>
                                <IconButton
                                    component={AgentLink}
                                    to={`/agents/${record.id}/show`}
                                    size="small"
                                    sx={{
                                        bgcolor: 'primary.main',
                                        color: 'white',
                                        '&:hover': {
                                            bgcolor: 'primary.dark'
                                        }
                                    }}
                                >
                                    <VisibilityIcon fontSize="small" />
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
    const isSmall = useMediaQuery((theme: Theme) =>
        theme.breakpoints.down('sm')
    );
    const translate = useTranslate();

    return (
        <List
            {...props}
            actions={<AgentListActions />}
            filters={<AgentFilter />}
            sx={{
                '& .RaList-main': {
                    padding: 0
                },
                '& .RaList-content': {
                    padding: 0
                },
                '& .RaLayout-content': {
                    padding: 0
                }
            }}
        >
            {isSmall ? (
                <AgentGrid />
            ) : (
                <Datagrid rowClick={false}>
                    <TextField
                        source="id"
                        label={translate('resources.agents.fields.id')}
                    />
                    <TextField
                        source="username"
                        label={translate('resources.agents.fields.username')}
                    />
                    <TextField
                        source="realname"
                        label={translate('resources.agents.fields.realname')}
                    />
                    <EmailField
                        source="mobile"
                        label={translate('resources.agents.fields.mobile')}
                    />

                    <FunctionField
                        label={translate('resources.agents.fields.balance')}
                        render={(record: any) =>
                            (record.balance || 0).toFixed(2)
                        }
                    />

                    <FunctionField
                        source="status"
                        label={translate('resources.agents.fields.status')}
                        render={(record: any) => (
                            <Chip
                                label={
                                    record.status === 'enabled'
                                        ? translate('resources.agents.status.enabled')
                                        : translate('resources.agents.status.disabled')
                                }
                                size="small"
                                color={
                                    record.status === 'enabled'
                                        ? 'success'
                                        : 'default'
                                }
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
                        <FieldGridItem>
                            <TextInput
                                source="radius_username"
                                label={translate('resources.agents.fields.radius_username')}
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
                        <FieldGridItem>
                            <TextInput
                                source="radius_username"
                                label={translate('resources.agents.fields.radius_username')}
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

// Helper hook for formatting (similar to products.tsx)
const useAgentFormatters = () => {
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';

    const formatTimestamp = (value?: string | number): string => {
        if (!value) return '-';
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) return '-';
        return date.toLocaleString();
    };

    const formatBalance = (balance?: number): string => {
        if (balance === undefined || balance === null) return '$0.00';
        return `${balance.toFixed(2)}`;
    };

    return { formatTimestamp, formatBalance, isRTL, translate };
};

const AgentStatusIndicator = ({ isEnabled }: { isEnabled: boolean }) => {
    const translate = useTranslate();
    return (
        <Chip
            icon={isEnabled ? <CheckCircleIcon sx={{ fontSize: '0.85rem !important' }} /> : <CancelIcon sx={{ fontSize: '0.85rem !important' }} />}
            label={isEnabled ? translate('resources.agents.status.active', { _: 'Active' }) : translate('resources.agents.status.inactive', { _: 'Inactive' })}
            size="small"
            color={isEnabled ? 'success' : 'default'}
            variant={isEnabled ? 'filled' : 'outlined'}
            sx={{ height: 22, fontWeight: 500, fontSize: '0.75rem' }}
        />
    );
};

const AgentHeaderCard = () => {
    const record = useRecordContext();
    const { formatBalance, translate } = useAgentFormatters();

    if (!record) return null;

    const isEnabled = record.status === 'enabled';

    return (
        <Card
            elevation={0}
            sx={{
                borderRadius: 4,
                background: theme =>
                    theme.palette.mode === 'dark'
                        ? isEnabled
                            ? `linear-gradient(135deg, ${alpha(theme.palette.primary.dark, 0.4)} 0%, ${alpha(theme.palette.info.dark, 0.3)} 100%)`
                            : `linear-gradient(135deg, ${alpha(theme.palette.grey[800], 0.5)} 0%, ${alpha(theme.palette.grey[700], 0.3)} 100%)`
                        : isEnabled
                            ? `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.1)} 0%, ${alpha(theme.palette.info.main, 0.08)} 100%)`
                            : `linear-gradient(135deg, ${alpha(theme.palette.grey[400], 0.15)} 0%, ${alpha(theme.palette.grey[300], 0.1)} 100%)`,
                border: theme => `1px solid ${alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.2)}`,
                overflow: 'hidden',
                position: 'relative',
            }}
        >
            <Box
                sx={{
                    position: 'absolute',
                    top: -50,
                    right: -50,
                    width: 200,
                    height: 200,
                    borderRadius: '50%',
                    background: theme => alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.1),
                    pointerEvents: 'none',
                }}
            />

            <CardContent sx={{ p: 3, position: 'relative', zIndex: 1 }}>
                <Box sx={{ display: 'flex', flexDirection: { xs: 'column', sm: 'row' }, justifyContent: 'space-between', alignItems: { xs: 'stretch', sm: 'flex-start' }, mb: 3, gap: 2 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                        <Avatar
                            sx={{
                                width: 64,
                                height: 64,
                                bgcolor: isEnabled ? 'primary.main' : 'grey.500',
                                fontSize: '1.5rem',
                                fontWeight: 700,
                                boxShadow: theme => `0 4px 14px ${alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.4)}`,
                            }}
                        >
                            {record.username?.charAt(0).toUpperCase() || 'A'}
                        </Avatar>
                        <Box>
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                                <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
                                    {record.realname || record.username || <EmptyValue message="Unknown Agent" />}
                                </Typography>
                                <AgentStatusIndicator isEnabled={isEnabled} />
                            </Box>
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                                <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                                    ID: {record.id}
                                </Typography>
                            </Box>
                        </Box>
                    </Box>
                    <Box
                        sx={{
                            display: 'flex',
                            flexDirection: 'column',
                            alignItems: { xs: 'stretch', sm: 'flex-end' },
                            gap: 1
                        }}
                    >
                        <Typography variant="caption" color="text.secondary">
                            {translate('resources.agents.fields.balance', { _: 'Balance' })}
                        </Typography>
                        <Typography variant="h4" sx={{ fontWeight: 700, color: 'primary.main' }}>
                            {formatBalance(record.balance)}
                        </Typography>
                    </Box>
                </Box>
            </CardContent>
        </Card>
    );
};

const AgentDetails = () => {
    const record = useRecordContext();
    const { formatTimestamp, formatBalance, isRTL, translate } = useAgentFormatters();

    if (!record) {
        return null;
    }

    return (
        <Box sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 }, direction: isRTL ? 'rtl' : 'ltr' }}>
            <Stack spacing={3}>
                <AgentHeaderCard />

                <DetailSectionCard
                    title={translate('resources.agents.sections.basic', { _: 'Basic Information' })}
                    description={translate('resources.agents.details.basic_desc', { _: 'Agent core profile details' })}
                    icon={<PersonIcon />}
                    color="primary"
                >
                    <Box
                        sx={{
                            display: 'grid',
                            gap: 2,
                            gridTemplateColumns: {
                                xs: 'repeat(1, 1fr)',
                                sm: 'repeat(2, 1fr)',
                            },
                        }}
                    >
                        <DetailItem
                            label={translate('resources.agents.fields.id', { _: 'ID' })}
                            value={record.id}
                        />
                        <DetailItem
                            label={translate('resources.agents.fields.username', { _: 'Username' })}
                            value={record.username}
                        />
                        <DetailItem
                            label={translate('resources.agents.fields.realname', { _: 'Real Name' })}
                            value={record.realname || '-'}
                        />
                        <DetailItem
                            label={translate('resources.agents.fields.status', { _: 'Status' })}
                            value={
                                <Chip
                                    label={record.status === 'enabled' ? translate('resources.agents.status.enabled', { _: 'Enabled' }) : translate('resources.agents.status.disabled', { _: 'Disabled' })}
                                    size="small"
                                    color={record.status === 'enabled' ? 'success' : 'default'}
                                />
                            }
                        />
                    </Box>
                </DetailSectionCard>

                <DetailSectionCard
                    title={translate('resources.agents.sections.contact', { _: 'Contact Information' })}
                    description={translate('resources.agents.details.contact_desc', { _: 'Agent contact details' })}
                    icon={<PhoneAndroidIcon />}
                    color="info"
                >
                    <Box
                        sx={{
                            display: 'grid',
                            gap: 2,
                            gridTemplateColumns: {
                                xs: 'repeat(1, 1fr)',
                                sm: 'repeat(2, 1fr)',
                            },
                        }}
                    >
                        <DetailItem
                            label={translate('resources.agents.fields.mobile', { _: 'Mobile' })}
                            value={record.mobile || '-'}
                        />
                        <DetailItem
                            label={translate('resources.agents.fields.email', { _: 'Email' })}
                            value={record.email || '-'}
                        />
                    </Box>
                </DetailSectionCard>

                <DetailSectionCard
                    title={translate('resources.agents.sections.financial', { _: 'Financial' })}
                    description={translate('resources.agents.details.financial_desc', { _: 'Agent balance and commission' })}
                    icon={<AccountBalanceWalletIcon />}
                    color="success"
                >
                    <Box
                        sx={{
                            display: 'grid',
                            gap: 2,
                            gridTemplateColumns: {
                                xs: 'repeat(1, 1fr)',
                                sm: 'repeat(2, 1fr)',
                            },
                        }}
                    >
                        <DetailItem
                            label={translate('resources.agents.fields.balance', { _: 'Balance' })}
                            value={formatBalance(record.balance)}
                            highlight
                        />
                        <DetailItem
                            label={translate('resources.agents.fields.commission_rate', { _: 'Commission Rate' })}
                            value={record.commission_rate ? `${(record.commission_rate * 100).toFixed(2)}%` : '-'}
                        />
                    </Box>
                </DetailSectionCard>

                <DetailSectionCard
                    title={translate('resources.agents.details.time_info', { _: 'Time Information' })}
                    description={translate('resources.agents.details.time_desc', { _: 'Creation and modification dates' })}
                    icon={<CalendarTodayIcon />}
                    color="warning"
                >
                    <Box
                        sx={{
                            display: 'grid',
                            gap: 2,
                            gridTemplateColumns: {
                                xs: 'repeat(1, 1fr)',
                                sm: 'repeat(2, 1fr)',
                            },
                        }}
                    >
                        <DetailItem
                            label={translate('resources.agents.fields.created_at', { _: 'Created At' })}
                            value={formatTimestamp(record.created_at)}
                        />
                        <DetailItem
                            label={translate('resources.agents.fields.updated_at', { _: 'Updated At' })}
                            value={formatTimestamp(record.updated_at)}
                        />
                    </Box>
                </DetailSectionCard>

                <DetailSectionCard
                    title={translate('resources.agents.sections.remark', { _: 'Remarks' })}
                    description={translate('resources.agents.details.remarks_desc', { _: 'Additional notes or descriptions' })}
                    icon={<NotesIcon />}
                    color="primary"
                >
                    <Box
                        sx={{
                            p: 2,
                            borderRadius: 2,
                            bgcolor: theme =>
                                theme.palette.mode === 'dark'
                                    ? 'rgba(255, 255, 255, 0.02)'
                                    : 'rgba(0, 0, 0, 0.02)',
                            border: theme => `1px solid ${theme.palette.divider}`,
                            minHeight: 80,
                        }}
                    >
                        <Typography
                            variant="body2"
                            sx={{
                                whiteSpace: 'pre-wrap',
                                wordBreak: 'break-word',
                                color: record.remark ? 'text.primary' : 'text.disabled',
                                fontStyle: record.remark ? 'normal' : 'italic',
                            }}
                        >
                            {record.remark || 'No remark added.'}
                        </Typography>
                    </Box>
                </DetailSectionCard>

                {record.id && (
                    <>
                        <DetailSectionCard
                            title={translate('resources.agents.hierarchy.sub_agents', { _: 'Sub Agents' })}
                            description={translate('resources.agents.hierarchy.sub_agents_desc', { _: 'View sub-agents under this agent' })}
                            icon={<BadgeIcon />}
                            color="primary"
                        >
                            <AgentHierarchyTree />
                        </DetailSectionCard>

                        <DetailSectionCard
                            title={translate('resources.agents.hierarchy.title', { _: 'Agent Hierarchy' })}
                            description={translate('resources.agents.hierarchy.description', { _: 'Manage agent parent relationships' })}
                            icon={<AccountTreeIcon />}
                            color="info"
                        >
                            <AgentHierarchyForm agentId={String(record.id)} />
                        </DetailSectionCard>
                    </>
                )}
            </Stack>
        </Box>
    );
};

export const AgentShow = (props: any) => (
    <Show {...props} emptyWhileLoading>
        <AgentDetails />
    </Show>
);
