import {
  List,
  Datagrid,
  TextField,
  NumberField,
  DateField,
  EditButton,
  ShowButton,
  DeleteButton,
  Create,
  SimpleForm,
  TextInput,
  NumberInput,
  SelectInput,
  ReferenceInput,
  ReferenceField,
  Edit,
  Show,
  required,
  useRecordContext,
  useTranslate,
  ListProps,
  ShowProps,
  CreateProps,
  EditProps,
  ListButton,
  useNotify,
  useRefresh,
  useListContext,
  RecordContextProvider,
  useLocale,
  FunctionField,
} from 'react-admin';
import { useWatch } from 'react-hook-form';
import { useMediaQuery, Theme, CardActions, Box, Card, CardContent, Stack, Avatar, Typography, Tooltip, IconButton, Chip, alpha } from '@mui/material';
import {
  Speed as SpeedIcon,
  Schedule as TimeIcon,
  Note as NoteIcon,
  ContentCopy as CopyIcon,
  Refresh as RefreshIcon,
  ArrowBack as BackIcon,
  Print as PrintIcon,
  CheckCircle as EnabledIcon,
  Cancel as DisabledIcon,
  AttachMoney as MoneyIcon,
  DataUsage as DataIcon,
} from '@mui/icons-material';
import {
  FormSection,
  formLayoutSx,
  DetailItem,
  DetailSectionCard,
  EmptyValue,
} from '../components';

const ProductTitle = () => {
  const record = useRecordContext();
  return <span>Product {record ? `"${record.name}"` : ''}</span>;
};


export const ProductGrid = () => {
  const { data, isLoading } = useListContext();
  const { isRtl, translate, formatRate, formatQuota } = useFormatters();

  if (isLoading || !data) return null;
  return (
    <Box display="grid" dir={isRtl ? 'rtl' : 'ltr'} gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)' }} gap={2} p={0} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}>
      {data.map(record => (
        <RecordContextProvider value={record} key={record.id}>
          <Card
            elevation={0}
            sx={{
              borderRadius: 3,
              border: theme => `1px solid ${theme.palette.divider}`,
              transition: 'box-shadow 0.2s',
              '&:hover': { boxShadow: 4 }
            }}
          >
            <CardContent sx={{ pb: 1 }}>
              <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                <Box display="flex" alignItems="center" gap={1.5}>
                  <Avatar sx={{ bgcolor: record.color || 'primary.main', width: 40, height: 40, fontWeight: 'bold' }}>
                    {record.name?.charAt(0).toUpperCase()}
                  </Avatar>
                  <Box>
                    <Typography variant="subtitle1" component="div" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
                      <TextField source="name" />
                    </Typography>
                    <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                      ID: {record.id}
                    </Typography>
                  </Box>
                </Box>
                <StatusIndicator isEnabled={record.status === 'enabled'} />
              </Box>

              <Box sx={{ bgcolor: theme => alpha(theme.palette.grey[500], 0.05), p: 1.5, borderRadius: 2, mb: 2 }}>
                <Typography variant="body2" color="text.secondary" sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                  <DataIcon fontSize="small" color="action" />
                  Profile: <strong><ReferenceField source="radius_profile_id" reference="radius-profiles"><TextField source="name" /></ReferenceField></strong>
                </Typography>
                <Box display="flex" justifyContent="space-between" mb={1}>
                  <Typography variant="body2" color="text.secondary">Price:</Typography>
                  <Typography variant="body2" sx={{ fontWeight: 'bold', color: 'success.main' }}>
                    <NumberField source="price" options={{ style: 'currency', currency: 'USD' }} />
                  </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between" mb={1}>
                  <Typography variant="body2" color="text.secondary">Quota:</Typography>
                  <Typography variant="body2" sx={{ fontWeight: 'bold' }}>
                    {formatQuota(record.data_quota)}
                  </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between">
                  <Typography variant="body2" color="text.secondary">Rates (U/D):</Typography>
                  <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                    {formatRate(record.up_rate)} / {formatRate(record.down_rate)}
                  </Typography>
                </Box>
              </Box>
            </CardContent>
            <CardActions sx={{ justifyContent: 'flex-end', borderTop: theme => `1px solid ${theme.palette.divider}`, px: 2, py: 1.5, gap: 1 }}>
              <EditButton label={translate('ra.action.edit', { _: 'Edit' })} size="small" variant="outlined" />
              <DeleteButton label={translate('ra.action.delete', { _: 'Delete' })} size="small" variant="outlined" />
              <ShowButton label={translate('ra.action.show', { _: 'Show' })} size="small" variant="outlined" />
            </CardActions>
          </Card>
        </RecordContextProvider>
      ))}
    </Box>
  );
};
export const ProductList = (props: ListProps) => {
  const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
  const { isRtl, translate } = useFormatters();
  
  return (
    <List {...props} sort={{ field: 'id', order: 'DESC' }}>
      {isSmall ? (
        <ProductGrid />
      ) : (
        <Datagrid rowClick="show" sx={{ direction: isRtl ? 'rtl' : 'ltr' }}>
          <TextField source="id" label={translate('resources.products.fields.id')} />
          <TextField source="name" label={translate('resources.products.fields.name')} />
          <ReferenceField source="radius_profile_id" reference="radius-profiles" label={translate('resources.products.fields.radius_profile_id')}>
            <TextField source="name" />
          </ReferenceField>
          <NumberField source="price" options={{ style: 'currency', currency: 'USD' }} label={translate('resources.products.fields.price')} />
          <NumberField source="up_rate" label={`${translate('resources.products.fields.up_rate')} (${translate('resources.products.units.kbps')})`} />
          <NumberField source="down_rate" label={`${translate('resources.products.fields.down_rate')} (${translate('resources.products.units.kbps')})`} />
          <NumberField source="data_quota" label={`${translate('resources.products.fields.data_quota')} (${translate('resources.products.units.mb')})`} />
          <FunctionField
            source="status"
            label={translate('resources.products.fields.status')}
            render={(record: any) => (
              <Chip
                label={record.status === 'enabled' ? translate('resources.products.status.enabled', { _: 'Enabled' }) : translate('resources.products.status.disabled', { _: 'Disabled' })}
                size="small"
                color={record.status === 'enabled' ? 'success' : 'default'}
                variant={record.status === 'enabled' ? 'filled' : 'outlined'}
                sx={{ fontWeight: 600, fontSize: '0.75rem' }}
              />
            )}
          />
          <DateField source="updated_at" showTime label={translate('resources.products.fields.updated_at')} />
          <EditButton label={translate('ra.action.edit', { _: 'Edit' })} variant="outlined" size="small" />
          <DeleteButton label={translate('ra.action.delete', { _: 'Delete' })} variant="outlined" size="small" />
        </Datagrid>
      )}
    </List>
  );
};

