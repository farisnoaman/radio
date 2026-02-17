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
import { useFormContext } from 'react-hook-form';
import React from 'react';
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

const ValidityInput = () => {
    const { setValue, getValues } = useFormContext();
    // Default to 'days' unless value is small (e.g. < 1 hour)
    const currentSeconds = getValues('validity_seconds') || 0;
    const initialUnit = currentSeconds > 0 && currentSeconds % 86400 === 0 ? 'days' :
        currentSeconds > 0 && currentSeconds % 3600 === 0 ? 'hours' : 'minutes';

    // Calculate initial value based on unit
    const initialValue = initialUnit === 'days' ? currentSeconds / 86400 :
        initialUnit === 'hours' ? currentSeconds / 3600 :
            currentSeconds / 60;

    const [unit, setUnit] = React.useState(initialUnit);
    const [val, setVal] = React.useState(initialValue > 0 ? initialValue : 30); // Default 30

    // Effect to update the actual source field when unit or val changes
    React.useEffect(() => {
        let multiplier = 60;
        if (unit === 'hours') multiplier = 3600;
        if (unit === 'days') multiplier = 86400;

        setValue('validity_seconds', val * multiplier);
    }, [unit, val, setValue]);

    return (
        <Box display="flex" width="100%" gap={2}>
            <Box flex={1}>
                <NumberInput
                    source="validity_value_virtual" // Virtual field
                    label="Validity Duration"
                    value={val}
                    onChange={(e) => setVal(Number(e.target.value))}
                    defaultValue={30}
                    fullWidth
                />
            </Box>
            <Box width="150px">
                <SelectInput
                    source="validity_unit_virtual" // Virtual field
                    label="Unit"
                    choices={[
                        { id: 'minutes', name: 'Minutes' },
                        { id: 'hours', name: 'Hours' },
                        { id: 'days', name: 'Days' },
                    ]}
                    value={unit}
                    onChange={(e) => setUnit(e.target.value)}
                    defaultValue="days"
                    fullWidth
                    disableValue="validity_seconds" // Don't submit this field directly if not needed, but react-admin might send it. It's fine.
                />
            </Box>
            <NumberInput source="validity_seconds" style={{ display: 'none' }} />
        </Box>
    );
};

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
            <Box display={{ xs: 'block', sm: 'flex', width: '100%' }}>
                <Box flex={1} mr={{ xs: 0, sm: '0.5em' }}>
                    <TextInput source="color" type="color" fullWidth label="Product Color" defaultValue="#1976d2" />
                </Box>
                <Box flex={1} ml={{ xs: 0, sm: '0.5em' }}>
                    <ValidityInput />
                </Box>
            </Box>
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
            <Box display={{ xs: 'block', sm: 'flex', width: '100%' }}>
                <Box flex={1} mr={{ xs: 0, sm: '0.5em' }}>
                    <TextInput source="color" type="color" fullWidth label="Product Color" />
                </Box>
                <Box flex={1} ml={{ xs: 0, sm: '0.5em' }}>
                    <ValidityInput />
                </Box>
            </Box>
            <SelectInput source="status" choices={[
                { id: 'enabled', name: 'Enabled' },
                { id: 'disabled', name: 'Disabled' },
            ]} fullWidth />
            <TextInput source="remark" multiline fullWidth />
        </SimpleForm>
    </Edit>
);
