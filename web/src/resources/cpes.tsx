import {
  List,
  Datagrid,
  TextField,
  DateField,
  SelectField,
  useListContext,
  TopToolbar,
  ExportButton,
  RefreshButton,
  useTranslate,
  useLocale,
} from 'react-admin';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Stack,
  Chip,
  Avatar,
  useTheme,
  useMediaQuery,
  alpha
} from '@mui/material';
import { Router as RouterIcon } from '@mui/icons-material';

const CpeListContent = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { data, total } = useListContext();
 
  if (!data || data.length === 0) {
    return null;
  }
 
  return (
    <Box>
      <Card elevation={0} sx={{ borderRadius: 2, border: theme => `1px solid ${theme.palette.divider}`, mb: 2 }}>
        <Box sx={{ px: 2, py: 1, bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.01)', borderBottom: theme => `1px solid ${theme.palette.divider}`, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="body2" color="text.secondary">
            Total: <strong>{total?.toLocaleString() || 0}</strong> Connected Routers
          </Typography>
        </Box>

        {isMobile ? (
          <Box sx={{ p: 2 }}>
            {data?.map((cpe) => (
              <Card key={cpe.id} variant="outlined" sx={{ mb: 2, borderRadius: 3 }}>
                <CardContent>
                  <Stack direction="row" alignItems="center" spacing={2} mb={2}>
                    <Avatar sx={{ bgcolor: 'secondary.main' }}><RouterIcon /></Avatar>
                    <Box sx={{ flexGrow: 1 }}>
                      <Typography variant="h6">{cpe.manufacturer}</Typography>
                      <Typography variant="caption" color="text.secondary">SN: {cpe.serial_number}</Typography>
                    </Box>
                    <Chip label={cpe.status} size="small" color={cpe.status === 'provisioned' ? 'success' : 'warning'} />
                  </Stack>
                  <Box sx={{ bgcolor: alpha(theme.palette.primary.main, 0.04), borderRadius: 2, p: 2 }}>
                    <Stack spacing={1}>
                      <Typography variant="caption"><strong>IP:</strong> {cpe.last_ip || '-'}</Typography>
                      <Typography variant="caption"><strong>Version:</strong> {cpe.software_version || '-'}</Typography>
                      <Typography variant="caption"><strong>Last Inform:</strong> {new Date(cpe.last_inform).toLocaleString()}</Typography>
                    </Stack>
                  </Box>
                </CardContent>
              </Card>
            ))}
          </Box>
        ) : (
          <Datagrid rowClick="show" bulkActionButtons={false} sx={{ '& .RaDatagrid-thead': { bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.02)' } }}>
            <TextField source="manufacturer" label="Brand" />
            <TextField source="serial_number" label="Serial Number" />
            <TextField source="product_class" label="Model" />
            <TextField source="last_ip" label="Last Seen IP" />
            <TextField source="software_version" label="Software" />
            <SelectField source="status" choices={[
              { id: 'pending', name: 'Pending' },
              { id: 'provisioned', name: 'Provisioned' },
              { id: 'failed', name: 'Failed' }
            ]} />
            <DateField source="last_inform" showTime label="Last Check-in" />
          </Datagrid>
        )}
      </Card>
    </Box>
  );
};

const CpeEmptyState = () => {
  const theme = useTheme();
  const translate = useTranslate();
  const locale = useLocale();
  const isRTL = locale === 'ar';
  
  return (
    <Box sx={{ p: 6, textAlign: 'center' }}>
      <Avatar sx={{ bgcolor: 'secondary.main', width: 64, height: 64, mx: 'auto', mb: 2 }}>
        <RouterIcon sx={{ fontSize: 32 }} />
      </Avatar>
      <Typography variant="h5" color="text.secondary" mb={2}>
        {translate('resources.cpes.empty.title', 'No CPE devices detected')}
      </Typography>
      <Typography variant="body2" color="text.secondary" mb={4} maxWidth="500px" mx="auto">
        {translate('resources.cpes.empty.description', 'This page shows TR-069-managed routers (CPE devices). To populate this page, configure your MikroTik to connect to the ACS server or implement API/SSH integration.')}
      </Typography>
      <Card sx={{ borderRadius: 2, border: 1, borderColor: theme.palette.divider, p: 3, maxWidth: '500px', mx: 'auto', textAlign: isRTL ? 'right' : 'left', direction: isRTL ? 'rtl' : 'ltr' }}>
        <Typography variant="body2" fontWeight={600} mb={2}>
          {translate('resources.cpes.empty.setupTitle', 'Quick Setup Guide:')}
        </Typography>
        <Stack direction="column" spacing={1.5}>
          <Box>
            <Typography variant="caption" color="text.secondary" display="block" mb={0.5}>
              {translate('resources.cpes.empty.step1', '1. Configure TR-069 on your MikroTik router:')}
            </Typography>
            <Typography 
              variant="caption" 
              sx={{ 
                fontFamily: 'monospace',
                bgcolor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.05)',
                p: 1,
                borderRadius: 1,
                display: 'block',
                direction: 'ltr',
                textAlign: 'left'
              }}
            >
              /tr069-client set acs-url=http://YOUR_SERVER_IP:7547/cpe enabled=yes
            </Typography>
          </Box>
          <Box>
            <Typography variant="caption" color="text.secondary" display="block" mb={0.5}>
              {translate('resources.cpes.empty.step2', '2. Trigger immediate connection:')}
            </Typography>
            <Typography 
              variant="caption" 
              sx={{ 
                fontFamily: 'monospace',
                bgcolor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.05)',
                p: 1,
                borderRadius: 1,
                display: 'block',
                direction: 'ltr',
                textAlign: 'left'
              }}
            >
              /tr069-client inform
            </Typography>
          </Box>
          <Typography variant="caption" color="text.secondary">
            {translate('resources.cpes.empty.note', 'Or implement API/SSH integration for direct device management (future feature).')}
          </Typography>
        </Stack>
      </Card>
      <Typography variant="caption" color="text.disabled" sx={{ mt: 3, display: 'block' }}>
        {translate('resources.cpes.empty.future', 'Future features: Bulk reboot, firmware upgrade, template provisioning, traffic monitoring')}
      </Typography>
    </Box>
  );
};

export const CpeList = () => (
  <List 
    actions={<TopToolbar><RefreshButton /><ExportButton /></TopToolbar>} 
    sort={{ field: 'last_inform', order: 'DESC' }}
    empty={<CpeEmptyState />}
  >
    <CpeListContent />
  </List>
);