const RTL_LANGUAGES = ['ar', 'he', 'fa', 'ur'];

// Helper to translate units
const useFormatters = () => {
  const translate = useTranslate();
  const locale = useLocale();
  const isRtl = RTL_LANGUAGES.includes(locale || '');

  const formatRate = (rate?: number): string => {
    if (rate === undefined || rate === null) return '-';
    if (rate === 0) return translate('resources.products.units.unlimited', { _: 'Unlimited' });
    if (rate >= 1024) {
      return `${(rate / 1024).toFixed(1)} ${translate('resources.products.units.mbps', { _: 'Mbps' })}`;
    }
    return `${rate} ${translate('resources.products.units.kbps', { _: 'Kbps' })}`;
  };

  const formatQuota = (quota?: number): string => {
    if (quota === undefined || quota === null) return '-';
    if (quota === 0) return translate('resources.products.units.unlimited', { _: 'Unlimited' });
    if (quota >= 1024) {
      return `${(quota / 1024).toFixed(1)} ${translate('resources.products.units.gb', { _: 'GB' })}`;
    }
    return `${quota} ${translate('resources.products.units.mb', { _: 'MB' })}`;
  };

  const formatValidity = (seconds?: number): string => {
    if (seconds === undefined || seconds === null) return '-';
    if (seconds === 0) return translate('resources.products.units.unlimited', { _: 'Unlimited' });
    if (seconds >= 86400 && seconds % 86400 === 0) return `${seconds / 86400} ${translate('resources.products.units.days', { _: 'Days' })}`;
    if (seconds >= 3600 && seconds % 3600 === 0) return `${seconds / 3600} ${translate('resources.products.units.hours', { _: 'Hours' })}`;
    return `${seconds / 60} ${translate('resources.products.units.minutes', { _: 'Minutes' })}`;
  };

  return { formatRate, formatQuota, formatValidity, isRtl, translate };
};

const formatTimestamp = (value?: string | number): string => {
  if (!value) {
    return '-';
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }
  return date.toLocaleString();
};

const StatusIndicator = ({ isEnabled }: { isEnabled: boolean }) => {
  const translate = useTranslate();
  return (
    <Chip
      icon={isEnabled ? <EnabledIcon sx={{ fontSize: '0.85rem !important' }} /> : <DisabledIcon sx={{ fontSize: '0.85rem !important' }} />}
      label={isEnabled ? translate('resources.products.status.enabled', { _: 'Enabled' }) : translate('resources.products.status.disabled', { _: 'Disabled' })}
      size="small"
      color={isEnabled ? 'success' : 'default'}
      variant={isEnabled ? 'filled' : 'outlined'}
      sx={{ height: 22, fontWeight: 500, fontSize: '0.75rem' }}
    />
  );
};

