import {
    Create,
    SimpleForm,
    TextInput,
    PasswordInput,
    NumberInput,
    required,
} from 'react-admin';
import { Typography, Divider, Box } from '@mui/material';

export const ServerCreate = () => {
    return (
        <Create title="Add Server">
            <SimpleForm sx={{ maxWidth: 1000 }}>
                <Box display="flex" flexDirection={{ xs: 'column', md: 'row' }} gap={4} width="100%">
                    <Box flex={1}>
                        <Box mb={2}>
                            <Typography variant="h6" color="primary">Router Info</Typography>
                            <Divider />
                        </Box>
                        <TextInput source="router_limit" label="Router Limit" fullWidth />
                        <TextInput source="name" label="Name" validate={required()} fullWidth />
                        <TextInput source="public_ip" label="Public IP" fullWidth />
                        <PasswordInput source="secret" label="Secret" fullWidth />
                        <TextInput source="username" label="Username" fullWidth />
                        <PasswordInput source="password" label="Password" fullWidth />
                        <TextInput source="ports" label="Ports" fullWidth helperText="Mikrotik API Port (e.g., 8728)" />
                    </Box>
                    <Box flex={1}>
                        <Box mb={2}>
                            <Typography variant="h6" color="primary">Database Information</Typography>
                            <Divider />
                        </Box>
                        <TextInput source="db_host" label="DB Host" fullWidth />
                        <NumberInput source="db_port" label="DB Port" fullWidth defaultValue={3306} />
                        <TextInput source="db_name" label="DB Name" fullWidth />
                        <TextInput source="db_username" label="DB Username" fullWidth />
                        <PasswordInput source="db_password" label="DB Password" fullWidth />
                    </Box>
                </Box>
            </SimpleForm>
        </Create>
    );
};
