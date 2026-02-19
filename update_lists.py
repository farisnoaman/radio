import re

def update_products():
    with open('web/src/resources/products.tsx', 'r') as f:
        content = f.read()
    
    if 'useMediaQuery' not in content:
        content = content.replace("import { Box, Card, CardContent", "import { useMediaQuery, Theme, CardActions, Box, Card, CardContent")
    if 'useListContext' not in content:
        content = content.replace("useRefresh\n} from 'react-admin';", "useRefresh,\n  useListContext,\n  RecordContextProvider\n} from 'react-admin';")
        
    grid_component = """
const ProductGrid = () => {
  const { data, isLoading } = useListContext();
  if (isLoading || !data) return null;
  return (
    <Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr' }} gap={2} p={1} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}>
      {data.map(record => (
        <RecordContextProvider value={record} key={record.id}>
          <Card elevation={2} sx={{ borderRadius: 2 }}>
            <CardContent>
              <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                <Typography variant="h6"><TextField source="name" /></Typography>
                <StatusIndicator isEnabled={record.status === 'enabled'} />
              </Box>
              <Typography variant="body2" color="text.secondary" mb={1}>
                Profile: <ReferenceField source="radius_profile_id" reference="radius-profiles"><TextField source="name" /></ReferenceField>
              </Typography>
              <Box display="flex" justifyContent="space-between" mb={0.5}>
                <Typography variant="body2">Price: <NumberField source="price" options={{ style: 'currency', currency: 'USD' }} /></Typography>
                <Typography variant="body2">Quota: {formatQuota(record.data_quota)}</Typography>
              </Box>
              <Typography variant="body2">
                Rates: {formatRate(record.up_rate)} / {formatRate(record.down_rate)}
              </Typography>
            </CardContent>
            <CardActions sx={{ justifyContent: 'flex-end', borderTop: '1px solid #efefef' }}>
              <EditButton />
              <DeleteButton />
            </CardActions>
          </Card>
        </RecordContextProvider>
      ))}
    </Box>
  );
};
"""
    if "const ProductGrid" not in content:
        # Insert before ProductList
        content = content.replace('export const ProductList', grid_component + '\nexport const ProductList')
        
        # Replace list component
        old_list = """export const ProductList = (props: ListProps) => (
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
);"""
        new_list = """export const ProductList = (props: ListProps) => {
  const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
  return (
    <List {...props} sort={{ field: 'id', order: 'DESC' }}>
      {isSmall ? (
        <ProductGrid />
      ) : (
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
      )}
    </List>
  );
};"""
        content = content.replace(old_list, new_list)
        
    with open('web/src/resources/products.tsx', 'w') as f:
        f.write(content)

def update_agents():
    with open('web/src/resources/agents.tsx', 'r') as f:
        content = f.read()

    # Imports
    if 'useMediaQuery' not in content:
        content = content.replace("import {\n    Dialog", "import { Box, Card, CardContent, CardActions, Typography, useMediaQuery, Theme } from '@mui/material';\nimport {\n    Dialog")
    if 'useListContext' not in content:
        content = content.replace("FunctionField,\n} from 'react-admin';", "FunctionField,\n    useListContext,\n    RecordContextProvider\n} from 'react-admin';")

    grid_component = """
const AgentGrid = () => {
    const { data, isLoading } = useListContext();
    if (isLoading || !data) return null;
    return (
        <Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr' }} gap={2} p={1} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}>
            {data.map(record => (
                <RecordContextProvider value={record} key={record.id}>
                    <Card elevation={2} sx={{ borderRadius: 2 }}>
                        <CardContent>
                            <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                                <Typography variant="h6"><TextField source="username" /></Typography>
                                <TextField source="status" sx={{ fontWeight: 'bold' }} />
                            </Box>
                            <Typography variant="body2" color="text.secondary" mb={1}>
                                <TextField source="realname" /> | <EmailField source="mobile" />
                            </Typography>
                            <Box display="flex" justifyContent="space-between">
                                <Typography variant="body2">
                                    Balance: <strong><FunctionField render={(r:any) => (r.balance||0).toFixed(2)} /></strong>
                                </Typography>
                                <Typography variant="body2" color="text.secondary">
                                    ID: <TextField source="id" />
                                </Typography>
                            </Box>
                        </CardContent>
                        <CardActions sx={{ justifyContent: 'flex-end', borderTop: '1px solid #efefef' }}>
                            <TopupButton />
                        </CardActions>
                    </Card>
                </RecordContextProvider>
            ))}
        </Box>
    );
};
"""

    if "const AgentGrid" not in content:
        content = content.replace('export const AgentList', grid_component + '\nexport const AgentList')
        old_list = """export const AgentList = (props: ListProps) => (
    <List {...props} actions={<AgentListActions />}>
        <Datagrid>
            <TextField source="id" />
            <TextField source="username" />
            <TextField source="realname" />
            <EmailField source="mobile" />
            <FunctionField
                label="Balance"
                render={(record: any) => (record.balance || 0).toFixed(2)}
            />
            <TextField source="status" />
            <TopupButton />
        </Datagrid>
    </List>
);"""
        new_list = """export const AgentList = (props: ListProps) => {
    const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
    return (
        <List {...props} actions={<AgentListActions />}>
            {isSmall ? (
                <AgentGrid />
            ) : (
                <Datagrid>
                    <TextField source="id" />
                    <TextField source="username" />
                    <TextField source="realname" />
                    <EmailField source="mobile" />
                    <FunctionField
                        label="Balance"
                        render={(record: any) => (record.balance || 0).toFixed(2)}
                    />
                    <TextField source="status" />
                    <TopupButton />
                </Datagrid>
            )}
        </List>
    );
};"""
        content = content.replace(old_list, new_list)

    with open('web/src/resources/agents.tsx', 'w') as f:
        f.write(content)

