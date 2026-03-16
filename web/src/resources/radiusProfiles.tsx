import {
  List,
  Datagrid,
  TextField,
  DateField,
  Edit,
  TextInput,
  Create,
  Show,
  BooleanInput,
  NumberInput,
  required,
  minLength,
  maxLength,
  useRecordContext,
  Toolbar,
  SaveButton,
  DeleteButton,
  ShowButton,
  SimpleForm,
  ToolbarProps,
  TopToolbar,
  ListButton,
  CreateButton,
  ExportButton,
  SortButton,
  useTranslate,
  useRefresh,
  useNotify,
  useListContext,
  useLocale,
  RaRecord,
  FunctionField,
  RecordContextProvider
} from 'react-admin';
import {
  Box,
  Chip,
  Typography,
  Card,
  CardContent,
  CardActions,
  Stack,
  Avatar,
  IconButton,
  Tooltip,
  Skeleton,
  useTheme,
  useMediaQuery,
  TextField as MuiTextField,
  alpha
} from '@mui/material';
import { useMemo, useCallback, useState, useEffect } from 'react';
import {
  Settings as ProfileIcon,
  Speed as SpeedIcon,
  Schedule as TimeIcon,
  Note as NoteIcon,
  ContentCopy as CopyIcon,
  Refresh as RefreshIcon,
  ArrowBack as BackIcon,
  Print as PrintIcon,
  FilterList as FilterIcon,
  Search as SearchIcon,
  Clear as ClearIcon,
  CheckCircle as EnabledIcon,
  Cancel as DisabledIcon,
  Wifi as NetworkIcon,
  Link as BindingIcon
} from '@mui/icons-material';
import {
  ServerPagination,
  ActiveFilters,
  FormSection,
  FieldGrid,
  FieldGridItem,
  formLayoutSx,
  controlWrapperSx,
  DetailItem,
  DetailSectionCard,
  EmptyValue
} from '../components';

const LARGE_LIST_PER_PAGE = 50;

// ============ 类型定义 ============

interface RadiusProfile extends RaRecord {
  name?: string;
  status?: 'enabled' | 'disabled';
  active_num?: number;
  up_rate?: number;
  down_rate?: number;
  addr_pool?: string;
  ipv6_prefix?: string;
  domain?: string;
  bind_mac?: boolean;
  bind_vlan?: boolean;
  remark?: string;
  created_at?: string;
  updated_at?: string;
}

// ============ 工具函数 ============

const formatTimestamp = (value?: string | number): string => {
  if (!value) {
    return '-';
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }
  return date.toLocaleString();
};

const formatRate = (translate: any, rate?: number): string => {
  if (rate === undefined || rate === null) return '-';
  if (rate === 0) return translate('resources.radius/profiles.units.unlimited', { _: 'Unlimited' });
  if (rate >= 1024) {
    return `${(rate / 1024).toFixed(1)} ${translate('resources.radius/profiles.units.mbps', { _: 'Mbps' })}`;
  }
  return `${rate} ${translate('resources.radius/profiles.units.kbps', { _: 'Kbps' })}`;
};

const formatQuota = (translate: any, quota?: number): string => {
  if (quota === undefined || quota === null) return '-';
  if (quota === 0) return translate('resources.radius/profiles.units.unlimited', { _: 'Unlimited' });
  if (quota >= 1024) {
    return `${(quota / 1024).toFixed(1)} ${translate('resources.radius/profiles.units.gb', { _: 'GB' })}`;
  }
  return `${quota} ${translate('resources.radius/profiles.units.mb', { _: 'MB' })}`;
};

// ============ 列表加载骨架屏 ============

const ProfileListSkeleton = ({ rows = 10 }: { rows?: number }) => (
  <Box sx={{ width: '100%' }}>
    {/* 搜索区域骨架屏 */}
    <Card
      elevation={0}
      sx={{
        mb: 2,
        borderRadius: 2,
        border: theme => `1px solid ${theme.palette.divider}`,
      }}
    >
      <CardContent sx={{ p: 2 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
          <Skeleton variant="rectangular" width={24} height={24} />
          <Skeleton variant="text" width={100} height={24} />
        </Box>
        <Box
          sx={{
            display: 'grid',
            gap: 2,
            gridTemplateColumns: {
              xs: '1fr',
              sm: 'repeat(2, 1fr)',
              md: 'repeat(4, 1fr)',
            },
          }}
        >
          {[...Array(4)].map((_, i) => (
            <Skeleton key={i} variant="rectangular" height={40} sx={{ borderRadius: 1 }} />
          ))}
        </Box>
      </CardContent>
    </Card>

    {/* 表格骨架屏 */}
    <Card
      elevation={0}
      sx={{
        borderRadius: 2,
        border: theme => `1px solid ${theme.palette.divider}`,
        overflow: 'hidden',
      }}
    >
      {/* 表头 */}
      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: 'repeat(8, 1fr)',
          gap: 1,
          p: 2,
          bgcolor: theme =>
            theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.02)',
          borderBottom: theme => `1px solid ${theme.palette.divider}`,
        }}
      >
        {[...Array(8)].map((_, i) => (
          <Skeleton key={i} variant="text" height={20} width="80%" />
        ))}
      </Box>

      {/* 表格行 */}
      {[...Array(rows)].map((_, rowIndex) => (
        <Box
          key={rowIndex}
          sx={{
            display: 'grid',
            gridTemplateColumns: 'repeat(8, 1fr)',
            gap: 1,
            p: 2,
            borderBottom: theme => `1px solid ${theme.palette.divider}`,
          }}
        >
          {[...Array(8)].map((_, colIndex) => (
            <Skeleton
              key={colIndex}
              variant="text"
              height={18}
              width={`${60 + Math.random() * 30}%`}
            />
          ))}
        </Box>
      ))}

      {/* 分页骨架屏 */}
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'flex-end',
          alignItems: 'center',
          gap: 2,
          p: 2,
        }}
      >
        <Skeleton variant="text" width={100} />
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Skeleton variant="circular" width={32} height={32} />
          <Skeleton variant="circular" width={32} height={32} />
        </Box>
      </Box>
    </Card>
  </Box>
);

