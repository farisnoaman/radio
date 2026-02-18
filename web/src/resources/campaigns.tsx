import {
    List,
    Datagrid,
    TextField,
    DateField,
    Create,
    SimpleForm,
    TextInput,
    NumberInput,
    ReferenceInput,
    SelectInput,
    required,
    useRecordContext,
    useNotify,
    useRefresh,
    Button,
    FunctionField
} from 'react-admin';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import { httpClient } from '../utils/apiClient';

const StartCampaignButton = () => {
    const record = useRecordContext();
    const notify = useNotify();
    const refresh = useRefresh();

    if (!record || record.status !== 'pending') return null;

    const handleStart = async (e: any) => {
        e.stopPropagation();
        try {
            await httpClient(`/campaigns/${record.id}/start`, { method: 'POST' });
            notify('Campaign started successfully', { type: 'success' });
            refresh();
        } catch (error: any) {
            notify(error?.body?.msg || 'Failed to start campaign', { type: 'error' });
        }
    };

    return (
        <Button label="Start" onClick={handleStart} color="primary">
            <PlayArrowIcon />
        </Button>
    );
};

const StatusField = (props: { source: string }) => {
    return (
        <FunctionField
            {...props}
            render={(record: any) => {
                let color = 'default';
                switch (record.status) {
                    case 'completed': color = 'success'; break;
                    case 'generating': color = 'info'; break;
                    case 'failed': color = 'error'; break;
                    case 'pending': color = 'warning'; break;
                }
                // Custom chip rendering or just text
                return <span style={{ color: color === 'default' ? 'inherit' : `var(--mui-palette-${color}-main)` }}>{record.status?.toUpperCase()}</span>;
            }}
        />
    );
};

export const CampaignList = () => (
    <List sort={{ field: 'id', order: 'DESC' }}>
        <Datagrid rowClick="show">
            <TextField source="id" />
            <TextField source="name" />
            <TextField source="prefix" />
            <TextField source="count" />
            <TextField source="value" label="Value" />
            <StatusField source="status" />
            <StartCampaignButton />
            <DateField source="created_at" showTime />
        </Datagrid>
    </List>
);

export const CampaignCreate = () => (
    <Create>
        <SimpleForm>
            <TextInput source="name" validate={[required()]} fullWidth />
            <TextInput source="prefix" label="Voucher Prefix" fullWidth />
            <NumberInput source="length" label="Code Length" defaultValue={12} validate={[required()]} />
            <NumberInput source="count" label="Quantity" validate={[required()]} />
            <ReferenceInput source="plan_id" reference="products" label="Product Plan">
                <SelectInput optionText="name" validate={[required()]} fullWidth />
            </ReferenceInput>
            <NumberInput source="value" label="Voucher Value" validate={[required()]} />
        </SimpleForm>
    </Create>
);