const printStyles = `
  @media print {
    body * {
      visibility: hidden;
    }
    .printable-content, .printable-content * {
      visibility: visible;
    }
    .printable-content {
      position: absolute;
      left: 0;
      top: 0;
      width: 100%;
      padding: 20px !important;
    }
    .no-print {
      display: none !important;
    }
  }
`;

const ProductFormBanner = () => {
  const name = useWatch({ name: 'name' });
  const color = useWatch({ name: 'color', defaultValue: '#1976d2' });
  const status = useWatch({ name: 'status', defaultValue: 'enabled' });
  const translate = useTranslate();
  
  const isEnabled = status === 'enabled';

  return (
    <Card
      elevation={0}
      sx={{
        borderRadius: 4,
        background: theme =>
          theme.palette.mode === 'dark'
            ? isEnabled
              ? `linear-gradient(135deg, ${alpha(color || theme.palette.primary.dark, 0.4)} 0%, ${alpha(theme.palette.info.dark, 0.3)} 100%)`
              : `linear-gradient(135deg, ${alpha(theme.palette.grey[800], 0.5)} 0%, ${alpha(theme.palette.grey[700], 0.3)} 100%)`
            : isEnabled
              ? `linear-gradient(135deg, ${alpha(color || theme.palette.primary.main, 0.1)} 0%, ${alpha(theme.palette.info.main, 0.08)} 100%)`
              : `linear-gradient(135deg, ${alpha(theme.palette.grey[400], 0.15)} 0%, ${alpha(theme.palette.grey[300], 0.1)} 100%)`,
        border: theme => `1px solid ${alpha(isEnabled ? color || theme.palette.primary.main : theme.palette.grey[500], 0.2)}`,
        overflow: 'hidden',
        position: 'relative',
        mb: 2,
      }}
    >
      <Box
        sx={{
          position: 'absolute',
          top: -50,
          right: -50,
          width: 200,
          height: 200,
          borderRadius: '50%',
          background: theme => alpha(isEnabled ? color || theme.palette.primary.main : theme.palette.grey[500], 0.1),
          pointerEvents: 'none',
        }}
      />

      <CardContent sx={{ p: 3, position: 'relative', zIndex: 1 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Avatar
            sx={{
              width: 56,
              height: 56,
              bgcolor: color || (isEnabled ? 'primary.main' : 'grey.500'),
              fontSize: '1.25rem',
              fontWeight: 700,
              boxShadow: theme => `0 4px 14px ${alpha(isEnabled ? color || theme.palette.primary.main : theme.palette.grey[500], 0.4)}`,
            }}
          >
            {name ? name.charAt(0).toUpperCase() : 'P'}
          </Avatar>
          <Box>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
              <Typography variant="h6" sx={{ fontWeight: 700, color: 'text.primary' }}>
                {name || translate('resources.products.fields.name', { _: 'Product Name' })}
              </Typography>
              <StatusIndicator isEnabled={isEnabled} />
            </Box>
            <Typography variant="body2" color="text.secondary">
              {translate('common.message.editing', { _: 'Editing' })}
            </Typography>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

const ProductHeaderCard = () => {
  const record = useRecordContext();
  const { formatRate, formatQuota } = useFormatters();
  const translate = useTranslate();
  const notify = useNotify();
  const refresh = useRefresh();

  if (!record) return null;

  const isEnabled = record.status === 'enabled';

  const handleCopy = (text: string, name: string) => {
    navigator.clipboard.writeText(text).then(() => {
      notify(translate('common.message.copied', { name, _: `Copied ${name}` }), { type: 'success' });
    });
  };

  const handleRefresh = () => {
    refresh();
    notify(translate('common.message.refreshed', { _: 'Data refreshed' }), { type: 'success' });
  };

  return (
    <Card
      elevation={0}
      sx={{
        borderRadius: 4,
        background: theme =>
          theme.palette.mode === 'dark'
            ? isEnabled
              ? `linear-gradient(135deg, ${alpha(theme.palette.primary.dark, 0.4)} 0%, ${alpha(theme.palette.info.dark, 0.3)} 100%)`
              : `linear-gradient(135deg, ${alpha(theme.palette.grey[800], 0.5)} 0%, ${alpha(theme.palette.grey[700], 0.3)} 100%)`
            : isEnabled
              ? `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.1)} 0%, ${alpha(theme.palette.info.main, 0.08)} 100%)`
              : `linear-gradient(135deg, ${alpha(theme.palette.grey[400], 0.15)} 0%, ${alpha(theme.palette.grey[300], 0.1)} 100%)`,
        border: theme => `1px solid ${alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.2)}`,
        overflow: 'hidden',
        position: 'relative',
      }}
    >
      <Box
        sx={{
          position: 'absolute',
          top: -50,
          right: -50,
          width: 200,
          height: 200,
          borderRadius: '50%',
          background: theme => alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.1),
          pointerEvents: 'none',
        }}
      />

      <CardContent sx={{ p: 3, position: 'relative', zIndex: 1 }}>
        <Box sx={{ display: 'flex', flexDirection: { xs: 'column', sm: 'row' }, justifyContent: 'space-between', alignItems: { xs: 'stretch', sm: 'flex-start' }, mb: 3, gap: 2 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Avatar
              sx={{
                width: 64,
                height: 64,
                bgcolor: record.color || (isEnabled ? 'primary.main' : 'grey.500'),
                fontSize: '1.5rem',
                fontWeight: 700,
                boxShadow: theme => `0 4px 14px ${alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.4)}`,
              }}
            >
              {record.name?.charAt(0).toUpperCase() || 'P'}
            </Avatar>
            <Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
                  {record.name || <EmptyValue message="Unknown Product" />}
                </Typography>
                <StatusIndicator isEnabled={isEnabled} />
              </Box>
              {record.name && (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                  <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                    ID: {record.id}
                  </Typography>
                  <Tooltip title="Copy Product ID">
                    <IconButton
                      size="small"
                      onClick={() => handleCopy(record.id?.toString() || '', 'Product ID')}
                      sx={{ p: 0.5 }}
                    >
                      <CopyIcon sx={{ fontSize: '0.75rem' }} />
                    </IconButton>
                  </Tooltip>
                </Box>
              )}
            </Box>
          </Box>

          <Box className="no-print" sx={{ display: 'flex', gap: 1, justifyContent: { xs: 'flex-end', sm: 'flex-start' } }}>
            <Tooltip title="Print Details">
              <IconButton
                onClick={() => window.print()}
                sx={{
                  bgcolor: theme => alpha(theme.palette.info.main, 0.1),
                  '&:hover': {
                    bgcolor: theme => alpha(theme.palette.info.main, 0.2),
                  },
                }}
              >
                <PrintIcon />
              </IconButton>
            </Tooltip>
            <Tooltip title="Refresh Data">
              <IconButton
                onClick={handleRefresh}
                sx={{
                  bgcolor: theme => alpha(theme.palette.primary.main, 0.1),
                  '&:hover': {
                    bgcolor: theme => alpha(theme.palette.primary.main, 0.2),
                  },
                }}
              >
                <RefreshIcon />
              </IconButton>
            </Tooltip>
            <ListButton
              label=""
              icon={<BackIcon />}
              sx={{
                minWidth: 'auto',
                bgcolor: theme => alpha(theme.palette.grey[500], 0.1),
                '&:hover': {
                  bgcolor: theme => alpha(theme.palette.grey[500], 0.2),
                },
              }}
            />
          </Box>
        </Box>

        <Box
          sx={{
            display: 'grid',
            gap: 2,
            gridTemplateColumns: {
              xs: 'repeat(2, 1fr)',
              sm: 'repeat(4, 1fr)',
            },
          }}
        >
          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <MoneyIcon sx={{ fontSize: '1.1rem', color: 'success.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.products.fields.price')}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              ${record.price?.toFixed(2) || '0.00'}
            </Typography>
          </Box>

          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <SpeedIcon sx={{ fontSize: '1.1rem', color: 'info.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.products.fields.up_rate')}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600, fontFamily: 'monospace' }}>
              {formatRate(record.up_rate)}
            </Typography>
          </Box>

          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <SpeedIcon sx={{ fontSize: '1.1rem', color: 'warning.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.products.fields.down_rate')}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600, fontFamily: 'monospace' }}>
              {formatRate(record.down_rate)}
            </Typography>
          </Box>

          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <DataIcon sx={{ fontSize: '1.1rem', color: 'error.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.products.fields.data_quota')}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600, fontFamily: 'monospace' }}>
              {formatQuota(record.data_quota)}
            </Typography>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

const ProductDetails = () => {
  const record = useRecordContext();
  const { formatRate, formatQuota, formatValidity, translate, isRtl } = useFormatters();

  if (!record) {
    return null;
  }

  const validitySeconds = record.validity_seconds || 0;
  const validityDisplay = formatValidity(validitySeconds);

  return (
    <>
      <style>{printStyles}</style>
      <Box className="printable-content" sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 }, direction: isRtl ? 'rtl' : 'ltr' }}>
        <Stack spacing={3}>
          <ProductHeaderCard />

          <DetailSectionCard
            title={translate('resources.products.section.pricing', { _: 'Pricing' })}
            description={translate('resources.products.details.pricing', { _: 'Product pricing details' })}
            icon={<MoneyIcon />}
            color="success"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.products.fields.price', { _: 'Price' })}
                value={`$${record.price?.toFixed(2) || '0.00'}`}
              />
              <DetailItem
                label={translate('resources.products.fields.cost_price', { _: 'Cost Price' })}
                value={`$${record.cost_price?.toFixed(2) || '0.00'}`}
              />
            </Box>
          </DetailSectionCard>

          <DetailSectionCard
            title={translate('resources.products.section.bandwidth', { _: 'Bandwidth & Quota' })}
            description={translate('resources.products.details.bandwidth_desc', { _: 'Limits configured for this product' })}
            icon={<SpeedIcon />}
            color="warning"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.products.fields.up_rate', { _: 'Upload Rate' })}
                value={formatRate(record.up_rate)}
              />
              <DetailItem
                label={translate('resources.products.fields.down_rate', { _: 'Download Rate' })}
                value={formatRate(record.down_rate)}
              />
              <DetailItem
                label={translate('resources.products.fields.data_quota', { _: 'Data Quota' })}
                value={formatQuota(record.data_quota)}
              />
              <DetailItem
                label={translate('resources.products.fields.validity', { _: 'Validity' })}
                value={validityDisplay}
              />
              <DetailItem
                label={translate('resources.products.fields.idle_timeout', { _: 'Idle Timeout' })}
                value={record.idle_timeout > 0 ? `${record.idle_timeout} ${translate('resources.products.units.seconds')}` : translate('resources.products.units.unlimited')}
              />
              <DetailItem
                label={translate('resources.products.fields.session_timeout', { _: 'Session Timeout' })}
                value={record.session_timeout > 0 ? `${record.session_timeout} ${translate('resources.products.units.seconds')}` : translate('resources.products.units.unlimited')}
              />
            </Box>
          </DetailSectionCard>

          <DetailSectionCard
            title={translate('resources.products.details.linked_profile', { _: 'Linked Profile' })}
            description={translate('resources.products.details.profile_desc', { _: 'The technical RADIUS profile attached to this product' })}
            icon={<DataIcon />}
            color="info"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.products.fields.radius_profile_id', { _: 'Radius Profile' })}
                value={
                  <ReferenceField source="radius_profile_id" reference="radius-profiles">
                    <TextField source="name" />
                  </ReferenceField>
                }
              />
            </Box>
          </DetailSectionCard>

          <DetailSectionCard
            title={translate('resources.products.details.time_info', { _: 'Time Information' })}
            description={translate('resources.products.details.time_desc', { _: 'Creation and modification dates' })}
            icon={<TimeIcon />}
            color="info"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.products.fields.created_at', { _: 'Created At' })}
                value={formatTimestamp(record.created_at)}
              />
              <DetailItem
                label={translate('resources.products.fields.updated_at', { _: 'Updated At' })}
                value={formatTimestamp(record.updated_at)}
              />
            </Box>
          </DetailSectionCard>

          <DetailSectionCard
            title={translate('resources.products.section.remark', { _: 'Remarks' })}
            description={translate('resources.products.details.remarks_desc', { _: 'Additional notes or descriptions' })}
            icon={<NoteIcon />}
            color="primary"
          >
            <Box
              sx={{
                p: 2,
                borderRadius: 2,
                bgcolor: theme =>
                  theme.palette.mode === 'dark'
                    ? 'rgba(255, 255, 255, 0.02)'
                    : 'rgba(0, 0, 0, 0.02)',
                border: theme => `1px solid ${theme.palette.divider}`,
                minHeight: 80,
              }}
            >
              <Typography
                variant="body2"
                sx={{
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  color: record.remark ? 'text.primary' : 'text.disabled',
                  fontStyle: record.remark ? 'normal' : 'italic',
                }}
              >
                {record.remark || 'No remark added.'}
              </Typography>
            </Box>
          </DetailSectionCard>
        </Stack>
      </Box>
    </>
  );
};