// ============ 空状态组件 ============

const ProfileEmptyState = () => {
  const translate = useTranslate();
  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        py: 8,
        color: 'text.secondary',
      }}
    >
      <ProfileIcon sx={{ fontSize: 64, opacity: 0.3, mb: 2 }} />
      <Typography variant="h6" sx={{ opacity: 0.6, mb: 1 }}>
        {translate('resources.radius/profiles.empty.title', { _: 'No Policies Yet' })}
      </Typography>
      <Typography variant="body2" sx={{ opacity: 0.5 }}>
        {translate('resources.radius/profiles.empty.description', { _: 'Click "Create" button to add your first billing policy' })}
      </Typography>
    </Box>
  );
};

// ============ 搜索表头区块组件 ============

const ProfileSearchHeaderCard = () => {
  const translate = useTranslate();
  const { filterValues, setFilters, displayedFilters } = useListContext();
  const [localFilters, setLocalFilters] = useState<Record<string, string>>({});

  useEffect(() => {
    const newLocalFilters: Record<string, string> = {};
    if (filterValues) {
      Object.entries(filterValues).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          newLocalFilters[key] = String(value);
        }
      });
    }
    setLocalFilters(newLocalFilters);
  }, [filterValues]);

  const handleFilterChange = useCallback(
    (field: string, value: string) => {
      setLocalFilters(prev => ({ ...prev, [field]: value }));
    },
    [],
  );

  const handleSearch = useCallback(() => {
    const newFilters: Record<string, string> = {};
    Object.entries(localFilters).forEach(([key, value]) => {
      if (value.trim()) {
        newFilters[key] = value.trim();
      }
    });
    setFilters(newFilters, displayedFilters);
  }, [localFilters, setFilters, displayedFilters]);

  const handleClear = useCallback(() => {
    setLocalFilters({});
    setFilters({}, displayedFilters);
  }, [setFilters, displayedFilters]);

  const handleKeyPress = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        handleSearch();
      }
    },
    [handleSearch],
  );

  const filterFields = [
    { key: 'name', label: translate('resources.radius/profiles.fields.name', { _: 'Policy Name' }) },
    { key: 'addr_pool', label: translate('resources.radius/profiles.fields.addr_pool', { _: 'Address Pool' }) },
    { key: 'domain', label: translate('resources.radius/profiles.fields.domain', { _: 'Domain' }) },
  ];

  return (
    <Card
      elevation={0}
      sx={{
        mb: 2,
        borderRadius: 2,
        border: theme => `1px solid ${theme.palette.divider}`,
        overflow: 'hidden',
      }}
    >
      <Box
        sx={{
          px: 2.5,
          py: 1.5,
          bgcolor: theme =>
            theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.03)' : 'rgba(0,0,0,0.02)',
          borderBottom: theme => `1px solid ${theme.palette.divider}`,
          display: 'flex',
          alignItems: 'center',
          gap: 1.5,
        }}
      >
        <FilterIcon sx={{ color: 'primary.main', fontSize: 20 }} />
        <Typography variant="subtitle2" sx={{ fontWeight: 600, color: 'text.primary' }}>
          {translate('resources.radius/profiles.filter.title', { _: 'Filter Conditions' })}
        </Typography>
      </Box>

      <CardContent sx={{ p: 2 }}>
        <Box
          sx={{
            display: 'grid',
            gap: 1.5,
            gridTemplateColumns: {
              xs: 'repeat(1, 1fr)',
              sm: 'repeat(2, 1fr)',
              md: 'repeat(4, 1fr)',
            },
            alignItems: 'center',
          }}
        >
          {filterFields.map(field => (
            <MuiTextField
              key={field.key}
              label={field.label}
              value={localFilters[field.key] || ''}
              onChange={e => handleFilterChange(field.key, e.target.value)}
              onKeyPress={handleKeyPress}
              size="small"
              fullWidth
              sx={{
                '& .MuiInputBase-root': {
                  borderRadius: 1.5,
                },
                '& .MuiInputLabel-root': {
                  lineHeight: '1.4375em', // Better vertical centering for labels
                }
              }}
            />
          ))}

          {/* 操作按钮 */}
          <Box sx={{ display: 'flex', gap: 0.5, justifyContent: 'flex-end' }}>
            <Tooltip title={translate('ra.action.clear_filters', { _: 'Clear Filters' })}>
              <IconButton
                onClick={handleClear}
                size="small"
                sx={{
                  bgcolor: theme => alpha(theme.palette.grey[500], 0.1),
                  '&:hover': {
                    bgcolor: theme => alpha(theme.palette.grey[500], 0.2),
                  },
                }}
              >
                <ClearIcon />
              </IconButton>
            </Tooltip>
            <Tooltip title={translate('ra.action.search', { _: 'Search' })}>
              <IconButton
                onClick={handleSearch}
                color="primary"
                sx={{
                  bgcolor: theme => alpha(theme.palette.primary.main, 0.1),
                  '&:hover': {
                    bgcolor: theme => alpha(theme.palette.primary.main, 0.2),
                  },
                }}
              >
                <SearchIcon />
              </IconButton>
            </Tooltip>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

