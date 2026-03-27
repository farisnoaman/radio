import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Switch,
  FormControlLabel,
  Chip,
  Stack,
  Divider,
  Alert,
  AlertTitle,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Save,
  Cancel,
  MonetizationOn,
} from '@mui/icons-material';
import { useState, useEffect } from 'react';
import { useTranslate } from 'react-admin';

const AVAILABLE_FEATURES = [
  { id: 'max_users', translationKey: 'feature_max_users', type: 'number' },
  { id: 'max_sessions', translationKey: 'feature_max_sessions', type: 'number' },
  { id: 'max_nas', translationKey: 'feature_max_nas', type: 'number' },
  { id: 'max_storage', translationKey: 'feature_max_storage', type: 'number' },
  { id: 'daily_backups', translationKey: 'feature_daily_backups', type: 'number' },
  { id: 'realtime_monitoring', translationKey: 'feature_realtime_monitoring', type: 'boolean' },
  { id: 'advanced_monitoring', translationKey: 'feature_advanced_monitoring', type: 'boolean' },
  { id: 'api_access', translationKey: 'feature_api_access', type: 'boolean' },
  { id: 'custom_branding', translationKey: 'feature_custom_branding', type: 'boolean' },
  { id: 'priority_support', translationKey: 'feature_priority_support', type: 'boolean' },
  { id: 'sla_guarantee', translationKey: 'feature_sla_guarantee', type: 'boolean' },
  { id: 'multi_location', translationKey: 'feature_multi_location', type: 'boolean' },
  { id: 'advanced_reporting', translationKey: 'feature_advanced_reporting', type: 'boolean' },
  { id: 'white_label', translationKey: 'feature_white_label', type: 'boolean' },
];

interface PricingPlan {
  id?: number;
  code: string;
  name: string;
  base_fee: number;
  included_users: number;
  overage_fee: number;
  max_users: number;
  features: string;
  is_active: boolean;
}

interface PlanFeature {
  id: string;
  value: number | boolean;
}

interface PricingPlansListProps {
  onEdit: (plan: PricingPlan) => void;
  onDelete: (plan: PricingPlan) => void;
}

