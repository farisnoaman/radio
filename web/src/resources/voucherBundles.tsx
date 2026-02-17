import {
    List,
    Datagrid,
    TextField,
    Create,
    SimpleForm,
    TextInput,
    NumberInput,
    ReferenceInput,
    SelectInput,
    required,
    ListProps,
    CreateProps,
    ArrayInput,
    SimpleFormIterator,
} from 'react-admin';

export const VoucherBundleList = (props: ListProps) => (
    <List {...props}>
        <Datagrid rowClick="edit">
            <TextField source="id" />
            <TextField source="name" />
            <TextField source="price" label="Price" />
            <TextField source="remark" />
        </Datagrid>
    </List>
);

export const VoucherBundleCreate = (props: CreateProps) => (
    <Create {...props}>
        <SimpleForm>
            <TextInput source="name" validate={[required()]} fullWidth />
            <NumberInput source="price" validate={[required()]} fullWidth />
            <TextInput source="remark" multiline fullWidth />

            <ArrayInput source="items" label="Bundle Items">
                <SimpleFormIterator inline>
                    <ReferenceInput source="product_id" reference="products" label="Product">
                        <SelectInput optionText="name" validate={[required()]} />
                    </ReferenceInput>
                    <NumberInput source="count" label="Count" defaultValue={1} min={1} validate={[required()]} />
                </SimpleFormIterator>
            </ArrayInput>
        </SimpleForm>
    </Create>
);
