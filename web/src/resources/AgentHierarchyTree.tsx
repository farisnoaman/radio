import { useState } from 'react';
import { Card, CardContent, Typography, Box, IconButton, Collapse } from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import { useApiQuery } from '../hooks/useApiQuery';
import { useParams } from 'react-router-dom';
import { useTranslate, useLocale } from 'react-admin';

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
  const translate = useTranslate();
  const locale = useLocale();
  const isRTL = locale === 'ar';
  const hasChildren = node.children && node.children.length > 0;

  return (
    <Box sx={{ [isRTL ? 'mr' : 'ml']: depth * 2 }}>
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          flexDirection: isRTL ? 'row-reverse' : 'row',
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
            sx={{ [isRTL ? 'ml' : 'mr']: 1 }}
          >
            {expanded ? <ExpandMoreIcon /> : (isRTL ? <ChevronLeftIcon /> : <ChevronRightIcon />)}
          </IconButton>
        ) : (
          <Box sx={{ width: 32 }} />
        )}
        <Box sx={{ flexGrow: 1, textAlign: isRTL ? 'right' : 'left' }}>
          <Typography variant="body2" sx={{ fontWeight: 'bold' }}>
            {node.name}
          </Typography>
          <Typography variant="caption" color="text.secondary">
            @{node.username} • {translate('resources.agents.fields.level')} {node.level} • {node.territory || translate('resources.agents.hierarchy.no_territory')}
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
            [isRTL ? 'mr' : 'ml']: 1
          }}
        >
          {Math.round(node.commission_rate * 100)}%
        </Box>
      </Box>
      <Collapse in={expanded}>
        {hasChildren && (
          <Box sx={{ [isRTL ? 'pr' : 'pl']: 2 }}>
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
  const translate = useTranslate();
  const locale = useLocale();
  const isRTL = locale === 'ar';

  const { data, isLoading, error } = useApiQuery<HierarchyNode>({
    path: `/agents/${id}/hierarchy-tree`,
    queryKey: ['agent-hierarchy-tree', id],
  });

  if (isLoading) {
    return (
      <Card dir={isRTL ? 'rtl' : 'ltr'}>
        <CardContent>
          <Typography>{translate('resources.agents.helpers.loading_hierarchy')}</Typography>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card dir={isRTL ? 'rtl' : 'ltr'}>
        <CardContent>
          <Typography color="error">{translate('resources.agents.hierarchy.error_loading')}</Typography>
        </CardContent>
      </Card>
    );
  }

  if (!data) {
    return (
      <Card dir={isRTL ? 'rtl' : 'ltr'}>
        <CardContent>
          <Typography>{translate('resources.agents.hierarchy.no_data')}</Typography>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card sx={{ mb: 2 }} dir={isRTL ? 'rtl' : 'ltr'}>
      <CardContent>
        <Typography variant="h6" gutterBottom sx={{ textAlign: isRTL ? 'right' : 'left' }}>
          {translate('resources.agents.hierarchy.title')}
        </Typography>
        <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1 }}>
          <HierarchyTreeNode node={data} />
        </Box>
      </CardContent>
    </Card>
  );
};
