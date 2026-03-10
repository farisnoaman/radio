# Agent Hierarchy Tree Component Design

## Overview
Implement a React component to visualize the agent hierarchy tree, integrating it into the existing agents list view.

## Goals
- Display agent hierarchy in a tree structure.
- Support expand/collapse functionality.
- Integrate seamlessly into both card (mobile) and list (desktop) views.

## Architecture

### 1. New Component: `AgentHierarchyTree`
**Location:** `/home/faris/Downloads/toughradius/toughradius/web/src/resources/AgentHierarchyTree.tsx`

**Props:**
- `agentId: number` (required): The ID of the agent whose hierarchy to display.

**State:**
- `expandedNodes: Set<number>`: Tracks which nodes are expanded.

**Data Fetching:**
- Uses `useApiQuery` hook.
- Endpoint: `/api/v1/agents/:id/hierarchy-tree`.
- Query Key: `['agent-hierarchy-tree', agentId]`.

**Structure:**
- Recursive component `HierarchyNode` to render each node in the tree.
- Each node displays:
  - Agent Name
  - Level
  - Commission Rate
  - Territory
  - Expand/Collapse toggle (if children exist)

**UI Components (Material-UI):**
- `Card`: Container for the tree.
- `Box`: Layout and nesting.
- `Typography`: Text display.
- `IconButton`: Expand/Collapse toggle.
- `Collapse`: Animate expand/collapse.
- `ChevronRight` / `ExpandMore` icons: Visual indicators.

**Styling:**
- Nested indentation for visual hierarchy.
- Visual cues for different levels (e.g., border-left color).

### 2. Integration in `agents.tsx`
**Location:** `/home/faris/Downloads/toughradius/toughradius/web/src/resources/agents.tsx`

**Changes to `AgentGrid` (Card View):**
1. Add a "Hierarchy" button in `CardActions`.
2. Add state `showHierarchy: boolean` to toggle visibility.
3. Conditionally render `AgentHierarchyTree` inside a `Collapse` component below the card content when `showHierarchy` is true.

**Changes to `AgentList` (List View):**
1. Add a custom column to the `Datagrid`.
2. The column contains a "View Hierarchy" button.
3. Clicking the button opens a `Dialog` (modal) containing the `AgentHierarchyTree`.

## Data Flow
1. User clicks "View Hierarchy" (or "Hierarchy" button in card view).
2. `AgentHierarchyTree` component is mounted/visible.
3. `useApiQuery` fetches data from `/api/v1/agents/:id/hierarchy-tree`.
4. Data is rendered recursively using the `HierarchyNode` component.
5. User can expand/collapse nodes by clicking the toggle button.

## Error Handling & Loading States
- **Loading:** Display a loading spinner or skeleton while data is being fetched.
- **Error:** Display an error message if the API request fails.
- **Empty:** Display a message if the agent has no sub-agents (empty hierarchy).

## Testing Strategy
- Verify the component renders correctly with mock data.
- Verify API calls are made with the correct parameters.
- Verify expand/collapse functionality works as expected.
- Verify integration in both `AgentGrid` and `AgentList`.

## Implementation Steps
1. Create `AgentHierarchyTree.tsx` component.
2. Implement recursive rendering logic.
3. Add expand/collapse state management.
4. Integrate into `AgentGrid` view.
5. Integrate into `AgentList` view (with Dialog).
6. Test and verify.