export const ProductShow = (props: ShowProps) => (
  <Show {...props} emptyWhileLoading>
    <ProductDetails />
  </Show>
);


const ValidityInput = () => {
  const record = useRecordContext();
  const { translate, isRtl } = useFormatters();
  
  // Set default values if record is present (for Edit).
  // React-admin will initialize the form with these defaultValues.
  let initUnit = 'days';
  let initVal: number | undefined = 30;

  if (record && record.validity_seconds !== undefined) {
    const seconds = record.validity_seconds;
    if (seconds === 0) {
      initUnit = 'days';
      initVal = undefined;
    } else if (seconds % 86400 === 0) {
      initUnit = 'days';
      initVal = seconds / 86400;
    } else if (seconds % 3600 === 0) {
      initUnit = 'hours';
      initVal = seconds / 3600;
    } else if (seconds % 60 === 0) {
      initUnit = 'minutes';
      initVal = seconds / 60;
    }
  }

  return (
    <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, width: '100%', direction: isRtl ? 'rtl' : 'ltr' }}>
      <Box>
        <NumberInput
          source="validity_value_virtual"
          label={translate('resources.products.fields.validity', { _: 'Validity Duration' })}
          placeholder="0"
          defaultValue={initVal}
          fullWidth
          size="small"
          inputProps={{ style: { textAlign: isRtl ? 'right' : 'left', direction: isRtl ? 'rtl' : 'ltr' } }}
          InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
        />
      </Box>
      <Box>
        <SelectInput
          source="validity_unit_virtual"
          label={translate('common.unit', { _: 'Unit' })}
          defaultValue={initUnit}
          choices={[
            { id: 'minutes', name: translate('resources.products.units.minutes', { _: 'Minutes' }) },
            { id: 'hours', name: translate('resources.products.units.hours', { _: 'Hours' }) },
            { id: 'days', name: translate('resources.products.units.days', { _: 'Days' }) },
          ]}
          fullWidth
          size="small"
          InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
        />
      </Box>
    </Box>
  );
};

