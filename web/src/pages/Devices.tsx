import React from 'react';
import {
  List,
  Datagrid,
  TextField,
  FunctionField,
  ShowButton,
  SimpleShowLayout,
  Show,
  Filter,
  SelectInput,
  useTranslate,
  useListContext,
  EditButton,
  DeleteButton,
  Button,
  useNotify,
  useRefresh,
  TopToolbar,
} from 'react-admin';
import { Box, Card, CardContent, Typography, useMediaQuery, useTheme, Chip, Avatar, Stack } from '@mui/material';
import RestartAltIcon from '@mui/icons-material/RestartAlt';
import RouterIcon from '@mui/icons-material/Router';
import WifiIcon from '@mui/icons-material/Wifi';
import LinkIcon from '@mui/icons-material/Link';
import SettingsEthernetIcon from '@mui/icons-material/SettingsEthernet';
import SecurityIcon from '@mui/icons-material/Security';
import DnsIcon from '@mui/icons-material/Dns';
import { useCallback } from 'react';

const DeviceTypeIcon = ({ type }: { type?: string }) => {
  const iconSize = 20;
  switch (type) {
    case 'router':
      return <RouterIcon sx={{ fontSize: iconSize }} />;
    case 'ap':
      return <WifiIcon sx={{ fontSize: iconSize }} />;
    case 'bridge':
      return <LinkIcon sx={{ fontSize: iconSize }} />;
    case 'switch':
      return <SettingsEthernetIcon sx={{ fontSize: iconSize }} />;
    case 'firewall':
      return <SecurityIcon sx={{ fontSize: iconSize }} />;
    default:
      return <DnsIcon sx={{ fontSize: iconSize }} />;
  }
};

const statusColor = (status?: string) => {
  switch ((status || '').toLowerCase()) {
    case 'online':
      return 'success';
    case 'offline':
      return 'error';
    default:
      return 'default';
  }
};

const deviceTypeChoices = [
  { id: 'router', name: 'resources.network/devices.types.router' },
  { id: 'ap', name: 'resources.network/devices.types.ap' },
  { id: 'bridge', name: 'resources.network/devices.types.bridge' },
  { id: 'switch', name: 'resources.network/devices.types.switch' },
  { id: 'firewall', name: 'resources.network/devices.types.firewall' },
  { id: 'other', name: 'resources.network/devices.types.other' },
];

const statusChoices = [
  { id: 'online', name: 'resources.network/devices.status.online' },
  { id: 'offline', name: 'resources.network/devices.status.offline' },
  { id: 'unknown', name: 'resources.network/devices.status.unknown' },
];

const DeviceFilter = (props: any) => {
  const t = useTranslate();
  return (
    <Filter {...props}>
      <SelectInput
        label={t('resources.network/devices.fields.device_type')}
        source="device_type"
        choices={deviceTypeChoices}
        alwaysOn
      />
      <SelectInput
        label={t('resources.network/devices.fields.status')}
        source="status"
        choices={statusChoices}
        alwaysOn
      />
    </Filter>
  );
};

const DeviceCardMobile = ({ device }: { device: any }) => {
  const t = useTranslate();
  const icon = <DeviceTypeIcon type={device.device_type} />;
  
  return (
    <Card sx={{ borderRadius: 2, mb: 2 }}>
      <CardContent>
        <Box display="flex" alignItems="center" justifyContent="space-between">
          <Box display="flex" alignItems="center" gap={2}>
            <Avatar sx={{ bgcolor: 'transparent' }}>{icon}</Avatar>
            <Box>
              <Typography variant="subtitle1" fontWeight={700}>{device.name}</Typography>
              <Typography variant="body2" color="text.secondary" dir="ltr">{device.ip_address}</Typography>
            </Box>
          </Box>
          <Chip 
            label={t(`resources.network/devices.status.${device.status}`)} 
            color={statusColor(device.status)} 
            size="small" 
          />
        </Box>
        <Stack direction="row" spacing={1} mt={1}>
          <Typography variant="caption" color="text.secondary">
            {t('resources.network/devices.fields.vendor')}: {device.vendor || '-'}
          </Typography>
          <Typography variant="caption" color="text.secondary">
            {t('resources.network/devices.fields.model')}: {device.model || '-'}
          </Typography>
        </Stack>
      </CardContent>
    </Card>
  );
};

const DeviceListContent = ({ isMobile }: { isMobile: boolean }) => {
  const t = useTranslate();
  const listContext = useListContext();
  const { ids, data, isLoading } = listContext as any;

  if (isLoading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body2" color="text.secondary">{t('ra.navigation.loading')}</Typography>
      </Box>
    );
  }

  if (isMobile) {
    return (
      <Box sx={{ p: 2 }}>
        {ids.map((id: string | number) => (
          <DeviceCardMobile key={id} device={data[id]} />
        ))}
        {ids.length === 0 && (
          <Box sx={{ p: 3, textAlign: 'center' }}>
            <Typography variant="body1" color="text.secondary">
              {t('resources.network/devices.empty.title')}
            </Typography>
            <Typography variant="body2" color="text.secondary" mt={1}>
              {t('resources.network/devices.empty.description')}
            </Typography>
          </Box>
        )}
      </Box>
    );
  }

  return (
    <Datagrid rowClick="show" bulkActionButtons={false}>
      <TextField source="name" label={t('resources.network/devices.fields.name')} />
      <TextField source="ip_address" label={t('resources.network/devices.fields.ip_address')} dir="ltr" />
      <FunctionField
        label={t('resources.network/devices.fields.device_type')}
        render={(record: any) => (
          <Box display="flex" alignItems="center" gap={1}>
            <DeviceTypeIcon type={record?.device_type} />
            <Typography variant="body2">
              {t(`resources.network/devices.types.${record?.device_type}`) || record?.device_type}
            </Typography>
          </Box>
        )}
      />
      <TextField source="vendor" label={t('resources.network/devices.fields.vendor')} />
      <TextField source="model" label={t('resources.network/devices.fields.model')} />
      <FunctionField
        label={t('resources.network/devices.fields.status')}
        render={(record: any) => (
          <Chip 
            label={t(`resources.network/devices.status.${record?.status}`)} 
            color={statusColor(record?.status)} 
            size="small" 
          />
        )}
      />
      <ShowButton />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  );
};

