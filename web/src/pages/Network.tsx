import React from 'react';
import { Box, Tabs, Tab } from '@mui/material';
import { useTranslate } from 'react-admin';
import { DeviceList } from './Devices';
import { LocationList } from './Locations';

export const NetworkPage = () => {
  const t = useTranslate();
  const [value, setValue] = React.useState(0);

  return (
    <Box sx={{ width: '100%' }}>
      <Tabs value={value} onChange={(_, newValue) => setValue(newValue)} textColor="primary">
        <Tab label={t('resources.network/devices.name')} />
        <Tab label={t('resources.network/locations.name')} />
      </Tabs>
      <Box sx={{ mt: 2, width: '100%' }}>
        {value === 0 ? <DeviceList resource="network/devices" /> : <LocationList resource="network/locations" />}
      </Box>
    </Box>
  );
};

export default NetworkPage;