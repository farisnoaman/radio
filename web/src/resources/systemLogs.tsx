import {
    Datagrid,
    DateField,
    List,
    TextField,
    TextInput
} from 'react-admin';

const logFilters = [
    <TextInput source="operator" label="Operator" alwaysOn />,
    <TextInput source="action" label="Action" alwaysOn />,
    <TextInput source="keyword" label="Keyword" alwaysOn />,
];

export const SystemLogList = () => (
    <List filters={logFilters} sort={{ field: 'id', order: 'DESC' }}>
        <Datagrid bulkActionButtons={false}>
            <TextField source="id" />
            <TextField source="opr_name" label="Operator" />
            <TextField source="opr_ip" label="IP Address" />
            <TextField source="opt_action" label="Action" />
            <TextField source="opt_desc" label="Description" />
            <DateField source="opt_time" label="Time" showTime />
        </Datagrid>
    </List>
);