def update_vouchers():
    with open('web/src/resources/vouchers.tsx', 'r') as f:
        content = f.read()

    # Imports Let's inject imports
    if 'useMediaQuery' not in content:
        content = content.replace("import { Box, Dialog", "import { useMediaQuery, Theme, Card, CardContent, CardActions, Box, Dialog")
    if 'RecordContextProvider' not in content:
        content = content.replace("useGetOne,\n    FunctionField,\n    BooleanInput,\n} from 'react-admin';", "useGetOne,\n    FunctionField,\n    BooleanInput,\n    RecordContextProvider,\n    useListContext\n} from 'react-admin';")

    batch_grid = """
const VoucherBatchGrid = () => {
    const { data, isLoading } = useListContext();
    if (isLoading || !data) return null;
    return (
        <Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr' }} gap={2} p={1} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}>
            {data.map(record => (
                <RecordContextProvider value={record} key={record.id}>
                    <Card elevation={2} sx={{ borderRadius: 2 }}>
                        <CardContent>
                            <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                                <Typography variant="h6"><TextField source="name" /></Typography>
                                <StatusField />
                            </Box>
                            <Typography variant="body2" color="text.secondary" mb={1}>
                                Product: <ReferenceField source="product_id" reference="products"><TextField source="name" /></ReferenceField>
                            </Typography>
                            <Box display="flex" justifyContent="space-between" mb={0.5}>
                                <Typography variant="body2">Count: <strong><TextField source="count" /></strong></Typography>
                                <Typography variant="body2" color="text.secondary">
                                    Agent: <ReferenceField source="agent_id" reference="agents" emptyText="System"><TextField source="realname" /></ReferenceField>
                                </Typography>
                            </Box>
                            <Typography variant="caption" color="text.secondary">
                                Exp: <DateField source="expire_time" showTime />
                            </Typography>
                        </CardContent>
                        <CardActions sx={{ justifyContent: 'flex-start', borderTop: '1px solid #efefef', flexWrap: 'wrap', gap: 1 }}>
                            <BatchActions />
                        </CardActions>
                    </Card>
                </RecordContextProvider>
            ))}
        </Box>
    );
};
"""

    if "const VoucherBatchGrid" not in content:
        content = content.replace('export const VoucherBatchList =', batch_grid + '\nexport const VoucherBatchList =')
        old_batch_list = """export const VoucherBatchList = (props: ListProps) => (
    <List {...props} sort={{ field: 'id', order: 'DESC' }}>
        <Datagrid>
            <TextField source="id" />
            <TextField source="name" />
            <ReferenceField source="product_id" reference="products">
                <TextField source="name" />
            </ReferenceField>
            <ReferenceField source="agent_id" reference="agents" emptyText="System">
                <TextField source="realname" />
            </ReferenceField>
            <TextField source="count" />
            <StatusField />
            <DateField source="expire_time" showTime label="Expiry Time" />
            <DateField source="created_at" showTime />
            <BatchActions />
        </Datagrid>
    </List>
);"""
        new_batch_list = """export const VoucherBatchList = (props: ListProps) => {
    const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
    return (
        <List {...props} sort={{ field: 'id', order: 'DESC' }}>
            {isSmall ? (
                <VoucherBatchGrid />
            ) : (
                <Datagrid>
                    <TextField source="id" />
                    <TextField source="name" />
                    <ReferenceField source="product_id" reference="products">
                        <TextField source="name" />
                    </ReferenceField>
                    <ReferenceField source="agent_id" reference="agents" emptyText="System">
                        <TextField source="realname" />
                    </ReferenceField>
                    <TextField source="count" />
                    <StatusField />
                    <DateField source="expire_time" showTime label="Expiry Time" />
                    <DateField source="created_at" showTime />
                    <BatchActions />
                </Datagrid>
            )}
        </List>
    );
};"""
        content = content.replace(old_batch_list, new_batch_list)

    voucher_grid = """
const VoucherGrid = () => {
    const { data, isLoading } = useListContext();
    if (isLoading || !data) return null;
    return (
        <Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr' }} gap={2} p={1} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}>
            {data.map(record => (
                <RecordContextProvider value={record} key={record.id}>
                    <Card elevation={2} sx={{ borderRadius: 2 }}>
                        <CardContent>
                            <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                                <Typography variant="h6"><TextField source="code" /></Typography>
                                <Typography variant="body2" sx={{ fontWeight: 'bold' }}><TextField source="status" /></Typography>
                            </Box>
                            <Typography variant="body2" color="text.secondary" mb={1}>
                                Batch: <ReferenceField source="batch_id" reference="voucher-batches"><TextField source="name" /></ReferenceField>
                            </Typography>
                            <Box display="flex" justifyContent="space-between" mb={0.5}>
                                <Typography variant="body2">Price: <TextField source="price" /></Typography>
                                <Typography variant="body2">PIN: <FunctionField render={(r:any) => r.require_pin ? (r.pin_view ? r.pin : '****') : 'N/A'} /></Typography>
                            </Box>
                            <Typography variant="caption" color="text.secondary">
                                Exp: <DateField source="expire_time" showTime />
                            </Typography>
                        </CardContent>
                        <CardActions sx={{ justifyContent: 'flex-end', borderTop: '1px solid #efefef' }}>
                            <RedeemButton />
                            <ExtendButton />
                        </CardActions>
                    </Card>
                </RecordContextProvider>
            ))}
        </Box>
    );
};
"""

    if "const VoucherGrid" not in content:
        content = content.replace('export const VoucherList =', voucher_grid + '\nexport const VoucherList =')
        old_list = """export const VoucherList = (props: ListProps) => (
    <List {...props} sort={{ field: 'id', order: 'DESC' }}>
        <Datagrid>
            <TextField source="id" />
            <TextField source="code" />
            <TextField source="status" />
            <ReferenceField source="batch_id" reference="voucher-batches">
                <TextField source="name" />
            </ReferenceField>
            <TextField source="price" />
            <FunctionField label="PIN" render={(record: any) => record.require_pin ? (record.pin_view ? record.pin : '****') : 'N/A'} />
            <RedeemButton />
            <ExtendButton />
            <DateField source="expire_time" showTime />
            <DateField source="created_at" showTime />
        </Datagrid>
    </List>
);"""
        new_list = """export const VoucherList = (props: ListProps) => {
    const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));
    return (
        <List {...props} sort={{ field: 'id', order: 'DESC' }}>
            {isSmall ? (
                <VoucherGrid />
            ) : (
                <Datagrid>
                    <TextField source="id" />
                    <TextField source="code" />
                    <TextField source="status" />
                    <ReferenceField source="batch_id" reference="voucher-batches">
                        <TextField source="name" />
                    </ReferenceField>
                    <TextField source="price" />
                    <FunctionField label="PIN" render={(record: any) => record.require_pin ? (record.pin_view ? record.pin : '****') : 'N/A'} />
                    <RedeemButton />
                    <ExtendButton />
                    <DateField source="expire_time" showTime />
                    <DateField source="created_at" showTime />
                </Datagrid>
            )}
        </List>
    );
};"""
        content = content.replace(old_list, new_list)

    with open('web/src/resources/vouchers.tsx', 'w') as f:
        f.write(content)

update_products()
update_agents()
update_vouchers()
