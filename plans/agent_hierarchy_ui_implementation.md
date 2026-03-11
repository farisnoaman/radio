# Agent Hierarchy UI Implementation Plan

## Overview
Add UI components to allow admins to assign agents to parent agents and set commission rates directly from the web interface.

## Backend API Analysis

### Existing Endpoints

| Endpoint | Method | Description | Request Body |
|----------|--------|-------------|--------------|
| `/api/v1/agents/:id/assign-parent` | POST | Assign agent to parent | `{agent_id, parent_id?, commission_rate?, territory?}` |
| `/api/v1/agents/:id/commission-rate` | PUT | Update commission rate | `{commission_rate: float}` |
| `/api/v1/agents/:id/hierarchy` | GET | Get agent's hierarchy info | - |
| `/api/v1/agents/roots` | GET | List root-level agents | - |

### Permission Check
- Both endpoints require `currentUser.Level == "super" || currentUser.Level == "admin"`
- Regular agents cannot access these endpoints

---

## Implementation Steps

### Step 1: Create API Hooks (web/src/hooks/)

Create `useAssignParent.ts` and `useUpdateCommissionRate.ts`:

```typescript
// useAssignParent.ts
export const useAssignParent = () => {
  const httpClient = useHttpClient();
  return useMutation({
    mutationFn: (data: { agent_id: number; parent_id?: number; commission_rate?: number; territory?: string }) =>
      httpClient(`/agents/${data.agent_id}/assign-parent`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  });
};

// useUpdateCommissionRate.ts
export const useUpdateCommissionRate = () => {
  const httpClient = useHttpClient();
  return useMutation({
    mutationFn: (data: { agent_id: number; commission_rate: number }) =>
      httpClient(`/agents/${data.agent_id}/commission-rate`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
  });
};
```

### Step 2: Create AgentHierarchyForm Component (web/src/resources/)

Create `AgentHierarchyForm.tsx` with:

```typescript
interface AgentHierarchyFormProps {
  agentId: number;
  currentHierarchy?: {
    parent_id?: number;
    commission_rate: number;
    territory: string;
    level: number;
  };
  onSuccess: () => void;
}
```

**Form Fields:**
- Parent Agent (Select dropdown with search) - fetch from `/api/v1/agents/roots`
- Commission Rate (Number input, 0-100%)
- Territory (Text input)

**Features:**
- Pre-populate with current values
- Validation (commission rate 0-100)
- Loading states
- Success/Error notifications

### Step 3: Integrate into AgentShow Page (web/src/resources/agents.tsx)

Add the form to the AgentShow component after the AgentHierarchyTree:

```typescript
export const AgentShow = (props: any) => (
  <Show {...props}>
    <SimpleShowLayout>
      {/* Existing fields */}
      <AgentHierarchyTree />
      
      {/* NEW: Hierarchy Management Form */}
      <Card sx={{ mt: 2 }}>
        <CardContent>
          <Typography variant="h6">Manage Hierarchy</Typography>
          <AgentHierarchyForm agentId={record.id} onSuccess={refresh} />
        </CardContent>
      </Card>
    </SimpleShowLayout>
  </Show>
);
```

### Step 4: Add i18n Translations

**en-US.ts:**
```typescript
'resources.agents.fields.parent': 'Parent Agent',
'resources.agents.fields.commission_rate': 'Commission Rate (%)',
'resources.agents.fields.territory': 'Territory',
'resources.agents.hierarchy.manage': 'Manage Hierarchy',
'resources.agents.hierarchy.assign': 'Assign Parent',
'resources.agents.hierarchy.update': 'Update Commission',
```

**zh-CN.ts:**
```typescript
'resources.agents.fields.parent': '上级代理',
'resources.agents.fields.commission_rate': '佣金比例 (%)',
'resources.agents.fields.territory': '区域',
'resources.agents.hierarchy.manage': '管理层级',
'resources.agents.hierarchy.assign': '分配上级',
'resources.agents.hierarchy.update': '更新佣金',
```

---

## Mermaid Diagram: Component Flow

```mermaid
graph TD
    A[Admin visits Agent Detail] --> B[AgentShow Component]
    B --> C[AgentHierarchyTree - Read Only]
    B --> D[AgentHierarchyForm - Edit]
    D --> E[Select Parent Agent]
    E --> F[/api/v1/agents/roots]
    D --> G[Enter Commission Rate]
    D --> H[Enter Territory]
    G --> I[Submit]
    I --> J{Update Type}
    J -->|Assign Parent| K[POST /api/v1/agents/:id/assign-parent]
    J -->|Update Rate| L[PUT /api/v1/agents/:id/commission-rate]
    K --> M[Success: Refresh + Notify]
    L --> M
    M --> N[Refresh AgentHierarchyTree]
```

---

## Files to Modify/Create

| File | Action | Description |
|------|--------|-------------|
| `web/src/hooks/useAssignParent.ts` | CREATE | Mutation hook for assign-parent |
| `web/src/hooks/useUpdateCommissionRate.ts` | CREATE | Mutation hook for commission-rate |
| `web/src/resources/AgentHierarchyForm.tsx` | CREATE | Form component |
| `web/src/resources/agents.tsx` | MODIFY | Add form to AgentShow |
| `web/src/i18n/en-US.ts` | MODIFY | Add translations |
| `web/src/i18n/zh-CN.ts` | MODIFY | Add translations |

---

## Testing Checklist

- [ ] Admin can see "Manage Hierarchy" form on Agent detail page
- [ ] Form pre-populates with current hierarchy values
- [ ] Admin can select a parent agent from dropdown
- [ ] Admin can set commission rate (0-100%)
- [ ] Admin can set territory
- [ ] Submitting assign-parent updates hierarchy and shows success
- [ ] Submitting commission-rate updates and shows success
- [ ] Error handling works for invalid inputs
- [ ] Non-admin users cannot access the form (backend check)
