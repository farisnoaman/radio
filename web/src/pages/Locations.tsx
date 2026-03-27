import {
  List,
  Datagrid,
  TextField,
  EditButton,
  Create,
  SimpleForm,
  TextInput,
  Edit,
  useTranslate,
  DeleteButton,
} from 'react-admin';


const LocationListGrid = () => {
  const t = useTranslate();
  return (
    <Datagrid rowClick="edit" bulkActionButtons={false}>
      <TextField source="name" label={t('resources.network/locations.fields.name')} />
      <TextField source="region" label={t('resources.network/locations.fields.region')} />
      <TextField source="address" label={t('resources.network/locations.fields.address')} />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  );
};

export const LocationList = (props: any) => {
  const t = useTranslate();
  
  return (
    <List {...props} title={t('resources.network/locations.name')}>
      <LocationListGrid />
    </List>
  );
};

export const LocationCreate = (props: any) => {
  const t = useTranslate();
  
  return (
    <Create {...props} title={t('resources.network/locations.create_title')}>
      <SimpleForm>
        <TextInput 
          source="name" 
          label={t('resources.network/locations.fields.name')} 
          required 
        />
        <TextInput 
          source="region" 
          label={t('resources.network/locations.fields.region')} 
        />
        <TextInput 
          source="address" 
          label={t('resources.network/locations.fields.address')}
          multiline
          rows={3}
        />
        <TextInput 
          source="lat" 
          label={t('resources.network/locations.fields.lat')}
          type="number"

        />
        <TextInput 
          source="lng" 
          label={t('resources.network/locations.fields.lng')}
          type="number"

        />
      </SimpleForm>
    </Create>
  );
};

export const LocationEdit = (props: any) => {
  const t = useTranslate();
  
  return (
    <Edit {...props} title={t('resources.network/locations.edit_title')}>
      <SimpleForm>
        <TextInput 
          source="name" 
          label={t('resources.network/locations.fields.name')} 
          required 
        />
        <TextInput 
          source="region" 
          label={t('resources.network/locations.fields.region')} 
        />
        <TextInput 
          source="address" 
          label={t('resources.network/locations.fields.address')}
          multiline
          rows={3}
        />
        <TextInput 
          source="lat" 
          label={t('resources.network/locations.fields.lat')}
          type="number"

        />
        <TextInput 
          source="lng" 
          label={t('resources.network/locations.fields.lng')}
          type="number"

        />
      </SimpleForm>
    </Edit>
  );
};

export default LocationList;
