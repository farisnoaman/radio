import {
  List,
  Datagrid,
  TextField,
  NumberField,
  DateField,
  FunctionField,
  useListContext,
  useRecordContext,
  useTranslate,
  TopToolbar,
  FilterButton,
  CreateButton,
  ExportButton,
  Button,
  useNotify,
  useRefresh,
  useUpdate,
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
  Divider,
  Button as MuiButton,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import {
  Backup,
  CloudDownload,
  Restore,
  Lock,
  CheckCircle,
  Pending,
  Error as ErrorIcon,
} from '@mui/icons-material';
import { useState } from 'react';
import { StatusBadge, MetricCard } from '../../components/saas';

interface BackupRecord {
  id: number;
  file_name: string;
  backup_type: 'automated' | 'manual';
  status: 'pending' | 'running' | 'completed' | 'failed';
  file_size?: number;
  duration?: number;
  encryption_enabled?: boolean;
  created_at?: string;
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
};

const BackupAside = () => {
  const { data } = useListContext<BackupRecord>();
  const translate = useTranslate();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  if (isMobile) return null;

  const totalBackups = data?.length || 0;
  const completedBackups = data?.filter((b) => b.status === 'completed').length || 0;
  const runningBackups = data?.filter((b) => b.status === 'running').length || 0;
  const failedBackups = data?.filter((b) => b.status === 'failed').length || 0;
  const totalSize = data?.reduce((sum, b) => sum + (b.file_size || 0), 0) || 0;

  return (
    <Box sx={{ width: 320, ml: 2, mb: 2 }}>
      <Stack spacing={2}>
        <MetricCard
          title={translate('backup.aside.total_backups')}
          value={totalBackups}
          icon={<Backup fontSize="small" />}
          variant="compact"
        />

        <MetricCard
          title={translate('backup.storage_statistics')}
          value={formatBytes(totalSize)}
          icon={<CloudDownload fontSize="small" />}
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
                {translate('backup.encryption_type')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#10b981', fontWeight: 700 }}>
              {completedBackups}
            </Typography>
          </CardContent>
        </Card>

        <Card
          sx={{
            background: 'linear-gradient(135deg, rgba(59, 130, 246, 0.08) 0%, rgba(59, 130, 246, 0.02) 100%)',
            border: '1px solid rgba(59, 130, 246, 0.2)',
            borderRadius: 2,
          }}
        >
          <CardContent sx={{ p: 2 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 1 }}>
              <Pending sx={{ color: '#3b82f6', fontSize: 20 }} />
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('monitoring.processing')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#3b82f6', fontWeight: 700 }}>
              {runningBackups}
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
                {translate('common.error')}
              </Typography>
            </Box>
            <Typography variant="h4" sx={{ color: '#ef4444', fontWeight: 700 }}>
              {failedBackups}
            </Typography>
          </CardContent>
        </Card>
      </Stack>
    </Box>
  );
};

const RestoreButton = () => {
  const record = useRecordContext<BackupRecord>();
  const [open, setOpen] = useState(false);
  const [, { isLoading }] = useUpdate();
  const notify = useNotify();
  const refresh = useRefresh();
  const translate = useTranslate();

  if (!record) return null;

  const handleRestore = async () => {
    try {
      notify(translate('backup.success_restore'), { type: 'success' });
      refresh();
      setOpen(false);
    } catch {
      notify(translate('backup.error_restore'), { type: 'error' });
    }
  };

  return (
    <>
      <MuiButton
        variant="outlined"
        size="small"
        startIcon={<Restore />}
        onClick={() => setOpen(true)}
        disabled={record.status !== 'completed'}
        sx={{ fontSize: '0.75rem' }}
      >
        {translate('backup.restore')}
      </MuiButton>
      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{translate('backup.confirm_restore')}</DialogTitle>
        <DialogContent>
          <Alert severity="warning">
            <AlertTitle>{translate('backup.confirm_restore_message')}</AlertTitle>
            <Box sx={{ mt: 2 }}>
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('backup.backup_id')}: {record.file_name}
              </Typography>
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('backup.size')}: {record.file_size ? `${(record.file_size / 1024 / 1024).toFixed(2)} MB` : 'N/A'}
              </Typography>
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('backup.created_at')}: {record.created_at ? new Date(record.created_at).toLocaleString() : 'N/A'}
              </Typography>
            </Box>
          </Alert>
        </DialogContent>
        <DialogActions>
          <Button label={translate('common.cancel')} onClick={() => setOpen(false)} />
          <Button
            label={translate('backup.confirm_restore')}
            onClick={handleRestore}
            disabled={isLoading}
            variant="contained"
            color="error"
          />
        </DialogActions>
      </Dialog>
    </>
  );
};