// ============ 状态组件 ============

const StatusIndicator = ({ isEnabled }: { isEnabled: boolean }) => {
  const translate = useTranslate();
  return (
    <Chip
      icon={isEnabled ? <EnabledIcon sx={{ fontSize: '0.85rem !important' }} /> : <DisabledIcon sx={{ fontSize: '0.85rem !important' }} />}
      label={isEnabled ? translate('resources.radius/profiles.status.enabled', { _: 'Enabled' }) : translate('resources.radius/profiles.status.disabled', { _: 'Disabled' })}
      size="small"
      color={isEnabled ? 'success' : 'default'}
      variant={isEnabled ? 'filled' : 'outlined'}
      sx={{ height: 22, fontWeight: 500, fontSize: '0.75rem' }}
    />
  );
};

const BooleanChip = ({ value, trueLabel, falseLabel }: { value?: boolean; trueLabel?: string; falseLabel?: string }) => {
  const translate = useTranslate();
  return (
    <Chip
      label={value ? (trueLabel || translate('common.yes', { _: 'Yes' })) : (falseLabel || translate('common.no', { _: 'No' }))}
      size="small"
      color={value ? 'success' : 'default'}
      variant="outlined"
      sx={{ height: 22, fontWeight: 500, fontSize: '0.75rem' }}
    />
  );
};

// ============ 增强版字段组件 ============

const ProfileNameField = () => {
  const record = useRecordContext<RadiusProfile>();
  if (!record) return null;

  const isEnabled = record.status === 'enabled';

  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
      <Avatar
        sx={{
          width: 32,
          height: 32,
          fontSize: '0.85rem',
          fontWeight: 600,
          bgcolor: isEnabled ? 'primary.main' : 'grey.400',
        }}
      >
        {record.name?.charAt(0).toUpperCase() || 'P'}
      </Avatar>
      <Box>
        <Typography
          variant="body2"
          sx={{ fontWeight: 600, color: 'text.primary', lineHeight: 1.3 }}
        >
          {record.name || '-'}
        </Typography>
        <StatusIndicator isEnabled={isEnabled} />
      </Box>
    </Box>
  );
};

const RateField = ({ source }: { source: 'up_rate' | 'down_rate' }) => {
  const record = useRecordContext<RadiusProfile>();
  const translate = useTranslate();
  if (!record) return null;

  const rate = record[source];
  const isUnlimited = rate === 0;
  return (
    <Chip
      label={formatRate(translate, rate)}
      size="small"
      color={isUnlimited ? 'success' : 'info'}
      variant={isUnlimited ? 'filled' : 'outlined'}
      sx={{ fontFamily: 'monospace', fontSize: '0.8rem', height: 24 }}
    />
  );
};

const QuotaField = () => {
  const record = useRecordContext<RadiusProfile>();
  const translate = useTranslate();
  if (!record) return null;

  const quota = record.data_quota;
  const isUnlimited = quota === 0;
  return (
    <Chip
      label={formatQuota(translate, quota)}
      size="small"
      color={isUnlimited ? 'success' : 'warning'}
      variant={isUnlimited ? 'filled' : 'outlined'}
      sx={{ fontFamily: 'monospace', fontSize: '0.8rem', height: 24 }}
    />
  );
};

// ============ 表单工具栏 ============

const ProfileFormToolbar = (props: ToolbarProps) => (
  <Toolbar {...props}>
    <SaveButton />
    <DeleteButton mutationMode="pessimistic" />
  </Toolbar>
);

// ============ 列表操作栏组件 ============

const ProfileListActions = () => {
  const translate = useTranslate();
  return (
    <TopToolbar>
      <SortButton
        fields={['created_at', 'name', 'up_rate', 'down_rate']}
        label={translate('ra.action.sort', { _: 'Sort' })}
      />
      <CreateButton />
      <ExportButton />
    </TopToolbar>
  );
};

// ============ 内部列表内容组件 ============


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
                       {translate('resources.radius/profiles.fields.active_num', { _: 'Sessions' })}: <strong><TextField source="active_num" /></strong>
                    </Typography>
                  </Box>
                </Box>
                <StatusIndicator isEnabled={record.status === 'enabled'} />
              </Box>
              
              <Box sx={{ bgcolor: theme => alpha(theme.palette.grey[500], 0.05), p: 1.5, borderRadius: 2, mb: 2 }}>
                <Box display="flex" justifyContent="space-between" mb={1}>
                  <Typography variant="body2" color="text.secondary">{translate('resources.radius/profiles.fields.data_quota', { _: 'Data Quota' })}:</Typography>
                  <Typography variant="body2" sx={{ fontWeight: 'bold' }}>
                    <QuotaField />
                  </Typography>
                </Box>
                <Box display="flex" justifyContent="space-between" mb={1}>
                   <Typography variant="body2" color="text.secondary">{translate('resources.radius/profiles.fields.addr_pool', { _: 'Address Pool' })}:</Typography>
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

