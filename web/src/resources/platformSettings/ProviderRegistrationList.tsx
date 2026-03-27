import {
  List,
  Datagrid,
  TextField,
  DateField,
  FunctionField,
  useListContext,
  TopToolbar,
  FilterButton,
  ExportButton,
  Button,
  useNotify,
  useRefresh,
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
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Alert,
  AlertTitle,
  Chip,
  TextField as MuiTextField,
  Button as MuiButton,
  useMediaQuery,
  useTheme,
  Divider,
} from '@mui/material';
import {
  CheckCircle,
  Cancel,
  Pending,
  Business,
} from '@mui/icons-material';
import { useState } from 'react';
import { fetchUtils } from 'react-admin';

interface RegistrationRecord {
  id: number;
  company_name: string;
  contact_name?: string;
  email?: string;
  business_type?: string;
  expected_users?: string;
  expected_nas?: string;
  status: 'pending' | 'approved' | 'rejected';
  created_at?: string;
}

const RegistrationAside = () => {
  const { data } = useListContext<RegistrationRecord>();
  const translate = useTranslate();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  if (isMobile) return null;

  if (!data) return null;

  const totalRequests = data.length;
  const pendingRequests = data.filter((r) => r && r.status === 'pending').length;
  const approvedRequests = data.filter((r) => r && r.status === 'approved').length;
  const rejectedRequests = data.filter((r) => r && r.status === 'rejected').length;

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
              {translate('providerRegistration.total_requests')}
            </Typography>
            <Typography variant="h4" sx={{ color: '#1e3a8a', fontWeight: 700 }}>
              {totalRequests}
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
              <Pending sx={{ color: '#f59e0b', fontSize: 20 }} />
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('providerRegistration.pending')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#f59e0b', fontWeight: 700 }}>
              {pendingRequests}
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
                {translate('providerRegistration.approved')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#10b981', fontWeight: 700 }}>
              {approvedRequests}
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
                {translate('providerRegistration.rejected')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#ef4444', fontWeight: 700 }}>
              {rejectedRequests}
            </Typography>
          </CardContent>
        </Card>
      </Stack>
    </Box>
  );
};

const ApprovalButtons = ({ record }: { record: RegistrationRecord }) => {
  const [approveOpen, setApproveOpen] = useState(false);
  const [rejectOpen, setRejectOpen] = useState(false);
  const [providerCode, setProviderCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const notify = useNotify();
  const refresh = useRefresh();
  const translate = useTranslate();

  const handleApprove = async () => {
    setIsLoading(true);
    try {
      const { fetchJson } = fetchUtils;
      const requestBody = {
        provider_code: providerCode || record.company_name.toLowerCase().replace(/\s+/g, '_'),
        max_users: record.expected_users || 1000,
        max_nas: record.expected_nas || 10,
      };

      const options: RequestInit = {
        method: 'POST',
        body: JSON.stringify(requestBody),
        headers: new Headers({
          'Content-Type': 'application/json',
        }),
      };

      const url = `${window.location.origin}/api/v1/providers/registrations/${record.id}/approve`;
      await fetchJson(url, options);

      notify(translate('providerRegistration.success_approve'), { type: 'success' });
      refresh();
      setApproveOpen(false);
    } catch (error: unknown) {
      const err = error as { message?: string; body?: { msg?: string; error?: string } };
      const errorMessage = err?.message || err?.body?.msg || err?.body?.error || 'Unknown error';
      notify(`${translate('providerRegistration.error_approve')}: ${errorMessage}`, { type: 'error', autoHideDuration: 5000 });
    } finally {
      setIsLoading(false);
    }
  };

  const handleReject = async () => {
    setIsLoading(true);
    try {
      const { fetchJson } = fetchUtils;
      const options: RequestInit = {
        method: 'POST',
        body: JSON.stringify({ reason: 'Registration rejected' }),
        headers: new Headers({
          'Content-Type': 'application/json',
        }),
      };

      await fetchJson(`${window.location.origin}/api/v1/providers/registrations/${record.id}/reject`, options);
      notify(translate('providerRegistration.success_reject'), { type: 'success' });
      refresh();
      setRejectOpen(false);
    } catch (error: unknown) {
      notify(translate('providerRegistration.error_reject'), { type: 'error' });
    } finally {
      setIsLoading(false);
    }
  };

  if (!record) return null;
  if (record.status !== 'pending') return null;

  return (
    <>
      <Button
        label={translate('providerRegistration.approve')}
        startIcon={<CheckCircle />}
        onClick={() => setApproveOpen(true)}
        size="small"
        sx={{
          mr: 1,
          backgroundColor: '#10b98115',
          color: '#10b981',
          '&:hover': { backgroundColor: '#10b98125' },
        }}
      />
      <Button
        label={translate('providerRegistration.reject')}
        startIcon={<Cancel />}
        onClick={() => setRejectOpen(true)}
        size="small"
        sx={{
          backgroundColor: '#ef444415',
          color: '#ef4444',
          '&:hover': { backgroundColor: '#ef444425' },
        }}
      />

      <Dialog open={approveOpen} onClose={() => setApproveOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{translate('providerRegistration.confirm_approve')}</DialogTitle>
        <DialogContent>
          <Alert severity="success" sx={{ mb: 2 }}>
            <AlertTitle>Confirm Approval</AlertTitle>
            You are about to approve the registration request from{' '}
            <strong>{record?.company_name || 'Unknown'}</strong>. This will create a new provider
            account with default quotas and pricing.
          </Alert>
          <Box sx={{ mt: 2 }}>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              Contact: {record?.contact_name || 'N/A'}
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              Email: {record?.email || 'N/A'}
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              Expected Users: {record?.expected_users || 'N/A'}
            </Typography>
          </Box>
          <Box sx={{ mt: 3 }}>
            <MuiTextField
              fullWidth
              label="Provider Code *"
              value={providerCode || (record?.company_name || '').toLowerCase().replace(/\s+/g, '_')}
              onChange={(e) => setProviderCode(e.target.value)}
              helperText="Unique identifier for the provider (e.g., TECHNET)"
              sx={{ mt: 1 }}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button label={translate('providerRegistration.cancel')} onClick={() => setApproveOpen(false)} />
          <Button
            label={translate('providerRegistration.approve_registration')}
            onClick={handleApprove}
            disabled={isLoading}
            variant="contained"
            sx={{
              backgroundColor: '#10b981',
              '&:hover': { backgroundColor: '#059669' },
            }}
          />
        </DialogActions>
      </Dialog>

      <Dialog open={rejectOpen} onClose={() => setRejectOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{translate('providerRegistration.confirm_reject')}</DialogTitle>
        <DialogContent>
          <Alert severity="error" sx={{ mb: 2 }}>
            <AlertTitle>Confirm Rejection</AlertTitle>
            You are about to reject the registration request from{' '}
            <strong>{record?.company_name || 'Unknown'}</strong>. This action cannot be undone.
          </Alert>
        </DialogContent>
        <DialogActions>
          <Button label={translate('providerRegistration.cancel')} onClick={() => setRejectOpen(false)} />
          <Button
            label={translate('providerRegistration.reject_registration')}
            onClick={handleReject}
            disabled={isLoading}
            variant="contained"
            color="error"
          />
        </DialogActions>
      </Dialog>
    </>
  );
};

const CardApprovalButtons = ({ record }: { record: RegistrationRecord }) => {
  const [approveOpen, setApproveOpen] = useState(false);
  const [rejectOpen, setRejectOpen] = useState(false);
  const [providerCode, setProviderCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const notify = useNotify();
  const refresh = useRefresh();
  const translate = useTranslate();

  const handleApprove = async () => {
    setIsLoading(true);
    try {
      const { fetchJson } = fetchUtils;
      const requestBody = {
        provider_code: providerCode || record.company_name.toLowerCase().replace(/\s+/g, '_'),
        max_users: record.expected_users || 1000,
        max_nas: record.expected_nas || 10,
      };
      const options: RequestInit = {
        method: 'POST',
        body: JSON.stringify(requestBody),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      };
      await fetchJson(`${window.location.origin}/api/v1/providers/registrations/${record.id}/approve`, options);
      notify(translate('providerRegistration.success_approve'), { type: 'success' });
      refresh();
      setApproveOpen(false);
    } catch (error: unknown) {
      const err = error as { message?: string; body?: { msg?: string; error?: string } };
      const errorMessage = err?.message || err?.body?.msg || err?.body?.error || 'Unknown error';
      notify(`${translate('providerRegistration.error_approve')}: ${errorMessage}`, { type: 'error', autoHideDuration: 5000 });
    } finally {
      setIsLoading(false);
    }
  };

  const handleReject = async () => {
    setIsLoading(true);
    try {
      const { fetchJson } = fetchUtils;
      await fetchJson(`${window.location.origin}/api/v1/providers/registrations/${record.id}/reject`, {
        method: 'POST',
        body: JSON.stringify({ reason: 'Registration rejected' }),
        headers: new Headers({ 'Content-Type': 'application/json' }),
      });
      notify(translate('providerRegistration.success_reject'), { type: 'success' });
      refresh();
      setRejectOpen(false);
    } catch (error: unknown) {
      notify(translate('providerRegistration.error_reject'), { type: 'error' });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <>
      <Box sx={{ display: 'flex', gap: 1.5 }}>
        <MuiButton
          variant="contained"
          size="small"
          startIcon={<CheckCircle />}
          onClick={() => setApproveOpen(true)}
          disabled={isLoading}
          sx={{
            backgroundColor: '#10b981',
            color: '#fff',
            fontWeight: 600,
            fontSize: '0.8rem',
            '&:hover': { backgroundColor: '#059669' },
          }}
        >
          {translate('providerRegistration.approve')}
        </MuiButton>
        <MuiButton
          variant="outlined"
          size="small"
          startIcon={<Cancel />}
          onClick={() => setRejectOpen(true)}
          disabled={isLoading}
          sx={{
            borderColor: '#ef4444',
            color: '#ef4444',
            fontWeight: 600,
            fontSize: '0.8rem',
            '&:hover': { borderColor: '#dc2626', backgroundColor: 'rgba(239,68,68,0.05)' },
          }}
        >
          {translate('providerRegistration.reject')}
        </MuiButton>
      </Box>

      <Dialog open={approveOpen} onClose={() => setApproveOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{translate('providerRegistration.confirm_approve')}</DialogTitle>
        <DialogContent>
          <Alert severity="success" sx={{ mb: 2 }}>
            <AlertTitle>Confirm Approval</AlertTitle>
            Approving registration request from{' '}
            <strong>{record?.company_name || 'Unknown'}</strong>.
          </Alert>
          <Box sx={{ mt: 2 }}>
            <Typography variant="body2"><strong>Contact:</strong> {record?.contact_name || 'N/A'}</Typography>
            <Typography variant="body2"><strong>Email:</strong> {record?.email || 'N/A'}</Typography>
            <Typography variant="body2"><strong>Expected Users:</strong> {record?.expected_users || 'N/A'}</Typography>
          </Box>
          <MuiTextField
            fullWidth
            label="Provider Code *"
            value={providerCode || (record?.company_name || '').toLowerCase().replace(/\s+/g, '_')}
            onChange={(e) => setProviderCode(e.target.value)}
            helperText="Unique identifier for the provider"
            sx={{ mt: 2 }}
          />
        </DialogContent>
        <DialogActions>
          <MuiButton onClick={() => setApproveOpen(false)}>{translate('providerRegistration.cancel')}</MuiButton>
          <MuiButton
            onClick={handleApprove}
            disabled={isLoading}
            variant="contained"
            sx={{ backgroundColor: '#10b981', '&:hover': { backgroundColor: '#059669' } }}
          >
            {translate('providerRegistration.approve_registration')}
          </MuiButton>
        </DialogActions>
      </Dialog>

      <Dialog open={rejectOpen} onClose={() => setRejectOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{translate('providerRegistration.confirm_reject')}</DialogTitle>
        <DialogContent>
          <Alert severity="error" sx={{ mb: 2 }}>
            <AlertTitle>Confirm Rejection</AlertTitle>
            Rejecting registration from{' '}
            <strong>{record?.company_name || 'Unknown'}</strong>. This cannot be undone.
          </Alert>
        </DialogContent>
        <DialogActions>
          <MuiButton onClick={() => setRejectOpen(false)}>{translate('providerRegistration.cancel')}</MuiButton>
          <MuiButton
            onClick={handleReject}
            disabled={isLoading}
            variant="contained"
            color="error"
          >
            {translate('providerRegistration.reject_registration')}
          </MuiButton>
        </DialogActions>
      </Dialog>
    </>
  );
};

const RegistrationActions = () => (
  <TopToolbar>
    <FilterButton />
    <ExportButton />
  </TopToolbar>
);

const statusColors: Record<string, string> = {
  pending: '#f59e0b',
  approved: '#10b981',
  rejected: '#ef4444',
};

const statusIcons: Record<string, React.ReactElement> = {
  pending: <Pending sx={{ fontSize: 14 }} />,
  approved: <CheckCircle sx={{ fontSize: 14 }} />,
  rejected: <Cancel sx={{ fontSize: 14 }} />,
};

const RegistrationCard = ({ record }: { record: RegistrationRecord }) => {
  const translate = useTranslate();
  const statusColor = statusColors[record.status] || '#6b7280';

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
              <Business sx={{ color: '#1e3a8a', fontSize: 20 }} />
            </Box>
            <Box sx={{ minWidth: 0, flex: 1 }}>
              <Typography variant="subtitle1" sx={{ fontWeight: 600, lineHeight: 1.3 }} noWrap>
                {record.company_name}
              </Typography>
              <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                {record.business_type}
              </Typography>
            </Box>
          </Box>
          <Chip
            icon={statusIcons[record.status]}
            label={translate(`providerRegistration.${record.status}`)}
            size="small"
            sx={{
              flexShrink: 0,
              bgcolor: `${statusColor}15`,
              color: statusColor,
              '& .MuiChip-icon': { color: statusColor },
            }}
            variant="outlined"
          />
        </Box>

        <Divider sx={{ my: 1 }} />

        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
            <Typography variant="caption" color="text.secondary">
              {translate('providerRegistration.contact')}
            </Typography>
            <Typography variant="caption" sx={{ fontWeight: 500 }}>
              {record.contact_name || '—'}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
            <Typography variant="caption" color="text.secondary">
              {translate('providerRegistration.email')}
            </Typography>
            <Typography variant="caption" sx={{ fontWeight: 500 }}>
              {record.email || '—'}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
            <Typography variant="caption" color="text.secondary">
              {translate('providerRegistration.expected_users')}
            </Typography>
            <Typography variant="caption" sx={{ fontWeight: 500 }}>
              {parseInt(record.expected_users || '0').toLocaleString()}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
            <Typography variant="caption" color="text.secondary">
              {translate('providerRegistration.submitted')}
            </Typography>
            <Typography variant="caption" sx={{ fontWeight: 500 }}>
              {record.created_at ? new Date(record.created_at).toLocaleDateString() : '—'}
            </Typography>
          </Box>
        </Box>

        {record.status === 'pending' && (
          <>
            <Divider sx={{ my: 1 }} />
            <CardApprovalButtons record={record} />
          </>
        )}
      </CardContent>
    </Card>
  );
};

const ProviderRegistrationListContent = () => {
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
          data.map((record) => (
            <RegistrationCard key={record.id} record={record} />
          ))
        )}
      </Stack>
    );
  }

  return (
    <Datagrid
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
      <TextField source="company_name" label={translate('providerRegistration.company')} />
      <TextField source="contact_name" label={translate('providerRegistration.contact')} />
      <TextField source="email" label={translate('providerRegistration.email')} />
      <TextField source="business_type" label={translate('providerRegistration.business_type')} />
      <FunctionField
        source="expected_users"
        label={translate('providerRegistration.expected_users')}
        render={(record: RegistrationRecord) => {
          if (!record) return '-';
          return `${parseInt(record.expected_users || '0').toLocaleString()} users`;
        }}
      />
      <FunctionField
        source="status"
        label={translate('providerRegistration.status')}
        render={(record: RegistrationRecord) => {
          if (!record || !record.status) return '-';
          return (
            <Chip
              label={translate(`providerRegistration.${record.status}`)}
              sx={{
                backgroundColor: `${statusColors[record.status] || '#6b7280'}15`,
                color: statusColors[record.status] || '#6b7280',
                border: `1px solid ${statusColors[record.status] || '#6b7280'}30`,
                fontWeight: 600,
                textTransform: 'capitalize',
              }}
            />
          );
        }}
      />
      <DateField source="created_at" label={translate('providerRegistration.submitted')} showTime />
      <FunctionField
        source="id"
        label=""
        render={(record: RegistrationRecord) => {
          if (!record) return null;
          return <ApprovalButtons record={record} />;
        }}
      />
    </Datagrid>
  );
};

export const ProviderRegistrationList = () => {
  const translate = useTranslate();

  const registrationFilters = [
    <SearchInput source="company_name" alwaysOn />,
    <SelectInput
      source="status"
      choices={[
        { id: 'pending', name: translate('providerRegistration.pending') },
        { id: 'approved', name: translate('providerRegistration.approved') },
        { id: 'rejected', name: translate('providerRegistration.rejected') },
      ]}
    />,
  ];

  return (
    <List
      aside={<RegistrationAside />}
      actions={<RegistrationActions />}
      filters={registrationFilters}
      perPage={25}
      resource="providers/registrations"
      empty={<div />}
    >
      <ProviderRegistrationListContent />
    </List>
  );
};

ProviderRegistrationList.displayName = 'ProviderRegistrationList';
