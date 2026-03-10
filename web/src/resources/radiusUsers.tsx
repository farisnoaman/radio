import { httpClient } from '../utils/apiClient';
import {
  List,
  Datagrid,
  TextField,
  DateField,
  Edit,
  TextInput,
  SelectInput,
  Create,
  Show,
  EmailField,
  BooleanInput,
  required,
  minLength,
  maxLength,
  email,
  useRecordContext,
  Toolbar,
  SaveButton,
  DeleteButton,
  SimpleForm,
  ToolbarProps,
  ReferenceInput,
  ReferenceField,
  TopToolbar,
  ListButton,
  CreateButton,
  ExportButton,
  useTranslate,
  useRefresh,
  useNotify,
  useListContext,
  SortButton,
  RaRecord,
  FunctionField,
  EditButton,
  Button,
  RecordContextProvider,
} from 'react-admin';
import UserAnonymizeDialog from '../components/UserAnonymizeDialog';
import NoAccountsIcon from '@mui/icons-material/NoAccounts';
import {
  Box,
  Chip,
  Card,
  CardContent,
  CardActions,
  Typography,
  IconButton,
  Tooltip,
  useMediaQuery,
  Paper,
  alpha,
  useTheme,
  Stack,
  Avatar,
  Skeleton,
  TextField as MuiTextField
} from '@mui/material';
import { Theme } from '@mui/material/styles';
import { ReactNode, useMemo, useCallback, useState, useEffect } from 'react';
import {
  Person as PersonIcon,
  ContactPhone as ContactIcon,
  Settings as SettingsIcon,
  Wifi as NetworkIcon,
  Schedule as TimeIcon,
  Note as NoteIcon,
  CheckCircle as EnabledIcon,
  Cancel as DisabledIcon,
  ContentCopy as CopyIcon,
  Refresh as RefreshIcon,
  ArrowBack as BackIcon,
  Print as PrintIcon,
  FilterList as FilterIcon,
  Search as SearchIcon,
  Clear as ClearIcon,
  Email as EmailIcon,
  Phone as PhoneIcon,
  CalendarToday as CalendarIcon,
  ReceiptLong as ReceiptLongIcon
} from '@mui/icons-material';
import { ServerPagination, ActiveFilters } from '../components';

const LARGE_LIST_PER_PAGE = 50;

// ============ 类型定义 ============

interface RadiusUser extends RaRecord {
  username?: string;
  password?: string;
  realname?: string;
  email?: string;
  mobile?: string;
  address?: string;
  status?: 'enabled' | 'disabled';
  profile_id?: string | number;
  expire_time?: string;
  ip_addr?: string;
  ipv6_addr?: string;
  billing_type?: 'prepaid' | 'postpaid';
  subscription_status?: 'active' | 'suspended' | 'canceled';
  next_billing_date?: string;
  monthly_fee?: number;
  price_per_gb?: number;
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

const formatExpireTime = (expireTime?: string): { text: string; color: 'success' | 'warning' | 'error' | 'default' } => {
  if (!expireTime) {
    return { text: 'Never expires', color: 'success' };
  }
  const expireDate = new Date(expireTime);
  const now = new Date();
  const diffDays = Math.ceil((expireDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));

  if (diffDays < 0) {
    return { text: `Expired ${Math.abs(diffDays)} days ago`, color: 'error' };
  } else if (diffDays <= 7) {
    return { text: `Expires in ${diffDays} days`, color: 'warning' };
  } else if (diffDays <= 30) {
    return { text: `Expires in ${diffDays} days`, color: 'default' };
  }
  return { text: expireDate.toLocaleDateString(), color: 'success' };
};

// ============ 美化的详情组件 ============

interface DetailItemProps {
  label: string;
  value?: ReactNode;
  highlight?: boolean;
}

const DetailItem = ({ label, value, highlight = false }: DetailItemProps) => (
  <Box
    sx={{
      display: 'flex',
      flexDirection: 'column',
      gap: 0.5,
      p: 1.5,
      borderRadius: 1.5,
      backgroundColor: theme =>
        highlight
          ? alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.15 : 0.06)
          : theme.palette.mode === 'dark'
            ? 'rgba(255, 255, 255, 0.02)'
            : 'rgba(0, 0, 0, 0.02)',
      border: theme =>
        highlight
          ? `1px solid ${alpha(theme.palette.primary.main, 0.3)}`
          : `1px solid ${theme.palette.divider}`,
      transition: 'all 0.2s ease',
      '&:hover': {
        backgroundColor: theme =>
          highlight
            ? alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.2 : 0.08)
            : theme.palette.mode === 'dark'
              ? 'rgba(255, 255, 255, 0.04)'
              : 'rgba(0, 0, 0, 0.03)',
      },
    }}
  >
    <Typography
      variant="caption"
      sx={{
        color: 'text.secondary',
        fontWeight: 500,
        fontSize: '0.85rem',
        textTransform: 'uppercase',
        letterSpacing: '0.5px',
      }}
    >
      {label}
    </Typography>
    <Typography
      variant="body2"
      sx={{
        fontWeight: highlight ? 600 : 500,
        color: highlight ? 'primary.main' : 'text.primary',
        wordBreak: 'break-word',
        fontSize: '1rem',
        lineHeight: 1.5,
      }}
    >
      {value ?? <span style={{ color: 'inherit', opacity: 0.4 }}>-</span>}
    </Typography>
  </Box>
);

