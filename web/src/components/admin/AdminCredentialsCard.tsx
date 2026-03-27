import React, { useState, useCallback } from 'react';
import {
  Card,
  CardContent,
  Box,
  Typography,
  TextField,
  IconButton,
  Chip,
  Button,
  Stack,
  CircularProgress,
  Alert,
  Tooltip,
} from '@mui/material';
import {
  Visibility as VisibilityIcon,
  VisibilityOff as VisibilityOffIcon,
  ContentCopy as CopyIcon,
  Refresh as RefreshIcon,
  RestartAlt as ResetIcon,
  AdminPanelSettings as AdminIcon,
  Edit as EditIcon,
} from '@mui/icons-material';
import { useNotify, useRefresh, useRecordContext, useTranslate } from 'react-admin';
import { apiRequest } from '../../utils/apiClient';
import { ResetPasswordDialog } from './ResetPasswordDialog';

interface AdminCredentials {
  username: string;
  password: string;
  enabled: boolean;
}

interface AdminCredentialsCardProps {
  onRefresh?: () => void;
}

export const AdminCredentialsCard: React.FC<AdminCredentialsCardProps> = ({
  onRefresh,
}) => {
  const record = useRecordContext();
  const notify = useNotify();
  const refresh = useRefresh();
  const translate = useTranslate();

  // Get providerId from record context
  const providerId = record?.id;

  const [credentials, setCredentials] = useState<AdminCredentials | null>(null);
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [resetting, setResetting] = useState(false);
  const [resetDialogOpen, setResetDialogOpen] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const getErrorMessage = useCallback((error: any) => {
    if (error?.body?.code === 'PROVIDER_NOT_FOUND') {
      return translate('provider.error_loading');
    }
    if (error?.body?.code === 'USERNAME_EXISTS') {
      return translate('provider.username_exists');
    }
    if (error?.body?.code === 'FORBIDDEN') {
      return translate('provider.update_error');
    }
    return error?.message || translate('provider.error_loading');
  }, [translate]);

  const fetchCredentials = useCallback(async () => {
    // Don't fetch if providerId is not available yet
    if (!providerId) {
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const data = await apiRequest<AdminCredentials>(
        `/platform/providers/${providerId}/admin-credentials`
      );
      setCredentials(data);
      setError(null);
    } catch (error: any) {
      setError(error);
      notify(`${translate('provider.error_loading')}: ${getErrorMessage(error)}`, {
        type: 'error',
      });
    } finally {
      setLoading(false);
    }
  }, [providerId, notify, getErrorMessage]);

  const handleCopy = useCallback(
    (text: string) => {
      navigator.clipboard.writeText(text);
      notify(translate('provider.copied'), { type: 'success' });
    },
    [notify, translate]
  );

  const handleReset = useCallback(async () => {
    // Don't proceed if providerId is not available
    if (!providerId) {
      notify(translate('provider.error_loading'), { type: 'warning' });
      return;
    }

    if (
      !window.confirm(translate('provider.reset_to_defaults_confirm'))
    ) {
      return;
    }

    setResetting(true);
    setError(null);
    try {
      const data = await apiRequest<AdminCredentials>(
        `/platform/providers/${providerId}/admin-credentials/reset`,
        {
          method: 'POST',
        }
      );
      setCredentials(data);
      setError(null);
      notify(translate('provider.reset_success'), {
        type: 'success',
      });
      if (onRefresh) {
        onRefresh();
      }
    } catch (error: any) {
      setError(error);
      notify(`${translate('provider.reset_error')}: ${getErrorMessage(error)}`, { type: 'error' });
    } finally {
      setResetting(false);
    }
  }, [providerId, notify, onRefresh, getErrorMessage, translate]);

  const handleResetDialogClose = useCallback(() => {
    setResetDialogOpen(false);
    // Refresh credentials after dialog closes
    fetchCredentials();
    if (onRefresh) {
      onRefresh();
    }
  }, [fetchCredentials, onRefresh]);

  React.useEffect(() => {
    fetchCredentials();
  }, [fetchCredentials]);

  if (loading) {
    return (
      <Card>
        <CardContent>
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              py: 4,
            }}
          >
            <CircularProgress size={24} />
            <Typography sx={{ ml: 2 }}>{translate('provider.loading_credentials')}</Typography>
          </Box>
        </CardContent>
      </Card>
    );
  }

  // Show loading state if provider record is not available yet
  if (!providerId) {
    return (
      <Card>
        <CardContent>
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              py: 4,
            }}
          >
            <CircularProgress size={24} />
            <Typography sx={{ ml: 2 }}>{translate('provider.loading_provider')}</Typography>
          </Box>
        </CardContent>
      </Card>
    );
  }

  if (error && !credentials) {
    return (
      <Card>
        <CardContent>
          <Alert
            severity="error"
            action={
              <Button color="inherit" size="small" onClick={fetchCredentials}>
                {translate('provider.retry')}
              </Button>
            }
          >
            {getErrorMessage(error)}
          </Alert>
        </CardContent>
      </Card>
    );
  }

  if (!credentials) {
    return (
      <Card>
        <CardContent>
          <Alert severity="info">
            {translate('provider.no_credentials')}
          </Alert>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card
      sx={{
        borderRadius: 2,
        border: theme => `1px solid ${theme.palette.divider}`,
      }}
    >
      <CardContent>
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'flex-start',
            mb: 3,
          }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
            <AdminIcon color="primary" sx={{ fontSize: 28 }} />
            <Box>
              <Typography variant="h6" sx={{ fontWeight: 600 }}>
                {translate('provider.admin_credentials_title')}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {translate('provider.admin_credentials_help')}
              </Typography>
            </Box>
          </Box>
          <Chip
            label={translate(credentials.enabled ? 'provider.enabled' : 'provider.disabled')}
            color={credentials.enabled ? 'success' : 'default'}
            size="small"
          />
        </Box>

        <Stack spacing={2.5}>
          {/* Username Field */}
          <Box>
            <Typography
              variant="body2"
              sx={{ mb: 1, fontWeight: 500, color: 'text.secondary' }}
            >
            {translate('provider.username')}
          </Typography>
            <Box sx={{ display: 'flex', gap: 1 }}>
              <TextField
                fullWidth
                size="small"
                value={credentials.username}
                disabled
                sx={{
                  '& .MuiInputBase-input.Mui-disabled': {
                    WebkitTextFillColor: 'text.primary',
                    color: 'text.primary',
                  },
                }}
              />
              <TooltipButton
                title={translate('provider.copy_username')}
                onClick={() => handleCopy(credentials.username)}
              >
                <CopyIcon />
              </TooltipButton>
            </Box>
          </Box>

          {/* Password Field */}
          <Box>
            <Typography
              variant="body2"
              sx={{ mb: 1, fontWeight: 500, color: 'text.secondary' }}
            >
            {translate('provider.current_password')}
          </Typography>
            <Box sx={{ display: 'flex', gap: 1 }}>
              <TextField
                fullWidth
                size="small"
                type={showPassword ? 'text' : 'password'}
                value={credentials.password}
                disabled
                sx={{
                  '& .MuiInputBase-input.Mui-disabled': {
                    WebkitTextFillColor: 'text.primary',
                    color: 'text.primary',
                  },
                }}
              />
              <TooltipButton
                title={showPassword ? translate('provider.hide_password') : translate('provider.show_password')}
                onClick={() => setShowPassword(!showPassword)}
              >
                {showPassword ? <VisibilityOffIcon /> : <VisibilityIcon />}
              </TooltipButton>
              <TooltipButton
                title={translate('provider.copy_password')}
                onClick={() => handleCopy(credentials.password)}
              >
                <CopyIcon />
              </TooltipButton>
            </Box>
          </Box>

          {/* Action Buttons */}
          <Box sx={{ display: 'flex', gap: 1.5, mt: 1 }}>
            <Button
              variant="outlined"
              size="small"
              startIcon={<EditIcon />}
              onClick={() => setResetDialogOpen(true)}
              sx={{ flex: 1 }}
            >
              {translate('provider.update_credentials')}
            </Button>
            <Button
              variant="outlined"
              size="small"
              startIcon={<ResetIcon />}
              onClick={handleReset}
              disabled={resetting}
              sx={{ flex: 1 }}
            >
              {resetting ? translate('provider.resetting') + '...' : translate('provider.reset_to_defaults')}
            </Button>
            <Button
              variant="outlined"
              size="small"
              startIcon={<RefreshIcon />}
              onClick={fetchCredentials}
              sx={{ flex: 1 }}
            >
              {translate('provider.refresh')}
            </Button>
          </Box>

          {/* Info Alert */}
          <Alert severity="info" sx={{ mt: 1 }}>
            <Typography variant="body2">
              {translate('provider.default_credentials_info')}
            </Typography>
          </Alert>
        </Stack>
      </CardContent>

      <ResetPasswordDialog
        open={resetDialogOpen}
        onClose={handleResetDialogClose}
        providerId={String(providerId)}
        onSuccess={() => {
          refresh();
          if (onRefresh) {
            onRefresh();
          }
        }}
      />
    </Card>
  );
};

// Helper component for tooltip buttons
const TooltipButton: React.FC<{
  title: string;
  onClick: () => void;
  children: React.ReactNode;
}> = ({ title, onClick, children }) => {
  return (
    <Tooltip title={title}>
      <IconButton
        size="small"
        onClick={onClick}
        sx={{
          border: theme => `1px solid ${theme.palette.divider}`,
          bgcolor: 'background.paper',
          '&:hover': {
            bgcolor: 'action.hover',
          },
        }}
      >
        {children}
      </IconButton>
    </Tooltip>
  );
};

export default AdminCredentialsCard;
