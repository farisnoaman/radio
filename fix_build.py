import re

with open('web/src/resources/products.tsx', 'r') as f:
    products_content = f.read()

# Fix ListButton to ShowButton
products_content = products_content.replace(
    "<ListButton component={Link} to={`/products/${record.id}/show`} label=\"View\" icon={<VisibilityIcon />} size=\"small\" variant=\"outlined\" />",
    "<ShowButton label=\"\" size=\"small\" />"
)

# Add ShowButton to imports
if "ShowButton" not in products_content:
    products_content = products_content.replace(
        "EditButton,",
        "EditButton,\n    ShowButton,"
    )

with open('web/src/resources/products.tsx', 'w') as f:
    f.write(products_content)


with open('web/src/resources/vouchers.tsx', 'r') as f:
    vouchers_content = f.read()

# Fix unused Avatar import
vouchers_content = vouchers_content.replace(
    ", Typography, Avatar } from '@mui/material';",
    ", Typography } from '@mui/material';"
)

with open('web/src/resources/vouchers.tsx', 'w') as f:
    f.write(vouchers_content)

print("Fixed build errors")
