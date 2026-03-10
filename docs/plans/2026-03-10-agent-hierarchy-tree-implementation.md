# Agent Hierarchy Tree Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the `AgentHierarchyTree` component and integrate it into the agents list view.

**Architecture:** Create a recursive React component to fetch and render the agent hierarchy tree. Integrate it into the existing `AgentGrid` (card view) and `AgentList` (list view) components.

**Tech Stack:** React, Material-UI, React-Admin, TanStack Query (react-query).

---

### Task 1: Create AgentHierarchyTree Component

**Files:**
- Create: `/home/faris/Downloads/toughradius/toughradius/web/src/resources/AgentHierarchyTree.tsx`

**Step 1: Write the component file**

```tsx
import React, { useState } from 'react';
import {
    Box,
    Card,
    CardContent,
    Typography,
    IconButton,
    Collapse,
    CircularProgress,
    Alert,
} from '@mui/material';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import { useApiQuery } from '../hooks/useApiQuery';

interface HierarchyNode {
    id: number;
    name: string;
    username: string;
    email: string;
    level: number;
    territory: string;
    commission_rate: number;
    children: HierarchyNode[];
}

interface AgentHierarchyTreeProps {
    agentId: number;
}

const HierarchyNodeComponent: React.FC<{ node: HierarchyNode; depth: number }> = ({ node, depth }) => {
    const [expanded, setExpanded] = useState(false);
    const hasChildren = node.children && node.children.length > 0;

    return (
        <Box sx={{ ml: depth * 2 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', py: 1 }}>
                {hasChildren && (
                    <IconButton size="small" onClick={() => setExpanded(!expanded)}>
                        {expanded ? <ExpandMoreIcon /> : <ChevronRightIcon />}
                    </IconButton>
                )}
                {!hasChildren && <Box sx={{ width: 40 }} />}
                <Box sx={{ flex: 1 }}>
                    <Typography variant="body2" sx={{ fontWeight: 'bold' }}>
                        {node.name}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                        Level: {node.level} | Rate: {(node.commission_rate * 100).toFixed(1)}% | Territory: {node.territory}
                    </Typography>
                </Box>
            </Box>
            {hasChildren && (
                <Collapse in={expanded}>
                    <Box sx={{ pl: 2 }}>
                        {node.children.map((child) => (
                            <HierarchyNodeComponent key={child.id} node={child} depth={depth + 1} />
                        ))}
                    </Box>
                </Collapse>
            )}
        </Box>
    );
};

export const AgentHierarchyTree: React.FC<AgentHierarchyTreeProps> = ({ agentId }) => {
    const { data, isLoading, error } = useApiQuery<HierarchyNode>({
        path: `/api/v1/agents/${agentId}/hierarchy-tree`,
        queryKey: ['agent-hierarchy-tree', agentId],
    });

    if (isLoading) {
        return (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
                <CircularProgress size={24} />
            </Box>
        );
    }

    if (error) {
        return (
            <Alert severity="error" sx={{ m: 1 }}>
                Failed to load hierarchy tree
            </Alert>
        );
    }

    if (!data) {
        return (
            <Alert severity="info" sx={{ m: 1 }}>
                No hierarchy data available
            </Alert>
        );
    }

    return (
        <Card variant="outlined" sx={{ mt: 1 }}>
            <CardContent sx={{ p: 1, '&:last-child': { pb: 1 } }}>
                <HierarchyNodeComponent node={data} depth={0} />
            </CardContent>
        </Card>
    );
};
```

**Step 2: Verify file exists**

Run: `ls /home/faris/Downloads/toughradius/toughradius/web/src/resources/AgentHierarchyTree.tsx`
Expected: File exists

**Step 3: Commit**

```bash
git add /home/faris/Downloads/toughradius/toughradius/web/src/resources/AgentHierarchyTree.tsx
git commit -m "feat: Create AgentHierarchyTree component"
```

---

### Task 2: Integrate into AgentGrid (Card View)

**Files:**
- Modify: `/home/faris/Downloads/toughradius/toughradius/web/src/resources/agents.tsx`

**Step 1: Read the current agents.tsx file**