const BackupActions = () => {
  const translate = useTranslate();
  return (
    <TopToolbar>
      <FilterButton />
      <CreateButton label={translate('backup.backup_now')} to="/platform/provider/backup/create" />
      <ExportButton />
    </TopToolbar>
  );
};

const BackupEmpty = () => {
  const translate = useTranslate();
  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        py: 8,
      }}
    >
      <Typography variant="h6" sx={{ mb: 2 }}>
        {translate('backup.no_backups') || 'No Backup Management yet.'}
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
        {translate('backup.do_you_want_to_add') || 'Do you want to add one?'}
      </Typography>
      <CreateButton label={translate('backup.create')} to="/platform/provider/backup/create" />
    </Box>
  );
};

const statusColors: Record<string, string> = {
  pending: '#f59e0b',
  running: '#3b82f6',
  completed: '#10b981',
  failed: '#ef4444',
};

const statusLabels: Record<string, string> = {
  pending: 'pending',
  running: 'running',
  completed: 'completed',
  failed: 'failed',
};

const BackupCard = ({ record }: { record: BackupRecord }) => {
  const translate = useTranslate();
  const [restoreOpen, setRestoreOpen] = useState(false);
  const notify = useNotify();
  const refresh = useRefresh();

  const statusColor = statusColors[record.status] || '#6b7280';

  const handleRestore = async () => {
    try {
      notify(translate('backup.success_restore'), { type: 'success' });
      refresh();
      setRestoreOpen(false);
    } catch {
      notify(translate('backup.error_restore'), { type: 'error' });
    }
  };

  return (
    <>
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
                <Backup sx={{ color: '#1e3a8a', fontSize: 20 }} />
              </Box>
              <Box sx={{ minWidth: 0, flex: 1 }}>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, lineHeight: 1.3 }} noWrap>
                  {record.file_name}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {record.created_at ? new Date(record.created_at).toLocaleString() : '—'}
                </Typography>
              </Box>
            </Box>
            <Chip
              label={translate(`backup.${statusLabels[record.status]}`)}
              size="small"
              sx={{
                flexShrink: 0,
                bgcolor: `${statusColor}15`,
                color: statusColor,
                fontWeight: 600,
              }}
            />
          </Box>

          <Divider sx={{ my: 1 }} />

          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
              <Typography variant="caption" color="text.secondary">
                {translate('backup.scope')}
              </Typography>
              <Typography variant="caption" sx={{ fontWeight: 500, textTransform: 'capitalize' }}>
                {record.backup_type || '—'}
              </Typography>
            </Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
              <Typography variant="caption" color="text.secondary">
                {translate('backup.size')}
              </Typography>
              <Typography variant="caption" sx={{ fontWeight: 500 }}>
                {record.file_size ? formatBytes(record.file_size) : '—'}
              </Typography>
            </Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
              <Typography variant="caption" color="text.secondary">
                {translate('backup.encrypted')}
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                {record.encryption_enabled ? (
                  <Lock sx={{ fontSize: 12, color: '#10b981' }} />
                ) : null}
                <Typography variant="caption" sx={{ fontWeight: 500, color: record.encryption_enabled ? '#10b981' : undefined }}>
                  {record.encryption_enabled ? translate('backup.encryption_type') : translate('common.none')}
                </Typography>
              </Box>
            </Box>
          </Box>

          {record.status === 'completed' && (
            <>
              <Divider sx={{ my: 1 }} />
              <MuiButton
                variant="outlined"
                color="primary"
                size="small"
                startIcon={<Restore />}
                onClick={() => setRestoreOpen(true)}
                sx={{ fontSize: '0.75rem' }}
              >
                {translate('backup.restore')}
              </MuiButton>
            </>
          )}
        </CardContent>
      </Card>

      <Dialog open={restoreOpen} onClose={() => setRestoreOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{translate('backup.confirm_restore')}</DialogTitle>
        <DialogContent>
          <Alert severity="warning">
            <AlertTitle>{translate('backup.confirm_restore_message')}</AlertTitle>
            <Box sx={{ mt: 2 }}>
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('backup.backup_id')}: {record.file_name}
              </Typography>
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('backup.size')}: {record.file_size ? `${(record.file_size / 1024 / 1024).toFixed(2)} MB` : 'N/A'}
              </Typography>
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('backup.created_at')}: {record.created_at ? new Date(record.created_at).toLocaleString() : 'N/A'}
              </Typography>
            </Box>
          </Alert>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRestoreOpen(false)}>{translate('common.cancel')}</Button>
          <Button onClick={handleRestore} variant="contained" color="error">
            {translate('backup.confirm_restore')}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

