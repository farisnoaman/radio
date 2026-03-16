import {
    List,
    Datagrid,
    TextField,
    DateField,
    NumberField,
    EditButton,
    useRecordContext,
    useTranslate,
    ChipField,
} from 'react-admin';

// Reusing RouterStatusField from previously defined but removing unused vars
const RouterStatusField = () => {
    const record = useRecordContext();
    const translate = useTranslate();

    if (!record) return null;

    const statusColors: Record<string, 'success' | 'error' | 'default'> = {
        online: 'success',
        offline: 'error',
        disabled: 'default',
    };

    const statusLabel = translate(`resources.network/servers.status.${record.router_status}`, {
        _: record.router_status,
    });

    return (
        <ChipField
            source="router_status"
            size="small"
            color={statusColors[record.router_status] || 'default'}
            record={{ router_status: statusLabel }}
        />
    );
};

export const ServerList = () => {
    const translate = useTranslate();

    return (
        <List
            title={translate('resources.network/servers.list_title', { _: 'Servers List' })}
            sort={{ field: 'id', order: 'DESC' }}
        >
            <Datagrid rowClick="edit" bulkActionButtons={false}>
                <TextField source="id" label={translate('resources.network/servers.fields.id', { _: 'SL' })} />
                <TextField source="db_name" label={translate('resources.network/servers.fields.db_name', { _: 'Database' })} />
                <TextField source="router_limit" label={translate('resources.network/servers.fields.router_limit', { _: 'Router Limit' })} />
                <RouterStatusField />
                <TextField source="name" label={translate('resources.network/servers.fields.name', { _: 'Server Name' })} />
                <TextField source="secret" label={translate('resources.network/servers.fields.secret', { _: 'Secret' })} />
                <TextField source="public_ip" label={translate('resources.network/servers.fields.public_ip', { _: 'Public IP' })} />
                <NumberField source="online_hotspot" label={translate('resources.network/servers.fields.online_hotspot', { _: 'Online Hotspot' })} />
                <NumberField source="online_pppoe" label={translate('resources.network/servers.fields.online_pppoe', { _: 'Online PPPoE' })} />
                <TextField source="ports" label={translate('resources.network/servers.fields.ports', { _: 'Ports' })} />
                <TextField source="username" label={translate('resources.network/servers.fields.username', { _: 'Username' })} />
                <DateField source="created_at" label={translate('resources.network/servers.fields.created_at', { _: 'Created Date' })} showTime />
                <EditButton />
            </Datagrid>
        </List>
    );
};
