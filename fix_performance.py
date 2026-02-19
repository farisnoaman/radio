import re

with open('web/src/pages/FinancialPerformance.tsx', 'r') as f:
    content = f.read()

# Add useMediaQuery hook
if "window" not in content and "useMediaQuery" not in content:
    content = content.replace("import { useTheme, alpha } from '@mui/material/styles';", "import { useTheme, alpha } from '@mui/material/styles';\nimport useMediaQuery from '@mui/material/useMediaQuery';")

if "const theme = useTheme();" in content and "const isMobile = useMediaQuery(theme.breakpoints.down('sm'));" not in content:
    content = content.replace("const theme = useTheme();", "const theme = useTheme();\n    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));")

# Fix Table 1: Detailed Agent Performance
old_agent_table = """                        <Typography variant="h6" sx={{ p: 2, fontWeight: 600 }}>Detailed Agent Performance</Typography>
                        <Divider />
                        <Table>
                            <TableHead>
                                <TableRow sx={{ backgroundColor: alpha(theme.palette.primary.main, 0.05) }}>
                                    <TableCell>Agent Name</TableCell>
                                    <TableCell>Username</TableCell>
                                    <TableCell align="right">Wallet Balance</TableCell>
                                    <TableCell align="right">Total Vouchers</TableCell>
                                    <TableCell align="right">Sold (Used)</TableCell>
                                    <TableCell align="right">Unused</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {(report.agents || []).map((agent) => (
                                    <TableRow key={agent.id} hover>
                                        <TableCell sx={{ fontWeight: 'bold' }}>{agent.name}</TableCell>
                                        <TableCell>{agent.username}</TableCell>
                                        <TableCell align="right" sx={{ fontWeight: 'bold', color: theme.palette.primary.main }}>
                                            ${agent.balance.toFixed(2)}
                                        </TableCell>
                                        <TableCell align="right">{agent.total_vouchers}</TableCell>
                                        <TableCell align="right">
                                            <Chip label={agent.used_vouchers} color="success" size="small" variant="outlined" sx={{ fontWeight: 600 }} />
                                        </TableCell>
                                        <TableCell align="right">
                                            <Chip label={agent.unused_vouchers} color="warning" size="small" variant="outlined" sx={{ fontWeight: 600 }} />
                                        </TableCell>
                                    </TableRow>
                                ))}
                                {(report.agents || []).length === 0 && (
                                    <TableRow>
                                        <TableCell colSpan={6} align="center" sx={{ py: 3, color: 'text.secondary' }}>No agents found</TableCell>
                                    </TableRow>
                                )}
                            </TableBody>
                        </Table>"""

