import re

with open('web/src/resources/systemLogs.tsx', 'r') as f:
    content = f.read()

# Add MUI and react-admin context imports
if "useMediaQuery" not in content:
    content = content.replace("} from 'react-admin';", "    useListContext,\n    RecordContextProvider\n} from 'react-admin';\nimport { Box, Card, CardContent, Typography, useMediaQuery, Theme, Chip } from '@mui/material';")

# Define the LogGrid component
log_grid = """
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
"""

if "const LogGrid = () => {" not in content:
    content = content.replace("export const SystemLogList = () => (", log_grid + "\nexport const SystemLogList = () => {\n    const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));\n    return (")
    # Change arrow function implicit return to explicit return block for SystemLogList
    content = content.replace("    </List>\n);", "    </List>\n    );\n};")


# Replace datagrid with conditional Datagrid or LogGrid
old_datagrid = """        <Datagrid bulkActionButtons={false}>
            <TextField source="id" />
            <TextField source="opr_name" label="Operator" />
            <TextField source="opr_ip" label="IP Address" />
            <TextField source="opt_action" label="Action" />
            <TextField source="opt_desc" label="Description" />
            <DateField source="opt_time" label="Time" showTime />
        </Datagrid>"""

new_datagrid = """        {isSmall ? (
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
        )}"""

if "{isSmall ? (" not in content:
    content = content.replace(old_datagrid, new_datagrid)

with open('web/src/resources/systemLogs.tsx', 'w') as f:
    f.write(content)

print("Updated systemLogs.tsx")
