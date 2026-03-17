import re

def enhance_products():
    with open('web/src/resources/products.tsx', 'r') as f:
        content = f.read()

    new_grid = """
const ProductGrid = () => {
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
            <CardContent sx={{ pb: 1 }}>
              <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                <Box display="flex" alignItems="center" gap={1.5}>
                  <Avatar sx={{ bgcolor: record.color || 'primary.main', width: 40, height: 40, fontWeight: 'bold' }}>
                    {record.name?.charAt(0).toUpperCase()}
                  </Avatar>
                  <Box>
                    <Typography variant="subtitle1" component="div" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
                      <TextField source="name" />
                    </Typography>
                    <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                      ID: {record.id}
                    </Typography>
                  </Box>
                </Box>
                <StatusIndicator isEnabled={record.status === 'enabled'} />
              </Box>
              
              <Box sx={{ bgcolor: theme => alpha(theme.palette.grey[500], 0.05), p: 1.5, borderRadius: 2, mb: 2 }}>
                <Typography variant="body2" color="text.secondary" sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                  <DataIcon fontSize="small" color="action" />
                  Profile: <strong><ReferenceField source="radius_profile_id" reference="radius-profiles"><TextField source="name" /></ReferenceField></strong>
                </Typography>
                <Box display="flex" justifyContent="space-between" mb={1}>
                  <Typography variant="body2" color="text.secondary">Price:</Typography>
                  <Typography variant="body2" sx={{ fontWeight: 'bold', color: 'success.main' }}>
                    <NumberField source="price" options={{ style: 'currency', currency: 'USD' }} />
                  </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between" mb={1}>
                  <Typography variant="body2" color="text.secondary">Quota:</Typography>
                  <Typography variant="body2" sx={{ fontWeight: 'bold' }}>
                    {formatQuota(record.data_quota)}
                  </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between">
                  <Typography variant="body2" color="text.secondary">Rates (U/D):</Typography>
                  <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                     {formatRate(record.up_rate)} / {formatRate(record.down_rate)}
                  </Typography>
                </Box>
              </Box>
            </CardContent>
            <CardActions sx={{ justifyContent: 'flex-end', borderTop: theme => `1px solid ${theme.palette.divider}`, px: 2, py: 1.5, gap: 1 }}>
              <EditButton label="" size="small" />
              <DeleteButton label="" size="small" />
              <ListButton component={Link} to={`/products/${record.id}/show`} label="View" icon={<VisibilityIcon />} size="small" variant="outlined" />
            </CardActions>
          </Card>
        </RecordContextProvider>
      ))}
    </Box>
  );
};
"""

    # We need to add VisibilityIcon and Link if we haven't already.
    # The user imports Link from react-router-dom and VisibilityIcon from mui/icons-material
    if 'VisibilityIcon' not in content:
        content = content.replace("DataUsage as DataIcon,\n} from '@mui/icons-material';", "DataUsage as DataIcon,\n  Visibility as VisibilityIcon,\n} from '@mui/icons-material';")
    if "import { Link } from 'react-router-dom';" not in content:
        content = "import { Link } from 'react-router-dom';\n" + content

    old_grid_pattern = r"const ProductGrid = \(\) => \{.+?return \(.+?</Box>\s*\);\s*\};\s*"
    
    if re.search(old_grid_pattern, content, flags=re.DOTALL):
        content = re.sub(old_grid_pattern, new_grid.strip() + "\n", content, flags=re.DOTALL)
        with open('web/src/resources/products.tsx', 'w') as f:
            f.write(content)
            print("Enhanced products grid")

