import {
  Show,
  SimpleShowLayout,
  TextField,
  DateField,
  FunctionField,
  TopToolbar,
  ListButton,
  Button,
  useRecordContext,
} from 'react-admin';
import {
  Box,
  Card,
  CardContent,
  Grid,
  Typography,
  Divider,
  Paper,
  Stack,
  Chip,
} from '@mui/material';
import { Download, Print } from '@mui/icons-material';

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value);
};

const InvoiceShowActions = () => (
  <TopToolbar>
    <ListButton />
    <Button label="Download PDF" startIcon={<Download />} />
    <Button label="Print" startIcon={<Print />} />
  </TopToolbar>
);

const InvoiceStatusBadge = ({ status }: { status: string }) => {
  const statusConfig: Record<string, { color: string; label: string }> = {
    paid: { color: '#10b981', label: 'Paid' },
    pending: { color: '#f59e0b', label: 'Pending' },
    overdue: { color: '#ef4444', label: 'Overdue' },
    draft: { color: '#6b7280', label: 'Draft' },
  };

  const config = statusConfig[status] || statusConfig.draft;

  return (
    <Chip
      label={config.label}
      sx={{
        backgroundColor: `${config.color}15`,
        color: config.color,
        border: `1px solid ${config.color}40`,
        fontWeight: 600,
        textTransform: 'capitalize',
      }}
    />
  );
};

const InvoiceBreakdown = () => {
  const record = useRecordContext();

  if (!record) return null;

  return (
    <Card
      sx={{
        background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(248,250,252,0.98) 100%)',
        border: '1px solid rgba(148, 163, 184, 0.1)',
        borderRadius: 2,
        height: '100%',
      }}
    >
      <CardContent>
        <Typography variant="h6" sx={{ fontWeight: 600, mb: 3 }}>
          Invoice Breakdown
        </Typography>

        <Stack spacing={2}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              Base Fee
            </Typography>
            <Typography variant="body1" sx={{ fontWeight: 500 }}>
              {formatCurrency(record.base_fee)}
            </Typography>
          </Box>

          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Box>
              <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                User Overage
              </Typography>
              <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                {record.overage_users} users × ${record.overage_fee}/user
              </Typography>
            </Box>
            <Typography variant="body1" sx={{ fontWeight: 500 }}>
              {formatCurrency(record.user_overage_fee)}
            </Typography>
          </Box>

          <Divider sx={{ my: 1 }} />

          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              Subtotal
            </Typography>
            <Typography variant="body1" sx={{ fontWeight: 600 }}>
              {formatCurrency(record.base_fee + record.user_overage_fee)}
            </Typography>
          </Box>

          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              Tax (15%)
            </Typography>
            <Typography variant="body1" sx={{ fontWeight: 500 }}>
              {formatCurrency(record.tax_amount)}
            </Typography>
          </Box>

          <Divider sx={{ my: 1 }} />

          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="h6" sx={{ fontWeight: 700 }}>
              Total
            </Typography>
            <Typography variant="h5" sx={{ fontWeight: 700, color: 'primary.main' }}>
              {formatCurrency(record.total_amount)}
            </Typography>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
};

const UsageStats = () => {
  const record = useRecordContext();

  if (!record) return null;

  return (
    <Card
      sx={{
        background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(248,250,252,0.98) 100%)',
        border: '1px solid rgba(148, 163, 184, 0.1)',
        borderRadius: 2,
        height: '100%',
      }}
    >
      <CardContent>
        <Typography variant="h6" sx={{ fontWeight: 600, mb: 3 }}>
          Usage Statistics
        </Typography>

        <Stack spacing={3}>
          <Box>
            <Typography variant="body2" sx={{ color: 'text.secondary', mb: 1 }}>
              Current Users
            </Typography>
            <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
              {record.current_users?.toLocaleString()}
            </Typography>
          </Box>

          <Box>
            <Typography variant="body2" sx={{ color: 'text.secondary', mb: 1 }}>
              Included Users
            </Typography>
            <Typography variant="h5" sx={{ fontWeight: 600, color: 'text.primary' }}>
              {record.included_users?.toLocaleString()}
            </Typography>
          </Box>

          <Box>
            <Typography variant="body2" sx={{ color: 'text.secondary', mb: 1 }}>
              Overage Users
            </Typography>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <Typography
                variant="h5"
                sx={{
                  fontWeight: 700,
                  color: record.overage_users > 0 ? '#f59e0b' : 'text.primary',
                }}
              >
                {record.overage_users?.toLocaleString()}
              </Typography>
              {record.overage_users > 0 && (
                <Chip
                  label="Over limit"
                  size="small"
                  sx={{
                    backgroundColor: '#f59e0b15',
                    color: '#f59e0b',
                    fontWeight: 600,
                  }}
                />
              )}
            </Box>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
};

export const InvoiceShow = () => {
  return (
    <Show actions={<InvoiceShowActions />}>
      <SimpleShowLayout>
        <Grid container spacing={3}>
          {/* Invoice Header */}
          <Grid size={12}>
            <Paper
              sx={{
                p: 3,
                background: 'linear-gradient(135deg, #1e3a8a 0%, #1e40af 100%)',
                color: 'white',
                borderRadius: 2,
              }}
            >
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Box>
                  <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
                    Invoice
                  </Typography>
                  <TextField source="invoice_number" sx={{ color: 'rgba(255,255,255,0.8)' }} />
                </Box>
                <FunctionField
                  source="status"
                  render={(record: any) => <InvoiceStatusBadge status={record.status} />}
                />
              </Box>
            </Paper>
          </Grid>

          {/* Invoice Details */}
          <Grid size={{ xs: 12, md: 6 }}>
            <InvoiceBreakdown />
          </Grid>

          <Grid size={{ xs: 12, md: 6 }}>
            <UsageStats />
          </Grid>

          {/* Billing Period */}
          <Grid size={12}>
            <Card
              sx={{
                background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(248,250,252,0.98) 100%)',
                border: '1px solid rgba(148, 163, 184, 0.1)',
                borderRadius: 2,
              }}
            >
              <CardContent>
                <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>
                  Billing Period
                </Typography>
                <Grid container spacing={2}>
                  <Grid size={6}>
                    <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                      Period Start
                    </Typography>
                    <DateField source="period_start" showTime />
                  </Grid>
                  <Grid size={6}>
                    <Typography variant="body2" sx={{ color: 'text.secondary' }}>
                      Period End
                    </Typography>
                    <DateField source="period_end" showTime />
                  </Grid>
                </Grid>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      </SimpleShowLayout>
    </Show>
  );
};

InvoiceShow.displayName = 'InvoiceShow';
