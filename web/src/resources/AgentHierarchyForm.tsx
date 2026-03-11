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
import { useTranslate } from 'react-admin';
import { useAssignParent, useAgentHierarchy } from '../hooks/useAgentHierarchy';
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
    const [parentId, setParentId] = useState<string>('');
    const [commissionRate, setCommissionRate] = useState<string>('0');
    const [territory, setTerritory] = useState<string>('');
    const [availableAgents, setAvailableAgents] = useState<AgentOption[]>([]);
    const [loadingAgents, setLoadingAgents] = useState<boolean>(true);

    // Fetch current hierarchy data
    const { data: hierarchy, isLoading: loadingHierarchy, refetch: refetchHierarchy } = useAgentHierarchy(agentId);

    // Assign parent mutation
    const assignParent = useAssignParent();

    // Load available agents for parent selection
    useEffect(() => {
        const fetchAgents = async () => {
            try {
                setLoadingAgents(true);
                // Fetch all agents excluding current agent
                const response = await httpClient(`/agents?filter={"level":"agent"}&perPage=1000`, {
                    method: 'GET',
                });
                const data = response.json as Promise<{ data: AgentOption[] }>;
                const result = await data;
                // Filter out the current agent
                const filtered = (result.data || []).filter((a: AgentOption) => a.id !== agentId);
                setAvailableAgents(filtered);
            } catch (error) {
                console.error('Failed to fetch agents:', error);
            } finally {
                setLoadingAgents(false);
            }
        };
        fetchAgents();
    }, [agentId]);

    // Set form values when hierarchy data loads
    useEffect(() => {
        if (hierarchy) {
            if (hierarchy.parent_id) {
                setParentId(String(hierarchy.parent_id));
            }
            setCommissionRate(String(hierarchy.commission_rate * 100));
            setTerritory(hierarchy.territory || '');
        }
    }, [hierarchy]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        // Validate agentId before submitting - now uses string for URL
        const agentIdStr = String(agentId);
        if (!agentIdStr || agentIdStr === '0' || agentIdStr === '') {
            console.error('Invalid agent ID:', agentId);
            return;
        }

        // Keep parentId as string to preserve precision - backend will parse it
        const parentIdValue = parentId && parentId !== '' ? parentId : null;
        const commissionRateNum = parseFloat(commissionRate) / 100; // Convert from percentage

        try {
            await assignParent.mutateAsync({
                agent_id: agentIdStr,
                parent_id: parentIdValue,
                commission_rate: commissionRateNum,
                territory: territory,
            });

            // Refresh hierarchy data
            refetchHierarchy();

            // Call success callback
            if (onSuccess) {
                onSuccess();
            }
        } catch (error) {
            console.error('Failed to assign parent:', error);
        }
    };

    const isLoading = loadingHierarchy || loadingAgents || assignParent.isPending;

    return (
        <Card variant="outlined" sx={{ mt: 2 }}>
            <CardContent>
                <Typography variant="h6" gutterBottom>
                    {translate('resources.agents.hierarchy.manage', { _: 'Manage Hierarchy' })}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    {translate('resources.agents.hierarchy.description', {
                        _: 'Assign a parent agent and set commission rate for this agent'
                    })}
                </Typography>

                <form onSubmit={handleSubmit}>
                    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                        {/* Parent Agent Selection */}
                        <FormControl fullWidth disabled={isLoading}>
                            <InputLabel id="parent-agent-label">
                                {translate('resources.agents.fields.parent', { _: 'Parent Agent' })}
                            </InputLabel>
                            <Select
                                labelId="parent-agent-label"
                                value={parentId}
                                label={translate('resources.agents.fields.parent', { _: 'Parent Agent' })}
                                onChange={(e) => setParentId(e.target.value)}
                            >
                                <MenuItem value="">
                                    <em>{translate('resources.agents.hierarchy.no_parent', { _: 'Root Agent (No Parent)' })}</em>
                                </MenuItem>
                                {availableAgents.map((agent) => (
                                    <MenuItem key={String(agent.id)} value={String(agent.id)}>
                                        {agent.realname || agent.username} ({String(agent.id)})
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>

                        {/* Commission Rate */}
                        <TextField
                            fullWidth
                            label={translate('resources.agents.fields.commission_rate', { _: 'Commission Rate (%)' })}
                            type="number"
                            value={commissionRate}
                            onChange={(e) => setCommissionRate(e.target.value)}
                            disabled={isLoading}
                            inputProps={{
                                min: 0,
                                max: 100,
                                step: 0.01,
                            }}
                            InputProps={{
                                endAdornment: <InputAdornment position="end">%</InputAdornment>,
                            }}
                        />

                        {/* Territory */}
                        <TextField
                            fullWidth
                            label={translate('resources.agents.fields.territory', { _: 'Territory' })}
                            value={territory}
                            onChange={(e) => setTerritory(e.target.value)}
                            disabled={isLoading}
                            placeholder={translate('resources.agents.hierarchy.territory_placeholder', { _: 'e.g., North Region, City Center' })}
                        />

                        {/* Current Hierarchy Info */}
                        {hierarchy && hierarchy.level !== undefined && (
                            <Alert severity="info">
                                <Typography variant="body2">
                                    <strong>{translate('resources.agents.hierarchy.current_level', { _: 'Current Level' })}:</strong> {hierarchy.level}
                                    {hierarchy.parent_id && (
                                        <> • <strong>{translate('resources.agents.fields.parent', { _: 'Parent' })}:</strong> ID {hierarchy.parent_id}</>
                                    )}
                                </Typography>
                            </Alert>
                        )}

                        {/* Error Message */}
                        {assignParent.isError && (
                            <Alert severity="error">
                                {translate('resources.agents.hierarchy.error', { _: 'Failed to update hierarchy' })}
                            </Alert>
                        )}

                        {/* Success Message */}
                        {assignParent.isSuccess && (
                            <Alert severity="success">
                                {translate('resources.agents.hierarchy.success', { _: 'Hierarchy updated successfully' })}
                            </Alert>
                        )}

                        {/* Submit Button */}
                        <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 1 }}>
                            <Button
                                type="submit"
                                variant="contained"
                                disabled={isLoading}
                                startIcon={isLoading ? <CircularProgress size={20} /> : <SaveIcon />}
                            >
                                {translate('resources.agents.hierarchy.save', { _: 'Save Hierarchy' })}
                            </Button>
                        </Box>
                    </Box>
                </form>
            </CardContent>
        </Card>
    );
};

export default AgentHierarchyForm;
