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
  RecordContextProvider
} from 'react-admin';
import { useFormContext } from 'react-hook-form';
import React from 'react';
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
  FieldGrid,
  FieldGridItem,
  formLayoutSx,
  DetailItem,
  DetailSectionCard,
  EmptyValue,
} from '../components';

const ProductTitle = () => {
  const record = useRecordContext();
  return <span>Product {record ? `"${record.name}"` : ''}</span>;
};


const ProductGrid = () => {
  const { data, isLoading } = useListContext();
  if (isLoading || !data) return null;
  return (
    <Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)' }} gap={2} p={2} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}>
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
              <EditButton label="" size="small" />
              <DeleteButton label="" size="small" />
              <ShowButton label="" size="small" />
            </CardActions>
          </Card>
        </RecordContextProvider>
      ))}
    </Box>
  );
};
export const ProductList = (props: ListProps) => {
  const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
  return (
    <List {...props} sort={{ field: 'id', order: 'DESC' }}>
      {isSmall ? (
        <ProductGrid />
      ) : (
        <Datagrid rowClick="show">
          <TextField source="id" />
          <TextField source="name" />
          <ReferenceField source="radius_profile_id" reference="radius-profiles">
            <TextField source="name" />
          </ReferenceField>
          <NumberField source="price" options={{ style: 'currency', currency: 'USD' }} />
          <NumberField source="up_rate" label="Up Rate (Kbps)" />
          <NumberField source="down_rate" label="Down Rate (Kbps)" />
          <NumberField source="data_quota" label="Quota (MB)" />
          <TextField source="status" />
          <DateField source="updated_at" showTime />
          <EditButton />
          <DeleteButton />
        </Datagrid>
      )}
    </List>
  );
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

const formatRate = (rate?: number): string => {
  if (rate === undefined || rate === null) return '-';
  if (rate === 0) return 'Unlimited';
  if (rate >= 1024) {
    return `${(rate / 1024).toFixed(1)} Mbps`;
  }
  return `${rate} Kbps`;
};

const formatQuota = (quota?: number): string => {
  if (quota === undefined || quota === null) return '-';
  if (quota === 0) return 'Unlimited';
  if (quota >= 1024) {
    return `${(quota / 1024).toFixed(1)} GB`;
  }
  return `${quota} MB`;
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

const ProductHeaderCard = () => {
  const record = useRecordContext();
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
                Price
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
                Upload
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
                Download
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
                Data Quota
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

  if (!record) {
    return null;
  }

  const validitySeconds = record.validity_seconds || 0;
  const validityDisplay = validitySeconds === 0 ? 'Unlimited' :
    validitySeconds >= 86400 && validitySeconds % 86400 === 0 ? `${validitySeconds / 86400} Days` :
      validitySeconds >= 3600 && validitySeconds % 3600 === 0 ? `${validitySeconds / 3600} Hours` :
        `${validitySeconds / 60} Minutes`;

  return (
    <>
      <style>{printStyles}</style>
      <Box className="printable-content" sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 } }}>
        <Stack spacing={3}>
          <ProductHeaderCard />

          <DetailSectionCard
            title="Pricing"
            description="Product pricing details"
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
                label="Price"
                value={`$${record.price?.toFixed(2) || '0.00'}`}
              />
              <DetailItem
                label="Cost Price"
                value={`$${record.cost_price?.toFixed(2) || '0.00'}`}
              />
            </Box>
          </DetailSectionCard>

          <DetailSectionCard
            title="Bandwidth & Quota"
            description="Limits configured for this product"
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
                label="Upload Rate"
                value={formatRate(record.up_rate)}
              />
              <DetailItem
                label="Download Rate"
                value={formatRate(record.down_rate)}
              />
              <DetailItem
                label="Data Quota"
                value={formatQuota(record.data_quota)}
              />
              <DetailItem
                label="Validity"
                value={validityDisplay}
              />
            </Box>
          </DetailSectionCard>

          <DetailSectionCard
            title="Linked Profile"
            description="The technical RADIUS profile attached to this product"
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
                label="Radius Profile"
                value={
                  <ReferenceField source="radius_profile_id" reference="radius-profiles">
                    <TextField source="name" />
                  </ReferenceField>
                }
              />
            </Box>
          </DetailSectionCard>

          <DetailSectionCard
            title="Time Information"
            description="Creation and modification dates"
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
                label="Created At"
                value={formatTimestamp(record.created_at)}
              />
              <DetailItem
                label="Updated At"
                value={formatTimestamp(record.updated_at)}
              />
            </Box>
          </DetailSectionCard>

          <DetailSectionCard
            title="Remarks"
            description="Additional notes or descriptions"
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
  const { setValue, getValues } = useFormContext();
  // Default to 'days' unless value is small (e.g. < 1 hour)
  const currentSeconds = getValues('validity_seconds') || 0;
  const initialUnit = currentSeconds > 0 && currentSeconds % 86400 === 0 ? 'days' :
    currentSeconds > 0 && currentSeconds % 3600 === 0 ? 'hours' : 'minutes';

  // Calculate initial value based on unit
  const initialValue = initialUnit === 'days' ? currentSeconds / 86400 :
    initialUnit === 'hours' ? currentSeconds / 3600 :
      currentSeconds / 60;

  const [unit, setUnit] = React.useState(initialUnit);
  const [val, setVal] = React.useState(initialValue > 0 ? initialValue : 30); // Default 30

  // Effect to update the actual source field when unit or val changes
  React.useEffect(() => {
    let multiplier = 60;
    if (unit === 'hours') multiplier = 3600;
    if (unit === 'days') multiplier = 86400;

    setValue('validity_seconds', val * multiplier);
  }, [unit, val, setValue]);

  return (
    <FieldGrid columns={{ xs: 1, sm: 2 }}>
      <FieldGridItem>
        <NumberInput
          source="validity_value_virtual" // Virtual field
          label="Validity Duration"
          value={val}
          onChange={(e) => setVal(Number(e.target.value))}
          defaultValue={30}
          fullWidth
        />
      </FieldGridItem>
      <FieldGridItem>
        <SelectInput
          source="validity_unit_virtual" // Virtual field
          label="Unit"
          choices={[
            { id: 'minutes', name: 'Minutes' },
            { id: 'hours', name: 'Hours' },
            { id: 'days', name: 'Days' },
          ]}
          value={unit}
          onChange={(e) => setUnit(e.target.value)}
          defaultValue="days"
          fullWidth
          disableValue="validity_seconds"
        />
      </FieldGridItem>
      <NumberInput source="validity_seconds" style={{ display: 'none' }} />
    </FieldGrid>
  );
};

