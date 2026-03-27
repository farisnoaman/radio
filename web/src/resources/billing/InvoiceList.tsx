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
  FilterButton,
  CreateButton,
  ExportButton,
} from 'react-admin';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Stack,
  useTheme,
  useMediaQuery,
} from '@mui/material';
import { Receipt, Pending, CheckCircle, Warning } from '@mui/icons-material';
import { MetricCard } from '../../components/saas';

const InvoiceAside = () => {
  const { data } = useListContext();
  const translate = useTranslate();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  const totalAmount = data?.reduce((sum: number, invoice: any) => sum + invoice.total_amount, 0) || 0;
  const paidAmount = data?.filter((i: any) => i.status === 'paid').reduce((sum: number, i: any) => sum + i.total_amount, 0) || 0;
  const pendingAmount = data?.filter((i: any) => i.status === 'pending').reduce((sum: number, i: any) => sum + i.total_amount, 0) || 0;
  const overdueAmount = data?.filter((i: any) => i.status === 'overdue').reduce((sum: number, i: any) => sum + i.total_amount, 0) || 0;

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount);
  };

  return (
    <Box sx={{ width: isMobile ? '100%' : 320, ml: isMobile ? 0 : 2, mb: 2 }}>
      <Stack spacing={2}>
        <MetricCard
          title={translate('billing.aside.total_revenue')}
          value={formatCurrency(totalAmount)}
          icon={<Receipt fontSize="small" />}
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
                {translate('billing.paid')}
              </Typography>
            </Box>
            <Typography variant="h5" sx={{ color: '#10b981', fontWeight: 700 }}>
              {formatCurrency(paidAmount)}
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
                {translate('billing.pending')}
              </Typography>
            </Box>
            <Typography variant="h5" sx={{ color: '#f59e0b', fontWeight: 700 }}>
              {formatCurrency(pendingAmount)}
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
              <Warning sx={{ color: '#ef4444', fontSize: 20 }} />
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {translate('billing.overdue')}
              </Typography>
            </Box>
            <Typography variant="h5" sx={{ color: '#ef4444', fontWeight: 700 }}>
              {formatCurrency(overdueAmount)}
            </Typography>
          </CardContent>
        </Card>
      </Stack>
    </Box>
  );
};

const InvoiceActions = () => {
  const translate = useTranslate();
  return (
    <TopToolbar>
      <FilterButton />
      <CreateButton label={translate('billing.generate_invoice')} />
      <ExportButton />
    </TopToolbar>
  );
};

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value);
};

const invoiceStatusColors: Record<string, string> = {
  paid: '#10b981',
  pending: '#f59e0b',
  overdue: '#ef4444',
  draft: '#6b7280',
};

export const InvoiceList = () => {
  const translate = useTranslate();

  return (
    <List
      aside={<InvoiceAside />}
      actions={<InvoiceActions />}
      perPage={25}
      sx={{
        '& .RaList-content': {
          bgcolor: 'background.default',
        },
      }}
    >
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
        <TextField source="invoice_number" label={translate('billing.invoice_id')} />
        <DateField source="created_at" label={translate('billing.issue_date')} />
        <FunctionField
          source="status"
          label={translate('billing.status')}
          render={(record: any) => (
            <Box
              sx={{
                display: 'inline-flex',
                alignItems: 'center',
                px: 2,
                py: 1,
                borderRadius: 1,
                backgroundColor: `${invoiceStatusColors[record.status] || '#6b7280'}15`,
                border: `1px solid ${invoiceStatusColors[record.status] || '#6b7280'}30`,
              }}
            >
              <Box
                sx={{
                  width: 6,
                  height: 6,
                  borderRadius: '50%',
                  backgroundColor: invoiceStatusColors[record.status] || '#6b7280',
                  mr: 1,
                }}
              />
              <Typography
                variant="body2"
                sx={{
                  fontWeight: 600,
                  color: invoiceStatusColors[record.status] || '#6b7280',
                  textTransform: 'capitalize',
                }}
              >
                {translate(`billing.${record.status}`)}
              </Typography>
            </Box>
          )}
        />
        <NumberField
          source="current_users"
          label={translate('billing.total_users')}
          options={{ minimumFractionDigits: 0 }}
        />
        <FunctionField
          source="base_fee"
          label={translate('billing.base_fee')}
          render={(record: any) => (
            <Typography variant="body2" sx={{ fontWeight: 500 }}>
              {formatCurrency(record.base_fee)}
            </Typography>
          )}
        />
        <FunctionField
          source="total_amount"
          label={translate('billing.total_amount')}
          render={(record: any) => (
            <Typography
              variant="body2"
              sx={{
                fontWeight: 700,
                color: 'text.primary',
                fontSize: 15,
              }}
            >
              {formatCurrency(record.total_amount)}
            </Typography>
          )}
        />
        <DateField source="due_date" label={translate('billing.due_date')} />
      </Datagrid>
    </List>
  );
};

InvoiceList.displayName = 'InvoiceList';
