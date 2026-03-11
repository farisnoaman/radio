import { useMutation, useQuery } from '@tanstack/react-query';
import { httpClient } from '../utils/apiClient';

// Request/Response types for Agent Hierarchy
export interface AssignParentRequest {
    agent_id: number | string;
    parent_id?: number | string | null;
    commission_rate?: number;
    territory?: string;
}

export interface UpdateCommissionRateRequest {
    agent_id: number;
    commission_rate: number;
}

export interface AgentHierarchy {
    agent_id: number;
    parent_id?: number;
    level: number;
    territory: string;
    commission_rate: number;
    status: string;
    created_at: string;
    updated_at: string;
    // Joined fields
    parent_name?: string;
    parent_username?: string;
}

export interface RootAgent {
    id: number;
    username: string;
    realname: string;
    email: string;
}

/**
 * Hook to assign an agent to a parent agent
 * POST /api/v1/agents/:id/assign-parent
 */
export const useAssignParent = () => {
    return useMutation({
        mutationFn: async (data: AssignParentRequest) => {
            // Use agent_id from URL parameter (not from body) to avoid JavaScript precision issues
            const response = await httpClient(`/agents/${data.agent_id}/assign-parent`, {
                method: 'POST',
                body: JSON.stringify({
                    // agent_id is now read from URL, not body
                    parent_id: data.parent_id != null ? String(data.parent_id) : null,
                    commission_rate: data.commission_rate,
                    territory: data.territory,
                }),
            });
            return response.json as Promise<AgentHierarchy>;
        },
    });
};

/**
 * Hook to update commission rate for an agent
 * PUT /api/v1/agents/:id/commission-rate
 */
export const useUpdateCommissionRate = () => {
    return useMutation({
        mutationFn: async (data: UpdateCommissionRateRequest) => {
            const response = await httpClient(`/agents/${data.agent_id}/commission-rate`, {
                method: 'PUT',
                body: JSON.stringify({
                    commission_rate: data.commission_rate,
                }),
            });
            return response.json as Promise<AgentHierarchy>;
        },
    });
};

/**
 * Hook to get an agent's hierarchy information
 * GET /api/v1/agents/:id/hierarchy
 */
export const useAgentHierarchy = (agentId: number | string) => {
    return useQuery({
        queryKey: ['agent-hierarchy', agentId],
        queryFn: async () => {
            const response = await httpClient(`/agents/${agentId}/hierarchy`, {
                method: 'GET',
            });
            return response.json as Promise<AgentHierarchy>;
        },
        enabled: !!agentId,
    });
};
