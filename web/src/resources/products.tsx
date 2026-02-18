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
    useTranslate,
    ListProps,
    ShowProps,
    CreateProps,
    EditProps,
} from 'react-admin';
import { useFormContext } from 'react-hook-form';
import React from 'react';
import { Box } from '@mui/material';
import {
    FormSection,
    FieldGrid,
    FieldGridItem,
    formLayoutSx,
} from '../components';

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
            <NumberField source="up_rate" label="Up Rate (Kbps)" />
            <NumberField source="down_rate" label="Down Rate (Kbps)" />
            <NumberField source="data_quota" label="Quota (MB)" />
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
            <Box display="flex" gap={2}>
                <NumberField source="price" options={{ style: 'currency', currency: 'USD' }} />
                <NumberField source="cost_price" options={{ style: 'currency', currency: 'USD' }} />
            </Box>
            <Box display="flex" gap={2}>
                <NumberField source="up_rate" label="Upload Rate (Kbps)" />
                <NumberField source="down_rate" label="Download Rate (Kbps)" />
            </Box>
            <Box display="flex" gap={2}>
                <NumberField source="data_quota" label="Data Quota (MB)" />
                <NumberField source="validity_seconds" label="Validity (Seconds)" />
            </Box>
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
        <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
                <NumberInput
                    source="validity_value_virtual" // Virtual field
                    label="Validity Duration"
                    value={val}
                    onChange={(e) => setVal(Number(e.target.value))}
                    defaultValue={30}
                    fullWidth
                />
            </FieldGridItem>
            <FieldGridItem>
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
                    disableValue="validity_seconds"
                />
            </FieldGridItem>
            <NumberInput source="validity_seconds" style={{ display: 'none' }} />
        </FieldGrid>
    );
};

const DataQuotaInput = () => {
    const { setValue, getValues } = useFormContext();
    const currentMB = getValues('data_quota') || 0;
    const initialUnit = currentMB > 0 && currentMB % 1024 === 0 ? 'GB' : 'MB';
    const initialValue = initialUnit === 'GB' ? currentMB / 1024 : currentMB;

    const [unit, setUnit] = React.useState(initialUnit);
    const [val, setVal] = React.useState(initialValue);

    React.useEffect(() => {
        const multiplier = unit === 'GB' ? 1024 : 1;
        setValue('data_quota', val * multiplier);
    }, [unit, val, setValue]);

    return (
        <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
                <NumberInput
                    source="data_quota_virtual"
                    label="Data Quota"
                    value={val}
                    onChange={(e) => setVal(Number(e.target.value))}
                    fullWidth
                />
            </FieldGridItem>
            <FieldGridItem>
                <SelectInput
                    source="data_quota_unit_virtual"
                    label="Unit"
                    choices={[
                        { id: 'MB', name: 'MB' },
                        { id: 'GB', name: 'GB' },
                    ]}
                    value={unit}
                    onChange={(e) => setUnit(e.target.value)}
                    fullWidth
                />
            </FieldGridItem>
            <NumberInput source="data_quota" style={{ display: 'none' }} />
        </FieldGrid>
    );
};