def enhance_agents():
    with open('web/src/resources/agents.tsx', 'r') as f:
        content = f.read()

    new_grid = """
const AgentGrid = () => {
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
                        <CardContent sx={{ pb: 1 }}>
                            <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                                <Box display="flex" alignItems="center" gap={1.5}>
                                    <Avatar sx={{ bgcolor: 'secondary.main', width: 40, height: 40, fontWeight: 'bold' }}>
                                        {record.username?.charAt(0).toUpperCase()}
                                    </Avatar>
                                    <Box>
                                        <Typography variant="subtitle1" component="div" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
                                            <TextField source="username" />
                                        </Typography>
                                        <Typography variant="caption" color="text.secondary">
                                            ID: {record.id}
                                        </Typography>
                                    </Box>
                                </Box>
                                <Chip 
                                    label={record.status === 'enabled' ? 'Active' : 'Disabled'} 
                                    size="small" 
                                    color={record.status === 'enabled' ? 'success' : 'default'}
                                    variant="outlined" 
                                />
                            </Box>
                            
                            <Box sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.03)', p: 1.5, borderRadius: 2, mb: 2 }}>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Name:</span>
                                    <strong style={{ textAlign: 'right' }}><TextField source="realname" /></strong>
                                </Typography>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Contact:</span>
                                    <strong style={{ textAlign: 'right' }}><EmailField source="mobile" /></strong>
                                </Typography>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                    <span style={{ color: 'text.secondary' }}>Wallet:</span>
                                    <Typography variant="subtitle2" sx={{ fontWeight: 'bold', color: 'primary.main', fontSize: '1.1rem' }}>
                                        <FunctionField render={(r:any) => `$${(r.balance||0).toFixed(2)}`} />
                                    </Typography>
                                </Typography>
                            </Box>
                        </CardContent>
                        <CardActions sx={{ justifyContent: 'flex-end', borderTop: theme => `1px solid ${theme.palette.divider}`, px: 2, py: 1.5 }}>
                            <TopupButton />
                        </CardActions>
                    </Card>
                </RecordContextProvider>
            ))}
        </Box>
    );
};
"""
    if 'Chip,' not in content:
        content = content.replace("CardActions, Typography, useMediaQuery, Theme } from '@mui/material';", "CardActions, Typography, useMediaQuery, Theme, Avatar, Chip } from '@mui/material';")

    old_grid_pattern = r"const AgentGrid = \(\) => \{.+?return \(.+?</Box>\s*\);\s*\};\s*"
    
    if re.search(old_grid_pattern, content, flags=re.DOTALL):
        content = re.sub(old_grid_pattern, new_grid.strip() + "\n", content, flags=re.DOTALL)
        with open('web/src/resources/agents.tsx', 'w') as f:
            f.write(content)
            print("Enhanced agents grid")