const PricingPlansList = ({ onEdit, onDelete }: PricingPlansListProps) => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const translate = useTranslate();

  const [plans] = useState<PricingPlan[]>([
    {
      id: 1, code: 'starter', name: 'Starter Plan', base_fee: 29, included_users: 100,
      overage_fee: 1, max_users: 500,
      features: JSON.stringify([
        { id: 'max_users', value: 500 }, { id: 'max_sessions', value: 100 },
        { id: 'max_nas', value: 10 }, { id: 'daily_backups', value: 1 },
        { id: 'realtime_monitoring', value: true },
      ]),
      is_active: true,
    },
    {
      id: 2, code: 'professional', name: 'Professional Plan', base_fee: 99, included_users: 500,
      overage_fee: 0.8, max_users: 5000,
      features: JSON.stringify([
        { id: 'max_users', value: 5000 }, { id: 'max_sessions', value: 1000 },
        { id: 'max_nas', value: 50 }, { id: 'daily_backups', value: 5 },
        { id: 'realtime_monitoring', value: true }, { id: 'advanced_monitoring', value: true },
        { id: 'api_access', value: true }, { id: 'custom_branding', value: true },
      ]),
      is_active: true,
    },
    {
      id: 3, code: 'enterprise', name: 'Enterprise Plan', base_fee: 299, included_users: 2000,
      overage_fee: 0.5, max_users: 50000,
      features: JSON.stringify([
        { id: 'max_users', value: 50000 }, { id: 'max_sessions', value: 10000 },
        { id: 'max_nas', value: 500 }, { id: 'daily_backups', value: 999 },
        { id: 'realtime_monitoring', value: true }, { id: 'advanced_monitoring', value: true },
        { id: 'api_access', value: true }, { id: 'custom_branding', value: true },
        { id: 'priority_support', value: true }, { id: 'sla_guarantee', value: true },
        { id: 'multi_location', value: true }, { id: 'advanced_reporting', value: true },
        { id: 'white_label', value: true },
      ]),
      is_active: true,
    },
  ]);

  if (isMobile) {
    return (
      <Stack spacing={1.5}>
        {plans.map((plan) => {
          const features: PlanFeature[] = JSON.parse(plan.features);
          const featureCount = features.length;

          return (
            <Card
              key={plan.id}
              sx={{
                border: '1px solid rgba(148, 163, 184, 0.12)',
                borderRadius: 2,
                transition: 'all 0.2s ease',
                '&:hover': {
                  borderColor: 'rgba(30, 58, 138, 0.3)',
                  boxShadow: '0 4px 12px rgba(0,0,0,0.08)',
                },
              }}
            >
              <CardContent sx={{ py: 1.5, px: 2 }}>
                <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 1, mb: 1 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, flex: 1, minWidth: 0 }}>
                    <Box
                      sx={{
                        p: 1,
                        borderRadius: 1.5,
                        bgcolor: 'rgba(30, 58, 138, 0.08)',
                        display: 'flex',
                        flexShrink: 0,
                      }}
                    >
                      <MonetizationOn sx={{ color: '#1e3a8a', fontSize: 20 }} />
                    </Box>
                    <Box sx={{ minWidth: 0, flex: 1 }}>
                      <Typography variant="subtitle1" sx={{ fontWeight: 600, lineHeight: 1.3 }}>
                        {plan.name}
                      </Typography>
                      <Chip label={plan.code} size="small" variant="outlined" sx={{ mt: 0.5 }} />
                    </Box>
                  </Box>
                  <Chip
                    label={plan.is_active
                      ? translate('platform_settings.status_active')
                      : translate('platform_settings.status_inactive')}
                    color={plan.is_active ? 'success' : 'default'}
                    size="small"
                  />
                </Box>

                <Divider sx={{ my: 1 }} />

                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="caption" color="text.secondary">
                      {translate('platform_settings.table_base_fee')}
                    </Typography>
                    <Typography variant="caption" sx={{ fontWeight: 600 }}>
                      ${plan.base_fee}/mo
                    </Typography>
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="caption" color="text.secondary">
                      {translate('platform_settings.table_included_users')}
                    </Typography>
                    <Typography variant="caption" sx={{ fontWeight: 500 }}>
                      {plan.included_users.toLocaleString()}
                    </Typography>
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="caption" color="text.secondary">
                      {translate('platform_settings.table_overage_fee')}
                    </Typography>
                    <Typography variant="caption" sx={{ fontWeight: 500 }}>
                      ${plan.overage_fee}/user
                    </Typography>
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="caption" color="text.secondary">
                      {translate('platform_settings.table_max_users')}
                    </Typography>
                    <Typography variant="caption" sx={{ fontWeight: 500 }}>
                      {plan.max_users.toLocaleString()}
                    </Typography>
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="caption" color="text.secondary">
                      {translate('platform_settings.table_features')}
                    </Typography>
                    <Typography variant="caption" sx={{ fontWeight: 500 }}>
                      {translate('platform_settings.features_count', { count: featureCount })}
                    </Typography>
                  </Box>
                </Box>

                <Divider sx={{ my: 1 }} />

                <Box sx={{ display: 'flex', gap: 1 }}>
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={<EditIcon />}
                    onClick={() => onEdit(plan)}
                    sx={{ fontSize: '0.75rem', flex: 1 }}
                  >
                    {translate('common.edit')}
                  </Button>
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={<DeleteIcon />}
                    onClick={() => onDelete(plan)}
                    color="error"
                    sx={{ fontSize: '0.75rem', flex: 1 }}
                  >
                    {translate('common.delete')}
                  </Button>
                </Box>
              </CardContent>
            </Card>
          );
        })}
      </Stack>
    );
  }

  return (
    <Box sx={{ overflowX: 'auto' }}>
      <Table size="medium">
        <TableHead>
          <TableRow>
            <TableCell sx={{ fontWeight: 700, whiteSpace: 'nowrap' }}>
              {translate('platform_settings.table_plan_name')}
            </TableCell>
            <TableCell sx={{ fontWeight: 700, whiteSpace: 'nowrap' }}>
              {translate('platform_settings.table_code')}
            </TableCell>
            <TableCell sx={{ fontWeight: 700, whiteSpace: 'nowrap' }}>
              {translate('platform_settings.table_base_fee')}
            </TableCell>
            <TableCell sx={{ fontWeight: 700, whiteSpace: 'nowrap' }}>
              {translate('platform_settings.table_included_users')}
            </TableCell>
            <TableCell sx={{ fontWeight: 700, whiteSpace: 'nowrap' }}>
              {translate('platform_settings.table_overage_fee')}
            </TableCell>
            <TableCell sx={{ fontWeight: 700, whiteSpace: 'nowrap' }}>
              {translate('platform_settings.table_max_users')}
            </TableCell>
            <TableCell sx={{ fontWeight: 700, whiteSpace: 'nowrap' }}>
              {translate('platform_settings.table_features')}
            </TableCell>
            <TableCell sx={{ fontWeight: 700, whiteSpace: 'nowrap' }}>
              {translate('platform_settings.table_status')}
            </TableCell>
            <TableCell sx={{ fontWeight: 700, whiteSpace: 'nowrap' }}>
              {translate('platform_settings.table_actions')}
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {plans.map((plan) => {
            const features: PlanFeature[] = JSON.parse(plan.features);
            const featureCount = features.length;

            return (
              <TableRow key={plan.id} hover>
                <TableCell sx={{ fontWeight: 600, whiteSpace: 'nowrap' }}>{plan.name}</TableCell>
                <TableCell>
                  <Chip label={plan.code} size="small" variant="outlined" />
                </TableCell>
                <TableCell sx={{ whiteSpace: 'nowrap' }}>${plan.base_fee}/mo</TableCell>
                <TableCell>{plan.included_users.toLocaleString()}</TableCell>
                <TableCell>${plan.overage_fee}/user</TableCell>
                <TableCell>{plan.max_users.toLocaleString()}</TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ fontSize: '0.75rem' }}>
                    {translate('platform_settings.features_count', { count: featureCount })}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Chip
                    label={plan.is_active
                      ? translate('platform_settings.status_active')
                      : translate('platform_settings.status_inactive')}
                    color={plan.is_active ? 'success' : 'default'}
                    size="small"
                  />
                </TableCell>
                <TableCell>
                  <Stack direction="row" spacing={0.5}>
                    <IconButton
                      size="small"
                      onClick={() => onEdit(plan)}
                      sx={{ color: '#1976d2' }}
                    >
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={() => onDelete(plan)}
                      sx={{ color: '#d32f2f' }}
                    >
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Stack>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </Box>
  );
};

interface PlanEditorProps {
  open: boolean;
  onClose: () => void;
  plan: PricingPlan | null;
  onSave: (plan: PricingPlan) => void;
}

const PlanEditor = ({ open, onClose, plan, onSave }: PlanEditorProps) => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const translate = useTranslate();

  const [formData, setFormData] = useState<PricingPlan>(
    plan || {
      code: '', name: '', base_fee: 0, included_users: 0, overage_fee: 0,
      max_users: 0, features: '[]', is_active: true,
    }
  );
  const [features, setFeatures] = useState<PlanFeature[]>(
    plan ? JSON.parse(plan.features) : []
  );

  useEffect(() => {
    if (plan) {
      setFormData(plan);
      setFeatures(JSON.parse(plan.features));
    }
  }, [plan]);

  const handleFeatureToggle = (featureId: string, value: number | boolean) => {
    setFeatures((prev) => {
      const exists = prev.find((f) => f.id === featureId);
      if (exists) {
        if (value === false || value === 0) {
          return prev.filter((f) => f.id !== featureId);
        }
        return prev.map((f) => (f.id === featureId ? { ...f, value } : f));
      }
      if (value !== false && value !== 0) {
        return [...prev, { id: featureId, value }];
      }
      return prev;
    });
  };

  const handleSave = () => {
    onSave({ ...formData, features: JSON.stringify(features) });
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle sx={{ fontWeight: 700, fontSize: isMobile ? '1.1rem' : undefined }}>
        {plan?.id
          ? translate('platform_settings.dialog_edit_title')
          : translate('platform_settings.dialog_add_title')}
      </DialogTitle>
      <DialogContent dividers>
        <Stack spacing={3} sx={{ mt: 2 }}>
          <Grid container spacing={2}>
            <Grid size={{ xs: 12, sm: 6 }}>
              <TextField
                fullWidth
                label={translate('platform_settings.plan_code')}
                value={formData.code}
                onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                helperText={translate('platform_settings.plan_code_help')}
                required
                size={isMobile ? 'small' : 'medium'}
              />
            </Grid>
            <Grid size={{ xs: 12, sm: 6 }}>
              <TextField
                fullWidth
                label={translate('platform_settings.plan_name')}
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                helperText={translate('platform_settings.plan_name_help')}
                required
                size={isMobile ? 'small' : 'medium'}
              />
            </Grid>
          </Grid>

          <Divider sx={{ my: 1 }} />
          <Typography
            variant="h6"
            sx={{ fontWeight: 600, fontSize: isMobile ? '0.95rem' : undefined }}
          >
            {translate('platform_settings.pricing_configuration')}
          </Typography>
          <Grid container spacing={2}>
            <Grid size={{ xs: 12, sm: 4 }}>
              <TextField
                fullWidth
                label={translate('platform_settings.base_fee_dollar')}
                type="number"
                value={formData.base_fee}
                onChange={(e) => setFormData({ ...formData, base_fee: parseFloat(e.target.value) })}
                InputProps={{ startAdornment: <Box sx={{ mr: 1 }}>$</Box> }}
                size={isMobile ? 'small' : 'medium'}
              />
            </Grid>
            <Grid size={{ xs: 12, sm: 4 }}>
              <TextField
                fullWidth
                label={translate('platform_settings.included_users_plan')}
                type="number"
                value={formData.included_users}
                onChange={(e) => setFormData({ ...formData, included_users: parseInt(e.target.value) })}
                size={isMobile ? 'small' : 'medium'}
              />
            </Grid>
            <Grid size={{ xs: 12, sm: 4 }}>
              <TextField
                fullWidth
                label={translate('platform_settings.overage_fee_dollar')}
                type="number"
                value={formData.overage_fee}
                onChange={(e) => setFormData({ ...formData, overage_fee: parseFloat(e.target.value) })}
                InputProps={{ startAdornment: <Box sx={{ mr: 1 }}>$</Box> }}
                size={isMobile ? 'small' : 'medium'}
              />
            </Grid>
          </Grid>

          <Divider sx={{ my: 1 }} />
          <Typography
            variant="h6"
            sx={{ fontWeight: 600, fontSize: isMobile ? '0.95rem' : undefined }}
          >
            {translate('platform_settings.plan_features')}
          </Typography>
          <Alert severity="info" sx={{ fontSize: isMobile ? '0.75rem' : undefined }}>
            {translate('platform_settings.plan_features_help')}
          </Alert>

          <Grid container spacing={2}>
            {AVAILABLE_FEATURES.map((feature) => {
              const featureData = features.find((f) => f.id === feature.id);
              const isEnabled = featureData !== undefined;

              return (
                <Grid size={{ xs: 12, sm: 6 }} key={feature.id}>
                  <Card
                    variant="outlined"
                    sx={{
                      borderColor: isEnabled ? 'primary.main' : 'divider',
                      bgcolor: isEnabled ? 'primary.50' : 'background.paper',
                    }}
                  >
                    <CardContent sx={{ p: isMobile ? 1.5 : 2 }}>
                      <Stack
                        direction="row"
                        alignItems="center"
                        justifyContent="space-between"
                        mb={isEnabled && feature.type === 'number' ? 1 : 0}
                      >
                        <FormControlLabel
                          control={
                            <Switch
                              checked={isEnabled}
                              onChange={(e) => {
                                if (e.target.checked) {
                                  handleFeatureToggle(
                                    feature.id,
                                    feature.type === 'boolean' ? true : 1
                                  );
                                } else {
                                  handleFeatureToggle(feature.id, false);
                                }
                              }}
                              size={isMobile ? 'small' : 'medium'}
                            />
                          }
                          label={translate(`platform_settings.${feature.translationKey}`)}
                          sx={{ mr: 0, '.MuiFormControlLabel-label': { fontSize: isMobile ? '0.8rem' : undefined } }}
                        />
                      </Stack>
                      {isEnabled && feature.type === 'number' && (
                        <TextField
                          fullWidth
                          size="small"
                          type="number"
                          value={featureData?.value || 0}
                          onChange={(e) => handleFeatureToggle(feature.id, parseInt(e.target.value))}
                        />
                      )}
                    </CardContent>
                  </Card>
                </Grid>
              );
            })}
          </Grid>

          <Divider sx={{ my: 1 }} />
          <FormControlLabel
            control={
              <Switch
                checked={formData.is_active}
                onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
              />
            }
            label={translate('platform_settings.active')}
          />
          <Typography variant="body2" sx={{ color: 'text.secondary', mt: 0.5, fontSize: isMobile ? '0.75rem' : undefined }}>
            {translate('platform_settings.active_help')}
          </Typography>
        </Stack>
      </DialogContent>
      <DialogActions sx={{ flexWrap: 'wrap', gap: 1, px: isMobile ? 2 : 3, pb: isMobile ? 2 : 3 }}>
        <Button onClick={onClose} startIcon={<Cancel />} size={isMobile ? 'small' : 'medium'}>
          {translate('platform_settings.cancel')}
        </Button>
        <Button
          onClick={handleSave}
          variant="contained"
          startIcon={<Save />}
          size={isMobile ? 'small' : 'medium'}
          sx={{
            background: 'linear-gradient(135deg, #1e3a8a 0%, #1e40af 100%)',
            '&:hover': {
              background: 'linear-gradient(135deg, #1e40af 0%, #2563eb 100%)',
            },
          }}
        >
          {translate('platform_settings.save_plan')}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export const PricingPlans = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const translate = useTranslate();
  const [editorOpen, setEditorOpen] = useState(false);
  const [selectedPlan, setSelectedPlan] = useState<PricingPlan | null>(null);

  const handleEdit = (plan: PricingPlan) => {
    setSelectedPlan(plan);
    setEditorOpen(true);
  };

  const handleAdd = () => {
    setSelectedPlan(null);
    setEditorOpen(true);
  };

  const handleDelete = (plan: PricingPlan) => {
    if (window.confirm(translate('platform_settings.delete_confirm', { name: plan.name }))) {
      console.log('Delete plan:', plan);
    }
  };

  const handleSave = (plan: PricingPlan) => {
    console.log('Save plan:', plan);
    setEditorOpen(false);
    setSelectedPlan(null);
  };

  return (
    <Box sx={{ py: isMobile ? 2 : 3 }}>
      <Box
        sx={{
          mb: 3,
          display: 'flex',
          flexDirection: isMobile ? 'column' : 'row',
          justifyContent: 'space-between',
          alignItems: isMobile ? 'stretch' : 'center',
          gap: isMobile ? 1.5 : 2,
        }}
      >
        <Box>
          <Typography
            variant={isMobile ? 'h6' : 'h5'}
            sx={{ fontWeight: 700, mb: 0.5 }}
          >
            {translate('platform_settings.pricing_plans_title')}
          </Typography>
          <Typography
            variant="body2"
            sx={{ color: 'text.secondary', display: { xs: 'none', sm: 'block' } }}
          >
            {translate('platform_settings.pricing_plans_subtitle')}
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleAdd}
          size={isMobile ? 'small' : 'medium'}
          sx={{
            background: 'linear-gradient(135deg, #1e3a8a 0%, #1e40af 100%)',
            '&:hover': {
              background: 'linear-gradient(135deg, #1e40af 0%, #2563eb 100%)',
            },
            whiteSpace: 'nowrap',
          }}
        >
          {translate('platform_settings.add_new_plan')}
        </Button>
      </Box>

      <Alert severity="info" sx={{ mb: 3, fontSize: isMobile ? '0.75rem' : undefined }}>
        <AlertTitle sx={{ fontSize: isMobile ? '0.8rem' : undefined }}>
          {translate('platform_settings.plan_features_title')}
        </AlertTitle>
        <Typography variant="body2">
          {translate('platform_settings.plan_features_message')}
        </Typography>
      </Alert>

      <PricingPlansList onEdit={handleEdit} onDelete={handleDelete} />

      <PlanEditor
        open={editorOpen}
        onClose={() => {
          setEditorOpen(false);
          setSelectedPlan(null);
        }}
        plan={selectedPlan}
        onSave={handleSave}
      />
    </Box>
  );
};

PricingPlans.displayName = 'PricingPlans';