const DataQuotaInput = () => {
  const { setValue, getValues } = useFormContext();
  const currentMB = getValues('data_quota') || 0;
  const initialUnit = currentMB > 0 && currentMB % 1024 === 0 ? 'GB' : 'MB';
  const initialValue = initialUnit === 'GB' ? currentMB / 1024 : currentMB;

  const [unit, setUnit] = React.useState(initialUnit);
  const [val, setVal] = React.useState(initialValue);

  React.useEffect(() => {
    const multiplier = unit === 'GB' ? 1024 : 1;
    setValue('data_quota', val * multiplier);
  }, [unit, val, setValue]);

  return (
    <FieldGrid columns={{ xs: 1, sm: 2 }}>
      <FieldGridItem>
        <NumberInput
          source="data_quota_virtual"
          label="Data Quota"
          value={val}
          onChange={(e) => setVal(Number(e.target.value))}
          fullWidth
        />
      </FieldGridItem>
      <FieldGridItem>
        <SelectInput
          source="data_quota_unit_virtual"
          label="Unit"
          choices={[
            { id: 'MB', name: 'MB' },
            { id: 'GB', name: 'GB' },
          ]}
          value={unit}
          onChange={(e) => setUnit(e.target.value)}
          fullWidth
        />
      </FieldGridItem>
      <NumberInput source="data_quota" style={{ display: 'none' }} />
    </FieldGrid>
  );
};

export const ProductCreate = (props: CreateProps) => {
  const translate = useTranslate();
  return (
    <Create {...props}>
      <SimpleForm sx={formLayoutSx}>
        <FormSection
          title={translate('resources.products.section.basic', { _: 'Basic Information' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput source="name" validate={[required()]} fullWidth />
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="radius_profile_id" reference="radius-profiles">
                <SelectInput optionText="name" validate={[required()]} fullWidth />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem>
              <TextInput source="color" type="color" fullWidth label="Product Color" defaultValue="#1976d2" />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput source="status" choices={[
                { id: 'enabled', name: 'Enabled' },
                { id: 'disabled', name: 'Disabled' },
              ]} defaultValue="enabled" fullWidth />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.products.section.pricing', { _: 'Pricing' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <NumberInput source="price" validate={[required()]} fullWidth />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput source="cost_price" validate={[required()]} fullWidth />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.products.section.bandwidth', { _: 'Bandwidth Limit' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <NumberInput source="up_rate" label="Upload Rate (Kbps)" defaultValue={0} fullWidth helperText="0 = Unlimited" />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput source="down_rate" label="Download Rate (Kbps)" defaultValue={0} fullWidth helperText="0 = Unlimited" />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.products.section.data_quota', { _: 'Data Quota' })}
        >
          <DataQuotaInput />
        </FormSection>

        <FormSection
          title={translate('resources.products.section.validity', { _: 'Validity Limit' })}
        >
          <ValidityInput />
        </FormSection>

        <FormSection
          title={translate('resources.products.section.remark', { _: 'Remark' })}
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput source="remark" multiline fullWidth minRows={3} />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Create>
  );
};

export const ProductEdit = (props: EditProps) => {
  const translate = useTranslate();
  return (
    <Edit {...props} title={<ProductTitle />}>
      <SimpleForm sx={formLayoutSx}>
        <FormSection
          title={translate('resources.products.section.basic', { _: 'Basic Information' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput source="id" disabled fullWidth />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput source="name" validate={[required()]} fullWidth />
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="radius_profile_id" reference="radius-profiles">
                <SelectInput optionText="name" validate={[required()]} fullWidth />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem>
              <TextInput source="color" type="color" fullWidth label="Product Color" />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput source="status" choices={[
                { id: 'enabled', name: 'Enabled' },
                { id: 'disabled', name: 'Disabled' },
              ]} fullWidth />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.products.section.pricing', { _: 'Pricing' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <NumberInput source="price" validate={[required()]} fullWidth />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput source="cost_price" validate={[required()]} fullWidth />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.products.section.bandwidth', { _: 'Bandwidth Limit' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <NumberInput source="up_rate" label="Upload Rate (Kbps)" fullWidth helperText="0 = Unlimited" />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput source="down_rate" label="Download Rate (Kbps)" fullWidth helperText="0 = Unlimited" />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.products.section.data_quota', { _: 'Data Quota' })}
        >
          <DataQuotaInput />
        </FormSection>

        <FormSection
          title={translate('resources.products.section.validity', { _: 'Validity Limit' })}
        >
          <ValidityInput />
        </FormSection>

        <FormSection
          title={translate('resources.products.section.remark', { _: 'Remark' })}
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput source="remark" multiline fullWidth minRows={3} />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};
