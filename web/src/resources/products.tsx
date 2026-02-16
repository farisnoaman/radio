import {
    List,
    Datagrid,
    TextField,
    NumberField,
    DateField,
    EditButton,
    DeleteButton,
    Create,
    SimpleForm,
    TextInput,
    NumberInput,
    SelectInput,
    ReferenceInput,
    ReferenceField,
    Edit,
    Show,
    SimpleShowLayout,
    required,
    useRecordContext,
    ListProps,
    ShowProps,
    CreateProps,
    EditProps,
} from 'react-admin';
import { Box } from '@mui/material';

const ProductTitle = () => {
    const record = useRecordContext();
    return <span>Product {record ? `"${record.name}"` : ''}</span>;
};

export const ProductList = (props: ListProps) => (
    <List {...props} sort={{ field: 'id', order: 'DESC' }}>
        <Datagrid rowClick="show">
            <TextField source="id" />
            <TextField source="name" />
            <ReferenceField source="radius_profile_id" reference="radius-profiles">
                <TextField source="name" />
            </ReferenceField>
            <NumberField source="price" options={{ style: 'currency', currency: 'USD' }} />
            <NumberField source="cost_price" options={{ style: 'currency', currency: 'USD' }} />
            <NumberField source="validity_seconds" />
            <TextField source="status" />
            <DateField source="updated_at" showTime />
            <EditButton />
            <DeleteButton />
        </Datagrid>
    </List>
);

export const ProductShow = (props: ShowProps) => (
    <Show {...props} title={<ProductTitle />}>
        <SimpleShowLayout>
            <TextField source="id" />
            <TextField source="name" />
            <ReferenceField source="radius_profile_id" reference="radius-profiles">
                <TextField source="name" />
            </ReferenceField>
            <NumberField source="price" options={{ style: 'currency', currency: 'USD' }} />
            <NumberField source="cost_price" options={{ style: 'currency', currency: 'USD' }} />
            <NumberField source="validity_seconds" />
            <TextField source="status" />
            <TextField source="remark" />
            <DateField source="created_at" showTime />
            <DateField source="updated_at" showTime />
        </SimpleShowLayout>
    </Show>
);

export const ProductCreate = (props: CreateProps) => (
    <Create {...props}>
        <SimpleForm>
            <TextInput source="name" validate={[required()]} fullWidth />
            <ReferenceInput source="radius_profile_id" reference="radius-profiles">
                <SelectInput optionText="name" validate={[required()]} />
            </ReferenceInput>
            <Box display={{ xs: 'block', sm: 'flex', width: '100%' }}>
                <Box flex={1} mr={{ xs: 0, sm: '0.5em' }}>
                    <NumberInput source="price" validate={[required()]} fullWidth />
                </Box>
                <Box flex={1} ml={{ xs: 0, sm: '0.5em' }}>
                    <NumberInput source="cost_price" validate={[required()]} fullWidth />
                </Box>
            </Box>
            <NumberInput source="validity_seconds" defaultValue={2592000} fullWidth helperText="30 days = 2592000 seconds" />
            <SelectInput source="status" choices={[
                { id: 'enabled', name: 'Enabled' },
                { id: 'disabled', name: 'Disabled' },
            ]} defaultValue="enabled" fullWidth />
            <TextInput source="remark" multiline fullWidth />
        </SimpleForm>
    </Create>
);

export const ProductEdit = (props: EditProps) => (
    <Edit {...props} title={<ProductTitle />}>
        <SimpleForm>
            <TextInput source="id" disabled />
            <TextInput source="name" validate={[required()]} fullWidth />
            <ReferenceInput source="radius_profile_id" reference="radius-profiles">
                <SelectInput optionText="name" validate={[required()]} />
            </ReferenceInput>
            <Box display={{ xs: 'block', sm: 'flex', width: '100%' }}>
                <Box flex={1} mr={{ xs: 0, sm: '0.5em' }}>
                    <NumberInput source="price" validate={[required()]} fullWidth />
                </Box>
                <Box flex={1} ml={{ xs: 0, sm: '0.5em' }}>
                    <NumberInput source="cost_price" validate={[required()]} fullWidth />
                </Box>
            </Box>
            <NumberInput source="validity_seconds" fullWidth />
            <SelectInput source="status" choices={[
                { id: 'enabled', name: 'Enabled' },
                { id: 'disabled', name: 'Disabled' },
            ]} fullWidth />
            <TextInput source="remark" multiline fullWidth />
        </SimpleForm>
    </Edit>
);
