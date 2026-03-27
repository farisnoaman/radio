import {
  List,
  Datagrid,
  TextField,
  NumberField,
  DateField,
  FunctionField,
  useListContext,
  useTranslate,
  TopToolbar,
  CreateButton,
  ExportButton,
  BulkActionsToolbar,
  BulkDeleteButton,
} from 'react-admin';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Stack,
  useMediaQuery,
  useTheme,
  Chip,
  Divider,
} from '@mui/material';
import {
  WifiOff,
  CheckCircle,
  Warning,
  Router,
  Speed,
  Memory,
} from '@mui/icons-material';
import { StatusBadge, MetricCard } from '../../components/saas';

const DeviceHealthAside = () => {
  const { data } = useListContext();
  const translate = useTranslate();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  if (isMobile) return null;

  const onlineCount = data?.filter((d: any) => d.status === 'online').length || 0;
  const offlineCount = data?.filter((d: any) => d.status === 'offline').length || 0;
  const warningCount = data?.filter((d: any) => d.status === 'warning').length || 0;

  return (
    <Box sx={{ width: 300, ml: 2, mb: 2 }}>
      <Stack spacing={2}>
        <MetricCard
          title={translate('monitoring.aside.total_devices')}
          value={data?.length || 0}
          icon={<WifiOff fontSize="small" />}
          variant="compact"
        />

        <Card
          sx={{
            background: 'linear-gradient(135deg, rgba(16, 185, 129, 0.08) 0%, rgba(16, 185, 129, 0.02) 100%)',
            border: '1px solid rgba(16, 185, 129, 0.2)',
            borderRadius: 2,
          }}
        >
          <CardContent sx={{ p: 2 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 1 }}>
              <CheckCircle sx={{ color: '#10b981', fontSize: 20 }} />
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('monitoring.online')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#10b981', fontWeight: 700 }}>
              {onlineCount}
            </Typography>
          </CardContent>
        </Card>

        <Card
          sx={{
            background: 'linear-gradient(135deg, rgba(239, 68, 68, 0.08) 0%, rgba(239, 68, 68, 0.02) 100%)',
            border: '1px solid rgba(239, 68, 68, 0.2)',
            borderRadius: 2,
          }}
        >
          <CardContent sx={{ p: 2 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 1 }}>
              <WifiOff sx={{ color: '#ef4444', fontSize: 20 }} />
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('monitoring.offline')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#ef4444', fontWeight: 700 }}>
              {offlineCount}
            </Typography>
          </CardContent>
        </Card>

        <Card
          sx={{
            background: 'linear-gradient(135deg, rgba(245, 158, 11, 0.08) 0%, rgba(245, 158, 11, 0.02) 100%)',
            border: '1px solid rgba(245, 158, 11, 0.2)',
            borderRadius: 2,
          }}
        >
          <CardContent sx={{ p: 2 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 1 }}>
              <Warning sx={{ color: '#f59e0b', fontSize: 20 }} />
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('monitoring.warning')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#f59e0b', fontWeight: 700 }}>
              {warningCount}
            </Typography>
          </CardContent>
        </Card>
      </Stack>
    </Box>
  );
};

const DeviceHealthFilters = () => null;

const DeviceHealthActions = () => {
  const translate = useTranslate();
  return (
    <TopToolbar>
      <CreateButton label={translate('monitoring.device_health_monitoring')} />
      <ExportButton />
    </TopToolbar>
  );
};

const deviceHealthListStyles = {
  '& .RaDatagrid-headerCell': {
    fontWeight: 600,
    backgroundColor: 'rgba(30, 58, 138, 0.04)',
    color: '#1e3a8a',
  },
  '& .RaDatagrid-row': {
    transition: 'all 0.2s ease',
    '&:hover': {
      backgroundColor: 'rgba(30, 58, 138, 0.04)',
    },
  },
};