const BackupListContent = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { data } = useListContext<BackupRecord>();
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
            <BackupCard key={record.id} record={record} />
          ))
        )}
      </Stack>
    );
  }

  return (
    <Datagrid
      rowClick="show"
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
      <TextField source="file_name" label={translate('backup.backup_id')} />
      <FunctionField
        source="backup_type"
        label={translate('backup.scope')}
        render={(record: BackupRecord) => (
          <Box
            sx={{
              display: 'inline-flex',
              alignItems: 'center',
              px: 2,
              py: 1,
              borderRadius: 1,
              backgroundColor:
                record.backup_type === 'automated'
                  ? 'rgba(59, 130, 246, 0.1)'
                  : 'rgba(16, 185, 129, 0.1)',
              border:
                record.backup_type === 'automated'
                  ? '1px solid rgba(59, 130, 246, 0.3)'
                  : '1px solid rgba(16, 185, 129, 0.3)',
            }}
          >
            <Typography
              variant="body2"
              sx={{
                fontWeight: 600,
                color: record.backup_type === 'automated' ? '#3b82f6' : '#10b981',
                textTransform: 'capitalize',
              }}
            >
              {record.backup_type}
            </Typography>
          </Box>
        )}
      />
      <FunctionField
        source="status"
        label={translate('backup.status')}
        render={(record: BackupRecord) => (
          <StatusBadge
            status={
              record.status === 'completed'
                ? 'online'
                : record.status === 'running'
                ? 'processing'
                : record.status === 'pending'
                ? 'warning'
                : 'error'
            }
            label={record.status}
          />
        )}
      />
      <FunctionField
        source="file_size"
        label={translate('backup.size')}
        render={(record: BackupRecord) => (
          <Typography variant="body2" sx={{ fontWeight: 500 }}>
            {record.file_size ? formatBytes(record.file_size) : 'N/A'}
          </Typography>
        )}
      />
      <FunctionField
        source="encryption_enabled"
        label={translate('backup.encrypted')}
        render={(record: BackupRecord) =>
          record.encryption_enabled ? (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, color: '#10b981' }}>
              <Lock sx={{ fontSize: 16 }} />
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('backup.encryption_type')}
              </Typography>
            </Box>
          ) : (
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {translate('common.none')}
            </Typography>
          )
        }
      />
      <NumberField source="duration" label={translate('backup.created_at')} />
      <DateField source="created_at" label={translate('backup.created_at')} showTime />
      <RestoreButton />
    </Datagrid>
  );
};

export const BackupList = () => {
  return (
    <List
      aside={<BackupAside />}
      actions={<BackupActions />}
      filters={[]}
      empty={<BackupEmpty />}
      perPage={25}
    >
      <BackupListContent />
    </List>
  );
};

BackupList.displayName = 'BackupList';
