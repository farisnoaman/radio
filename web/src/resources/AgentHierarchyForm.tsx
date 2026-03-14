import { useState, useEffect } from 'react';
import {
    Box,
    Card,
    CardContent,
    Typography,
    TextField,
    Button,
    FormControl,
    InputLabel,
    Select,
    MenuItem,
    CircularProgress,
    Alert,
    InputAdornment,
} from '@mui/material';
import SaveIcon from '@mui/icons-material/Save';
import { useTranslate, useLocale, useNotify, useRefresh } from 'react-admin';
import { httpClient } from '../utils/apiClient';

interface AgentHierarchyFormProps {
    agentId: number | string;
    onSuccess?: () => void;
}

interface AgentOption {
    id: number | string;
    username: string;
    realname: string;
}

export const AgentHierarchyForm: React.FC<AgentHierarchyFormProps> = ({
    agentId,
    onSuccess,
}) => {
    const translate = useTranslate();
    const locale = useLocale();
    const isRTL = locale === 'ar';
    const notify = useNotify();
    const refresh = useRefresh();

    const [parentId, setParentId] = useState<string>('');
    const [commissionRate, setCommissionRate] = useState<string>('');
    const [territory, setTerritory] = useState<string>('');
    const [availableAgents, setAvailableAgents] = useState<AgentOption[]>([]);
    
    const [loading, setLoading] = useState<boolean>(true);
    const [submitting, setSubmitting] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);

    // Load initial data
    useEffect(() => {
        const loadData = async () => {
            try {
                setLoading(true);
                // Fetch current hierarchy
                const hierarchyRes = await httpClient(`/agents/${agentId}/hierarchy`, { method: 'GET' });
                const hierarchy = hierarchyRes.json as any;
                if (hierarchy) {
                    if (hierarchy.parent_id) setParentId(String(hierarchy.parent_id));
                    setCommissionRate(String((hierarchy.commission_rate || 0) * 100));
                    setTerritory(hierarchy.territory || '');
                }

                // Fetch available parent agents
                const agentsRes = await httpClient(`/agents?filter={"level":"agent"}&perPage=1000`, { method: 'GET' });
                const agentsData = agentsRes.json as any;
                const filtered = (agentsData.data || []).filter((a: any) => String(a.id) !== String(agentId));
                setAvailableAgents(filtered);
            } catch (err: any) {
                console.error('Failed to load hierarchy data:', err);
                setError(translate('resources.agents.hierarchy.error_loading'));
            } finally {
                setLoading(false);
            }
        };

        if (agentId) {
            loadData();
        }
    }, [agentId, translate]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setSubmitting(true);
        try {
            const parentIdValue = parentId && parentId !== '' ? parentId : null;
            await httpClient(`/agents/${agentId}/assign-parent`, {
                method: 'POST',
                body: JSON.stringify({
                    parent_id: parentIdValue,
                    commission_rate: parseFloat(commissionRate) / 100,
                    territory: territory,
                }),
            });
            notify('resources.agents.hierarchy.success', { type: 'success' });
            refresh();
            if (onSuccess) onSuccess();
        } catch (err: any) {
            notify(err.message || 'resources.agents.hierarchy.error', { type: 'error' });
        } finally {
            setSubmitting(false);
        }
    };

    if (loading) {
        return (
            <Box display="flex" justifyContent="center" py={4}>
                <CircularProgress size={24} />
            </Box>
        );
    }

    const inputLabelSx = isRTL ? {
        transformOrigin: 'top right',
        left: 'auto',
        right: 28,
        textAlign: 'right',
        '&.MuiInputLabel-shrink': {
            transform: 'translate(14px, -6px) scale(0.75)',
            right: 24,
        }
    } : {};

    const inputSx = { textAlign: (isRTL ? 'right' : 'left') as any };

    return (
        <Card variant="outlined" sx={{ mt: 2 }} dir={isRTL ? 'rtl' : 'ltr'}>
            <CardContent>
                <Typography variant="h6" gutterBottom sx={{ textAlign: isRTL ? 'right' : 'left' }}>
                    {translate('resources.agents.hierarchy.manage')}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2, textAlign: isRTL ? 'right' : 'left' }}>
                    {translate('resources.agents.hierarchy.description')}
                </Typography>

                <Box component="form" onSubmit={handleSubmit} sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
                    {error && <Alert severity="error">{error}</Alert>}

                    <FormControl fullWidth size="small" variant="outlined">
                        <InputLabel id="parent-id-label" sx={inputLabelSx}>
                            {translate('resources.agents.fields.parent')}
                        </InputLabel>
                        <Select
                            labelId="parent-id-label"
                            value={parentId}
                            label={translate('resources.agents.fields.parent')}
                            onChange={(e) => setParentId(e.target.value)}
                            sx={inputSx}
                        >
                            <MenuItem value="" sx={{ direction: isRTL ? 'rtl' : 'ltr', textAlign: isRTL ? 'right' : 'left' }}>
                                <em>{translate('resources.agents.hierarchy.no_parent')}</em>
                            </MenuItem>
                            {availableAgents.map((agent) => (
                                <MenuItem key={String(agent.id)} value={String(agent.id)} sx={{ direction: isRTL ? 'rtl' : 'ltr', textAlign: isRTL ? 'right' : 'left' }}>
                                    {agent.realname || agent.username} ({String(agent.id).slice(-8)})
                                </MenuItem>
                            ))}
                        </Select>
                    </FormControl>

                    <TextField
                        label={translate('resources.agents.fields.commission_rate')}
                        type="number"
                        fullWidth
                        size="small"
                        value={commissionRate}
                        placeholder="0"
                        onChange={(e) => setCommissionRate(e.target.value)}
                        InputProps={{
                            endAdornment: <InputAdornment position="end">%</InputAdornment>,
                            sx: inputSx
                        }}
                        inputProps={{ min: 0, max: 100, step: 0.01, style: inputSx }}
                        InputLabelProps={{ sx: inputLabelSx }}
                    />

                    <TextField
                        label={translate('resources.agents.fields.territory')}
                        fullWidth
                        size="small"
                        value={territory}
                        onChange={(e) => setTerritory(e.target.value)}
                        placeholder={translate('resources.agents.hierarchy.territory_placeholder')}
                        InputProps={{ sx: inputSx }}
                        inputProps={{ style: inputSx }}
                        InputLabelProps={{ sx: inputLabelSx }}
                    />

                    <Box sx={{ display: 'flex', justifyContent: isRTL ? 'flex-start' : 'flex-end', mt: 1 }}>
                        <Button
                            type="submit"
                            variant="contained"
                            disabled={submitting}
                            startIcon={submitting ? <CircularProgress size={20} /> : <SaveIcon />}
                        >
                            {submitting ? translate('resources.agents.actions.saving') : translate('resources.agents.actions.save')}
                        </Button>
                        <Button onClick={onSuccess} sx={{ ml: isRTL ? 0 : 1, mr: isRTL ? 1 : 0 }}>
                            {translate('resources.agents.actions.cancel')}
                        </Button>
                    </Box>
                </Box>
            </CardContent>
        </Card>
    );
};

export default AgentHierarchyForm;
