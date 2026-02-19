import re

# 1. Fix radiusProfiles.tsx import
with open('web/src/resources/radiusProfiles.tsx', 'r') as f:
    r_content = f.read()

if "CardActions" not in r_content[:1500]: # Check in the top imports
    r_content = r_content.replace(
        "CardContent,\n  Stack,",
        "CardContent,\n  CardActions,\n  Stack,"
    )
    with open('web/src/resources/radiusProfiles.tsx', 'w') as f:
        f.write(r_content)


# 2. Fix SystemConfigPage.tsx layout
with open('web/src/pages/SystemConfigPage.tsx', 'r') as f:
    c_content = f.read()

# Make the action buttons wrap and gap properly, instead of mr: 2
old_buttons_box = """      {/* 操作按钮 */}
      <Box sx={{ mb: 3 }}>
        <Button
          variant="contained"
          startIcon={<SaveIcon />}
          onClick={handleSave}
          disabled={saveMutation.isPending || isLoading}
          sx={{ mr: 2 }}
        >
          {saveMutation.isPending ? translate('pages.system_config.saving') : translate('pages.system_config.save')}
        </Button>
        <Button
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={() => setResetDialogOpen(true)}
          disabled={saveMutation.isPending || isLoading}
          sx={{ mr: 2 }}
        >
          {translate('pages.system_config.reset')}
        </Button>
        <Button
          variant="text"
          startIcon={<RefreshIcon />}
          onClick={handleReload}
          disabled={isLoading}
        >
          {isLoading ? translate('pages.system_config.loading') : translate('pages.system_config.reload')}
        </Button>
      </Box>"""

new_buttons_box = """      {/* 操作按钮 */}
      <Box sx={{ mb: 3, display: 'flex', flexWrap: 'wrap', gap: 2 }}>
        <Button
          variant="contained"
          startIcon={<SaveIcon />}
          onClick={handleSave}
          disabled={saveMutation.isPending || isLoading}
        >
          {saveMutation.isPending ? translate('pages.system_config.saving') : translate('pages.system_config.save')}
        </Button>
        <Button
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={() => setResetDialogOpen(true)}
          disabled={saveMutation.isPending || isLoading}
        >
          {translate('pages.system_config.reset')}
        </Button>
        <Button
          variant="text"
          startIcon={<RefreshIcon />}
          onClick={handleReload}
          disabled={isLoading}
        >
          {isLoading ? translate('pages.system_config.loading') : translate('pages.system_config.reload')}
        </Button>
      </Box>"""

c_content = c_content.replace(old_buttons_box, new_buttons_box)

# Make the config grid responsive
old_grid = "<Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))', gap: 3 }}>"
new_grid = "<Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: 'repeat(auto-fit, minmax(320px, 1fr))' }, gap: 3 }}>"

c_content = c_content.replace(old_grid, new_grid)

# Adjust typography for mobile
old_h4 = """        <Typography variant="h4" gutterBottom>"""
new_h4 = """        <Typography variant="h4" sx={{ fontSize: { xs: '1.5rem', sm: '2.125rem' }, fontWeight: { xs: 600, sm: 400 } }} gutterBottom>"""

c_content = c_content.replace(old_h4, new_h4)

with open('web/src/pages/SystemConfigPage.tsx', 'w') as f:
    f.write(c_content)

print("Applied responsive tweaks to SystemConfigPage and fixed radiusProfiles import.")
