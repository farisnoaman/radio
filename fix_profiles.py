import re

with open('web/src/resources/radiusProfiles.tsx', 'r') as f:
    content = f.read()

# Add necessary imports if missing
if "RecordContextProvider" not in content:
    content = content.replace("  FunctionField\n} from 'react-admin';", "  FunctionField,\n  RecordContextProvider\n} from 'react-admin';")
if "ShowButton" not in content:
    content = content.replace("  DeleteButton,", "  DeleteButton,\n  ShowButton,")

# Create ProfileGrid component
profile_grid = """
const ProfileGrid = () => {
  const { data, isLoading } = useListContext<RadiusProfile>();
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
                    {record.name?.charAt(0).toUpperCase() || 'P'}
                  </Avatar>
                  <Box>
                    <Typography variant="subtitle1" component="div" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
                      <TextField source="name" />
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                       {translate('resources.radius/profiles.fields.active_num', { _: '并发数' })}: <strong><TextField source="active_num" /></strong>
                    </Typography>
                  </Box>
                </Box>
                <StatusIndicator isEnabled={record.status === 'enabled'} />
              </Box>
              
              <Box sx={{ bgcolor: theme => alpha(theme.palette.grey[500], 0.05), p: 1.5, borderRadius: 2, mb: 2 }}>
                <Box display="flex" justifyContent="space-between" mb={1}>
                  <Typography variant="body2" color="text.secondary">{translate('resources.radius/profiles.fields.data_quota', { _: '数据配额' })}:</Typography>
                  <Typography variant="body2" sx={{ fontWeight: 'bold' }}>
                    <QuotaField />
                  </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between" mb={1}>
                   <Typography variant="body2" color="text.secondary">{translate('resources.radius/profiles.fields.addr_pool', { _: '地址池' })}:</Typography>
                   <Typography variant="body2" sx={{ fontWeight: 'bold', color: 'primary.main' }}>
                       <TextField source="addr_pool" emptyText="N/A" />
                   </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between">
                  <Typography variant="body2" color="text.secondary">Rates (U/D):</Typography>
                  <Box sx={{ display: 'flex', gap: 1 }}>
                    <RateField source="up_rate" />
                    <RateField source="down_rate" />
                  </Box>
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

if "const ProfileGrid = () => {" not in content:
    content = content.replace("const ProfileListContent = () => {", profile_grid + "\nconst ProfileListContent = () => {")

# Modify ProfileListContent to use ProfileGrid on mobile
old_datagrid = """          <Datagrid rowClick="show" bulkActionButtons={false}>
            <FunctionField
              source="name"
              label={translate('resources.radius/profiles.fields.name', { _: '策略名称' })}
              render={() => <ProfileNameField />}
            />
            <TextField
              source="active_num"
              label={translate('resources.radius/profiles.fields.active_num', { _: '并发数' })}
            />
            <FunctionField
              source="up_rate"
              label={translate('resources.radius/profiles.fields.up_rate', { _: '上行速率' })}
              render={() => <RateField source="up_rate" />}
            />
            <FunctionField
              source="down_rate"
              label={translate('resources.radius/profiles.fields.down_rate', { _: '下行速率' })}
              render={() => <RateField source="down_rate" />}
            />
            <FunctionField
              source="data_quota"
              label={translate('resources.radius/profiles.fields.data_quota', { _: '数据配额' })}
              render={() => <QuotaField />}
            />
            <TextField
              source="addr_pool"
              label={translate('resources.radius/profiles.fields.addr_pool', { _: '地址池' })}
            />
            <TextField
              source="domain"
              label={translate('resources.radius/profiles.fields.domain', { _: '域名' })}
            />
            <DateField
              source="created_at"
              label={translate('resources.radius/profiles.fields.created_at', { _: '创建时间' })}
              showTime
            />
          </Datagrid>"""

new_datagrid = """          {isMobile ? (
            <ProfileGrid />
          ) : (
          <Datagrid rowClick="show" bulkActionButtons={false}>
            <FunctionField
              source="name"
              label={translate('resources.radius/profiles.fields.name', { _: '策略名称' })}
              render={() => <ProfileNameField />}
            />
            <TextField
              source="active_num"
              label={translate('resources.radius/profiles.fields.active_num', { _: '并发数' })}
            />
            <FunctionField
              source="up_rate"
              label={translate('resources.radius/profiles.fields.up_rate', { _: '上行速率' })}
              render={() => <RateField source="up_rate" />}
            />
            <FunctionField
              source="down_rate"
              label={translate('resources.radius/profiles.fields.down_rate', { _: '下行速率' })}
              render={() => <RateField source="down_rate" />}
            />
            <FunctionField
              source="data_quota"
              label={translate('resources.radius/profiles.fields.data_quota', { _: '数据配额' })}
              render={() => <QuotaField />}
            />
            <TextField
              source="addr_pool"
              label={translate('resources.radius/profiles.fields.addr_pool', { _: '地址池' })}
            />
            <TextField
              source="domain"
              label={translate('resources.radius/profiles.fields.domain', { _: '域名' })}
            />
            <DateField
              source="created_at"
              label={translate('resources.radius/profiles.fields.created_at', { _: '创建时间' })}
              showTime
            />
          </Datagrid>
          )}"""

if "{isMobile ? (" not in content:
    content = content.replace(old_datagrid, new_datagrid)

with open('web/src/resources/radiusProfiles.tsx', 'w') as f:
    f.write(content)

print("Updated radiusProfiles.tsx")

