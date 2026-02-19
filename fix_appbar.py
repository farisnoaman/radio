import re

with open('web/src/components/CustomAppBar.tsx', 'r') as f:
    content = f.read()

# Add useLocation, useNavigate imports
if "react-router-dom" not in content:
    content = "import { useLocation, useNavigate } from 'react-router-dom';\n" + content
if "ArrowBackIcon" not in content:
    content = content.replace("import MenuOpenIcon from '@mui/icons-material/MenuOpen';", "import MenuOpenIcon from '@mui/icons-material/MenuOpen';\nimport ArrowBackIcon from '@mui/icons-material/ArrowBack';")

# Add location and navigate hooks to CustomAppBar
if "const location = useLocation();" not in content:
    content = content.replace("const [sidebarOpen, setSidebarOpen] = useSidebarState();", "const [sidebarOpen, setSidebarOpen] = useSidebarState();\n  const location = useLocation();\n  const navigate = useNavigate();")

# Determine if we should show a back button
# E.g. paths like /products/1/show or /products/create
back_button_logic = """
  // Determine if we are on a page that should have a back button (not a top-level list page)
  const isSubPage = location.pathname !== '/' && location.pathname.split('/').length > 2;

  const handleBack = () => {
    navigate(-1);
  };
"""
if "const isSubPage =" not in content:
    content = content.replace("const handleToggleSidebar = () => {", back_button_logic + "\n  const handleToggleSidebar = () => {")

# Modify the AppBar to include the back button and restyle the Title
old_stack = """        <Stack direction="row" spacing={1} alignItems="center">
          {/* 侧边栏展开/收起按钮 */}
          <Tooltip title={sidebarOpen ? translate('appbar.collapse_menu') : translate('appbar.expand_menu')}>
            <IconButton 
              size="medium"
              onClick={handleToggleSidebar}
              sx={{
                color: isDark ? '#f1f5f9' : '#6b7280',
                transition: 'all 0.2s ease',
                '&:hover': {
                  backgroundColor: isDark 
                    ? 'rgba(255, 255, 255, 0.1)'
                    : 'rgba(0, 0, 0, 0.05)',
                },
              }}
            >
              {sidebarOpen ? <MenuOpenIcon /> : <MenuIcon />}
            </IconButton>
          </Tooltip>
          <Typography 
            variant="h6" 
            sx={{ 
              fontSize: 18, 
              fontWeight: 700, 
              color: isDark ? '#f1f5f9' : '#1f2937',
              letterSpacing: '0.5px',
            }}
          >
            TOUGHRADIUS
          </Typography>
        </Stack>"""

new_stack = """        <Stack direction="row" spacing={0} alignItems="center">
          {isSubPage ? (
            <Tooltip title={translate('ra.action.back') || 'Back'}>
              <IconButton 
                size="medium"
                onClick={handleBack}
                sx={{
                  color: isDark ? '#f1f5f9' : '#6b7280',
                  mr: 1
                }}
              >
                <ArrowBackIcon />
              </IconButton>
            </Tooltip>
          ) : (
            <Tooltip title={sidebarOpen ? translate('appbar.collapse_menu') : translate('appbar.expand_menu')}>
              <IconButton 
                size="medium"
                onClick={handleToggleSidebar}
                sx={{
                  color: isDark ? '#f1f5f9' : '#6b7280',
                  mr: 1,
                  transition: 'all 0.2s ease',
                  '&:hover': {
                    backgroundColor: isDark 
                      ? 'rgba(255, 255, 255, 0.1)'
                      : 'rgba(0, 0, 0, 0.05)',
                  },
                }}
              >
                {sidebarOpen ? <MenuOpenIcon /> : <MenuIcon />}
              </IconButton>
            </Tooltip>
          )}

          {/* Logo only on larger screens when it's not a subpage */}
          <Box sx={{ display: { xs: isSubPage ? 'none' : 'block', sm: 'block' } }}>
            <Typography 
              variant="h6" 
              sx={{ 
                fontSize: 18, 
                fontWeight: 700, 
                color: isDark ? '#f1f5f9' : '#1f2937',
                letterSpacing: '0.5px',
              }}
            >
              TOUGHRADIUS
            </Typography>
          </Box>
        </Stack>"""

content = content.replace(old_stack, new_stack)

# Add styling to the AppBar to make the TitlePortal look better
appbar_sx_addition = """        '& #react-admin-title': {
          fontSize: { xs: '1.1rem', sm: '1.25rem' },
          fontWeight: 700,
          whiteSpace: 'nowrap',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          maxWidth: { xs: '140px', sm: '300px' },
          flex: 1,
          textAlign: 'center',
          color: isDark ? '#f1f5f9' : '#1f2937',
          position: 'absolute',
          left: '50%',
          transform: 'translateX(-50%)',
          letterSpacing: '0.3px',
        },"""

if "'& #react-admin-title'" not in content:
    content = content.replace("transition: 'all 0.3s ease',", "transition: 'all 0.3s ease',\n" + appbar_sx_addition)

with open('web/src/components/CustomAppBar.tsx', 'w') as f:
    f.write(content)

print("Modified CustomAppBar.tsx")