const GenerateBillButton = ({ username }: { username: string }) => {
  const translate = useTranslate();
  const notify = useNotify();
  const refresh = useRefresh();
  const [loading, setLoading] = useState(false);

  const handleClick = async () => {
    setLoading(true);
    try {
      await httpClient(`/radius/users/${username}/bill`, { method: 'POST' });
      notify('resources.radius/users.notifications.bill_generated', { type: 'success' });
      refresh();
    } catch (error) {
      notify('resources.radius/users.notifications.bill_failed', { type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Button
      label={translate('resources.radius/users.actions.generate_bill', { _: 'Generate Bill' })}
      onClick={handleClick}
      disabled={loading}
      variant="contained"
      color="secondary"
      sx={{ mt: 1 }}
    >
      <ReceiptLongIcon />
    </Button>
  );
};

interface DetailSectionCardProps {
  title: string;
  description?: string;
  icon: ReactNode;
  children: ReactNode;
  color?: 'primary' | 'success' | 'warning' | 'info' | 'error';
}

const DetailSectionCard = ({
  title,
  description,
  icon,
  children,
  color = 'primary',
}: DetailSectionCardProps) => (
  <Card
    elevation={0}
    sx={{
      borderRadius: 3,
      border: theme => `1px solid ${theme.palette.divider}`,
      overflow: 'hidden',
      transition: 'all 0.2s ease',
      '&:hover': {
        boxShadow: theme =>
          theme.palette.mode === 'dark'
            ? '0 4px 20px rgba(0, 0, 0, 0.3)'
            : '0 4px 20px rgba(0, 0, 0, 0.08)',
      },
    }}
  >
    <Box
      sx={{
        px: { xs: 1.5, sm: 2.5 },
        py: { xs: 1, sm: 2 },
        backgroundColor: theme =>
          alpha(
            theme.palette[color].main,
            theme.palette.mode === 'dark' ? 0.15 : 0.06
          ),
        borderBottom: theme =>
          `1px solid ${alpha(theme.palette[color].main, 0.2)}`,
      }}
    >
      <Box sx={{ display: 'flex', alignItems: 'center', gap: { xs: 1, sm: 1.5 } }}>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: { xs: 32, sm: 36 },
            height: { xs: 32, sm: 36 },
            borderRadius: 2,
            backgroundColor: theme =>
              alpha(theme.palette[color].main, theme.palette.mode === 'dark' ? 0.3 : 0.15),
            color: `${color}.main`,
          }}
        >
          {icon}
        </Box>
        <Box>
          <Typography
            variant="subtitle1"
            sx={{
              fontWeight: 600,
              color: `${color}.main`,
              fontSize: { xs: '1rem', sm: '1.1rem' },
            }}
          >
            {title}
          </Typography>
          {description && (
            <Typography
              variant="body2"
              sx={{
                color: 'text.secondary',
                fontSize: '0.9rem',
                mt: 0.25,
                display: { xs: 'none', sm: 'block' }
              }}
            >
              {description}
            </Typography>
          )}
        </Box>
      </Box>
    </Box>
    <CardContent sx={{ p: { xs: 1.5, sm: 2.5 } }}>{children}</CardContent>
  </Card>
);

// 空态组件
interface EmptyStateProps {
  message?: string;
}

const EmptyValue = ({ message = 'No data' }: EmptyStateProps) => (
  <Box
    sx={{
      display: 'flex',
      alignItems: 'center',
      gap: 0.5,
      color: 'text.disabled',
      fontStyle: 'italic',
      fontSize: '0.85rem',
    }}
  >
    <Typography variant="body2" sx={{ opacity: 0.6 }}>
      {message}
    </Typography>
  </Box>
);

// ============ 表单组件 ============

interface FormSectionProps {
  title: string;
  description?: string;
  children: ReactNode;
}

const FormSection = ({ title, description, children }: FormSectionProps) => (
  <Paper
    elevation={0}
    sx={{
      p: 3,
      mb: 3,
      borderRadius: 2,
      border: theme => `1px solid ${theme.palette.divider}`,
      backgroundColor: theme => theme.palette.background.paper,
      width: '100%'
    }}
  >
    <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
      {title}
    </Typography>
    {description && (
      <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, mb: 1 }}>
        {description}
      </Typography>
    )}
    <Box sx={{ mt: 2, width: '100%' }}>
      {children}
    </Box>
  </Paper>
);

type ColumnConfig = {
  xs?: number;
  sm?: number;
  md?: number;
  lg?: number;
  xl?: number;
};

interface FieldGridProps {
  children: ReactNode;
  columns?: ColumnConfig;
  gap?: number;
}

const defaultColumns: Required<Pick<ColumnConfig, 'xs' | 'sm' | 'md' | 'lg'>> = {
  xs: 1,
  sm: 2,
  md: 3,
  lg: 3
};

const FieldGrid = ({
  children,
  columns = {},
  gap = 2
}: FieldGridProps) => {
  const resolved = {
    xs: columns.xs ?? defaultColumns.xs,
    sm: columns.sm ?? defaultColumns.sm,
    md: columns.md ?? defaultColumns.md,
    lg: columns.lg ?? defaultColumns.lg
  };

  return (
    <Box
      sx={{
        display: 'grid',
        gap,
        width: '100%',
        alignItems: 'stretch',
        justifyItems: 'stretch',
        gridTemplateColumns: {
          xs: `repeat(${resolved.xs}, minmax(0, 1fr))`,
          sm: `repeat(${resolved.sm}, minmax(0, 1fr))`,
          md: `repeat(${resolved.md}, minmax(0, 1fr))`,
          lg: `repeat(${resolved.lg}, minmax(0, 1fr))`
        }
      }}
    >
      {children}
    </Box>
  );
};

interface FieldGridItemProps {
  children: ReactNode;
  span?: ColumnConfig;
}

const FieldGridItem = ({
  children,
  span = {}
}: FieldGridItemProps) => {
  const resolved = {
    xs: span.xs ?? 1,
    sm: span.sm ?? span.xs ?? 1,
    md: span.md ?? span.sm ?? span.xs ?? 1,
    lg: span.lg ?? span.md ?? span.sm ?? span.xs ?? 1
  };

  return (
    <Box
      sx={{
        width: '100%',
        gridColumn: {
          xs: `span ${resolved.xs}`,
          sm: `span ${resolved.sm}`,
          md: `span ${resolved.md}`,
          lg: `span ${resolved.lg}`
        }
      }}
    >
      {children}
    </Box>
  );
};

