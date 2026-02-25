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
    const { data, isLoading, total } = useListContext();
  
    if (isLoading) return null; // Simplified for brevity in this initial implementation
  
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

export const CpeList = () => (
    <List actions={<TopToolbar><RefreshButton /><ExportButton /></TopToolbar>} sort={{ field: 'last_inform', order: 'DESC' }}>
        <CpeListContent />
    </List>
);
