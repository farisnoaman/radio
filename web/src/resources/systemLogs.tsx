import {
    Datagrid,
    DateField,
    List,
    TextField,
    TextInput,
    useListContext,
    RecordContextProvider
} from 'react-admin';
import { Box, Card, CardContent, Typography, useMediaQuery, Theme, Chip } from '@mui/material';

const logFilters = [
    <TextInput source="operator" label="Operator" alwaysOn />,
    <TextInput source="action" label="Action" alwaysOn />,
    <TextInput source="keyword" label="Keyword" alwaysOn />,
];


const LogGrid = () => {
    const { data, isLoading } = useListContext();
    if (isLoading || !data) return null;
    return (
        <Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)' }} gap={2} p={2} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}>
            {data.map(record => (
                <RecordContextProvider value={record} key={record.id}>
                    <Card
                        elevation={0}
                        sx={{
                            borderRadius: 3,
                            border: theme => `1px solid ${theme.palette.divider}`,
                            transition: 'box-shadow 0.2s',
                            '&:hover': { boxShadow: 4 }
                        }}
                    >
                        <CardContent sx={{ pb: 2 }}>
                            <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={1.5}>
                                <Box>
                                    <Typography variant="subtitle2" color="primary.main" sx={{ fontWeight: 600 }}>
                                        #{record.id} - {record.opr_name}
                                    </Typography>
                                    <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 0.5 }}>
                                        <DateField source="opt_time" showTime />
                                    </Typography>
                                </Box>
                                <Chip
                                    label={<TextField source="opt_action" />}
                                    size="small"
                                    color="info"
                                    variant="outlined"
                                    sx={{ fontWeight: 500 }}
                                />
                            </Box>

                            <Box sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.03)', p: 1.5, borderRadius: 2 }}>
                                <Typography variant="body2" sx={{ wordBreak: 'break-word', whiteSpace: 'pre-wrap', mb: 1 }}>
                                    {record.opt_desc}
                                </Typography>
                                <Typography variant="caption" sx={{ fontFamily: 'monospace', color: 'text.secondary' }}>
                                    IP: {record.opr_ip}
                                </Typography>
                            </Box>
                        </CardContent>
                    </Card>
                </RecordContextProvider>
            ))}
        </Box>
    );
};

export const SystemLogList = () => {
    const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
    return (
        <List filters={logFilters} sort={{ field: 'id', order: 'DESC' }}>
            {isSmall ? (
                <LogGrid />
            ) : (
                <Datagrid bulkActionButtons={false}>
                    <TextField source="id" />
                    <TextField source="opr_name" label="Operator" />
                    <TextField source="opr_ip" label="IP Address" />
                    <TextField source="opt_action" label="Action" />
                    <TextField source="opt_desc" label="Description" />
                    <DateField source="opt_time" label="Time" showTime />
                </Datagrid>
            )}
        </List>
    );
};
