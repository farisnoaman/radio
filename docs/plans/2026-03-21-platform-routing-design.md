# Platform Routing Reorganization Design

## Overview
Move all platform-level pages under `/platform` route with separate Admin instance, layout, and menu.

## Architecture

### Current State
- Platform pages scattered: `/platform/dashboard`, `/quotas`, `/providers`, `/providers/registrations`
- All resources in single Admin component
- Shared layout between platform and provider operations

### Target Structure

```
BrowserRouter
├── /landing/*              → LandingI18nProvider → LandingPage (public)
├── /platform/*             → PlatformAdmin (separate Admin instance)
│   ├── /platform/dashboard          → PlatformDashboard
│   ├── /platform/monitoring         → MonitoringDashboard
│   ├── /platform/monitoring/devices → DeviceHealthList (Resource)
│   ├── /platform/quotas             → QuotaList (Resource)
│   ├── /platform/providers          → ProviderList (Resource)
│   ├── /platform/registrations      → ProviderRegistrationList (Resource)
│   └── /platform/backups            → BackupList (Resource)
└── /*                      → MainAdmin (provider-specific operations)
    ├── /radius/users
    ├── /products
    ├── /agents
    └── ... (provider resources)
```

## Component Structure

### New Files
1. **PlatformLayout.tsx** - Platform-specific layout with custom sidebar
2. **PlatformMenu.tsx** - Platform navigation menu

### Modify Files
1. **App.tsx** - Create separate PlatformAdmin route
2. **CustomMenu.tsx** - Remove platform items from main menu

## Implementation Steps

1. Create PlatformLayout component
2. Create PlatformMenu component
3. Reorganize App.tsx routing
4. Update CustomMenu.tsx
5. Test all routes
6. Update translations

## Benefits
- Clear separation: Platform vs Provider concerns
- Different UX for platform admins
- Independent routing and layouts
- Provider admins see only relevant sections