const controlWrapperSx = {
  border: (theme: Theme) => `1px solid ${theme.palette.divider}`,
  borderRadius: 2,
  px: 2,
  py: 1.5,
  height: '100%',
  display: 'flex',
  alignItems: 'center',
  '& .MuiFormControl-root': {
    width: '100%',
    margin: 0
  },
  '& .MuiFormControlLabel-root': {
    margin: 0,
    width: '100%'
  }
};

const formLayoutSx = {
  width: '100%',
  maxWidth: 'none',
  mx: 0,
  px: { xs: 1, sm: 2, md: 2 },
  '& .RaSimpleForm-main': {
    width: '100%',
    maxWidth: 'none',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-start'
  },
  '& .RaSimpleForm-content': {
    width: '100%',
    maxWidth: 'none',
    px: 0
  },
  '& .RaSimpleForm-form': {
    width: '100%',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-start'
  },
  '& .RaSimpleForm-input': {
    width: '100%'
  }
};

// 简化后的自定义工具栏（仅展示保存与删除）
const UserFormToolbar = (props: ToolbarProps) => (
  <Toolbar {...props}>
    <SaveButton />
    <DeleteButton mutationMode="pessimistic" />
  </Toolbar>
);

// ============ 列表加载骨架屏 ============

const RadiusUserListSkeleton = ({ rows = 10 }: { rows?: number }) => (
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
              md: 'repeat(3, 1fr)',
              lg: 'repeat(6, 1fr)',
            },
          }}
        >
          {[...Array(6)].map((_, i) => (
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
          gridTemplateColumns: 'repeat(9, 1fr)',
          gap: 1,
          p: 2,
          bgcolor: theme =>
            theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.02)',
          borderBottom: theme => `1px solid ${theme.palette.divider}`,
        }}
      >
        {[...Array(9)].map((_, i) => (
          <Skeleton key={i} variant="text" height={20} width="80%" />
        ))}
      </Box>

      {/* 表格行 */}
      {[...Array(rows)].map((_, rowIndex) => (
        <Box
          key={rowIndex}
          sx={{
            display: 'grid',
            gridTemplateColumns: 'repeat(9, 1fr)',
            gap: 1,
            p: 2,
            borderBottom: theme => `1px solid ${theme.palette.divider}`,
          }}
        >
          {[...Array(9)].map((_, colIndex) => (
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

const UserEmptyListState = () => {
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
      <PersonIcon sx={{ fontSize: 64, opacity: 0.3, mb: 2 }} />
      <Typography variant="h6" sx={{ opacity: 0.6, mb: 1 }}>
        {translate('resources.radius/users.empty.title', { _: '暂无用户' })}
      </Typography>
      <Typography variant="body2" sx={{ opacity: 0.5 }}>
        {translate('resources.radius/users.empty.description', { _: '点击"新建"按钮添加第一个RADIUS用户' })}
      </Typography>
    </Box>
  );
};

// ============ 搜索表头区块组件 ============

const UserSearchHeaderCard = () => {
  const translate = useTranslate();
  const { filterValues, setFilters, displayedFilters } = useListContext();
  const [localFilters, setLocalFilters] = useState<Record<string, string>>({});

  // 同步外部筛选值到本地状态
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
    { key: 'username', label: translate('resources.radius/users.fields.username', { _: '用户名' }) },
    { key: 'realname', label: translate('resources.radius/users.fields.realname', { _: '真实姓名' }) },
    { key: 'email', label: translate('resources.radius/users.fields.email', { _: '邮箱' }) },
    { key: 'mobile', label: translate('resources.radius/users.fields.mobile', { _: '手机号' }) },
    { key: 'ip_addr', label: translate('resources.radius/users.fields.ip_addr', { _: 'IP地址' }) },
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
          {translate('resources.radius/users.filter.title', { _: '筛选条件' })}
        </Typography>
      </Box>

      <CardContent sx={{ p: 2 }}>
        <Box
          sx={{
            display: 'grid',
            gap: 1.5,
            gridTemplateColumns: {
              xs: 'repeat(2, 1fr)',
              sm: 'repeat(3, 1fr)',
              md: 'repeat(4, 1fr)',
              lg: 'repeat(6, 1fr)',
            },
            alignItems: 'end',
          }}
        >
          {/* 文本筛选字段 */}
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
              }}
            />
          ))}

          {/* 操作按钮 */}
          <Box sx={{ display: 'flex', gap: 0.5, justifyContent: 'flex-end' }}>
            <Tooltip title={translate('ra.action.clear_filters', { _: '清除筛选' })}>
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
            <Tooltip title={translate('ra.action.search', { _: '搜索' })}>
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

// ============ 状态指示器组件 ============

const StatusIndicator = ({ isEnabled }: { isEnabled: boolean }) => {
  const translate = useTranslate();
  return (
    <Chip
      icon={isEnabled ? <EnabledIcon sx={{ fontSize: '0.85rem !important' }} /> : <DisabledIcon sx={{ fontSize: '0.85rem !important' }} />}
      label={isEnabled ? translate('resources.radius/users.status.enabled', { _: '启用' }) : translate('resources.radius/users.status.disabled', { _: '禁用' })}
      size="small"
      color={isEnabled ? 'success' : 'default'}
      variant={isEnabled ? 'filled' : 'outlined'}
      sx={{ height: 22, fontWeight: 500, fontSize: '0.75rem' }}
    />
  );
};

// ============ 增强版 Datagrid 字段组件 ============

const UsernameField = () => {
  const record = useRecordContext<RadiusUser>();
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
        {record.username?.charAt(0).toUpperCase() || 'U'}
      </Avatar>
      <Box>
        <Typography
          variant="body2"
          sx={{ fontWeight: 600, color: 'text.primary', lineHeight: 1.3 }}
        >
          {record.username || '-'}
        </Typography>
        <StatusIndicator isEnabled={isEnabled} />
      </Box>
    </Box>
  );
};