const DataQuotaInput = () => {
  const record = useRecordContext();
  const { translate, isRtl } = useFormatters();

  let initUnit = 'MB';
  let initVal: number | undefined = undefined;

  if (record && record.data_quota !== undefined) {
    const mb = record.data_quota;
    if (mb > 0 && mb % 1024 === 0) {
      initUnit = 'GB';
      initVal = mb / 1024;
    } else {
      initUnit = 'MB';
      initVal = mb;
    }
  }

  return (
    <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, width: '100%', direction: isRtl ? 'rtl' : 'ltr' }}>
      <Box>
        <NumberInput
          source="data_quota_virtual"
          label={translate('resources.products.fields.data_quota', { _: 'Data Quota' })}
          placeholder="0"
          defaultValue={initVal}
          fullWidth
          size="small"
          inputProps={{ style: { textAlign: isRtl ? 'right' : 'left', direction: isRtl ? 'rtl' : 'ltr' } }}
          InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
        />
      </Box>
      <Box>
        <SelectInput
          source="data_quota_unit_virtual"
          label={translate('common.unit', { _: 'Unit' })}
          defaultValue={initUnit}
          choices={[
            { id: 'MB', name: translate('resources.products.units.mb', { _: 'MB' }) },
            { id: 'GB', name: translate('resources.products.units.gb', { _: 'GB' }) },
          ]}
          fullWidth
          size="small"
          InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
        />
      </Box>
    </Box>
  );
};