new_agent_table = """                        <Typography variant="h6" sx={{ p: 2, fontWeight: 600 }}>Detailed Agent Performance</Typography>
                        <Divider />
                        {isMobile ? (
                            <Box sx={{ p: 2 }}>
                                {(report.agents || []).map((agent) => (
                                    <Card variant="outlined" key={agent.id} sx={{ mb: 2, borderRadius: 2 }}>
                                        <CardContent sx={{ pb: 2 }}>
                                            <Stack direction="row" justifyContent="space-between" mb={1}>
                                                <Typography variant="subtitle1" fontWeight="bold">{agent.name}</Typography>
                                                <Typography variant="subtitle1" fontWeight="bold" color="primary.main">${agent.balance.toFixed(2)}</Typography>
                                            </Stack>
                                            <Typography variant="body2" color="text.secondary" mb={2}>@{agent.username}</Typography>
                                            <Grid container spacing={1}>
                                                <Grid item xs={4}>
                                                    <Typography variant="caption" color="text.secondary">Total</Typography>
                                                    <Typography variant="body2" fontWeight="bold">{agent.total_vouchers}</Typography>
                                                </Grid>
                                                <Grid item xs={4}>
                                                    <Typography variant="caption" color="text.secondary">Sold</Typography>
                                                    <Box><Chip label={agent.used_vouchers} color="success" size="small" variant="outlined" sx={{ fontWeight: 600, height: 20 }} /></Box>
                                                </Grid>
                                                <Grid item xs={4}>
                                                    <Typography variant="caption" color="text.secondary">Unused</Typography>
                                                    <Box><Chip label={agent.unused_vouchers} color="warning" size="small" variant="outlined" sx={{ fontWeight: 600, height: 20 }} /></Box>
                                                </Grid>
                                            </Grid>
                                        </CardContent>
                                    </Card>
                                ))}
                                {(report.agents || []).length === 0 && (
                                    <Typography color="text.secondary" textAlign="center" py={3}>No agents found</Typography>
                                )}
                            </Box>
                        ) : (
                        <Table>
                            <TableHead>
                                <TableRow sx={{ backgroundColor: alpha(theme.palette.primary.main, 0.05) }}>
                                    <TableCell>Agent Name</TableCell>
                                    <TableCell>Username</TableCell>
                                    <TableCell align="right">Wallet Balance</TableCell>
                                    <TableCell align="right">Total Vouchers</TableCell>
                                    <TableCell align="right">Sold (Used)</TableCell>
                                    <TableCell align="right">Unused</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {(report.agents || []).map((agent) => (
                                    <TableRow key={agent.id} hover>
                                        <TableCell sx={{ fontWeight: 'bold' }}>{agent.name}</TableCell>
                                        <TableCell>{agent.username}</TableCell>
                                        <TableCell align="right" sx={{ fontWeight: 'bold', color: theme.palette.primary.main }}>
                                            ${agent.balance.toFixed(2)}
                                        </TableCell>
                                        <TableCell align="right">{agent.total_vouchers}</TableCell>
                                        <TableCell align="right">
                                            <Chip label={agent.used_vouchers} color="success" size="small" variant="outlined" sx={{ fontWeight: 600 }} />
                                        </TableCell>
                                        <TableCell align="right">
                                            <Chip label={agent.unused_vouchers} color="warning" size="small" variant="outlined" sx={{ fontWeight: 600 }} />
                                        </TableCell>
                                    </TableRow>
                                ))}
                                {(report.agents || []).length === 0 && (
                                    <TableRow>
                                        <TableCell colSpan={6} align="center" sx={{ py: 3, color: 'text.secondary' }}>No agents found</TableCell>
                                    </TableRow>
                                )}
                            </TableBody>
                        </Table>
                        )}"""

content = content.replace(old_agent_table, new_agent_table)


# Fix Table 2: Admin Batch Details
old_batch_table = """                            <Table>
                                <TableHead>
                                    <TableRow sx={{ backgroundColor: alpha(theme.palette.primary.main, 0.05) }}>
                                        <TableCell>Batch Name</TableCell>
                                        <TableCell>Product</TableCell>
                                        <TableCell align="right">Count</TableCell>
                                        <TableCell align="right">Sold</TableCell>
                                        <TableCell align="right">Unused</TableCell>
                                        <TableCell align="right">Total Value</TableCell>
                                        <TableCell align="right">Generated At</TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {(report.admin.batches || []).map((batch) => (
                                        <TableRow key={batch.id} hover>
                                            <TableCell sx={{ fontWeight: 'bold' }}>{batch.name}</TableCell>
                                            <TableCell>
                                                <Chip label={batch.product_name} size="small" sx={{ borderRadius: 1 }} />
                                            </TableCell>
                                            <TableCell align="right">{batch.count}</TableCell>
                                            <TableCell align="right">
                                                <Typography variant="body2" color="success.main" fontWeight={600}>{batch.used_vouchers}</Typography>
                                            </TableCell>
                                            <TableCell align="right">
                                                <Typography variant="body2" color="warning.main" fontWeight={600}>{batch.unused_vouchers}</Typography>
                                            </TableCell>
                                            <TableCell align="right" sx={{ fontWeight: 'bold' }}>
                                                ${batch.total_cost.toFixed(2)}
                                            </TableCell>
                                            <TableCell align="right" sx={{ color: 'text.secondary', fontSize: '0.875rem' }}>
                                                {new Date(batch.created_at).toLocaleString()}
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                    {(report.admin.batches || []).length === 0 && (
                                        <TableRow>
                                            <TableCell colSpan={7} align="center" sx={{ py: 3, color: 'text.secondary' }}>No admin batches found</TableCell>
                                        </TableRow>
                                    )}
                                </TableBody>
                            </Table>"""

