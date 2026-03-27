import {
  List,
  Datagrid,
  TextField,
  NumberField,
  DateField,
  FunctionField,
  useListContext,
  TopToolbar,
  FilterButton,
  CreateButton,
  ExportButton,
  SearchInput,
  SelectInput,
  useTranslate,
} from 'react-admin';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Stack,
  Chip,
  Divider,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import {
  CheckCircle,
  Cancel,
  Business,
  People,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { StatusBadge } from '../../components/saas';

const ProviderAside = () => {
  const { data } = useListContext();
  const translate = useTranslate();

  const totalProviders = data?.length || 0;
  const activeProviders = data?.filter((p: any) => p.status === 'active').length || 0;
  const suspendedProviders = data?.filter((p: any) => p.status === 'suspended').length || 0;

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
              {translate('provider.total_providers')}
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
                {translate('provider.active')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#10b981', fontWeight: 700 }}>
              {activeProviders}
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
              <Cancel sx={{ color: '#ef4444', fontSize: 20 }} />
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('provider.suspended')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#ef4444', fontWeight: 700 }}>
              {suspendedProviders}
            </Typography>
          </CardContent>
        </Card>
      </Stack>
    </Box>
  );
};

  const ProviderCard = ({ record }: { record: any }) => {
    const navigate = useNavigate();
    const translate = useTranslate();

  return (
    <Card
      onClick={() => navigate(`/platform/providers/${record.id}/show`)}
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
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 1, mb: 1 }}>
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
                {record.name}
              </Typography>
              <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                {record.code}
              </Typography>
            </Box>
          </Box>
          <Chip
            label={translate(`provider.${record.status}`)}
            size="small"
            color={record.status === 'active' ? 'success' : 'error'}
            variant="outlined"
            sx={{ flexShrink: 0 }}
          />
        </Box>

        <Divider sx={{ my: 1.5 }} />

        <Box sx={{ display: 'flex', alignItems: 'center', gap: 3 }}>
          <Box>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.25 }}>
              <People sx={{ fontSize: 12, verticalAlign: 'middle', mr: 0.25 }} />
              Max Users
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.max_users?.toLocaleString() ?? '—'}
            </Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.25 }}>
              NAS
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.max_nas ?? '—'}
            </Typography>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

const ProviderListContent = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { data } = useListContext();
  const translate = useTranslate();

  if (isMobile) {
    return (
      <Stack spacing={1.5}>
        {(!data || data.length === 0) ? (
          <Box sx={{ textAlign: 'center', py: 4 }}>
            <Typography variant="body2" color="text.secondary">
              {translate('ra.navigation.no_results')}
            </Typography>
          </Box>
        ) : (
          data.map((record: any) => (
            <ProviderCard key={record.id} record={record} />
          ))
        )}
      </Stack>
    );
  }

  return (
    <Datagrid
      rowClick={(id) => `/platform/providers/${id}/show`}
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
    >
      <TextField source="name" label={translate('provider.name')} />
      <TextField source="code" label={translate('provider.code')} />
      <FunctionField
        source="status"
        label={translate('provider.status')}
        render={(record: any) => (
          <StatusBadge
            status={record.status === 'active' ? 'online' : 'error'}
            label={translate(`provider.${record.status}`)}
          />
        )}
      />
      <NumberField source="max_users" label={translate('provider.max_users')} />
      <NumberField source="max_nas" label={translate('provider.max_nas')} />
      <DateField source="created_at" label={translate('provider.created_at')} showTime />
    </Datagrid>
  );
};

  const ProviderActions = () => {
    const translate = useTranslate();
  return (
    <TopToolbar>
      <FilterButton />
      <CreateButton label={translate('provider.create')} to="/platform/providers/create" />
      <ExportButton />
    </TopToolbar>
  );
};

export const ProviderList = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const translate = useTranslate();

  const providerFilters = [
    <SearchInput source="name" alwaysOn />,
    <SelectInput
      source="status"
      choices={[
        { id: 'active', name: translate('provider.active') },
        { id: 'suspended', name: translate('provider.suspended') },
        { id: 'pending', name: translate('provider.pending') },
      ]}
    />,
  ];

  return (
    <List
      aside={isMobile ? undefined : <ProviderAside />}
      actions={<ProviderActions />}
      filters={providerFilters}
      perPage={25}
    >
      <ProviderListContent />
    </List>
  );
};

ProviderList.displayName = 'ProviderList';