const DeviceCard = ({ record }: { record: any }) => {
  const translate = useTranslate();

  const statusColor = record.status === 'online' ? '#10b981' : record.status === 'warning' ? '#f59e0b' : '#ef4444';
  const StatusIcon = record.status === 'online' ? CheckCircle : record.status === 'warning' ? Warning : WifiOff;

  return (
    <Card
      sx={{
        border: '1px solid rgba(148, 163, 184, 0.12)',
        borderRadius: 2,
        transition: 'all 0.2s ease',
        '&:hover': {
          borderColor: 'rgba(30, 58, 138, 0.3)',
          boxShadow: '0 4px 12px rgba(0,0,0,0.08)',
          transform: 'translateY(-1px)',
        },
      }}
    >
      <CardContent sx={{ py: 1.5, px: 2 }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 1, mb: 1.5 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, flex: 1, minWidth: 0 }}>
            <Box
              sx={{
                p: 1,
                borderRadius: 1.5,
                bgcolor: 'rgba(30, 58, 138, 0.08)',
                display: 'flex',
                flexShrink: 0,
              }}
            >
              <Router sx={{ color: '#1e3a8a', fontSize: 20 }} />
            </Box>
            <Box sx={{ minWidth: 0, flex: 1 }}>
              <Typography variant="subtitle1" sx={{ fontWeight: 600, lineHeight: 1.3 }} noWrap>
                {record.name}
              </Typography>
              <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                {record.ip}
              </Typography>
            </Box>
          </Box>
          <Chip
            icon={<StatusIcon sx={{ fontSize: 14 }} />}
            label={translate(`monitoring.${record.status}`)}
            size="small"
            sx={{
              flexShrink: 0,
              bgcolor: `${statusColor}15`,
              color: statusColor,
              borderColor: statusColor,
              '& .MuiChip-icon': { color: statusColor },
            }}
            variant="outlined"
          />
        </Box>

        <Divider sx={{ my: 1 }} />

        <Box sx={{ display: 'flex', gap: 2 }}>
          <Box>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.25 }}>
              <Speed sx={{ fontSize: 11, verticalAlign: 'middle', mr: 0.25 }} />
              CPU
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.cpu_usage ?? '—'}%
            </Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.25 }}>
              <Memory sx={{ fontSize: 11, verticalAlign: 'middle', mr: 0.25 }} />
              Memory
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.memory_usage ?? '—'}%
            </Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.25 }}>
              Sessions
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.active_sessions ?? '—'}
            </Typography>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

const DeviceHealthListContent = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { data } = useListContext();
  const translate = useTranslate();

  if (isMobile) {
    return (
      <Stack spacing={1.5} sx={{ p: 2 }}>
        {(!data || data.length === 0) ? (
          <Box sx={{ textAlign: 'center', py: 4 }}>
            <Typography variant="body2" color="text.secondary">
              {translate('ra.navigation.no_results')}
            </Typography>
          </Box>
        ) : (
          data.map((record: any) => (
            <DeviceCard key={record.id} record={record} />
          ))
        )}
      </Stack>
    );
  }

  return (
    <Box sx={deviceHealthListStyles}>
      <Datagrid
        bulkActionButtons={<BulkActionsToolbar><BulkDeleteButton /></BulkActionsToolbar>}
        rowClick="show"
        sx={{
          bgcolor: 'background.paper',
          borderRadius: 2,
          overflow: 'hidden',
          boxShadow: '0 1px 3px rgba(0,0,0,0.12)',
        }}
      >
        <TextField source="name" label={translate('monitoring.devices')} />
        <TextField source="ip" label={translate('common.ip_address')} />
        <FunctionField
          source="status"
          label={translate('monitoring.status')}
          render={(record: any) => (
            <StatusBadge
              status={record.status === 'online' ? 'online' : 'offline'}
              label={record.status}
            />
          )}
        />
        <NumberField source="cpu_usage" label={translate('monitoring.cpu_usage')} options={{ style: 'unit', unit: '%' }} />
        <NumberField source="memory_usage" label={translate('monitoring.memory_usage')} options={{ style: 'unit', unit: '%' }} />
        <NumberField source="active_sessions" label={translate('monitoring.active_sessions')} />
        <DateField source="last_check" label={translate('monitoring.last_check')} showTime />
      </Datagrid>
    </Box>
  );
};

export const DeviceHealthList = () => {
  return (
    <List
      aside={<DeviceHealthAside />}
      actions={<DeviceHealthActions />}
      filters={<DeviceHealthFilters />}
      perPage={25}
      sx={{
        '& .RaList-content': {
          bgcolor: 'background.default',
        },
      }}
    >
      <DeviceHealthListContent />
    </List>
  );
};

DeviceHealthList.displayName = 'DeviceHealthList';