// Form transformations to compute validity_seconds and data_quota before saving
const transformProduct = (data: any) => {
  const transformed = { ...data };

  // Calculate Validity
  if (data.validity_value_virtual !== undefined && data.validity_unit_virtual !== undefined) {
    let multiplier = 60;
    if (data.validity_unit_virtual === 'hours') multiplier = 3600;
    if (data.validity_unit_virtual === 'days') multiplier = 86400;
    transformed.validity_seconds = data.validity_value_virtual * multiplier;
  }

  // Calculate Data Quota
  if (data.data_quota_virtual !== undefined && data.data_quota_unit_virtual !== undefined) {
    const multiplier = data.data_quota_unit_virtual === 'GB' ? 1024 : 1;
    transformed.data_quota = data.data_quota_virtual * multiplier;
  }

  // Clean up virtual fields
  delete transformed.validity_value_virtual;
  delete transformed.validity_unit_virtual;
  delete transformed.data_quota_virtual;
  delete transformed.data_quota_unit_virtual;

  return transformed;
};

export const ProductCreate = (props: CreateProps) => {
  const { translate, isRtl } = useFormatters();
  
  const textInputProps = { style: { textAlign: isRtl ? 'right' : 'left', direction: isRtl ? 'rtl' : 'ltr' } } as const;
  const numInputProps = { style: { textAlign: isRtl ? 'right' : 'left', direction: isRtl ? 'rtl' : 'ltr' } } as const;

  return (
    <Create {...props} transform={transformProduct}>
      <SimpleForm sx={formLayoutSx}>
        <ProductFormBanner />
        
        {/* Row 1: Basic Information */}
        <FormSection
          title={translate('resources.products.section.basic', { _: 'Basic Information' })}
        >
          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, width: '100%', direction: isRtl ? 'rtl' : 'ltr' }}>
            <Box>
              <TextInput 
                source="name" 
                validate={[required()]} 
                fullWidth 
                size="small"
                label={translate('resources.products.fields.name', { _: 'Product Name' })} 
                inputProps={textInputProps} 
                InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
              />
            </Box>
            <Box>
              <ReferenceInput source="radius_profile_id" reference="radius-profiles">
                <SelectInput 
                  optionText="name" 
                  validate={[required()]} 
                  fullWidth 
                  size="small"
                  label={translate('resources.products.fields.radius_profile_id', { _: 'RADIUS Profile' })} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </ReferenceInput>
            </Box>
            <Box>
              <TextInput 
                source="color" 
                type="color" 
                fullWidth 
                size="small"
                label={translate('resources.products.fields.color', { _: 'Product Color' })} 
                defaultValue="#1976d2" 
                InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
              />
            </Box>
            <Box>
              <SelectInput 
                source="status" 
                label={translate('resources.products.fields.status', { _: 'Status' })} 
                choices={[
                  { id: 'enabled', name: translate('resources.products.status.enabled', { _: 'Enabled' }) },
                  { id: 'disabled', name: translate('resources.products.status.disabled', { _: 'Disabled' }) },
                ]} 
                defaultValue="enabled" 
                fullWidth 
                size="small"
                InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
              />
            </Box>
          </Box>
        </FormSection>

        {/* Row 2: Pricing & Bandwidth */}
        <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 2, width: '100%' }}>
          <FormSection
            title={translate('resources.products.section.pricing', { _: 'Pricing' })}
          >
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, width: '100%', direction: isRtl ? 'rtl' : 'ltr' }}>
              <Box>
                <NumberInput 
                  source="price" 
                  validate={[required()]} 
                  fullWidth 
                  size="small"
                  label={translate('resources.products.fields.price', { _: 'Price' })} 
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
              <Box>
                <NumberInput 
                  source="cost_price" 
                  validate={[required()]} 
                  fullWidth 
                  size="small"
                  label={translate('resources.products.fields.cost_price', { _: 'Cost Price' })} 
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
            </Box>
          </FormSection>

          <FormSection
            title={translate('resources.products.section.bandwidth', { _: 'Bandwidth Limit' })}
          >
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, width: '100%', direction: isRtl ? 'rtl' : 'ltr' }}>
              <Box>
                <NumberInput 
                  source="up_rate" 
                  label={translate('resources.products.fields.up_rate', { _: 'Upload Rate' })} 
                  placeholder="0" 
                  fullWidth 
                  size="small"
                  helperText={`0 = ${translate('resources.products.units.unlimited')}`} 
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
              <Box>
                <NumberInput 
                  source="down_rate" 
                  label={translate('resources.products.fields.down_rate', { _: 'Download Rate' })} 
                  placeholder="0" 
                  fullWidth 
                  size="small"
                  helperText={`0 = ${translate('resources.products.units.unlimited')}`} 
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
            </Box>
          </FormSection>
        </Box>

        {/* Row 3: Quota & Validity Leveled */}
        <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 2, width: '100%' }}>
          <FormSection
            title={translate('resources.products.section.data_quota', { _: 'Data Quota' })}
          >
            <DataQuotaInput />
          </FormSection>

          <FormSection
            title={translate('resources.products.section.validity', { _: 'Validity & Session' })}
          >
            <ValidityInput />
            <Box sx={{ mt: 2, display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, width: '100%', direction: isRtl ? 'rtl' : 'ltr' }}>
              <Box>
                <NumberInput 
                  source="idle_timeout" 
                  label={translate('resources.products.fields.idle_timeout', { _: 'Idle Timeout' })} 
                  placeholder="0" 
                  fullWidth 
                  size="small"
                  helperText={translate('resources.products.fields.idle_timeout_helper', { _: 'Seconds of inactivity before logout' })}
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
              <Box>
                <NumberInput 
                  source="session_timeout" 
                  label={translate('resources.products.fields.session_timeout', { _: 'Session Timeout' })} 
                  placeholder="0" 
                  fullWidth 
                  size="small"
                  helperText={translate('resources.products.fields.session_timeout_helper', { _: 'Maximum session duration in seconds' })}
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
            </Box>
          </FormSection>
        </Box>

        {/* Row 4: Remark */}
        <FormSection
          title={translate('resources.products.section.remark', { _: 'Remark' })}
        >
          <TextInput 
            source="remark" 
            label={translate('resources.products.fields.remark', { _: 'Remark' })} 
            multiline 
            fullWidth 
            size="small"
            minRows={2} 
            inputProps={textInputProps} 
            InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
          />
        </FormSection>
      </SimpleForm>
    </Create>
  );
};