const ExpireTimeField = () => {
  const record = useRecordContext<RadiusUser>();
  if (!record) return null;

  const expireInfo = formatExpireTime(record.expire_time);

  return (
    <Chip
      label={expireInfo.text}
      size="small"
      color={expireInfo.color}
      variant="outlined"
      sx={{ fontWeight: 500, fontSize: '0.75rem' }}
    />
  );
};

const IpAddressField = () => {
  const record = useRecordContext<RadiusUser>();
  if (!record?.ip_addr) return <Typography variant="body2" color="text.secondary">-</Typography>;

  return (
    <Chip
      label={record.ip_addr}
      size="small"
      color="info"
      variant="outlined"
      sx={{ fontFamily: 'monospace', fontSize: '0.8rem', height: 24 }}
    />
  );
};

// ============ 列表操作栏组件 ============

const UserListActions = () => {
  const translate = useTranslate();
  return (
    <TopToolbar>
      <SortButton
        fields={['created_at', 'expire_time', 'username']}
        label={translate('ra.action.sort', { _: '排序' })}
      />
      <CreateButton />
      <ExportButton />
    </TopToolbar>
  );
};

// ============ 内部列表内容组件 ============

const AnonymizeButton = () => {
  const record = useRecordContext<RadiusUser>();
  const [open, setOpen] = useState(false);

  if (!record) return null;

  const handleOpen = (e: any) => {
    e.stopPropagation();
    setOpen(true);
  };

  return (
    <>
      <Tooltip title="Anonymize User (GDPR)">
        <IconButton onClick={handleOpen} size="small" color="default">
          <NoAccountsIcon fontSize="small" />
        </IconButton>
      </Tooltip>
      {open && (
        <UserAnonymizeDialog
          open={open}
          onClose={() => setOpen(false)}
          username={record.username || ''}
        />
      )}
    </>
  );
};

// ============ Mobile Card Grid Component ============

const UserCardGrid = () => {
  const { data, isLoading } = useListContext<RadiusUser>();
  const theme = useTheme();
  
  if (isLoading || !data) return null;

  return (
    <Box 
      display="grid" 
      gridTemplateColumns={{ xs: '1fr', sm: '1fr', md: 'repeat(2, 1fr)', lg: 'repeat(3, 1fr)' }} 
      gap={2} 
      p={0}
      sx={{ bgcolor: theme.palette.mode === 'dark' ? 'transparent' : 'rgba(0,0,0,0.02)' }}
    >
      {data.map(record => (
        <RecordContextProvider value={record} key={record.id}>
            <Card 
              elevation={2} 
              sx={{ 
                borderRadius: 2.5,
                transition: 'all 0.3s ease',
                cursor: 'pointer',
                '&:hover': { 
                  transform: 'translateY(-4px)',
                  boxShadow: 6,
                  bgcolor: theme => theme.palette.mode === 'dark' 
                    ? 'rgba(255,255,255,0.08)' 
                    : 'rgba(25, 118, 210, 0.04)'
                }
              }}
              onClick={() => {
                window.location.href = `#/radius/users/${record.id}/show`;
              }}
          >
            <CardContent sx={{ pb: 1, pt: 1.5, px: 2 }}>
              {/* Header: Avatar, Username, Status */}
              <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={1.5}>
                <Box display="flex" alignItems="center" gap={1.5}>
                  <Avatar 
                    sx={{ 
                      bgcolor: record.status === 'enabled' ? 'primary.main' : 'grey.400', 
                      width: 44, 
                      height: 44, 
                      fontSize: '1rem', 
                      fontWeight: 'bold' 
                    }}
                  >
                    {record.username?.charAt(0).toUpperCase() || 'U'}
                  </Avatar>
                  <Box minWidth={0}>
                    <Typography 
                      variant="subtitle1" 
                      component="div" 
                      sx={{ 
                        fontWeight: 700, 
                        lineHeight: 1.2,
                        overflow: 'hidden', 
                        textOverflow: 'ellipsis', 
                        whiteSpace: 'nowrap',
                        maxWidth: 150
                      }}
                    >
                      {record.username || '-'}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      #{String(record.id).slice(-6)}
                    </Typography>
                  </Box>
                </Box>
                <StatusIndicator isEnabled={record.status === 'enabled'} />
              </Box>



              {/* Stats Row */}
              <Box display="flex" justifyContent="space-between" gap={1} mb={1}>
                <Box flex={1}>
                  <Typography variant="caption" color="text.secondary">Profile</Typography>
                  <Typography variant="body2" fontWeight="bold" noWrap>
                    {record.profile_id || '-'}
                  </Typography>
                </Box>
                <Box flex={1}>
                  <Typography variant="caption" color="text.secondary">Type</Typography>
                  <Typography variant="body2" fontWeight="bold" noWrap>
                    {record.billing_type || '-'}
                  </Typography>
                </Box>
                <Box flex={1}>
                  <Typography variant="caption" color="text.secondary">Expire</Typography>
                  <Typography variant="body2" noWrap>
                    {formatExpireTime(record.expire_time).text}
                  </Typography>
                </Box>
              </Box>
            </CardContent>

            {/* Actions */}
            <CardActions sx={{ 
              justifyContent: 'flex-end', 
              borderTop: theme => `1px solid ${theme.palette.divider}`, 
              px: 1.5, 
              py: 1,
              gap: 0.5,
              bgcolor: theme => theme.palette.mode === 'dark' ? 'rgba(0,0,0,0.2)' : 'rgba(0,0,0,0.02)'
            }}>
              <Box sx={{ display: 'flex', gap: 0.5, alignItems: 'center' }}>
                <AnonymizeButton />
                <EditButton label="" size="small" />
              </Box>
            </CardActions>
          </Card>
        </RecordContextProvider>
      ))}
    </Box>
  );
};

