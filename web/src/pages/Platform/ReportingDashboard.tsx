import { useState } from 'react';
import {
  Box,
  Typography,
  Tabs,
  Tab,
  Button,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { Grid } from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import PeopleIcon from '@mui/icons-material/People';
import SessionIcon from '@mui/icons-material/SensorsOutlined';
import DataUsageIcon from '@mui/icons-material/DataUsage';
import { useApiQuery } from '../../hooks/useApiQuery';
import { SummaryCard } from '../../components/Platform/SummaryCard';
import { NetworkStatusWidget } from '../../components/Platform/NetworkStatusWidget';
import { AgentFinancialWidget } from '../../components/Platform/AgentFinancialWidget';
import { IssuesReporterWidget } from '../../components/Platform/IssuesReporterWidget';
import { FraudAlertWidget } from '../../components/Platform/FraudAlertWidget';

interface ReportingSummary {
  period: string;
  start_date: string;
  end_date: string;
  users: {
    total_users: number;
    active_users: number;
    new_monthly_users: number;
    new_voucher_users: number;
  };
  sessions: {
    total_sessions: number;
    active_sessions: number;
  };
  data: {
    monthly_data_used_gb: number;
    voucher_data_used_gb: number;
  };
  network: {
    nodes: { active: number; total: number };
    servers: { active: number; total: number };
  };
  agents: {
    total_agents: number;
    total_batches: number;
    revenue: number;
    mrr: number;
  };
  issues: {
    device_issues: number;
    network_issues: number;
  };
}

export default function ReportingDashboard() {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const [period, setPeriod] = useState('daily');

  const { data: summary, isLoading } = useApiQuery<ReportingSummary>({
    path: `/api/v1/reporting/summary?period=${period}`,
    queryKey: ['reporting', 'summary', period],
    enabled: true,
  });

  const handleExport = () => {
    window.open(`/api/v1/reporting/export?period=${period}`, '_blank');
  };

  return (
    <Box sx={{ p: isMobile ? 1 : 3 }}>
      <Box
        sx={{
          display: 'flex',
          flexDirection: isMobile ? 'column' : 'row',
          justifyContent: 'space-between',
          alignItems: isMobile ? 'stretch' : 'center',
          gap: isMobile ? 1 : 2,
          mb: 3,
        }}
      >
        <Typography
          variant={isMobile ? 'h6' : 'h5'}
          sx={{ fontWeight: 'bold' }}
        >
          Reporting Dashboard
        </Typography>
        <Box
          sx={{
            display: 'flex',
            flexDirection: isMobile ? 'column' : 'row',
            alignItems: 'center',
            gap: 1,
          }}
        >
          <Tabs
            value={period}
            onChange={(_, v) => setPeriod(v)}
            variant={isMobile ? 'scrollable' : 'standard'}
            scrollButtons={isMobile ? 'auto' : false}
          >
            <Tab label="Daily" value="daily" />
            <Tab label="Weekly" value="weekly" />
            <Tab label="Monthly" value="monthly" />
          </Tabs>
          <Button
            startIcon={<DownloadIcon />}
            variant="outlined"
            onClick={handleExport}
            size={isMobile ? 'small' : 'medium'}
          >
            Export
          </Button>
        </Box>
      </Box>

      <Grid container spacing={isMobile ? 1.5 : 2} sx={{ mb: isMobile ? 2 : 3 }}>
        <Grid size={{ xs: 6, sm: 6, md: 3 }}>
          <SummaryCard
            title="Total Users"
            value={summary?.users?.total_users ?? 0}
            loading={isLoading}
            color="#1976d2"
            icon={!isMobile && <PeopleIcon color="primary" />}
          />
        </Grid>
        <Grid size={{ xs: 6, sm: 6, md: 3 }}>
          <SummaryCard
            title="New Monthly"
            value={summary?.users?.new_monthly_users ?? 0}
            loading={isLoading}
            color="#2e7d32"
            icon={!isMobile && <PeopleIcon color="success" />}
          />
        </Grid>
        <Grid size={{ xs: 6, sm: 6, md: 3 }}>
          <SummaryCard
            title="Active Sessions"
            value={summary?.sessions?.active_sessions ?? 0}
            loading={isLoading}
            color="#ed6c02"
            icon={!isMobile && <SessionIcon color="warning" />}
          />
        </Grid>
        <Grid size={{ xs: 6, sm: 6, md: 3 }}>
          <SummaryCard
            title="Data (GB)"
            value={(summary?.data?.monthly_data_used_gb ?? 0).toFixed(1)}
            loading={isLoading}
            color="#9c27b0"
            icon={!isMobile && <DataUsageIcon color="secondary" />}
          />
        </Grid>
      </Grid>

      <Box sx={{ mb: isMobile ? 2 : 3 }}>
        <NetworkStatusWidget />
      </Box>

      <Box sx={{ mb: isMobile ? 2 : 3 }}>
        <AgentFinancialWidget />
      </Box>

      <Grid container spacing={isMobile ? 1.5 : 2}>
        <Grid size={{ xs: 12, sm: 12, md: 6 }}>
          <IssuesReporterWidget />
        </Grid>
        <Grid size={{ xs: 12, sm: 12, md: 6 }}>
          <FraudAlertWidget />
        </Grid>
      </Grid>
    </Box>
  );
}