def enhance_vouchers():
    with open('web/src/resources/vouchers.tsx', 'r') as f:
        content = f.read()

    new_batch_grid = """
const VoucherBatchGrid = () => {
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
                        <CardContent sx={{ pb: 1 }}>
                            <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                                <Box>
                                    <Typography variant="subtitle1" component="div" sx={{ fontWeight: 700, lineHeight: 1.2, mb: 0.5 }}>
                                        <TextField source="name" />
                                    </Typography>
                                    <Typography variant="caption" color="text.secondary">
                                        BATCH ID: {record.id}
                                    </Typography>
                                </Box>
                                <StatusField />
                            </Box>
                            
                            <Box sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.03)', p: 1.5, borderRadius: 2, mb: 2 }}>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Product:</span>
                                    <strong style={{ textAlign: 'right' }}><ReferenceField source="product_id" reference="products"><TextField source="name" /></ReferenceField></strong>
                                </Typography>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Count:</span>
                                    <strong style={{ textAlign: 'right', fontSize: '1.1em' }}><TextField source="count" /></strong>
                                </Typography>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Agent:</span>
                                    <strong style={{ textAlign: 'right' }}><ReferenceField source="agent_id" reference="agents" emptyText="System"><TextField source="realname" /></ReferenceField></strong>
                                </Typography>
                                <Typography variant="caption" sx={{ display: 'flex', justifyContent: 'space-between', color: 'error.main' }}>
                                    <span>Expires:</span>
                                    <DateField source="expire_time" showTime />
                                </Typography>
                            </Box>
                        </CardContent>
                        <CardActions sx={{ justifyContent: 'flex-start', borderTop: theme => `1px solid ${theme.palette.divider}`, px: 2, py: 1.5, flexWrap: 'wrap', gap: 1 }}>
                            <BatchActions />
                        </CardActions>
                    </Card>
                </RecordContextProvider>
            ))}
        </Box>
    );
};
"""

    new_voucher_grid = """
const VoucherGrid = () => {
    const { data, isLoading } = useListContext();
    if (isLoading || !data) return null;
    return (
        <Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)', lg: 'repeat(4, 1fr)' }} gap={2} p={2} sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}>
            {data.map(record => (
                <RecordContextProvider value={record} key={record.id}>
                    <Card 
                        elevation={0} 
                        sx={{ 
                            borderRadius: 3, 
                            border: theme => `1px solid ${theme.palette.divider}`,
                            transition: 'box-shadow 0.2s',
                            '&:hover': { boxShadow: 4 },
                            position: 'relative',
                            overflow: 'hidden'
                        }}
                    >
                        {/* Decorative side accent */}
                        <Box sx={{ position: 'absolute', left: 0, top: 0, bottom: 0, width: 4, bgcolor: record.status === 'unused' ? 'success.main' : record.status === 'used' ? 'error.main' : 'warning.main' }} />
                        
                        <CardContent sx={{ pb: 1, pl: 3 }}>
                            <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                                <Typography variant="h6" component="div" sx={{ fontFamily: 'monospace', fontWeight: 700, letterSpacing: 1 }}>
                                    <TextField source="code" />
                                </Typography>
                                <Chip 
                                    label={record.status.toUpperCase()} 
                                    size="small" 
                                    color={record.status === 'unused' ? 'success' : record.status === 'used' ? 'error' : 'default'}
                                    variant={record.status === 'unused' ? 'filled' : 'outlined'}
                                    sx={{ fontWeight: 'bold', fontSize: '0.7rem' }}
                                />
                            </Box>
                            
                            <Box sx={{ bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.03)', p: 1.5, borderRadius: 2, mb: 2 }}>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Batch:</span>
                                    <strong style={{ maxWidth: '120px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                                        <ReferenceField source="batch_id" reference="voucher-batches"><TextField source="name" /></ReferenceField>
                                    </strong>
                                </Typography>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>Price:</span>
                                    <strong style={{ textAlign: 'right', color: 'success.main' }}>
                                        $<TextField source="price" />
                                    </strong>
                                </Typography>
                                <Typography variant="body2" sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                                    <span style={{ color: 'text.secondary' }}>PIN:</span>
                                    <strong style={{ fontFamily: 'monospace', letterSpacing: 2 }}>
                                        <FunctionField render={(r:any) => r.require_pin ? (r.pin_view ? r.pin : '****') : 'N/A'} />
                                    </strong>
                                </Typography>
                                <Typography variant="caption" sx={{ display: 'flex', justifyContent: 'space-between', color: 'text.secondary', mt: 1, pt: 1, borderTop: '1px dashed rgba(150,150,150,0.3)' }}>
                                    <span>Exp:</span>
                                    <DateField source="expire_time" showTime />
                                </Typography>
                            </Box>
                        </CardContent>
                        <CardActions sx={{ justifyContent: 'flex-end', borderTop: theme => `1px solid ${theme.palette.divider}`, px: 2, py: 1.5 }}>
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

    old_batch_grid_pattern = r"const VoucherBatchGrid = \(\) => \{.+?return \(.+?</Box>\s*\);\s*\};\s*"
    if re.search(old_batch_grid_pattern, content, flags=re.DOTALL):
        content = re.sub(old_batch_grid_pattern, new_batch_grid.strip() + "\n", content, flags=re.DOTALL)
        
    old_voucher_grid_pattern = r"const VoucherGrid = \(\) => \{.+?return \(.+?</Box>\s*\);\s*\};\s*"
    if re.search(old_voucher_grid_pattern, content, flags=re.DOTALL):
        content = re.sub(old_voucher_grid_pattern, new_voucher_grid.strip() + "\n", content, flags=re.DOTALL)
        
    if 'Avatar' not in content:
        content = content.replace("DialogTitle, DialogContent, DialogActions, Button as MuiButton, Typography", "DialogTitle, DialogContent, DialogActions, Button as MuiButton, Typography, Avatar")
        
    with open('web/src/resources/vouchers.tsx', 'w') as f:
        f.write(content)
        print("Enhanced vouchers grids")


enhance_products()
enhance_agents()
enhance_vouchers()