new_batch_table = """                            {isMobile ? (
                                <Box sx={{ p: 2 }}>
                                    {(report.admin.batches || []).map((batch) => (
                                        <Card variant="outlined" key={batch.id} sx={{ mb: 2, borderRadius: 2 }}>
                                            <CardContent sx={{ pb: 2 }}>
                                                <Stack direction="row" justifyContent="space-between" mb={1}>
                                                    <Typography variant="subtitle1" fontWeight="bold">{batch.name}</Typography>
                                                    <Typography variant="subtitle1" fontWeight="bold" color="primary.main">${batch.total_cost.toFixed(2)}</Typography>
                                                </Stack>
                                                <Box mb={2}>
                                                    <Chip label={batch.product_name} size="small" sx={{ borderRadius: 1 }} />
                                                </Box>
                                                <Grid container spacing={1}>
                                                    <Grid item xs={4}>
                                                        <Typography variant="caption" color="text.secondary">Total</Typography>
                                                        <Typography variant="body2" fontWeight="bold">{batch.count}</Typography>
                                                    </Grid>
                                                    <Grid item xs={4}>
                                                        <Typography variant="caption" color="text.secondary">Sold</Typography>
                                                        <Typography variant="body2" color="success.main" fontWeight={600}>{batch.used_vouchers}</Typography>
                                                    </Grid>
                                                    <Grid item xs={4}>
                                                        <Typography variant="caption" color="text.secondary">Unused</Typography>
                                                        <Typography variant="body2" color="warning.main" fontWeight={600}>{batch.unused_vouchers}</Typography>
                                                    </Grid>
                                                </Grid>
                                                <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 2, textAlign: 'right' }}>
                                                    {new Date(batch.created_at).toLocaleString()}
                                                </Typography>
                                            </CardContent>
                                        </Card>
                                    ))}
                                    {(report.admin.batches || []).length === 0 && (
                                        <Typography color="text.secondary" textAlign="center" py={3}>No admin batches found</Typography>
                                    )}
                                </Box>
                            ) : (
                            <Table>
                                <TableHead>
                                    <TableRow sx={{ backgroundColor: alpha(theme.palette.primary.main, 0.05) }}>
                                        <TableCell>Batch Name</TableCell>
                                        <TableCell>Product</TableCell>
                                        <TableCell align="right">Count</TableCell>
                                        <TableCell align="right">Sold</TableCell>
                                        <TableCell align="right">Unused</TableCell>
                                        <TableCell align="right">Total Value</TableCell>
                                        <TableCell align="right">Generated At</TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {(report.admin.batches || []).map((batch) => (
                                        <TableRow key={batch.id} hover>
                                            <TableCell sx={{ fontWeight: 'bold' }}>{batch.name}</TableCell>
                                            <TableCell>
                                                <Chip label={batch.product_name} size="small" sx={{ borderRadius: 1 }} />
                                            </TableCell>
                                            <TableCell align="right">{batch.count}</TableCell>
                                            <TableCell align="right">
                                                <Typography variant="body2" color="success.main" fontWeight={600}>{batch.used_vouchers}</Typography>
                                            </TableCell>
                                            <TableCell align="right">
                                                <Typography variant="body2" color="warning.main" fontWeight={600}>{batch.unused_vouchers}</Typography>
                                            </TableCell>
                                            <TableCell align="right" sx={{ fontWeight: 'bold' }}>
                                                ${batch.total_cost.toFixed(2)}
                                            </TableCell>
                                            <TableCell align="right" sx={{ color: 'text.secondary', fontSize: '0.875rem' }}>
                                                {new Date(batch.created_at).toLocaleString()}
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                    {(report.admin.batches || []).length === 0 && (
                                        <TableRow>
                                            <TableCell colSpan={7} align="center" sx={{ py: 3, color: 'text.secondary' }}>No admin batches found</TableCell>
                                        </TableRow>
                                    )}
                                </TableBody>
                            </Table>
                            )}"""

content = content.replace(old_batch_table, new_batch_table)

with open('web/src/pages/FinancialPerformance.tsx', 'w') as f:
    f.write(content)

print("Updated financial performance tables")
