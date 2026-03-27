import {
  List,
  Datagrid,
  TextField,
  NumberField,
  FunctionField,
  useListContext,
  useGetList,
  TopToolbar,
  FilterButton,
  ExportButton,
  SearchInput,
  SelectInput,
} from 'react-admin';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Stack,
  LinearProgress,
  Chip,
  Divider,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import {
  CheckCircle,
  Warning,
  Error as ErrorIcon,
  Business,
  People,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useTranslate } from 'react-admin';
import { StatusBadge } from '../../components/saas';

const QuotaAside = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const { data } = useListContext();
  const translate = useTranslate();

  if (isMobile) return null;

  const totalProviders = data?.length || 0;
  const healthyProviders = data?.filter((p: any) => {
    const userPercent = p.utilization?.users_percent || 0;
    const sessionPercent = p.utilization?.sessions_percent || 0;
    return userPercent < 80 && sessionPercent < 80;
  }).length || 0;
  const warningProviders = data?.filter((p: any) => {
    const userPercent = p.utilization?.users_percent || 0;
    const sessionPercent = p.utilization?.sessions_percent || 0;
    return (userPercent >= 80 && userPercent < 100) || (sessionPercent >= 80 && sessionPercent < 100);
  }).length || 0;
  const criticalProviders = data?.filter((p: any) => {
    const userPercent = p.utilization?.users_percent || 0;
    const sessionPercent = p.utilization?.sessions_percent || 0;
    return userPercent >= 100 || sessionPercent >= 100;
  }).length || 0;

  return (
    <Box sx={{ width: 300, ml: 2, mb: 2 }}>
      <Stack spacing={2}>
        <Card
          sx={{
            background: 'linear-gradient(135deg, rgba(30, 58, 138, 0.08) 0%, rgba(30, 58, 138, 0.02) 100%)',
            border: '1px solid rgba(30, 58, 138, 0.2)',
            borderRadius: 2,
          }}
        >
          <CardContent sx={{ p: 2 }}>
            <Typography variant="body2" sx={{ fontWeight: 600, mb: 1 }}>
              {translate('quota.title')}
            </Typography>
            <Typography variant="h4" sx={{ color: '#1e3a8a', fontWeight: 700 }}>
              {totalProviders}
            </Typography>
          </CardContent>
        </Card>

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
                {translate('quota.healthy')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#10b981', fontWeight: 700 }}>
              {healthyProviders}
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
                {translate('quota.warning')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#f59e0b', fontWeight: 700 }}>
              {warningProviders}
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
              <ErrorIcon sx={{ color: '#ef4444', fontSize: 20 }} />
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('quota.critical')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#ef4444', fontWeight: 700 }}>
              {criticalProviders}
            </Typography>
          </CardContent>
        </Card>
      </Stack>
    </Box>
  );
};

const UsageBar = ({ percent, label }: { percent: number; label: string; current?: number; max?: number }) => {
  const color = percent >= 100 ? '#ef4444' : percent >= 80 ? '#f59e0b' : '#10b981';

  return (
    <Box sx={{ mb: 1 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
        <Typography variant="caption" color="text.secondary">
          {label}
        </Typography>
        <Typography variant="caption" sx={{ fontWeight: 600, color }}>
          {percent.toFixed(0)}%
        </Typography>
      </Box>
      <LinearProgress
        variant="determinate"
        value={Math.min(percent, 100)}
        sx={{
          height: 5,
          borderRadius: 3,
          backgroundColor: 'rgba(0,0,0,0.08)',
          '& .MuiLinearProgress-bar': { backgroundColor: color },
        }}
      />
    </Box>
  );
};

const QuotaCard = ({ record }: { record: any }) => {
  const translate = useTranslate();
  const navigate = useNavigate();

  const userPercent = record.utilization?.users_percent || 0;
  const sessionPercent = record.utilization?.sessions_percent || 0;
  const maxPercent = Math.max(userPercent, sessionPercent);

  const statusColor = maxPercent >= 100 ? 'error' : maxPercent >= 80 ? 'warning' : 'success';
  const statusLabel =
    maxPercent >= 100 ? translate('quota.critical') :
    maxPercent >= 80 ? translate('quota.warning') :
    translate('quota.healthy');

  return (
    <Card
      onClick={() => navigate(`/platform/quotas/${record.id}/show`)}
      sx={{
        cursor: 'pointer',
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
              <Business sx={{ color: '#1e3a8a', fontSize: 20 }} />
            </Box>
            <Box sx={{ minWidth: 0, flex: 1 }}>
              <Typography variant="subtitle1" sx={{ fontWeight: 600, lineHeight: 1.3 }} noWrap>
                {record.provider_name}
              </Typography>
              <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                {record.provider_code}
              </Typography>
            </Box>
          </Box>

          <Chip
            label={statusLabel}
            size="small"
            color={statusColor}
            variant="outlined"
            sx={{ flexShrink: 0 }}
          />
        </Box>

        <Box sx={{ mb: 1 }}>
          <UsageBar
            label={`${translate('quota.users')}: ${record.usage?.current_users ?? 0} / ${record.quota?.max_users ?? 1000}`}
            percent={userPercent}
            current={record.usage?.current_users ?? 0}
            max={record.quota?.max_users ?? 1000}
          />
          <UsageBar
            label={`${translate('quota.online_sessions')}: ${record.usage?.current_online_users ?? 0} / ${record.quota?.max_online_users ?? 500}`}
            percent={sessionPercent}
            current={record.usage?.current_online_users ?? 0}
            max={record.quota?.max_online_users ?? 500}
          />
        </Box>

        <Divider sx={{ my: 1 }} />

        <Box sx={{ display: 'flex', gap: 2 }}>
          <Box>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
              <People sx={{ fontSize: 11, verticalAlign: 'middle', mr: 0.25 }} />
              {translate('quota.active')}
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.usage?.current_users?.toLocaleString() ?? '—'}
            </Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
              {translate('quota.sessions')}
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.usage?.current_online_users?.toLocaleString() ?? '—'}
            </Typography>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

const QuotaActions = () => {
  return (
    <TopToolbar>
      <FilterButton />
      <ExportButton />
    </TopToolbar>
  );
};

export const QuotaList = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const translate = useTranslate();

  const { data: providers, isLoading } = useGetList('admin/providers', {
    pagination: { page: 1, perPage: 100 },
    sort: { field: 'id', order: 'ASC' },
  });

  const quotaFilters = [
    <SearchInput source="provider_name" alwaysOn />,
    <SelectInput
      source="status"
      choices={[
        { id: 'healthy', name: translate('quota.healthy') },
        { id: 'warning', name: translate('quota.warning') },
        { id: 'critical', name: translate('quota.critical') },
      ]}
    />,
  ];

  const enrichedData = providers?.map((provider: any) => {
    const maxUsers = provider.max_users || 1000;
    const maxNas = provider.max_nas || 100;
    const currentUsers = provider.usage?.current_users || 0;
    const currentSessions = provider.usage?.current_online_users || 0;
    const currentNas = provider.usage?.current_nas || 0;
    const usersPercent = provider.utilization?.users_percent !== undefined
      ? provider.utilization.users_percent
      : (currentUsers / maxUsers) * 100;
    const sessionsPercent = provider.utilization?.sessions_percent !== undefined
      ? provider.utilization.sessions_percent
      : (currentSessions / maxUsers) * 100;
    const nasPercent = provider.utilization?.nas_percent !== undefined
      ? provider.utilization.nas_percent
      : (currentNas / maxNas) * 100;

    return {
      ...provider,
      quota: {
        max_users: maxUsers,
        max_online_users: maxUsers,
        max_nas: maxNas,
      },
      usage: {
        current_users: currentUsers,
        current_online_users: currentSessions,
        current_nas: currentNas,
      },
      utilization: {
        users_percent: usersPercent,
        sessions_percent: sessionsPercent,
        nas_percent: nasPercent,
      },
    };
  }) || [];

  return (
    <List
      aside={<QuotaAside />}
      actions={<QuotaActions />}
      filters={quotaFilters}
      perPage={25}
      loading={isLoading}
      sx={{
        '& .RaList-content': {
          bgcolor: 'background.default',
        },
      }}
    >
      {isMobile ? (
        <Stack spacing={1.5} sx={{ p: 2 }}>
          {enrichedData.length === 0 ? (
            <Box sx={{ textAlign: 'center', py: 4 }}>
              <Typography variant="body2" color="text.secondary">
                {translate('ra.navigation.no_results')}
              </Typography>
            </Box>
          ) : (
            enrichedData.map((record: any) => (
              <QuotaCard key={record.id} record={record} />
            ))
          )}
        </Stack>
      ) : (
        <Datagrid
          rowClick={(id) => `/platform/quotas/${id}/show`}
          sx={{
            bgcolor: 'background.paper',
            borderRadius: 2,
            overflow: 'hidden',
            boxShadow: '0 1px 3px rgba(0,0,0,0.12)',
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
          }}
          data={enrichedData}
        >
          <TextField source="provider_name" label={translate('provider.name')} />
          <TextField source="provider_code" label={translate('provider.code')} />
          <FunctionField
            source="utilization"
            label={translate('quota.utilization')}
            render={(record: any) => {
              const userPercent = record.utilization?.users_percent || 0;
              const sessionPercent = record.utilization?.sessions_percent || 0;
              const maxPercent = Math.max(userPercent, sessionPercent);
              const color = maxPercent >= 100 ? '#ef4444' : maxPercent >= 80 ? '#f59e0b' : '#10b981';
              return (
                <Box sx={{ width: '100%' }}>
                  <Typography variant="caption" sx={{ color, fontWeight: 600 }}>
                    {maxPercent.toFixed(1)}%
                  </Typography>
                  <LinearProgress
                    variant="determinate"
                    value={Math.min(maxPercent, 100)}
                    sx={{
                      height: 4,
                      borderRadius: 2,
                      backgroundColor: 'rgba(0,0,0,0.08)',
                      '& .MuiLinearProgress-bar': { backgroundColor: color },
                    }}
                  />
                </Box>
              );
            }}
          />
          <NumberField source="usage.current_users" label={translate('quota.users')} />
          <NumberField source="usage.current_online_users" label={translate('quota.online_sessions')} />
          <FunctionField
            source="status"
            label={translate('quota.status')}
            render={(record: any) => {
              const userPercent = record.utilization?.users_percent || 0;
              const sessionPercent = record.utilization?.sessions_percent || 0;
              const maxPercent = Math.max(userPercent, sessionPercent);
              const status: 'online' | 'error' | 'warning' =
                maxPercent >= 100 ? 'error' :
                maxPercent >= 80 ? 'warning' : 'online';
              const label = maxPercent >= 100 ? translate('quota.critical') :
                maxPercent >= 80 ? translate('quota.warning') :
                translate('quota.healthy');
              return <StatusBadge status={status} label={label} />;
            }}
          />
        </Datagrid>
      )}
    </List>
  );
};

QuotaList.displayName = 'QuotaList';
