import re

with open('/home/faris/Downloads/toughradius/toughradius/web/src/resources/products.tsx', 'r') as f:
    content = f.read()

import_react_admin_old = """
    EditProps,
} from 'react-admin';"""
import_react_admin_new = """
    EditProps,
    ListButton,
    useNotify,
    useRefresh
} from 'react-admin';"""

content = content.replace(import_react_admin_old, import_react_admin_new)

import_mui_old = """
import { Box } from '@mui/material';
import {
    FormSection,
    FieldGrid,
    FieldGridItem,
    formLayoutSx,
} from '../components';"""

import_mui_new = """
import { Box, Card, CardContent, Stack, Avatar, Typography, Tooltip, IconButton, Chip, alpha } from '@mui/material';
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
} from '../components';"""

content = content.replace(import_mui_old.strip(), import_mui_new.strip())

product_show_old = """export const ProductShow = (props: ShowProps) => (
    <Show {...props} title={<ProductTitle />}>
        <SimpleShowLayout>
            <TextField source="id" />
            <TextField source="name" />
            <ReferenceField source="radius_profile_id" reference="radius-profiles">
                <TextField source="name" />
            </ReferenceField>
            <Box display="flex" gap={2}>
                <NumberField source="price" options={{ style: 'currency', currency: 'USD' }} />
                <NumberField source="cost_price" options={{ style: 'currency', currency: 'USD' }} />
            </Box>
            <Box display="flex" gap={2}>
                <NumberField source="up_rate" label="Upload Rate (Kbps)" />
                <NumberField source="down_rate" label="Download Rate (Kbps)" />
            </Box>
            <Box display="flex" gap={2}>
                <NumberField source="data_quota" label="Data Quota (MB)" />
                <NumberField source="validity_seconds" label="Validity (Seconds)" />
            </Box>
            <TextField source="status" />
            <TextField source="remark" />
            <DateField source="created_at" showTime />
            <DateField source="updated_at" showTime />
        </SimpleShowLayout>
    </Show>
);"""

product_show_new = """const formatTimestamp = (value?: string | number): string => {
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
    const translate = useTranslate();

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
                        color="secondary"
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
);"""

if product_show_old in content:
    content = content.replace(product_show_old, product_show_new)
    with open('/home/faris/Downloads/toughradius/toughradius/web/src/resources/products.tsx', 'w') as f:
        f.write(content)
    print("Successfully replaced content")
else:
    print("Could not find old product show content perfectly.")