const ProfileListContent = () => {
  const translate = useTranslate();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { data, isLoading, total } = useListContext<RadiusProfile>();

  const fieldLabels = useMemo(
    () => ({
      name: translate('resources.radius/profiles.fields.name', { _: 'Policy Name' }),
      addr_pool: translate('resources.radius/profiles.fields.addr_pool', { _: 'Address Pool' }),
      domain: translate('resources.radius/profiles.fields.domain', { _: 'Domain' }),
      status: translate('resources.radius/profiles.fields.status', { _: 'Status' }),
    }),
    [translate],
  );

  const statusLabels = useMemo(
    () => ({
      enabled: translate('resources.radius/profiles.status.enabled', { _: 'Enabled' }),
      disabled: translate('resources.radius/profiles.status.disabled', { _: 'Disabled' }),
    }),
    [translate],
  );

  if (isLoading) {
    return <ProfileListSkeleton />;
  }

  if (!data || data.length === 0) {
    return (
      <Box>
        <ProfileSearchHeaderCard />
        <Card
          elevation={0}
          sx={{
            borderRadius: 2,
            border: theme => `1px solid ${theme.palette.divider}`,
          }}
        >
          <ProfileEmptyState />
        </Card>
      </Box>
    );
  }

  return (
    <Box>
      {/* 搜索区块 */}
      <ProfileSearchHeaderCard />

      {/* 活动筛选标签 */}
      <ActiveFilters fieldLabels={fieldLabels} valueLabels={{ status: statusLabels }} />

      {/* 表格容器 */}
      <Card
        elevation={0}
        sx={{
          borderRadius: 2,
          border: theme => `1px solid ${theme.palette.divider}`,
          overflow: 'hidden',
        }}
      >
        {/* 表格统计信息 */}
        <Box
          sx={{
            px: 2,
            py: 1,
            bgcolor: theme =>
              theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.01)',
            borderBottom: theme => `1px solid ${theme.palette.divider}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
            {translate('resources.radius/profiles.total_count', { 
              count: total?.toLocaleString() || 0, 
              _: 'Total %{count} policies' 
            })}
        </Box>

        {/* 响应式表格 */}
        <Box
          sx={{
            overflowX: 'auto',
            '& .RaDatagrid-root': {
              minWidth: isMobile ? 900 : 'auto',
            },
            '& .RaDatagrid-thead': {
              position: 'sticky',
              top: 0,
              zIndex: 1,
              bgcolor: theme =>
                theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.02)',
              '& th': {
                fontWeight: 600,
                fontSize: '0.8rem',
                color: 'text.secondary',
                textTransform: 'uppercase',
                letterSpacing: '0.5px',
                py: 1.5,
                borderBottom: theme => `2px solid ${theme.palette.divider}`,
              },
            },
            '& .RaDatagrid-tbody': {
              '& tr': {
                transition: 'background-color 0.15s ease',
                cursor: 'pointer',
                '&:hover': {
                  bgcolor: theme =>
                    theme.palette.mode === 'dark'
                      ? 'rgba(255,255,255,0.05)'
                      : 'rgba(25, 118, 210, 0.04)',
                },
                '&:nth-of-type(odd)': {
                  bgcolor: theme =>
                    theme.palette.mode === 'dark'
                      ? 'rgba(255,255,255,0.01)'
                      : 'rgba(0,0,0,0.01)',
                },
              },
              '& td': {
                py: 1.5,
                fontSize: '0.875rem',
                borderBottom: theme => `1px solid ${alpha(theme.palette.divider, 0.5)}`,
              },
            },
          }}
        >
          {isMobile ? (
            <ProfileGrid />
          ) : (
          <Datagrid rowClick="show" bulkActionButtons={false}>
            <FunctionField
              source="name"
              label={translate('resources.radius/profiles.fields.name', { _: 'Policy Name' })}
              render={() => <ProfileNameField />}
            />
            <TextField
              source="active_num"
              label={translate('resources.radius/profiles.fields.active_num', { _: 'Sessions' })}
            />
            <FunctionField
              source="up_rate"
              label={translate('resources.radius/profiles.fields.up_rate', { _: 'Upload Rate' })}
              render={() => <RateField source="up_rate" />}
            />
            <FunctionField
              source="down_rate"
              label={translate('resources.radius/profiles.fields.down_rate', { _: 'Download Rate' })}
              render={() => <RateField source="down_rate" />}
            />
            <FunctionField
              source="data_quota"
              label={translate('resources.radius/profiles.fields.data_quota', { _: 'Data Quota' })}
              render={() => <QuotaField />}
            />
            <TextField
              source="addr_pool"
              label={translate('resources.radius/profiles.fields.addr_pool', { _: 'Address Pool' })}
            />
            <TextField
              source="domain"
              label={translate('resources.radius/profiles.fields.domain', { _: 'Domain' })}
            />
            <DateField
              source="created_at"
              label={translate('resources.radius/profiles.fields.created_at', { _: 'Created At' })}
              showTime
            />
          </Datagrid>
          )}
        </Box>
      </Card>
    </Box>
  );
};

// RADIUS 计费策略列表
export const RadiusProfileList = () => {
  return (
    <List
      actions={<ProfileListActions />}
      sort={{ field: 'created_at', order: 'DESC' }}
      perPage={LARGE_LIST_PER_PAGE}
      pagination={<ServerPagination />}
      empty={false}
    >
      <ProfileListContent />
    </List>
  );
};

// ============ 编辑页面 ============

export const RadiusProfileEdit = () => {
  const translate = useTranslate();
  const locale = useLocale();
  const isRTL = locale === 'ar';

  const inputLabelProps = {
    sx: {
      transformOrigin: isRTL ? 'top right' : 'top left',
      left: isRTL ? 'auto' : 0,
      right: isRTL ? 24 : 'auto',
    }
  };

  const textInputProps = {
    style: { textAlign: (isRTL ? 'right' : 'left') as any },
    dir: isRTL ? 'rtl' : 'ltr'
  };

  const numInputProps = {
    style: { textAlign: (isRTL ? 'right' : 'left') as any, direction: isRTL ? 'rtl' : 'ltr' as any }
  };

  return (
    <Edit>
      <SimpleForm toolbar={<ProfileFormToolbar />} sx={{ ...formLayoutSx, direction: isRTL ? 'rtl' : 'ltr' }}>
        <FormSection
          title={translate('resources.radius/profiles.sections.basic.title', { _: 'Basic Information' })}
          description={translate('resources.radius/profiles.sections.basic.description', { _: 'Basic configuration of the policy' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="id"
                disabled
                label={translate('resources.radius/profiles.fields.id', { _: 'Policy ID' })}
                helperText={translate('resources.radius/profiles.helpers.id', { _: 'System generated unique identifier' })}
                fullWidth
                size="small"
                inputProps={textInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="name"
                label={translate('resources.radius/profiles.fields.name', { _: 'Policy Name' })}
                validate={[required(), minLength(2), maxLength(50)]}
                helperText={translate('resources.radius/profiles.helpers.name', { _: '2-50 characters' })}
                fullWidth
                size="small"
                inputProps={textInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="status"
                  label={translate('resources.radius/profiles.fields.status_enabled', { _: 'Enabled Status' })}
                  helperText={translate('resources.radius/profiles.helpers.status', { _: 'Whether this policy is enabled' })}
                />
              </Box>
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.rate_control.title', { _: 'Rate Control' })}
          description={translate('resources.radius/profiles.sections.rate_control.description', { _: 'Concurrent session limit and bandwidth rates' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <NumberInput
                source="active_num"
                label={translate('resources.radius/profiles.fields.active_num', { _: 'Sessions' })}
                min={0}
                placeholder="1"
                helperText={translate('resources.radius/profiles.helpers.active_num', { _: 'Maximum concurrent sessions allowed' })}
                fullWidth
                size="small"
                inputProps={numInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="up_rate"
                label={translate('resources.radius/profiles.fields.up_rate', { _: 'Upload Rate (Kbps)' })}
                min={0}
                placeholder="1024"
                helperText={translate('resources.radius/profiles.helpers.up_rate', { _: 'Upload bandwidth limit' })}
                fullWidth
                size="small"
                inputProps={numInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="down_rate"
                label={translate('resources.radius/profiles.fields.down_rate', { _: 'Download Rate (Kbps)' })}
                min={0}
                placeholder="1024"
                helperText={translate('resources.radius/profiles.helpers.down_rate', { _: 'Download bandwidth limit' })}
                fullWidth
                size="small"
                inputProps={numInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>

            <FieldGridItem>
              <NumberInput
                source="data_quota"
                label={translate('resources.radius/profiles.fields.data_quota', { _: 'Data Quota (MB)' })}
                min={0}
                placeholder="0"
                helperText={translate('resources.radius/profiles.helpers.data_quota', { _: 'Total data quota (MB)' })}
                fullWidth
                size="small"
                inputProps={numInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.network.title', { _: 'Network Configuration' })}
          description={translate('resources.radius/profiles.sections.network.description', { _: 'IP address pool and domain settings' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="addr_pool"
                label={translate('resources.radius/profiles.fields.addr_pool', { _: 'Address Pool' })}
                helperText={translate('resources.radius/profiles.helpers.addr_pool', { _: 'RADIUS Address Pool Name' })}
                fullWidth
                size="small"
                inputProps={textInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="ipv6_prefix"
                label={translate('resources.radius/profiles.fields.ipv6_prefix', { _: 'IPv6 Prefix' })}
                helperText={translate('resources.radius/profiles.helpers.ipv6_prefix', { _: 'IPv6 prefix delegation' })}
                fullWidth
                size="small"
                inputProps={textInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="domain"
                label={translate('resources.radius/profiles.fields.domain', { _: 'Domain' })}
                helperText={translate('resources.radius/profiles.helpers.domain', { _: 'User authentication domain' })}
                fullWidth
                size="small"
                inputProps={textInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.binding.title', { _: 'Binding Policy' })}
          description={translate('resources.radius/profiles.sections.binding.description', { _: 'MAC and VLAN binding configuration' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="bind_mac"
                  label={translate('resources.radius/profiles.fields.bind_mac', { _: 'Bind MAC' })}
                  helperText={translate('resources.radius/profiles.helpers.bind_mac', { _: 'Whether to enable MAC binding' })}
                />
              </Box>
            </FieldGridItem>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="bind_vlan"
                  label={translate('resources.radius/profiles.fields.bind_vlan', { _: 'Bind VLAN' })}
                  helperText={translate('resources.radius/profiles.helpers.bind_vlan', { _: 'Whether to enable VLAN binding' })}
                />
              </Box>
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.remark.title', { _: 'Remarks' })}
          description={translate('resources.radius/profiles.sections.remark.description', { _: 'Additional notes and comments' })}
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput
                source="remark"
                label={translate('resources.radius/profiles.fields.remark', { _: 'Remark' })}
                multiline
                minRows={3}
                fullWidth
                size="small"
                inputProps={textInputProps}
                InputLabelProps={inputLabelProps}
                helperText={translate('resources.radius/profiles.helpers.remark', { _: 'Optional remark' })}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};

// ============ 创建页面 ============

export const RadiusProfileCreate = () => {
  const translate = useTranslate();
  const locale = useLocale();
  const isRTL = locale === 'ar';

  const inputLabelProps = {
    sx: {
      transformOrigin: isRTL ? 'top right' : 'top left',
      left: isRTL ? 'auto' : 0,
      right: isRTL ? 24 : 'auto',
    }
  };

  const textInputProps = {
    style: { textAlign: (isRTL ? 'right' : 'left') as any },
    dir: isRTL ? 'rtl' : 'ltr'
  };

  const numInputProps = {
    style: { textAlign: (isRTL ? 'right' : 'left') as any, direction: isRTL ? 'rtl' : 'ltr' as any }
  };

  return (
    <Create>
      <SimpleForm sx={{ ...formLayoutSx, direction: isRTL ? 'rtl' : 'ltr' }}>
        <FormSection
          title={translate('resources.radius/profiles.sections.basic.title', { _: 'Basic Information' })}
          description={translate('resources.radius/profiles.sections.basic.description', { _: 'Basic configuration of the policy' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="name"
                label={translate('resources.radius/profiles.fields.name', { _: 'Policy Name' })}
                validate={[required(), minLength(2), maxLength(50)]}
                helperText={translate('resources.radius/profiles.helpers.name', { _: '2-50 characters' })}
                fullWidth
                size="small"
                inputProps={textInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="status"
                  label={translate('resources.radius/profiles.fields.status_enabled', { _: 'Enabled Status' })}
                  defaultValue={true}
                  helperText={translate('resources.radius/profiles.helpers.status', { _: 'Whether this policy is enabled' })}
                />
              </Box>
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.rate_control.title', { _: 'Rate Control' })}
          description={translate('resources.radius/profiles.sections.rate_control.description', { _: 'Concurrent session limit and bandwidth rates' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <NumberInput
                source="active_num"
                label={translate('resources.radius/profiles.fields.active_num', { _: 'Sessions' })}
                min={0}
                placeholder="1"
                helperText={translate('resources.radius/profiles.helpers.active_num', { _: 'Maximum concurrent sessions allowed' })}
                fullWidth
                size="small"
                inputProps={numInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="up_rate"
                label={translate('resources.radius/profiles.fields.up_rate', { _: 'Upload Rate (Kbps)' })}
                min={0}
                placeholder="1024"
                helperText={translate('resources.radius/profiles.helpers.up_rate', { _: 'Upload bandwidth limit' })}
                fullWidth
                size="small"
                inputProps={numInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="down_rate"
                label={translate('resources.radius/profiles.fields.down_rate', { _: 'Download Rate (Kbps)' })}
                min={0}
                placeholder="1024"
                helperText={translate('resources.radius/profiles.helpers.down_rate', { _: 'Download bandwidth limit' })}
                fullWidth
                size="small"
                inputProps={numInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>

            <FieldGridItem>
              <NumberInput
                source="data_quota"
                label={translate('resources.radius/profiles.fields.data_quota', { _: 'Data Quota (MB)' })}
                min={0}
                placeholder="0"
                helperText={translate('resources.radius/profiles.helpers.data_quota', { _: 'Total data quota (MB)' })}
                fullWidth
                size="small"
                inputProps={numInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.network.title', { _: 'Network Configuration' })}
          description={translate('resources.radius/profiles.sections.network.description', { _: 'IP address pool and domain settings' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="addr_pool"
                label={translate('resources.radius/profiles.fields.addr_pool', { _: 'Address Pool' })}
                helperText={translate('resources.radius/profiles.helpers.addr_pool', { _: 'RADIUS Address Pool Name' })}
                fullWidth
                size="small"
                inputProps={textInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="ipv6_prefix"
                label={translate('resources.radius/profiles.fields.ipv6_prefix', { _: 'IPv6 Prefix' })}
                helperText={translate('resources.radius/profiles.helpers.ipv6_prefix', { _: 'IPv6 prefix delegation' })}
                fullWidth
                size="small"
                inputProps={textInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="domain"
                label={translate('resources.radius/profiles.fields.domain', { _: 'Domain' })}
                helperText={translate('resources.radius/profiles.helpers.domain', { _: 'User authentication domain' })}
                fullWidth
                size="small"
                inputProps={textInputProps}
                InputLabelProps={inputLabelProps}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.binding.title', { _: 'Binding Policy' })}
          description={translate('resources.radius/profiles.sections.binding.description', { _: 'MAC and VLAN binding configuration' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="bind_mac"
                  label={translate('resources.radius/profiles.fields.bind_mac', { _: 'Bind MAC' })}
                  defaultValue={false}
                  helperText={translate('resources.radius/profiles.helpers.bind_mac', { _: 'Whether to enable MAC binding' })}
                />
              </Box>
            </FieldGridItem>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="bind_vlan"
                  label={translate('resources.radius/profiles.fields.bind_vlan', { _: 'Bind VLAN' })}
                  defaultValue={false}
                  helperText={translate('resources.radius/profiles.helpers.bind_vlan', { _: 'Whether to enable VLAN binding' })}
                />
              </Box>
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.remark.title', { _: 'Remarks' })}
          description={translate('resources.radius/profiles.sections.remark.description', { _: 'Additional notes and comments' })}
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput
                source="remark"
                label={translate('resources.radius/profiles.fields.remark', { _: 'Remark' })}
                multiline
                minRows={3}
                fullWidth
                size="small"
                inputProps={textInputProps}
                InputLabelProps={inputLabelProps}
                helperText={translate('resources.radius/profiles.helpers.remark', { _: 'Optional remark' })}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Create>
  );
};

// ============ 详情页顶部概览卡片 ============

const ProfileHeaderCard = () => {
  const record = useRecordContext<RadiusProfile>();
  const translate = useTranslate();
  const notify = useNotify();
  const refresh = useRefresh();

  const handleCopy = useCallback((text: string, label: string) => {
    navigator.clipboard.writeText(text);
    notify(translate('resources.radius/profiles.copied', { label, _: '%{label} copied to clipboard' }), { type: 'info' });
  }, [notify, translate]);

  const handleRefresh = useCallback(() => {
    refresh();
    notify(translate('resources.radius/profiles.data_refreshed', { _: 'Data refreshed' }), { type: 'info' });
  }, [refresh, notify, translate]);

  if (!record) return null;

  const isEnabled = record.status === 'enabled';

  return (
    <Card
      elevation={0}
      sx={{
        borderRadius: 4,
        background: theme =>
          theme.palette.mode === 'dark'
            ? isEnabled
              ? `linear-gradient(135deg, ${alpha(theme.palette.primary.dark, 0.4)} 0%, ${alpha(theme.palette.info.dark, 0.3)} 100%)`
              : `linear-gradient(135deg, ${alpha(theme.palette.grey[800], 0.5)} 0%, ${alpha(theme.palette.grey[700], 0.3)} 100%)`
            : isEnabled
              ? `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.1)} 0%, ${alpha(theme.palette.info.main, 0.08)} 100%)`
              : `linear-gradient(135deg, ${alpha(theme.palette.grey[400], 0.15)} 0%, ${alpha(theme.palette.grey[300], 0.1)} 100%)`,
        border: theme => `1px solid ${alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.2)}`,
        overflow: 'hidden',
        position: 'relative',
      }}
    >
      {/* 装饰背景 */}
      <Box
        sx={{
          position: 'absolute',
          top: -50,
          right: -50,
          width: 200,
          height: 200,
          borderRadius: '50%',
          background: theme => alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.1),
          pointerEvents: 'none',
        }}
      />

      <CardContent sx={{ p: 3, position: 'relative', zIndex: 1 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 3 }}>
          {/* 左侧：策略信息 */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Avatar
              sx={{
                width: 64,
                height: 64,
                bgcolor: isEnabled ? 'primary.main' : 'grey.500',
                fontSize: '1.5rem',
                fontWeight: 700,
                boxShadow: theme => `0 4px 14px ${alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.4)}`,
              }}
            >
              {record.name?.charAt(0).toUpperCase() || 'P'}
            </Avatar>
            <Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
                  {record.name || <EmptyValue message={translate('resources.radius/profiles.fields.unknown_policy', { _: 'Unknown Policy' })} />}
                </Typography>
                <StatusIndicator isEnabled={isEnabled} />
              </Box>
              {record.name && (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                  <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                    ID: {record.id}
                  </Typography>
                  <Tooltip title={translate('resources.radius/profiles.copy_name', { _: 'Copy Policy Name' })}>
                    <IconButton
                      size="small"
                      onClick={() => handleCopy(record.name!, translate('resources.radius/profiles.fields.name', { _: 'Policy Name' }))}
                      sx={{ p: 0.5 }}
                    >
                      <CopyIcon sx={{ fontSize: '0.75rem' }} />
                    </IconButton>
                  </Tooltip>
                </Box>
              )}
            </Box>
          </Box>

          {/* 右侧：操作按钮 */}
          <Box className="no-print" sx={{ display: 'flex', gap: 1 }}>
            <Tooltip title={translate('resources.radius/profiles.print_details', { _: 'Print Details' })}>
              <IconButton
                onClick={() => window.print()}
                sx={{
                  bgcolor: theme => alpha(theme.palette.info.main, 0.1),
                  '&:hover': {
                    bgcolor: theme => alpha(theme.palette.info.main, 0.2),
                  },
                }}
              >
                <PrintIcon />
              </IconButton>
            </Tooltip>
            <Tooltip title={translate('resources.radius/profiles.refresh_data', { _: 'Refresh Data' })}>
              <IconButton
                onClick={handleRefresh}
                sx={{
                  bgcolor: theme => alpha(theme.palette.primary.main, 0.1),
                  '&:hover': {
                    bgcolor: theme => alpha(theme.palette.primary.main, 0.2),
                  },
                }}
              >
                <RefreshIcon />
              </IconButton>
            </Tooltip>
            <ListButton
              label=""
              icon={<BackIcon />}
              sx={{
                minWidth: 'auto',
                bgcolor: theme => alpha(theme.palette.grey[500], 0.1),
                '&:hover': {
                  bgcolor: theme => alpha(theme.palette.grey[500], 0.2),
                },
              }}
            />
          </Box>
        </Box>

        {/* 快速统计 */}
        <Box
          sx={{
            display: 'grid',
            gap: 2,
            gridTemplateColumns: {
              xs: 'repeat(2, 1fr)',
              sm: 'repeat(4, 1fr)',
            },
          }}
        >
          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <SpeedIcon sx={{ fontSize: '1.1rem', color: 'info.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.radius/profiles.fields.active_num', { _: 'Sessions' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.active_num || 0}
            </Typography>
          </Box>

          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <SpeedIcon sx={{ fontSize: '1.1rem', color: 'success.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.radius/profiles.fields.up_rate', { _: 'Upload Rate' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600, fontFamily: 'monospace' }}>
              {formatRate(translate, record.up_rate)}
            </Typography>
          </Box>

          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <SpeedIcon sx={{ fontSize: '1.1rem', color: 'warning.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.radius/profiles.fields.down_rate', { _: 'Download Rate' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600, fontFamily: 'monospace' }}>
              {formatRate(translate, record.down_rate)}
            </Typography>
          </Box>

          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <NetworkIcon sx={{ fontSize: '1.1rem', color: 'primary.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.radius/profiles.fields.addr_pool', { _: 'Address Pool' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.addr_pool || '-'}
            </Typography>
          </Box>

          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <SpeedIcon sx={{ fontSize: '1.1rem', color: 'error.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.radius/profiles.fields.data_quota', { _: 'Data Quota' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600, fontFamily: 'monospace' }}>
              {formatQuota(translate, record.data_quota)}
            </Typography>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

// 打印样式
const printStyles = `
  @media print {
    body * {
      visibility: hidden;
    }
    .printable-content, .printable-content * {
      visibility: visible;
    }
    .printable-content {
      position: absolute;
      left: 0;
      top: 0;
      width: 100%;
      padding: 20px !important;
    }
    .no-print {
      display: none !important;
    }
  }
