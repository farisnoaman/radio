import {
    Edit,
    SimpleForm,
    TextInput,
    PasswordInput,
    NumberInput,
    required,
    useTranslate,
    useLocale,
} from 'react-admin';
import { Typography, Divider, Box } from '@mui/material';

export const ServerEdit = () => {
    const translate = useTranslate();
    const locale = useLocale();
    const isRtl = locale === 'ar';

    const textInputProps = { style: { textAlign: isRtl ? 'right' : 'left', direction: isRtl ? 'rtl' : 'ltr' } } as const;

    return (
        <Edit title={translate('resources.network/servers.edit_title', { _: 'Edit Server' })}>
            <SimpleForm sx={{ maxWidth: 1000, direction: isRtl ? 'rtl' : 'ltr' }}>
                <Box display="flex" flexDirection={{ xs: 'column', md: 'row' }} gap={4} width="100%">
                    <Box flex={1}>
                        <Box mb={2}>
                            <Typography variant="h6" color="primary">
                                {translate('resources.network/servers.sections.router_info', { _: 'Router Info' })}
                            </Typography>
                            <Divider />
                        </Box>
                        <TextInput
                            source="router_limit"
                            label={translate('resources.network/servers.fields.router_limit', { _: 'Router Limit' })}
                            fullWidth
                            inputProps={textInputProps}
                        />
                        <TextInput
                            source="name"
                            label={translate('resources.network/servers.fields.name', { _: 'Name' })}
                            validate={required()}
                            fullWidth
                            inputProps={textInputProps}
                        />
                        <TextInput
                            source="public_ip"
                            label={translate('resources.network/servers.fields.public_ip_address', { _: 'Public IP' })}
                            fullWidth
                            inputProps={textInputProps}
                        />
                        <PasswordInput
                            source="secret"
                            label={translate('resources.network/servers.fields.secret', { _: 'Secret' })}
                            fullWidth
                            inputProps={textInputProps}
                        />
                        <TextInput
                            source="username"
                            label={translate('resources.network/servers.fields.username', { _: 'Username' })}
                            fullWidth
                            inputProps={textInputProps}
                        />
                        <PasswordInput
                            source="password"
                            label={translate('resources.network/servers.fields.password', { _: 'Password' })}
                            fullWidth
                            inputProps={textInputProps}
                        />
                        <TextInput
                            source="ports"
                            label={translate('resources.network/servers.fields.ports', { _: 'Ports' })}
                            fullWidth
                            helperText={translate('resources.network/servers.helper.ports', { _: 'Mikrotik API Port (e.g., 8728)' })}
                            inputProps={textInputProps}
                        />
                    </Box>
                    <Box flex={1}>
                        <Box mb={2}>
                            <Typography variant="h6" color="primary">
                                {translate('resources.network/servers.sections.database_info', { _: 'Database Information' })}
                            </Typography>
                            <Divider />
                        </Box>
                        <TextInput
                            source="db_host"
                            label={translate('resources.network/servers.fields.db_host', { _: 'DB Host' })}
                            fullWidth
                            inputProps={textInputProps}
                        />
                        <NumberInput
                            source="db_port"
                            label={translate('resources.network/servers.fields.db_port', { _: 'DB Port' })}
                            fullWidth
                            inputProps={textInputProps}
                        />
                        <TextInput
                            source="db_name"
                            label={translate('resources.network/servers.fields.db_name', { _: 'DB Name' })}
                            fullWidth
                            inputProps={textInputProps}
                        />
                        <TextInput
                            source="db_username"
                            label={translate('resources.network/servers.fields.db_username', { _: 'DB Username' })}
                            fullWidth
                            inputProps={textInputProps}
                        />
                        <PasswordInput
                            source="db_password"
                            label={translate('resources.network/servers.fields.db_password', { _: 'DB Password' })}
                            fullWidth
                            inputProps={textInputProps}
                        />
                    </Box>
                </Box>
            </SimpleForm>
        </Edit>
    );
};
