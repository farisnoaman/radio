# Platform Routing Architecture

## Overview
Platform-level management pages are organized under `/platform` route with separate Admin instance.

## Platform Routes
- `/platform/dashboard` - Platform dashboard
- `/platform/monitoring` - Platform-wide monitoring metrics
- `/platform/monitoring/devices` - Device health monitoring
- `/platform/quotas` - Resource quota management
- `/platform/providers` - Provider management
- `/platform/registrations` - Provider registration approval
- `/platform/backups` - Backup management

## Main Admin Routes
- All provider-specific operations (RADIUS, billing, users, products, etc.)
- Accessible at root level routes

## Key Differences
- **Platform Admin**: Teal-colored sidebar (#0f766e), multi-tenant management
- **Main Admin**: Blue-colored sidebar (#1e40af), provider-specific operations

## Access Control
- Platform routes require super admin permission
- Provider admins can only access main admin routes relevant to their provider

## RTL Support
- Arabic and other RTL languages are fully supported
- Sidebar automatically positions on the right side for RTL languages
- Text direction and layout adjust automatically based on locale
