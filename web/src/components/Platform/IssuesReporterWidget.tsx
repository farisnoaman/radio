import {
  Card,
  CardContent,
  Typography,
  Box,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Chip,
  Skeleton,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { useApiQuery } from '../../hooks/useApiQuery';

interface Issue {
  id: number;
  device_type: string;
  device_name: string;
  issue_type: string;
  status: string;
  created_at: string;
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'open': return 'error';
    case 'resolved': return 'success';
    case 'ignored': return 'default';
    default: return 'default';
  }
};

export const IssuesReporterWidget = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const { data: issues, isLoading } = useApiQuery<Issue[]>({
    path: '/api/v1/reporting/issues',
    queryKey: ['reporting', 'issues'],
    enabled: true,
  });

  if (isLoading) {
    return (
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>Device &amp; Network Issues</Typography>
          <Box>
            <Skeleton variant="text" />
            <Skeleton variant="text" />
            <Skeleton variant="text" />
          </Box>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent sx={{ py: isMobile ? 1.5 : 2 }}>
        <Typography
          variant="h6"
          gutterBottom
          sx={{ fontSize: isMobile ? '1rem' : undefined }}
        >
          Device &amp; Network Issues
        </Typography>
        {!issues || issues.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            No open issues
          </Typography>
        ) : (
          <Box sx={{ overflowX: 'auto' }}>
            <Table size="small" sx={{ minWidth: isMobile ? 400 : undefined }}>
              <TableHead>
                <TableRow>
                  <TableCell>Device</TableCell>
                  {!isMobile && <TableCell>Type</TableCell>}
                  <TableCell>Issue</TableCell>
                  <TableCell>Status</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {issues.slice(0, 10).map((issue) => (
                  <TableRow key={issue.id}>
                    <TableCell>{issue.device_name}</TableCell>
                    {!isMobile && <TableCell>{issue.device_type}</TableCell>}
                    <TableCell>{issue.issue_type}</TableCell>
                    <TableCell>
                      <Chip
                        label={issue.status}
                        color={getStatusColor(issue.status)}
                        size="small"
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Box>
        )}
      </CardContent>
    </Card>
  );
};
