import {
  Card,
  CardContent,
  Typography,
  Box,
  Skeleton,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { useApiQuery } from '../../hooks/useApiQuery';

interface AgentMetrics {
  total_agents: number;
  total_batches: number;
  revenue: number;
  mrr: number;
}

export const AgentFinancialWidget = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const { data, isLoading } = useApiQuery<AgentMetrics>({
    path: '/api/v1/reporting/agents',
    queryKey: ['reporting', 'agents'],
    enabled: true,
  });

  return (
    <Card>
      <CardContent sx={{ py: isMobile ? 1.5 : 2 }}>
        <Typography
          variant="h6"
          gutterBottom
          sx={{ fontSize: isMobile ? '1rem' : undefined }}
        >
          Agent Financial Summary
        </Typography>
        {isLoading ? (
          <Box>
            <Skeleton variant="text" width="40%" />
            <Skeleton variant="text" width="30%" height={40} />
          </Box>
        ) : (
          <Box
            sx={{
              display: 'flex',
              gap: isMobile ? 2 : 3,
              flexWrap: 'wrap',
            }}
          >
            <Box>
              <Typography variant="caption" color="text.secondary">
                Total Agents
              </Typography>
              <Typography variant={isMobile ? 'h5' : 'h4'}>
                {data?.total_agents ?? 0}
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary">
                Revenue
              </Typography>
              <Typography variant={isMobile ? 'h5' : 'h4'}>
                ${(data?.revenue ?? 0).toFixed(2)}
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary">
                MRR
              </Typography>
              <Typography variant={isMobile ? 'h5' : 'h4'}>
                ${(data?.mrr ?? 0).toFixed(2)}
              </Typography>
            </Box>
          </Box>
        )}
      </CardContent>
    </Card>
  );
};