`;

// ============ 策略详情内容 ============

const ProfileDetails = () => {
  const record = useRecordContext<RadiusProfile>();
  const translate = useTranslate();

  if (!record) {
    return null;
  }

  return (
    <>
      <style>{printStyles}</style>
      <Box className="printable-content" sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 } }}>
        <Stack spacing={3}>
          {/* 顶部概览卡片 */}
          <ProfileHeaderCard />

          {/* 网络配置 */}
          <DetailSectionCard
            title={translate('resources.radius/profiles.sections.network.title', { _: 'Network Configuration' })}
            description={translate('resources.radius/profiles.sections.network.description', { _: 'IPv6 and Domain Configuration' })}
            icon={<NetworkIcon />}
            color="success"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.radius/profiles.fields.ipv6_prefix', { _: 'IPv6 Prefix' })}
                value={record.ipv6_prefix || <EmptyValue />}
              />
              <DetailItem
                label={translate('resources.radius/profiles.fields.domain', { _: 'Domain' })}
                value={record.domain || <EmptyValue />}
              />
            </Box>
          </DetailSectionCard>

          {/* 绑定策略 */}
          <DetailSectionCard
            title={translate('resources.radius/profiles.sections.binding.title', { _: 'Binding Policy' })}
            description={translate('resources.radius/profiles.sections.binding.description', { _: 'MAC and VLAN Binding Configuration' })}
            icon={<BindingIcon />}
            color="warning"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.radius/profiles.fields.bind_mac', { _: 'Bind MAC' })}
                value={<BooleanChip value={record.bind_mac} />}
              />
              <DetailItem
                label={translate('resources.radius/profiles.fields.bind_vlan', { _: 'Bind VLAN' })}
                value={<BooleanChip value={record.bind_vlan} />}
              />
            </Box>
          </DetailSectionCard>

          {/* 时间信息 */}
          <DetailSectionCard
            title={translate('resources.radius/profiles.sections.timestamps.title', { _: 'Time Information' })}
            description={translate('resources.radius/profiles.sections.timestamps.description', { _: 'Creation and Update Time' })}
            icon={<TimeIcon />}
            color="info"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.radius/profiles.fields.created_at', { _: 'Created At' })}
                value={formatTimestamp(record.created_at)}
              />
              <DetailItem
                label={translate('resources.radius/profiles.fields.updated_at', { _: 'Updated At' })}
                value={formatTimestamp(record.updated_at)}
              />
            </Box>
          </DetailSectionCard>

          {/* 备注信息 */}
          <DetailSectionCard
            title={translate('resources.radius/profiles.sections.remark.title', { _: 'Remarks' })}
            description={translate('resources.radius/profiles.sections.remark.description', { _: 'Additional notes and comments' })}
            icon={<NoteIcon />}
            color="primary"
          >
            <Box
              sx={{
                p: 2,
                borderRadius: 2,
                bgcolor: theme =>
                  theme.palette.mode === 'dark'
                    ? 'rgba(255, 255, 255, 0.02)'
                    : 'rgba(0, 0, 0, 0.02)',
                border: theme => `1px solid ${theme.palette.divider}`,
                minHeight: 80,
              }}
            >
              <Typography
                variant="body2"
                sx={{
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  color: record.remark ? 'text.primary' : 'text.disabled',
                  fontStyle: record.remark ? 'normal' : 'italic',
                }}
              >
                {record.remark || translate('resources.radius/profiles.empty.no_remark', { _: 'No remark information' })}
              </Typography>
            </Box>
          </DetailSectionCard>
        </Stack>
      </Box>
    </>
  );
};

// RADIUS 计费策略详情
export const RadiusProfileShow = () => {
  return (
    <Show>
      <ProfileDetails />
    </Show>
  );
};