const RadiusUserListContent = () => {
  const translate = useTranslate();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { data, isLoading, total } = useListContext<RadiusUser>();

  // 活动筛选器标签配置
  const fieldLabels = useMemo(
    () => ({
      username: translate('resources.radius/users.fields.username', { _: '用户名' }),
      realname: translate('resources.radius/users.fields.realname', { _: '真实姓名' }),
      email: translate('resources.radius/users.fields.email', { _: '邮箱' }),
      mobile: translate('resources.radius/users.fields.mobile', { _: '手机号' }),
      ip_addr: translate('resources.radius/users.fields.ip_addr', { _: 'IP地址' }),
      status: translate('resources.radius/users.fields.status', { _: '状态' }),
    }),
    [translate],
  );

  const statusLabels = useMemo(
    () => ({
      enabled: translate('resources.radius/users.status.enabled', { _: '启用' }),
      disabled: translate('resources.radius/users.status.disabled', { _: '禁用' }),
    }),
    [translate],
  );

  if (isLoading) {
    return <RadiusUserListSkeleton />;
  }

  if (!data || data.length === 0) {
    return (
      <Box>
        <UserSearchHeaderCard />
        <Card
          elevation={0}
          sx={{
            borderRadius: 2,
            border: theme => `1px solid ${theme.palette.divider}`,
          }}
        >
          <UserEmptyListState />
        </Card>
      </Box>
    );
  }

  return (
    <Box>
      {/* 搜索区块 */}
      <UserSearchHeaderCard />

      {/* 活动筛选标签 */}
      <ActiveFilters fieldLabels={fieldLabels} valueLabels={{ status: statusLabels }} />

      {isMobile ? (
        /* Mobile Card Grid View - with p={0} */
        <Box p={0}>
          {/* Statistics bar */}
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
            <Typography variant="body2" color="text.secondary">
              {translate('resources.radius/users.total_users', { _: 'Total %{count} users', count: total?.toLocaleString() || 0 })}
            </Typography>
          </Box>
          <UserCardGrid />
        </Box>
      ) : (
        /* Desktop Table View */
        /* 表格容器 */
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
            <Typography variant="body2" color="text.secondary">
              {translate('resources.radius/users.total_users', { _: 'Total %{count} users', count: total?.toLocaleString() || 0 })}
            </Typography>
          </Box>

          {/* 响应式表格 */}
          <Box
            sx={{
              overflowX: 'auto',
              '& .RaDatagrid-root': {
                minWidth: isMobile ? 1000 : 'auto',
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
            <Datagrid rowClick="show" bulkActionButtons={false}>
              <FunctionField
                source="username"
                label={translate('resources.radius/users.fields.username', { _: '用户名' })}
                render={() => <UsernameField />}
              />
              <TextField
                source="realname"
                label={translate('resources.radius/users.fields.realname', { _: '真实姓名' })}
              />
              <EmailField
                source="email"
                label={translate('resources.radius/users.fields.email', { _: '邮箱' })}
              />
              <TextField
                source="mobile"
                label={translate('resources.radius/users.fields.mobile', { _: '手机号' })}
              />
              <FunctionField
                source="ip_addr"
                label={translate('resources.radius/users.fields.ip_addr', { _: 'IP地址' })}
                render={() => <IpAddressField />}
              />
              <ReferenceField
                source="profile_id"
                reference="radius/profiles"
                label={translate('resources.radius/users.fields.profile_id', { _: '计费策略' })}
              >
                <TextField source="name" />
              </ReferenceField>
              <FunctionField
                source="expire_time"
                label={translate('resources.radius/users.fields.expire_time', { _: '过期时间' })}
                render={() => <ExpireTimeField />}
              />
              <TextField
                source="billing_type"
                label={translate('resources.radius/users.fields.billing_type', { _: '计费类型' })}
              />
              <DateField
                source="created_at"
                label={translate('resources.radius/users.fields.created_at', { _: '创建时间' })}
                showTime
              />
              <FunctionField
                render={() => (
                  <Box sx={{ display: 'flex', gap: 1 }}>
                    <EditButton />
                    <AnonymizeButton />
                  </Box>
                )}
              />
            </Datagrid>
          </Box>
        </Card>
      )}
    </Box>
  );
};

// RADIUS 用户列表
export const RadiusUserList = () => {
  return (
    <List
      actions={<UserListActions />}
      sort={{ field: 'created_at', order: 'DESC' }}
      perPage={LARGE_LIST_PER_PAGE}
      pagination={<ServerPagination />}
      empty={false}
    >
      <RadiusUserListContent />
    </List>
  );
};

// RADIUS 用户编辑
export const RadiusUserEdit = () => {
  const translate = useTranslate();
  return (
    <Edit>
      <SimpleForm toolbar={<UserFormToolbar />} sx={formLayoutSx}>
        <FormSection
          title={translate('resources.radius/users.sections.authentication')}
          description={translate('resources.radius/users.sections.authentication_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="id"
                disabled
                label={translate('resources.radius/users.fields.id')}
                helperText={translate('resources.radius/users.helpers.id')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="username"
                label={translate('resources.radius/users.fields.username')}
                validate={[required(), minLength(3), maxLength(50)]}
                helperText={translate('resources.radius/users.helpers.username')}
                autoComplete="username"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="password"
                label={translate('resources.radius/users.fields.password')}
                type="password"
                validate={[minLength(6), maxLength(128)]}
                helperText={translate('resources.radius/users.helpers.password')}
                autoComplete="new-password"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="realname"
                label={translate('resources.radius/users.fields.realname')}
                validate={[maxLength(100)]}
                helperText={translate('resources.radius/users.helpers.realname')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.contact')}
          description={translate('resources.radius/users.sections.contact_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="email"
                label={translate('resources.radius/users.fields.email')}
                type="email"
                validate={[email(), maxLength(100)]}
                helperText={translate('resources.radius/users.helpers.email')}
                autoComplete="email"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="mobile"
                label={translate('resources.radius/users.fields.mobile')}
                validate={[maxLength(20)]}
                helperText={translate('resources.radius/users.helpers.mobile')}
                autoComplete="tel"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="address"
                label={translate('resources.radius/users.fields.address')}
                multiline
                minRows={2}
                helperText={translate('resources.radius/users.helpers.address')}
                autoComplete="street-address"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.service')}
          description={translate('resources.radius/users.sections.service_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="status"
                  label={translate('resources.radius/users.fields.status')}
                  helperText={translate('resources.radius/users.helpers.status')}
                />
              </Box>
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="profile_id" reference="radius/profiles">
                <SelectInput
                  label={translate('resources.radius/users.fields.profile_id')}
                  optionText="name"
                  helperText={translate('resources.radius/users.helpers.profile_id')}
                  fullWidth
                  size="small"
                />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="expire_time"
                label={translate('resources.radius/users.fields.expire_time')}
                type="datetime-local"
                helperText={translate('resources.radius/users.helpers.expire_time')}
                fullWidth
                size="small"
                InputLabelProps={{ shrink: true }}
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="billing_type"
                label={translate('resources.radius/users.fields.billing_type')}
                choices={[
                  { id: 'prepaid', name: 'Prepaid' },
                  { id: 'postpaid', name: 'Postpaid' },
                ]}
                defaultValue="prepaid"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="subscription_status"
                label={translate('resources.radius/users.fields.subscription_status')}
                choices={[
                  { id: 'active', name: 'Active' },
                  { id: 'suspended', name: 'Suspended' },
                  { id: 'canceled', name: 'Canceled' },
                ]}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="monthly_fee"
                label={translate('resources.radius/users.fields.monthly_fee')}
                type="number"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="price_per_gb"
                label={translate('resources.radius/users.fields.price_per_gb')}
                type="number"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="next_billing_date"
                label={translate('resources.radius/users.fields.next_billing_date')}
                type="date"
                fullWidth
                size="small"
                InputLabelProps={{ shrink: true }}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.network')}
          description={translate('resources.radius/users.sections.network_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="ip_addr"
                label={translate('resources.radius/users.fields.ip_addr')}
                helperText={translate('resources.radius/users.helpers.ip_addr')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="ipv6_addr"
                label={translate('resources.radius/users.fields.ipv6_addr')}
                helperText={translate('resources.radius/users.helpers.ipv6_addr')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.remark')}
          description={translate('resources.radius/users.sections.remark_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="remark"
                label={translate('resources.radius/users.fields.remark')}
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText={translate('resources.radius/users.helpers.remark')}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};

// RADIUS 用户创建
export const RadiusUserCreate = () => {
  const translate = useTranslate();
  return (
    <Create>
      <SimpleForm sx={formLayoutSx}>
        <FormSection
          title={translate('resources.radius/users.sections.authentication')}
          description={translate('resources.radius/users.sections.authentication_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="username"
                label={translate('resources.radius/users.fields.username')}
                validate={[required(), minLength(3), maxLength(50)]}
                helperText={translate('resources.radius/users.helpers.username')}
                autoComplete="username"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="password"
                label={translate('resources.radius/users.fields.password')}
                type="password"
                validate={[required(), minLength(6), maxLength(128)]}
                helperText={translate('resources.radius/users.helpers.password_create')}
                autoComplete="new-password"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="realname"
                label={translate('resources.radius/users.fields.realname')}
                validate={[maxLength(100)]}
                helperText={translate('resources.radius/users.helpers.realname')}
                autoComplete="name"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.contact')}
          description={translate('resources.radius/users.sections.contact_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="email"
                label={translate('resources.radius/users.fields.email')}
                type="email"
                validate={[email(), maxLength(100)]}
                helperText={translate('resources.radius/users.helpers.email')}
                autoComplete="email"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="mobile"
                label={translate('resources.radius/users.fields.mobile')}
                validate={[maxLength(20)]}
                helperText={translate('resources.radius/users.helpers.mobile')}
                autoComplete="tel"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="address"
                label={translate('resources.radius/users.fields.address')}
                multiline
                minRows={2}
                helperText={translate('resources.radius/users.helpers.address')}
                autoComplete="street-address"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.service')}
          description={translate('resources.radius/users.sections.service_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="status"
                  label={translate('resources.radius/users.fields.status')}
                  defaultValue={true}
                  helperText={translate('resources.radius/users.helpers.status')}
                />
              </Box>
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="profile_id" reference="radius/profiles">
                <SelectInput
                  label={translate('resources.radius/users.fields.profile_id')}
                  optionText="name"
                  helperText={translate('resources.radius/users.helpers.profile_id')}
                  fullWidth
                  size="small"
                />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="expire_time"
                label={translate('resources.radius/users.fields.expire_time')}
                type="datetime-local"
                helperText={translate('resources.radius/users.helpers.expire_time')}
                fullWidth
                size="small"
                InputLabelProps={{ shrink: true }}
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="billing_type"
                label={translate('resources.radius/users.fields.billing_type')}
                choices={[
                  { id: 'prepaid', name: 'Prepaid' },
                  { id: 'postpaid', name: 'Postpaid' },
                ]}
                defaultValue="prepaid"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="monthly_fee"
                label={translate('resources.radius/users.fields.monthly_fee')}
                type="number"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="price_per_gb"
                label={translate('resources.radius/users.fields.price_per_gb')}
                type="number"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.network')}
          description={translate('resources.radius/users.sections.network_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="ip_addr"
                label={translate('resources.radius/users.fields.ip_addr')}
                helperText={translate('resources.radius/users.helpers.ip_addr')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="ipv6_addr"
                label={translate('resources.radius/users.fields.ipv6_addr')}
                helperText={translate('resources.radius/users.helpers.ipv6_addr')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.remark')}
          description={translate('resources.radius/users.sections.remark_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="remark"
                label={translate('resources.radius/users.fields.remark')}
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText={translate('resources.radius/users.helpers.remark')}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Create>
  );
};

// ============ 顶部概览卡片 ============

const UserHeaderCard = () => {
  const record = useRecordContext<RadiusUser>();
  const translate = useTranslate();
  const notify = useNotify();
  const refresh = useRefresh();

  const handleCopy = useCallback((text: string, label: string) => {
    navigator.clipboard.writeText(text);
    notify(translate('resources.radius/users.copied', { _: '%{label} copied to clipboard', label }), { type: 'info' });
  }, [notify, translate]);

  const handleRefresh = useCallback(() => {
    refresh();
    notify(translate('resources.radius/users.data_refreshed', { _: 'Data refreshed' }), { type: 'info' });
  }, [refresh, notify, translate]);

  if (!record) return null;

  const isEnabled = record.status === 'enabled';
  const expireInfo = formatExpireTime(record.expire_time);

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

      <CardContent sx={{ p: { xs: 2, sm: 3 }, position: 'relative', zIndex: 1 }}>
        <Box sx={{ display: 'flex', flexDirection: { xs: 'column', sm: 'row' }, justifyContent: 'space-between', alignItems: { xs: 'stretch', sm: 'flex-start' }, gap: { xs: 2, sm: 0 }, mb: { xs: 2, sm: 3 } }}>
          {/* Left: User Info */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: { xs: 1.5, sm: 2 } }}>
            <Avatar
              sx={{
                width: { xs: 48, sm: 64 },
                height: { xs: 48, sm: 64 },
                bgcolor: isEnabled ? 'primary.main' : 'grey.500',
                fontSize: { xs: '1.2rem', sm: '1.5rem' },
                fontWeight: 700,
                boxShadow: theme => `0 4px 14px ${alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.4)}`,
              }}
            >
              {record.username?.charAt(0).toUpperCase() || 'U'}
            </Avatar>
            <Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5, flexWrap: 'wrap' }}>
                <Typography variant="h6" sx={{ fontWeight: 700, color: 'text.primary', fontSize: { xs: '1.1rem', sm: '1.5rem' } }}>
                  {record.username || <EmptyValue message={translate('resources.radius/users.fields.unknown_user', { _: 'Unknown user' })} />}
                </Typography>
                {isEnabled ? (
                  <Chip
                    icon={<EnabledIcon sx={{ fontSize: '1rem !important' }} />}
                    label={translate('resources.radius/users.status.enabled', { _: '启用' })}
                    size="small"
                    color="success"
                    sx={{ fontWeight: 600, height: 24 }}
                  />
                ) : (
                  <Chip
                    icon={<DisabledIcon sx={{ fontSize: '1rem !important' }} />}
                    label={translate('resources.radius/users.status.disabled', { _: '禁用' })}
                    size="small"
                    color="default"
                    variant="outlined"
                    sx={{ fontWeight: 600, height: 24 }}
                  />
                )}
              </Box>
              {record.realname && (
                <Typography variant="body2" color="text.secondary">
                  {record.realname}
                </Typography>
              )}
              {record.username && (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                  <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                    {translate('resources.radius/users.fields.id', { _: 'User ID' })}: {record.id}
                  </Typography>
                  <Tooltip title={translate('resources.radius/users.copy_username', { _: 'Copy Username' })}>
                    <IconButton
                      size="small"
                      onClick={() => handleCopy(record.username!, translate('resources.radius/users.fields.username', { _: 'Username' }))}
                      sx={{ p: 0.5 }}
                    >
                      <CopyIcon sx={{ fontSize: '0.75rem' }} />
                    </IconButton>
                  </Tooltip>
                </Box>
              )}
            </Box>
          </Box>

          {/* Right: Action buttons */}
          <Box className="no-print" sx={{ display: 'flex', gap: 1, justifyContent: { xs: 'flex-end', sm: 'flex-start' } }}>
            <Tooltip title={translate('resources.radius/users.print_details', { _: 'Print Details' })}>
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
            <Tooltip title={translate('resources.radius/users.refresh_data', { _: 'Refresh Data' })}>
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
            gap: 1.5,
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
              <EmailIcon sx={{ fontSize: '1.1rem', color: 'info.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.radius/users.fields.email', { _: '邮箱' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600, wordBreak: 'break-all' }}>
              {record.email || '-'}
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
              <PhoneIcon sx={{ fontSize: '1.1rem', color: 'success.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.radius/users.fields.mobile', { _: '手机号' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.mobile || '-'}
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
              <NetworkIcon sx={{ fontSize: '1.1rem', color: 'warning.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.radius/users.fields.ip_addr', { _: 'IP地址' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600, fontFamily: 'monospace' }}>
              {record.ip_addr || '-'}
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
              <CalendarIcon sx={{ fontSize: '1.1rem', color: expireInfo.color === 'error' ? 'error.main' : expireInfo.color === 'warning' ? 'warning.main' : 'success.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.radius/users.fields.expire_time', { _: '过期时间' })}
              </Typography>
            </Box>
            <Chip
              label={expireInfo.text}
              size="small"
              color={expireInfo.color}
              sx={{ fontWeight: 600 }}
            />
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

// ============ 用户详情内容 ============

const UserDetails = () => {
  const record = useRecordContext<RadiusUser>();
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
          <UserHeaderCard />

          {/* 基本信息 */}
          <DetailSectionCard
            title={translate('resources.radius/users.sections.basic', { _: '基本信息' })}
            description={translate('resources.radius/users.sections.basic_desc', { _: '用户的身份认证信息' })}
            icon={<PersonIcon />}
            color="primary"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                  md: 'repeat(3, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.radius/users.fields.username', { _: '用户名' })}
                value={record.username}
                highlight
              />
              <DetailItem
                label={translate('resources.radius/users.fields.realname', { _: '真实姓名' })}
                value={record.realname || <EmptyValue />}
              />
              <DetailItem
                label={translate('resources.radius/users.fields.status', { _: '状态' })}
                value={
                  <Chip
                    icon={record.status === 'enabled' ? <EnabledIcon sx={{ fontSize: '0.9rem !important' }} /> : <DisabledIcon sx={{ fontSize: '0.9rem !important' }} />}
                    label={record.status === 'enabled' ? translate('resources.radius/users.status.enabled', { _: '启用' }) : translate('resources.radius/users.status.disabled', { _: '禁用' })}
                    size="small"
                    color={record.status === 'enabled' ? 'success' : 'default'}
                    sx={{ fontWeight: 600 }}
                  />
                }
                highlight
              />
            </Box>
          </DetailSectionCard>

          {/* 联系方式 */}
          <DetailSectionCard
            title={translate('resources.radius/users.sections.contact', { _: '联系方式' })}
            description={translate('resources.radius/users.sections.contact_desc', { _: '联系信息和地址' })}
            icon={<ContactIcon />}
            color="info"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                  md: 'repeat(3, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.radius/users.fields.email', { _: '邮箱' })}
                value={record.email || <EmptyValue />}
              />
              <DetailItem
                label={translate('resources.radius/users.fields.mobile', { _: '手机号' })}
                value={record.mobile || <EmptyValue />}
              />
              <DetailItem
                label={translate('resources.radius/users.fields.address', { _: '地址' })}
                value={record.address || <EmptyValue />}
              />
            </Box>
          </DetailSectionCard>

          {/* 服务配置 */}
          <DetailSectionCard
            title={translate('resources.radius/users.sections.service', { _: '服务配置' })}
            description={translate('resources.radius/users.sections.service_desc', { _: 'RADIUS服务和权限设置' })}
            icon={<SettingsIcon />}
            color="success"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                  md: 'repeat(3, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.radius/users.fields.profile_id', { _: '计费策略' })}
                value={
                  record.profile_id ? (
                    <ReferenceField source="profile_id" reference="radius/profiles" link="show">
                      <TextField source="name" />
                    </ReferenceField>
                  ) : (
                    <EmptyValue message={translate('resources.radius/users.fields.unassigned', { _: 'Unassigned' })} />
                  )
                }
                highlight
              />
              <DetailItem
                label={translate('resources.radius/users.fields.expire_time', { _: '过期时间' })}
                value={
                  (() => {
                    const info = formatExpireTime(record.expire_time);
                    return (
                      <Chip
                        label={info.text}
                        size="small"
                        color={info.color}
                        sx={{ fontWeight: 600 }}
                      />
                    );
                  })()
                }
                highlight
              />
              <DetailItem
                label={translate('resources.radius/users.fields.billing_type', { _: 'Billing Type' })}
                value={record.billing_type}
              />
              {record.billing_type === 'postpaid' && (
                <>
                  <DetailItem
                    label={translate('resources.radius/users.fields.subscription_status', { _: 'Sub Status' })}
                    value={record.subscription_status}
                  />
                  <DetailItem
                    label={translate('resources.radius/users.fields.monthly_fee', { _: 'Monthly Fee' })}
                    value={record.monthly_fee}
                  />
                  <DetailItem
                    label={translate('resources.radius/users.fields.price_per_gb', { _: 'Price per GB' })}
                    value={record.price_per_gb}
                  />
                  <DetailItem
                    label={translate('resources.radius/users.fields.next_billing_date', { _: 'Next Billing' })}
                    value={record.next_billing_date ? new Date(record.next_billing_date).toLocaleDateString() : '-'}
                  />
                </>
              )}
            </Box>
          </DetailSectionCard>

          {record.billing_type === 'postpaid' && record.username && (
            <Card variant="outlined" sx={{ borderRadius: 2, bgcolor: theme => alpha(theme.palette.primary.main, 0.05) }}>
              <CardContent>
                <Typography variant="subtitle2" gutterBottom>
                  {translate('resources.radius/users.billing_actions', { _: 'Billing Actions' })}
                </Typography>
                <GenerateBillButton username={record.username} />
              </CardContent>
            </Card>
          )}

          {/* 网络配置 */}
          <DetailSectionCard
            title={translate('resources.radius/users.sections.network', { _: '网络配置' })}
            description={translate('resources.radius/users.sections.network_desc', { _: 'IP地址分配设置' })}
            icon={<NetworkIcon />}
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
                label={translate('resources.radius/users.fields.ip_addr', { _: 'IPv4地址' })}
                value={
                  record.ip_addr ? (
                    <Chip
                      label={record.ip_addr}
                      size="small"
                      color="info"
                      variant="outlined"
                      sx={{ fontFamily: 'monospace' }}
                    />
                  ) : (
                    <EmptyValue message={translate('resources.radius/users.fields.unassigned', { _: 'Unassigned' })} />
                  )
                }
              />
              <DetailItem
                label={translate('resources.radius/users.fields.ipv6_addr', { _: 'IPv6地址' })}
                value={
                  record.ipv6_addr ? (
                    <Chip
                      label={record.ipv6_addr}
                      size="small"
                      color="info"
                      variant="outlined"
                      sx={{ fontFamily: 'monospace', fontSize: '0.75rem' }}
                    />
                  ) : (
                    <EmptyValue message={translate('resources.radius/users.fields.unassigned', { _: 'Unassigned' })} />
                  )
                }
              />
            </Box>
          </DetailSectionCard>

          {/* 时间信息 */}
          <DetailSectionCard
            title={translate('resources.radius/users.sections.timing', { _: '时间信息' })}
            description={translate('resources.radius/users.sections.timing_desc', { _: '创建和更新时间' })}
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
                label={translate('resources.radius/users.fields.created_at', { _: '创建时间' })}
                value={formatTimestamp(record.created_at)}
              />
              <DetailItem
                label={translate('resources.radius/users.fields.updated_at', { _: '更新时间' })}
                value={formatTimestamp(record.updated_at)}
              />
            </Box>
          </DetailSectionCard>

          {/* 备注信息 */}
          <DetailSectionCard
            title={translate('resources.radius/users.sections.remark', { _: '备注信息' })}
            description={translate('resources.radius/users.sections.remark_desc', { _: '额外的说明和备注' })}
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
                {record.remark || translate('resources.radius/users.empty.no_remark', { _: '无备注信息' })}
              </Typography>
            </Box>
          </DetailSectionCard>
        </Stack>
      </Box>
    </>
  );
};

// RADIUS 用户详情
export const RadiusUserShow = () => {
  return (
    <Show>
      <UserDetails />
    </Show>
  );
};