const DeviceListActions = () => {
  const t = useTranslate();
  const refresh = useRefresh();
  return (
    <TopToolbar>
      <Button
        label={t('ra.action.refresh')}
        onClick={() => refresh()}
      >
        {t('ra.action.refresh')}
      </Button>
    </TopToolbar>
  );
};

export const DeviceList = (props: any) => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  
  return (
    <List 
      {...props} 
      filters={<DeviceFilter />} 
      sort={{ field: 'name', order: 'ASC' }}
      actions={<DeviceListActions />}
    >
      <DeviceListContent isMobile={isMobile} />
    </List>
  );
};

export const DeviceShow = (props: any) => {
  const t = useTranslate();
  const notify = useNotify();
  const [rebooting, setRebooting] = React.useState(false);

  const handleReboot = useCallback(async () => {
    const id = props.id;
    if (!id) return;
    
    setRebooting(true);
    try {
      const response = await fetch(`/api/v1/devices/${id}/reboot`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
      if (response.ok) {
        notify(t('resources.network/devices.reboot_success'));
      } else {
        const data = await response.json();
        notify(data.message || t('resources.network/devices.reboot_failed'), { type: 'error' });
      }
    } catch (e) {
      notify(t('resources.network/devices.reboot_failed'), { type: 'error' });
    } finally {
      setRebooting(false);
    }
  }, [props.id, notify, t]);

  return (
    <Show {...props} title={t('resources.network/devices.show_title')}>
      <SimpleShowLayout>
        <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider' }}>
          <Typography variant="h5" component="span">
            {t('resources.network/devices.show_title')}
          </Typography>
        </Box>
        
        <Box sx={{ p: 3 }}>
          <Stack spacing={3}>
            <Box>
              <Typography variant="h6" gutterBottom>{t('resources.network/devices.sections.basic.title')}</Typography>
              <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2 }}>
                <Box>
                  <Typography variant="caption" color="text.secondary">{t('resources.network/devices.fields.name')}</Typography>
                  <Typography variant="body1">{props.record?.name}</Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary" dir="ltr">{t('resources.network/devices.fields.ip_address')}</Typography>
                  <Typography variant="body1" dir="ltr">{props.record?.ip_address}</Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">{t('resources.network/devices.fields.device_type')}</Typography>
                  <Box display="flex" alignItems="center" gap={1}>
                    <DeviceTypeIcon type={props.record?.device_type} />
                    <Typography variant="body1">
                      {t(`resources.network/devices.types.${props.record?.device_type}`) || props.record?.device_type}
                    </Typography>
                  </Box>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">{t('resources.network/devices.fields.status')}</Typography>
                  <Box mt={0.5}>
                    <Chip 
                      label={t(`resources.network/devices.status.${props.record?.status}`)} 
                      color={statusColor(props.record?.status)} 
                      size="small" 
                    />
                  </Box>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">{t('resources.network/devices.fields.vendor')}</Typography>
                  <Typography variant="body1">{props.record?.vendor || '-'}</Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">{t('resources.network/devices.fields.model')}</Typography>
                  <Typography variant="body1">{props.record?.model || '-'}</Typography>
                </Box>
              </Box>
            </Box>

            <Box>
              <Typography variant="h6" gutterBottom>{t('resources.network/devices.sections.hardware.title')}</Typography>
              <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2 }}>
                <Box>
                  <Typography variant="caption" color="text.secondary">{t('resources.network/devices.fields.serial_number')}</Typography>
                  <Typography variant="body1" dir="ltr">{props.record?.serial_number || '-'}</Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">{t('resources.network/devices.fields.firmware_version')}</Typography>
                  <Typography variant="body1" dir="ltr">{props.record?.firmware_version || '-'}</Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">{t('resources.network/devices.fields.mac_address')}</Typography>
                  <Typography variant="body1" dir="ltr">{props.record?.mac_address || '-'}</Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">{t('resources.network/devices.fields.last_seen')}</Typography>
                  <Typography variant="body1">
                    {props.record?.last_seen ? new Date(props.record.last_seen).toLocaleString() : '-'}
                  </Typography>
                </Box>
              </Box>
            </Box>

            <Box>
              <Typography variant="h6" gutterBottom>{t('resources.network/devices.sections.actions.title')}</Typography>
              <Stack direction="row" spacing={2}>
                <Button
                  variant="contained"
                  color="warning"
                  startIcon={<RestartAltIcon />}
                  onClick={handleReboot}
                  disabled={rebooting}
                  label={rebooting ? t('resources.network/devices.rebooting') : t('resources.network/devices.reboot')}
                />
              </Stack>
            </Box>
          </Stack>
        </Box>
      </SimpleShowLayout>
    </Show>
  );
};

export default DeviceList;
