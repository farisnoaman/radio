import React, { useState, useCallback } from 'react';
import {
  Card,
  CardContent,
  Box,
  TextField,
  Typography,
  FormControlLabel,
  Checkbox,
  Collapse,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import { useTranslate } from 'react-admin';

interface AdminCredentialsSectionProps {
  username?: string;
  password?: string;
  onUsernameChange: (username: string) => void;
  onPasswordChange: (password: string) => void;
  disabled?: boolean;
}

export const AdminCredentialsSection: React.FC<AdminCredentialsSectionProps> = ({
  username = '',
  password = '',
  onUsernameChange,
  onPasswordChange,
  disabled = false,
}) => {
  const translate = useTranslate();
  const [expanded, setExpanded] = useState(false);
  const [useDefaults, setUseDefaults] = useState(true);
  const [usernameError, setUsernameError] = useState('');
  const [passwordError, setPasswordError] = useState('');

  const validateUsername = useCallback((value: string) => {
    if (value && !/^[a-zA-Z0-9]+$/.test(value)) {
      setUsernameError(translate('provider.username_invalid'));
      return false;
    }
    if (value && (value.length < 3 || value.length > 50)) {
      setUsernameError(translate('provider.username_too_short') + ' / ' + translate('provider.username_too_long'));
      return false;
    }
    setUsernameError('');
    return true;
  }, [translate]);

  const validatePassword = useCallback((value: string) => {
    if (value && value.length < 6) {
      setPasswordError(translate('provider.password_too_short'));
      return false;
    }
    setPasswordError('');
    return true;
  }, [translate]);

  const handleToggleDefaults = (checked: boolean) => {
    setUseDefaults(checked);
    if (checked) {
      onUsernameChange('admin');
      onPasswordChange('123456');
      setUsernameError('');
      setPasswordError('');
    }
  };

  const handleUsernameChange = useCallback((value: string) => {
    onUsernameChange(value);
    validateUsername(value);
  }, [onUsernameChange, validateUsername]);

  const handlePasswordChange = useCallback((value: string) => {
    onPasswordChange(value);
    validatePassword(value);
  }, [onPasswordChange, validatePassword]);

  return (
    <Card>
      <Box
        sx={{
          p: 2,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          cursor: 'pointer',
          bgcolor: 'grey.50',
        }}
        onClick={() => setExpanded(!expanded)}
      >
        <Typography variant="h6">
          {translate('provider.admin_credentials')}
        </Typography>
        {expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
      </Box>

      <Collapse in={expanded}>
        <CardContent>
          <Box sx={{ mb: 2 }}>
            <FormControlLabel
              control={
                <Checkbox
                  checked={useDefaults}
                  onChange={(e) => handleToggleDefaults(e.target.checked)}
                  disabled={disabled}
                />
              }
              label={translate('provider.use_defaults')}
            />
          </Box>

          <TextField
            fullWidth
            label={translate('provider.username')}
            value={username}
            onChange={(e) => handleUsernameChange(e.target.value)}
            disabled={disabled || useDefaults}
            sx={{ mb: 2 }}
            error={!!usernameError}
            helperText={usernameError || translate('provider.username_help')}
            placeholder={useDefaults ? translate('provider.username_placeholder') : ''}
          />

          <TextField
            fullWidth
            label={translate('provider.password')}
            type="password"
            value={password}
            onChange={(e) => handlePasswordChange(e.target.value)}
            disabled={disabled || useDefaults}
            error={!!passwordError}
            helperText={passwordError || translate('provider.password_help')}
            placeholder={useDefaults ? translate('provider.password_placeholder') : ''}
          />

          {!useDefaults && (
            <Typography variant="caption" color="textSecondary">
              {translate('provider.custom_warning')}
            </Typography>
          )}
        </CardContent>
      </Collapse>
    </Card>
  );
};

export default AdminCredentialsSection;
