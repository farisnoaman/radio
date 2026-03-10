import { useState } from 'react';
import { Card, CardContent, Typography, Box, IconButton, Collapse } from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import { useApiQuery } from '../hooks/useApiQuery';
import { useParams } from 'react-router-dom';

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

const HierarchyTreeNode = ({ node, depth = 0 }: { node: HierarchyNode; depth?: number }) => {
  const [expanded, setExpanded] = useState(depth === 0);
  const hasChildren = node.children && node.children.length > 0;

  return (
    <Box sx={{ ml: depth * 2 }}>
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          py: 1,
          px: 1,
          borderBottom: '1px solid',
          borderColor: 'divider',
        }}
      >
        {hasChildren ? (
          <IconButton
            size="small"
            onClick={() => setExpanded(!expanded)}
            sx={{ mr: 1 }}
          >
            {expanded ? <ExpandMoreIcon /> : <ChevronRightIcon />}
          </IconButton>
        ) : (
          <Box sx={{ width: 32 }} />
        )}
        <Box sx={{ flexGrow: 1 }}>
          <Typography variant="body2" sx={{ fontWeight: 'bold' }}>
            {node.name}
          </Typography>
          <Typography variant="caption" color="text.secondary">
            @{node.username} • Level {node.level} • {node.territory || 'No territory'}
          </Typography>
        </Box>
        <Box
          sx={{
            bgcolor: 'primary.light',
            color: 'white',
            px: 1,
            py: 0.5,
            borderRadius: 1,
            fontSize: '0.75rem',
          }}
        >
          {Math.round(node.commission_rate * 100)}%
        </Box>
      </Box>
      <Collapse in={expanded}>
        {hasChildren && (
          <Box sx={{ pl: 2 }}>
            {node.children.map((child) => (
              <HierarchyTreeNode
                key={child.id}
                node={child}
                depth={depth + 1}
              />
            ))}
          </Box>
        )}
      </Collapse>
    </Box>
  );
};

export const AgentHierarchyTree = () => {
  const { id } = useParams<{ id: string }>();
  const { data, isLoading, error } = useApiQuery<HierarchyNode>({
    path: `/agents/${id}/hierarchy-tree`,
    queryKey: ['agent-hierarchy-tree', id],
  });

  if (isLoading) {
    return (
      <Card>
        <CardContent>
          <Typography>Loading hierarchy...</Typography>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent>
          <Typography color="error">Failed to load hierarchy</Typography>
        </CardContent>
      </Card>
    );
  }

  if (!data) {
    return (
      <Card>
        <CardContent>
          <Typography>No hierarchy data available</Typography>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card sx={{ mb: 2 }}>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          Agent Hierarchy
        </Typography>
        <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1 }}>
          <HierarchyTreeNode node={data} />
        </Box>
      </CardContent>
    </Card>
  );
};
