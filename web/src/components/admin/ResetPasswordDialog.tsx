import React, { useState, useCallback } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Box,
  Typography,
  Alert,
  AlertTitle,
  CircularProgress,
  Paper,
  Divider,
} from '@mui/material';
import {
  VpnKey as VpnKeyIcon,
  CheckCircle as CheckCircleIcon,
  Warning as WarningIcon,
} from '@mui/icons-material';
import { useNotify, useTranslate, useLocale } from 'react-admin';
import { apiRequest } from '../../utils/apiClient';

interface ResetPasswordDialogProps {
  open: boolean;
  onClose: () => void;
  providerId: string;
  onSuccess?: () => void;
}

interface FormData {
  username: string;
  password: string;
  confirmPassword: string;
}

interface FormErrors {
  username?: string;
  password?: string;
  confirmPassword?: string;
}

interface CredentialsResponse {
  username: string;
  password: string;
}

export const ResetPasswordDialog: React.FC<ResetPasswordDialogProps> = ({
  open,
  onClose,
  providerId,
  onSuccess,
}) => {
  const notify = useNotify();
  const translate = useTranslate();
  const locale = useLocale();
  const isRTL = locale === 'ar';

  const [formData, setFormData] = useState<FormData>({
    username: '',
    password: '',
    confirmPassword: '',
  });

  const [errors, setErrors] = useState<FormErrors>({});
  const [loading, setLoading] = useState(false);
  const [showSuccess, setShowSuccess] = useState(false);
  const [credentials, setCredentials] = useState<CredentialsResponse | null>(null);

  const getPasswordStrength = useCallback((password: string) => {
    if (!password) return { strength: 0, label: '' };

    let strength = 0;
    if (password.length >= 8) strength++;
    if (password.length >= 12) strength++;
    if (/[a-z]/.test(password) && /[A-Z]/.test(password)) strength++;
    if (/\d/.test(password)) strength++;
    if (/[^a-zA-Z0-9]/.test(password)) strength++;

    const labels = [
      '',
      translate('provider.strength_weak'),
      translate('provider.strength_fair'),
      translate('provider.strength_good'),
      translate('provider.strength_strong'),
      translate('provider.strength_very_strong'),
    ];
    return { strength, label: labels[strength] };
  }, [translate]);

  const validateForm = useCallback((): boolean => {
    const newErrors: FormErrors = {};

    // Username validation: 3-50 alphanumeric characters
    if (!formData.username) {
      newErrors.username = translate('provider.username_required', {
        _: 'Username is required',
      });
    } else if (formData.username.length < 3) {
      newErrors.username = translate('provider.username_too_short', {
        _: 'Username must be at least 3 characters',
      });
    } else if (formData.username.length > 50) {
      newErrors.username = translate('provider.username_too_long', {
        _: 'Username must not exceed 50 characters',
      });
    } else if (!/^[a-zA-Z0-9]+$/.test(formData.username)) {
      newErrors.username = translate('provider.username_invalid', {
        _: 'Username must contain only letters and numbers',
      });
    }

    // Password validation: minimum 6 characters
    if (!formData.password) {
      newErrors.password = translate('provider.password_required', {
        _: 'Password is required',
      });
    } else if (formData.password.length < 6) {
      newErrors.password = translate('provider.password_too_short', {
        _: 'Password must be at least 6 characters',
      });
    }

    // Confirm password validation
    if (!formData.confirmPassword) {
      newErrors.confirmPassword = translate('provider.confirm_password', {
        _: 'Please confirm your password',
      });
    } else if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = translate('provider.password_mismatch', {
        _: 'Passwords do not match',
      });
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [formData, translate]);

  const handleSubmit = useCallback(async () => {
    if (!validateForm()) {
      return;
    }

    setLoading(true);

    try {
      const result = await apiRequest<CredentialsResponse>(
        `/platform/providers/${providerId}/admin-credentials`,
        {
          method: 'PUT',
          body: JSON.stringify({
            username: formData.username,
            password: formData.password,
          }),
        }
      );

      setCredentials(result);
      setShowSuccess(true);
      notify(
        translate('provider.admin_update_success', {
          _: 'Admin credentials updated successfully',
        }),
        { type: 'success' }
      );

      if (onSuccess) {
        onSuccess();
      }
    } catch (error: any) {
      console.error('Failed to update credentials:', error);
      notify(
        error.message ||
          translate('provider.admin_update_error', {
            _: 'Failed to update admin credentials',
          }),
        { type: 'error' }
      );
    } finally {
      setLoading(false);
    }
  }, [formData, providerId, validateForm, notify, translate, onSuccess]);

  const handleClose = useCallback(() => {
    // Reset state
    setFormData({ username: '', password: '', confirmPassword: '' });
    setErrors({});
    setShowSuccess(false);
    setCredentials(null);
    onClose();
  }, [onClose]);

  const handleFieldChange = useCallback(
    (field: keyof FormData) => (event: React.ChangeEvent<HTMLInputElement>) => {
      setFormData((prev) => ({ ...prev, [field]: event.target.value }));
      // Clear error for this field when user starts typing
      if (errors[field]) {
        setErrors((prev) => ({ ...prev, [field]: undefined }));
      }
    },
    [errors]
  );

  return (
    <Dialog
      open={open}
      onClose={loading ? undefined : handleClose}
      maxWidth="sm"
      fullWidth
      dir={isRTL ? 'rtl' : 'ltr'}
    >
      <DialogTitle
        sx={{
          display: 'flex',
          alignItems: 'center',
          gap: 1,
        }}
      >
        {showSuccess ? <CheckCircleIcon color="success" /> : <VpnKeyIcon color="primary" />}
        {showSuccess
          ? translate('provider.credentials_updated')
          : translate('provider.update_credentials_title')}
      </DialogTitle>

      <DialogContent>
        {!showSuccess ? (
          <>
            <Alert severity="warning" sx={{ mb: 3 }}>
              <AlertTitle>
                {translate('provider.save_warning')}
              </AlertTitle>
              <Typography variant="body2">
                {translate('provider.custom_warning')}
              </Typography>
            </Alert>

            <TextField
              fullWidth
              label={translate('provider.new_username')}
              value={formData.username}
              onChange={handleFieldChange('username')}
              error={!!errors.username}
              helperText={
                errors.username ||
                translate('provider.username_help')
              }
              disabled={loading}
              autoComplete="off"
              sx={{ mb: 2 }}
            />

            <TextField
              fullWidth
              label={translate('provider.new_password')}
              type="password"
              value={formData.password}
              onChange={handleFieldChange('password')}
              error={!!errors.password}
              helperText={
                errors.password ||
                translate('provider.password_help')
              }
              disabled={loading}
              autoComplete="new-password"
              sx={{ mb: 1 }}
            />

            {formData.password && !errors.password && (
              <Box sx={{ mt: 1, mb: 2 }}>
                <Typography variant="caption" color="textSecondary">
                  {translate('provider.password_strength')}: {getPasswordStrength(formData.password).label}
                </Typography>
                <Box
                  sx={{
                    height: 4,
                    bgcolor: 'grey.300',
                    borderRadius: 2,
                    mt: 0.5,
                  }}
                >
                  <Box
                    sx={{
                      height: '100%',
                      width: `${(getPasswordStrength(formData.password).strength / 5) * 100}%`,
                      bgcolor: getPasswordStrength(formData.password).strength <= 2 ? 'error.main' :
                               getPasswordStrength(formData.password).strength === 3 ? 'warning.main' : 'success.main',
                      borderRadius: 2,
                      transition: 'width 0.3s',
                    }}
                  />
                </Box>
              </Box>
            )}

            <TextField
              fullWidth
              label={translate('provider.confirm_password')}
              type="password"
              value={formData.confirmPassword}
              onChange={handleFieldChange('confirmPassword')}
              error={!!errors.confirmPassword}
              helperText={errors.confirmPassword}
              disabled={loading}
              autoComplete="new-password"
            />
          </>
        ) : (
          <Box>
            <Alert severity="success" sx={{ mb: 3 }}>
              <AlertTitle>
                {translate('provider.credentials_updated')}
              </AlertTitle>
              <Typography variant="body2">
                {translate('provider.credentials_updated_message')}
              </Typography>
            </Alert>

            <Paper
              variant="outlined"
              sx={{
                p: 3,
                bgcolor: 'grey.50',
                border: 2,
                borderColor: 'primary.main',
              }}
            >
              <Typography
                variant="subtitle2"
                color="textSecondary"
                gutterBottom
                sx={{ fontWeight: 600 }}
              >
                {translate('provider.new_credentials')}
              </Typography>

              <Divider sx={{ my: 2 }} />

              <Box sx={{ mb: 2 }}>
                <Typography
                  variant="caption"
                  color="textSecondary"
                  sx={{ display: 'block', mb: 0.5 }}
                >
                  {translate('provider.username')}
                </Typography>
                <Typography
                  variant="body1"
                  sx={{
                    fontFamily: 'monospace',
                    fontWeight: 600,
                    fontSize: '1.1rem',
                    p: 1,
                    bgcolor: 'background.paper',
                    borderRadius: 1,
                  }}
                >
                  {credentials?.username}
                </Typography>
              </Box>

              <Box>
                <Typography
                  variant="caption"
                  color="textSecondary"
                  sx={{ display: 'block', mb: 0.5 }}
                >
                  {translate('provider.password')}
                </Typography>
                <Typography
                  variant="body1"
                  sx={{
                    fontFamily: 'monospace',
                    fontWeight: 600,
                    fontSize: '1.1rem',
                    p: 1,
                    bgcolor: 'background.paper',
                    borderRadius: 1,
                    letterSpacing: 2,
                  }}
                >
                  {credentials?.password}
                </Typography>
              </Box>
            </Paper>

            <Alert severity="error" sx={{ mt: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1 }}>
                <WarningIcon fontSize="small" sx={{ mt: 0.2 }} />
                <Typography variant="body2">
                  {translate('provider.save_warning')}
                </Typography>
              </Box>
            </Alert>
          </Box>
        )}
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 2 }}>
        {!showSuccess ? (
          <>
            <Button
              onClick={handleClose}
              disabled={loading}
              variant="outlined"
              color="inherit"
            >
              {translate('ra.action.cancel')}
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={loading}
              variant="contained"
              startIcon={loading ? <CircularProgress size={20} /> : <VpnKeyIcon />}
            >
              {loading
                ? translate('provider.updating_credentials')
                : translate('provider.update_credentials')}
            </Button>
          </>
        ) : (
          <Button
            onClick={handleClose}
            variant="contained"
            color="primary"
            autoFocus
          >
            {translate('ra.action.confirm')}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
};

export default ResetPasswordDialog;