Read `/home/faris/Downloads/toughradius/toughradius/web/src/resources/agents.tsx` to understand the exact structure.

**Step 2: Add import for AgentHierarchyTree and Collapse**

Add to imports:
```tsx
import { Box, Card, CardContent, Typography, useMediaQuery, Theme, Avatar, Chip, Collapse } from '@mui/material';
import AgentHierarchyTree from './AgentHierarchyTree';
```

**Step 3: Modify AgentGrid component**

Inside the `AgentGrid` component, inside the `CardActions` section:
1. Add state `const [showHierarchy, setShowHierarchy] = useState(false);`
2. Add a button to toggle hierarchy:
```tsx
<Button 
    label="Hierarchy" 
    onClick={() => setShowHierarchy(!showHierarchy)} 
    startIcon={showHierarchy ? <ExpandMoreIcon /> : <ChevronRightIcon />}
>
    Hierarchy
</Button>
```
3. Add the `AgentHierarchyTree` component inside a `Collapse` below the card content:
```tsx
<Collapse in={showHierarchy}>
    <AgentHierarchyTree agentId={record.id} />
</Collapse>
```

**Step 4: Update CardActions to include the new button**

Ensure the button is added to the `CardActions` component.

**Step 5: Run lint/typecheck**

Run: `npm run lint` (or equivalent) in `/home/faris/Downloads/toughradius/toughradius/web`
Expected: No errors

**Step 6: Commit**

```bash
git add /home/faris/Downloads/toughradius/toughradius/web/src/resources/agents.tsx
git commit -m "feat: Integrate AgentHierarchyTree into AgentGrid"
```

---

### Task 3: Integrate into AgentList (List View)

**Files:**
- Modify: `/home/faris/Downloads/toughradius/toughradius/web/src/resources/agents.tsx`

**Step 1: Add imports for Dialog and Button**

Add to imports:
```tsx
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button as MuiButton,
} from '@mui/material';
```

**Step 2: Create a HierarchyButton component**

Add this component inside `agents.tsx`:
```tsx
const HierarchyButton = () => {
    const record = useRecordContext();
    const [open, setOpen] = useState(false);

    if (!record) return null;

    return (
        <>
            <Button label="View Hierarchy" onClick={() => setOpen(true)} />
            <Dialog open={open} onClose={() => setOpen(false)} maxWidth="md" fullWidth>
                <DialogTitle>Hierarchy Tree - {record.username}</DialogTitle>
                <DialogContent>
                    <AgentHierarchyTree agentId={record.id} />
                </DialogContent>
                <DialogActions>
                    <MuiButton onClick={() => setOpen(false)}>Close</MuiButton>
                </DialogActions>
            </Dialog>
        </>
    );
};
```

**Step 3: Add HierarchyButton to Datagrid**

In the `AgentList` component, inside the `<Datagrid>`:
Add: `<HierarchyButton />`

**Step 4: Run lint/typecheck**

Run: `npm run lint` in `/home/faris/Downloads/toughradius/toughradius/web`
Expected: No errors

**Step 5: Commit**

```bash
git add /home/faris/Downloads/toughradius/toughradius/web/src/resources/agents.tsx
git commit -m "feat: Integrate AgentHierarchyTree into AgentList"
```

---

### Task 4: Verify Implementation

**Files:**
- Test: Manual verification in browser

**Step 1: Start the development server**

Run: `npm run dev` in `/home/faris/Downloads/toughradius/toughradius/web`

**Step 2: Verify AgentGrid integration**
1. Navigate to Agents list (mobile view or narrow window).
2. Click "Hierarchy" button on a card.
3. Verify tree expands/collapses.
4. Verify data loads correctly.

**Step 3: Verify AgentList integration**
1. Navigate to Agents list (desktop view).
2. Click "View Hierarchy" button in the datagrid.
3. Verify modal opens.
4. Verify tree renders correctly inside modal.

**Step 4: Commit verification (optional)**

If manual testing is successful, commit any final adjustments.

```bash
git add .
git commit -m "chore: Verify AgentHierarchyTree implementation"
```
