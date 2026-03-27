import {
  List,
  Datagrid,
  TextField,
  EditButton,
  DeleteButton,
  Create,
  Edit,
  SimpleForm,
  TextInput,
  SelectInput,
  ArrayInput,
  SimpleFormIterator,
  BooleanInput,
  useTranslate,
} from 'react-admin';

const vendorChoices = [
  { id: 'mikrotik', name: 'mikrotik' },
  { id: 'cisco', name: 'cisco' },
  { id: 'huawei', name: 'huawei' },
  { id: 'juniper', name: 'juniper' },
  { id: 'ubiquiti', name: 'ubiquiti' },
  { id: 'tplink', name: 'tplink' },
  { id: 'other', name: 'other' },
];

const valueTypeChoices = [
  { id: 'string', name: 'resources.network/nas-templates.value_types.string' },
  { id: 'integer', name: 'resources.network/nas-templates.value_types.integer' },
  { id: 'ipaddr', name: 'resources.network/nas-templates.value_types.ipaddr' },
];

const TemplateAttributeInput = () => {
  const t = useTranslate();
  return (
    <ArrayInput source="attributes" label={t('resources.network/nas-templates.fields.attributes')}>
      <SimpleFormIterator>
        <TextInput
          source="attr_name"
          label={t('resources.network/nas-templates.fields.attr_name')}
          fullWidth
        />
        <TextInput
          source="vendor_attr"
          label={t('resources.network/nas-templates.fields.vendor_attr')}
          fullWidth
        />
        <SelectInput
          source="value_type"
          label={t('resources.network/nas-templates.fields.value_type')}
          choices={valueTypeChoices}
          fullWidth
        />
        <BooleanInput
          source="is_required"
          label={t('resources.network/nas-templates.fields.is_required')}
        />
        <TextInput
          source="default_value"
          label={t('resources.network/nas-templates.fields.default_value')}
          fullWidth
        />
      </SimpleFormIterator>
    </ArrayInput>
  );
};

export const NasTemplateList = () => {
  const t = useTranslate();
  return (
    <List title={t('resources.network/nas-templates.name')}>
      <Datagrid rowClick="edit" bulkActionButtons={false}>
        <TextField source="id" label={t('resources.network/nas-templates.fields.id')} />
        <TextField source="vendor_code" label={t('resources.network/nas-templates.fields.vendor_code')} />
        <TextField source="name" label={t('resources.network/nas-templates.fields.name')} />
        <TextField
          source="is_default"
          label={t('resources.network/nas-templates.fields.is_default')}
        />
        <EditButton />
        <DeleteButton />
      </Datagrid>
    </List>
  );
};

export const NasTemplateCreate = () => {
  const t = useTranslate();
  return (
    <Create title={t('resources.network/nas-templates.create_title')}>
      <SimpleForm>
        <SelectInput
          source="vendor_code"
          label={t('resources.network/nas-templates.fields.vendor_code')}
          choices={vendorChoices}
          fullWidth
        />
        <TextInput
          source="name"
          label={t('resources.network/nas-templates.fields.name')}
          fullWidth
        />
        <BooleanInput
          source="is_default"
          label={t('resources.network/nas-templates.fields.is_default')}
        />
        <TemplateAttributeInput />
        <TextInput
          source="remark"
          label={t('resources.network/nas-templates.fields.remark')}
          multiline
          fullWidth
        />
      </SimpleForm>
    </Create>
  );
};

export const NasTemplateEdit = () => {
  const t = useTranslate();
  return (
    <Edit title={t('resources.network/nas-templates.edit_title')}>
      <SimpleForm>
        <TextInput
          source="id"
          label={t('resources.network/nas-templates.fields.id')}
          disabled
        />
        <SelectInput
          source="vendor_code"
          label={t('resources.network/nas-templates.fields.vendor_code')}
          choices={vendorChoices}
          fullWidth
        />
        <TextInput
          source="name"
          label={t('resources.network/nas-templates.fields.name')}
          fullWidth
        />
        <BooleanInput
          source="is_default"
          label={t('resources.network/nas-templates.fields.is_default')}
        />
        <TemplateAttributeInput />
        <TextInput
          source="remark"
          label={t('resources.network/nas-templates.fields.remark')}
          multiline
          fullWidth
        />
      </SimpleForm>
    </Edit>
  );
};
