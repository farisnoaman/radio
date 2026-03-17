import re

# 1. Update CustomLayout.tsx
with open('web/src/components/CustomLayout.tsx', 'r') as f:
    layout_content = f.read()

if "TitlePortal" not in layout_content:
    layout_content = layout_content.replace(
        "import { Layout, LayoutProps } from 'react-admin';",
        "import { Layout, LayoutProps, TitlePortal } from 'react-admin';\nimport { Box, Typography, useMediaQuery } from '@mui/material';"
    )
    
if "const isSmall =" not in layout_content:
    layout_content = layout_content.replace(
        "export const CustomLayout = ({ sx, ...rest }: CustomLayoutProps) => (",
        "export const CustomLayout = ({ sx, children, ...rest }: CustomLayoutProps & { children?: React.ReactNode }) => {\n  const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));\n  return ("
    )
    # also add closing brace
    layout_content = layout_content.replace(
        "  />\n);",
        "  >\n    {isSmall && (\n      <Box sx={{ \n          position: 'absolute', \n          top: 16, \n          left: 16, \n          zIndex: 10, \n          maxWidth: 'calc(100% - 130px)',\n          pointerEvents: 'none'\n      }}>\n          <Typography variant=\"h6\" sx={{ \n              fontWeight: 'bold', \n              whiteSpace: 'nowrap',\n              overflow: 'hidden',\n              textOverflow: 'ellipsis',\n              pointerEvents: 'auto',\n              color: 'text.primary'\n          }}>\n              <TitlePortal />\n          </Typography>\n      </Box>\n    )}\n    {children}\n  </Layout>\n);"
    )
    
    # Fix position: relative on the content area so absolute positioning works if needed (though top:16 relative to what? RaLayout-content is not offset parent natively)
    # Actually, we should put position: relative on RaLayout-content
    if "'& .RaLayout-content': {" in layout_content:
        layout_content = layout_content.replace(
            "'& .RaLayout-content': {",
            "'& .RaLayout-content': {\n          position: 'relative',"
        )
        
with open('web/src/components/CustomLayout.tsx', 'w') as f:
    f.write(layout_content)


# 2. Update CustomAppBar.tsx
with open('web/src/components/CustomAppBar.tsx', 'r') as f:
    appbar_content = f.read()

if "const isSmall =" not in appbar_content:
    appbar_content = appbar_content.replace(
        "const navigate = useNavigate();",
        "const navigate = useNavigate();\n  const isSmall = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));"
    )

if "<TitlePortal />" in appbar_content:
    appbar_content = appbar_content.replace(
        "<TitlePortal />",
        "{!isSmall && <TitlePortal />}"
    )

with open('web/src/components/CustomAppBar.tsx', 'w') as f:
    f.write(appbar_content)

print("Updated layout and appbar")

