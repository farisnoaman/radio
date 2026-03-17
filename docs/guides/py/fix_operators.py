import re

with open('web/src/resources/operators.tsx', 'r') as f:
    content = f.read()

# Add essential imports to the top
if "RecordContextProvider" not in content:
    content = content.replace("  FunctionField\n} from 'react-admin';", "  FunctionField,\n  RecordContextProvider\n} from 'react-admin';")
if "ShowButton" not in content:
    content = content.replace("  DeleteButton,", "  DeleteButton,\n  ShowButton,")

# Add the OperatorGrid Component
operator_grid = """
const OperatorGrid = () => {
  const { data, isLoading } = useListContext<Operator>();
  const translate = useTranslate();
  
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
            <CardContent sx={{ pb: 1 }}>
              <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                <Box display="flex" alignItems="center" gap={1.5}>
                  <Avatar sx={{ bgcolor: record.status === 'enabled' ? 'primary.main' : 'grey.400', width: 40, height: 40, fontWeight: 'bold' }}>
                    {record.username?.charAt(0).toUpperCase() || 'O'}
                  </Avatar>
                  <Box>
                    <Typography variant="subtitle1" component="div" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
                      {record.username}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                       {record.realname}
                    </Typography>
                  </Box>
                </Box>
                <StatusIndicator isEnabled={record.status === 'enabled'} />
              </Box>
              
              <Box sx={{ bgcolor: theme => alpha(theme.palette.grey[500], 0.05), p: 1.5, borderRadius: 2, mb: 1 }}>
                <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                  <Typography variant="body2" color="text.secondary">{translate('resources.system/operators.fields.email', { _: '邮箱' })}:</Typography>
                  <Typography variant="body2" sx={{ fontWeight: 'bold' }}>
                    <EmailField source="email" />
                  </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                  <Typography variant="body2" color="text.secondary">{translate('resources.system/operators.fields.mobile', { _: '手机号' })}:</Typography>
                  <Typography variant="body2" sx={{ fontWeight: 'bold', fontFamily: 'monospace' }}>
                    <TextField source="mobile" emptyText="N/A" />
                  </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between" alignItems="center">
                   <Typography variant="body2" color="text.secondary">{translate('resources.system/operators.fields.level', { _: '权限级别' })}:</Typography>
                   <LevelField />
                </Box>
              </Box>
            </CardContent>
            <CardActions sx={{ justifyContent: 'flex-end', borderTop: theme => `1px solid ${theme.palette.divider}`, px: 2, py: 1.5, gap: 1 }}>
              <ShowButton label="" size="small" />
            </CardActions>
          </Card>
        </RecordContextProvider>
      ))}
    </Box>
  );
};
"""

# Inject OperatorGrid component before OperatorListContent
if "const OperatorGrid =" not in content:
    content = content.replace("const OperatorListContent = () => {", operator_grid + "\nconst OperatorListContent = () => {")
    content = content.replace("  CardContent,\n  Stack,", "  CardContent,\n  CardActions,\n  Stack,")

# Replace Datagrid in OperatorListContent with responsive switch
old_datagrid = """        <Box
          sx={{
            overflowX: 'auto',
            '& .RaDatagrid-root': {
              minWidth: isMobile ? 900 : 'auto',
            },"""
            
new_datagrid = """        <Box
          sx={{
            overflowX: 'auto',
            '& .RaDatagrid-root': {
              minWidth: 'auto',
            },"""

content = content.replace(old_datagrid, new_datagrid)

old_datagrid_component = """          <Datagrid rowClick="show" bulkActionButtons={false}>
            <FunctionField
              source="username"
              label={translate('resources.system/operators.fields.username', { _: '用户名' })}
              render={() => <OperatorNameField />}
            />
            <TextField
              source="realname"
              label={translate('resources.system/operators.fields.realname', { _: '真实姓名' })}
            />
            <EmailField
              source="email"
              label={translate('resources.system/operators.fields.email', { _: '邮箱' })}
            />
            <TextField
              source="mobile"
              label={translate('resources.system/operators.fields.mobile', { _: '手机号' })}
            />
            <FunctionField
              source="level"
              label={translate('resources.system/operators.fields.level', { _: '权限级别' })}
              render={() => <LevelField />}
            />
            <DateField
              source="last_login"
              label={translate('resources.system/operators.fields.last_login', { _: '最后登录' })}
              showTime
            />
            <DateField
              source="created_at"
              label={translate('resources.system/operators.fields.created_at', { _: '创建时间' })}
              showTime
            />
          </Datagrid>"""

new_datagrid_component = """          {isMobile ? (
            <OperatorGrid />
          ) : (
          <Datagrid rowClick="show" bulkActionButtons={false}>
            <FunctionField
              source="username"
              label={translate('resources.system/operators.fields.username', { _: '用户名' })}
              render={() => <OperatorNameField />}
            />
            <TextField
              source="realname"
              label={translate('resources.system/operators.fields.realname', { _: '真实姓名' })}
            />
            <EmailField
              source="email"
              label={translate('resources.system/operators.fields.email', { _: '邮箱' })}
            />
            <TextField
              source="mobile"
              label={translate('resources.system/operators.fields.mobile', { _: '手机号' })}
            />
            <FunctionField
              source="level"
              label={translate('resources.system/operators.fields.level', { _: '权限级别' })}
              render={() => <LevelField />}
            />
            <DateField
              source="last_login"
              label={translate('resources.system/operators.fields.last_login', { _: '最后登录' })}
              showTime
            />
            <DateField
              source="created_at"
              label={translate('resources.system/operators.fields.created_at', { _: '创建时间' })}
              showTime
            />
          </Datagrid>
          )}"""

if "{isMobile ? (" not in content:
    content = content.replace(old_datagrid_component, new_datagrid_component)

with open('web/src/resources/operators.tsx', 'w') as f:
    f.write(content)

print("Updated operators table layout")
