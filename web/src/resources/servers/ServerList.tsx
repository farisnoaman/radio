import {
    List,
    Datagrid,
    TextField,
    DateField,
    NumberField,
    EditButton,
    useRecordContext,
    ChipField,
} from 'react-admin';

// Reusing RouterStatusField from previously defined but removing unused vars
const RouterStatusField = () => {
    const record = useRecordContext();
    if (!record) return null;

    const statusColors: Record<string, 'success' | 'error' | 'default'> = {
        online: 'success',
        offline: 'error',
        disabled: 'default',
    };

    return (
        <ChipField
            source="router_status"
            size="small"
            color={statusColors[record.router_status] || 'default'}
        />
    );
};

export const ServerList = () => {
    return (
        <List
            title="Servers List"
            sort={{ field: 'id', order: 'DESC' }}
        >
            <Datagrid rowClick="edit" bulkActionButtons={false}>
                <TextField source="id" label="SL" />
                <TextField source="db_name" label="Database" />
                <TextField source="router_limit" label="Router Limit" />
                <RouterStatusField />
                <TextField source="name" label="Server Name" />
                <TextField source="secret" label="Secret" />
                <TextField source="public_ip" label="Public IP" />
                <NumberField source="online_hotspot" label="Online Hotspot" />
                <NumberField source="online_pppoe" label="Online PPPoE" />
                <TextField source="ports" label="Ports" />
                <TextField source="username" label="Username" />
                <DateField source="created_at" label="Created Date" showTime />
                <EditButton />
            </Datagrid>
        </List>
    );
};