export const ProductCreate = (props: CreateProps) => {
    const translate = useTranslate();
    return (
        <Create {...props}>
            <SimpleForm sx={formLayoutSx}>
                <FormSection
                    title={translate('resources.products.section.basic', { _: 'Basic Information' })}
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <TextInput source="name" validate={[required()]} fullWidth />
                        </FieldGridItem>
                        <FieldGridItem>
                            <ReferenceInput source="radius_profile_id" reference="radius-profiles">
                                <SelectInput optionText="name" validate={[required()]} fullWidth />
                            </ReferenceInput>
                        </FieldGridItem>
                        <FieldGridItem>
                            <TextInput source="color" type="color" fullWidth label="Product Color" defaultValue="#1976d2" />
                        </FieldGridItem>
                        <FieldGridItem>
                            <SelectInput source="status" choices={[
                                { id: 'enabled', name: 'Enabled' },
                                { id: 'disabled', name: 'Disabled' },
                            ]} defaultValue="enabled" fullWidth />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title={translate('resources.products.section.pricing', { _: 'Pricing' })}
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <NumberInput source="price" validate={[required()]} fullWidth />
                        </FieldGridItem>
                        <FieldGridItem>
                            <NumberInput source="cost_price" validate={[required()]} fullWidth />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title={translate('resources.products.section.bandwidth', { _: 'Bandwidth Limit' })}
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <NumberInput source="up_rate" label="Upload Rate (Kbps)" defaultValue={0} fullWidth helperText="0 = Unlimited" />
                        </FieldGridItem>
                        <FieldGridItem>
                            <NumberInput source="down_rate" label="Download Rate (Kbps)" defaultValue={0} fullWidth helperText="0 = Unlimited" />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title={translate('resources.products.section.data_quota', { _: 'Data Quota' })}
                >
                    <DataQuotaInput />
                </FormSection>

                <FormSection
                    title={translate('resources.products.section.validity', { _: 'Validity Limit' })}
                >
                    <ValidityInput />
                </FormSection>

                <FormSection
                    title={translate('resources.products.section.remark', { _: 'Remark' })}
                >
                    <FieldGrid columns={{ xs: 1 }}>
                        <FieldGridItem>
                            <TextInput source="remark" multiline fullWidth minRows={3} />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>
            </SimpleForm>
        </Create>
    );
};

export const ProductEdit = (props: EditProps) => {
    const translate = useTranslate();
    return (
        <Edit {...props} title={<ProductTitle />}>
            <SimpleForm sx={formLayoutSx}>
                <FormSection
                    title={translate('resources.products.section.basic', { _: 'Basic Information' })}
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <TextInput source="id" disabled fullWidth />
                        </FieldGridItem>
                        <FieldGridItem>
                            <TextInput source="name" validate={[required()]} fullWidth />
                        </FieldGridItem>
                        <FieldGridItem>
                            <ReferenceInput source="radius_profile_id" reference="radius-profiles">
                                <SelectInput optionText="name" validate={[required()]} fullWidth />
                            </ReferenceInput>
                        </FieldGridItem>
                        <FieldGridItem>
                            <TextInput source="color" type="color" fullWidth label="Product Color" />
                        </FieldGridItem>
                        <FieldGridItem>
                            <SelectInput source="status" choices={[
                                { id: 'enabled', name: 'Enabled' },
                                { id: 'disabled', name: 'Disabled' },
                            ]} fullWidth />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title={translate('resources.products.section.pricing', { _: 'Pricing' })}
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <NumberInput source="price" validate={[required()]} fullWidth />
                        </FieldGridItem>
                        <FieldGridItem>
                            <NumberInput source="cost_price" validate={[required()]} fullWidth />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title={translate('resources.products.section.bandwidth', { _: 'Bandwidth Limit' })}
                >
                    <FieldGrid columns={{ xs: 1, sm: 2 }}>
                        <FieldGridItem>
                            <NumberInput source="up_rate" label="Upload Rate (Kbps)" fullWidth helperText="0 = Unlimited" />
                        </FieldGridItem>
                        <FieldGridItem>
                            <NumberInput source="down_rate" label="Download Rate (Kbps)" fullWidth helperText="0 = Unlimited" />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>

                <FormSection
                    title={translate('resources.products.section.data_quota', { _: 'Data Quota' })}
                >
                    <DataQuotaInput />
                </FormSection>

                <FormSection
                    title={translate('resources.products.section.validity', { _: 'Validity Limit' })}
                >
                    <ValidityInput />
                </FormSection>

                <FormSection
                    title={translate('resources.products.section.remark', { _: 'Remark' })}
                >
                    <FieldGrid columns={{ xs: 1 }}>
                        <FieldGridItem>
                            <TextInput source="remark" multiline fullWidth minRows={3} />
                        </FieldGridItem>
                    </FieldGrid>
                </FormSection>
            </SimpleForm>
        </Edit>
    );
};