export const ProductEdit = (props: EditProps) => {
  const { translate, isRtl } = useFormatters();

  const textInputProps = { style: { textAlign: isRtl ? 'right' : 'left', direction: isRtl ? 'rtl' : 'ltr' } } as const;
  const numInputProps = { style: { textAlign: isRtl ? 'right' : 'left', direction: isRtl ? 'rtl' : 'ltr' } } as const;

  return (
    <Edit {...props} title={<ProductTitle />} transform={transformProduct}>
      <SimpleForm sx={formLayoutSx}>
        <ProductFormBanner />
        
        {/* Row 1: Basic Information */}
        <FormSection
          title={translate('resources.products.section.basic', { _: 'Basic Information' })}
        >
          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, width: '100%', direction: isRtl ? 'rtl' : 'ltr' }}>
            <Box>
              <TextInput 
                source="id" 
                disabled 
                fullWidth 
                size="small"
                label={translate('resources.products.fields.id', { _: 'ID' })} 
                inputProps={textInputProps} 
                InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
              />
            </Box>
            <Box>
              <TextInput 
                source="name" 
                validate={[required()]} 
                fullWidth 
                size="small"
                label={translate('resources.products.fields.name', { _: 'Product Name' })} 
                inputProps={textInputProps} 
                InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
              />
            </Box>
            <Box>
              <ReferenceInput source="radius_profile_id" reference="radius-profiles">
                <SelectInput 
                  optionText="name" 
                  validate={[required()]} 
                  fullWidth 
                  size="small"
                  label={translate('resources.products.fields.radius_profile_id', { _: 'RADIUS Profile' })} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </ReferenceInput>
            </Box>
            <Box>
              <TextInput 
                source="color" 
                type="color" 
                fullWidth 
                size="small"
                label={translate('resources.products.fields.color', { _: 'Product Color' })} 
                InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
              />
            </Box>
            <Box sx={{ gridColumn: '1 / -1' }}>
              <SelectInput 
                source="status" 
                label={translate('resources.products.fields.status', { _: 'Status' })} 
                choices={[
                  { id: 'enabled', name: translate('resources.products.status.enabled', { _: 'Enabled' }) },
                  { id: 'disabled', name: translate('resources.products.status.disabled', { _: 'Disabled' }) },
                ]} 
                fullWidth 
                size="small"
                InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
              />
            </Box>
          </Box>
        </FormSection>

        {/* Row 2: Pricing & Bandwidth */}
        <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 2, width: '100%' }}>
          <FormSection
            title={translate('resources.products.section.pricing', { _: 'Pricing' })}
          >
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, width: '100%', direction: isRtl ? 'rtl' : 'ltr' }}>
              <Box>
                <NumberInput 
                  source="price" 
                  validate={[required()]} 
                  fullWidth 
                  size="small"
                  label={translate('resources.products.fields.price', { _: 'Price' })} 
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
              <Box>
                <NumberInput 
                  source="cost_price" 
                  validate={[required()]} 
                  fullWidth 
                  size="small"
                  label={translate('resources.products.fields.cost_price', { _: 'Cost Price' })} 
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
            </Box>
          </FormSection>

          <FormSection
            title={translate('resources.products.section.bandwidth', { _: 'Bandwidth Limit' })}
          >
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, width: '100%', direction: isRtl ? 'rtl' : 'ltr' }}>
              <Box>
                <NumberInput 
                  source="up_rate" 
                  label={translate('resources.products.fields.up_rate', { _: 'Upload Rate' })} 
                  placeholder="0" 
                  fullWidth 
                  size="small"
                  helperText={`0 = ${translate('resources.products.units.unlimited')}`} 
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
              <Box>
                <NumberInput 
                  source="down_rate" 
                  label={translate('resources.products.fields.down_rate', { _: 'Download Rate' })} 
                  placeholder="0" 
                  fullWidth 
                  size="small"
                  helperText={`0 = ${translate('resources.products.units.unlimited')}`} 
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
            </Box>
          </FormSection>
        </Box>

        {/* Row 3: Quota & Validity Leveled */}
        <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 2, width: '100%' }}>
          <FormSection
            title={translate('resources.products.section.data_quota', { _: 'Data Quota' })}
          >
            <DataQuotaInput />
          </FormSection>

          <FormSection
            title={translate('resources.products.section.validity', { _: 'Validity & Session' })}
          >
            <ValidityInput />
            <Box sx={{ mt: 2, display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2, width: '100%', direction: isRtl ? 'rtl' : 'ltr' }}>
              <Box>
                <NumberInput 
                  source="idle_timeout" 
                  label={translate('resources.products.fields.idle_timeout', { _: 'Idle Timeout' })} 
                  placeholder="0" 
                  fullWidth 
                  size="small"
                  helperText={translate('resources.products.fields.idle_timeout_helper', { _: 'Seconds of inactivity before logout' })}
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
              <Box>
                <NumberInput 
                  source="session_timeout" 
                  label={translate('resources.products.fields.session_timeout', { _: 'Session Timeout' })} 
                  placeholder="0" 
                  fullWidth 
                  size="small"
                  helperText={translate('resources.products.fields.session_timeout_helper', { _: 'Maximum session duration in seconds' })}
                  inputProps={numInputProps} 
                  InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
                />
              </Box>
            </Box>
          </FormSection>
        </Box>

        {/* Row 4: Remark */}
        <FormSection
          title={translate('resources.products.section.remark', { _: 'Remark' })}
        >
          <TextInput 
            source="remark" 
            label={translate('resources.products.fields.remark', { _: 'Remark' })} 
            multiline 
            fullWidth 
            size="small"
            minRows={2} 
            inputProps={textInputProps} 
            InputLabelProps={{ sx: { transformOrigin: isRtl ? 'top right' : 'top left', left: isRtl ? 'auto' : 0, right: isRtl ? 24 : 'auto' } }}
          />
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};
