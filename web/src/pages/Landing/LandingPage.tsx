import {
  Box,
  Container,
  Typography,
  Button,
  Card,
  CardContent,
  Stack,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Alert,
  AlertTitle,
  Paper,
  Avatar,
} from '@mui/material';
import { useEffect } from 'react';
import { SelectChangeEvent } from '@mui/material';
import {
  Rocket,
  Speed,
  Support,
  CheckCircle,
  TrendingUp,
  CloudDone,
  Router,
  Shield,
  Send,
  Person,
  Dashboard as DashboardIcon,
} from '@mui/icons-material';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLandingTranslate } from '../../contexts/LandingI18nProvider';
import { LandingLanguageSwitcher } from '../../components/LandingLanguageSwitcher';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import IconButton from '@mui/material/IconButton';
import CloseIcon from '@mui/icons-material/Close';
import { fetchUtils } from 'react-admin';

const features = [
  {
    icon: <Speed />,
    titleKey: 'landing.feature_lightning_fast',
    descriptionKey: 'landing.feature_lightning_fast_desc',
  },
  {
    icon: <CloudDone />,
    titleKey: 'landing.feature_multi_tenant',
    descriptionKey: 'landing.feature_multi_tenant_desc',
  },
  {
    icon: <Shield />,
    titleKey: 'landing.feature_security',
    descriptionKey: 'landing.feature_security_desc',
  },
  {
    icon: <TrendingUp />,
    titleKey: 'landing.feature_scaling',
    descriptionKey: 'landing.feature_scaling_desc',
  },
  {
    icon: <Router />,
    titleKey: 'landing.feature_network',
    descriptionKey: 'landing.feature_network_desc',
  },
  {
    icon: <Support />,
    titleKey: 'landing.feature_monitoring',
    descriptionKey: 'landing.feature_monitoring_desc',
  },
];

const pricingTiers = [
  {
    nameKey: 'landing.starter_plan',
    users: '1,000',
    sessions: '500',
    devices: '10',
    storage: '10 GB',
    price: '$99',
    period: 'landing.starter_period',
    features: [
      'landing.starter_users',
      'landing.starter_sessions',
      'landing.starter_devices',
      'landing.starter_storage',
      'landing.starter_support',
      'landing.starter_backups',
      'landing.starter_monitoring',
    ],
    recommended: false,
  },
  {
    nameKey: 'landing.professional_plan',
    users: '5,000',
    sessions: '1,500',
    devices: '50',
    storage: '50 GB',
    price: '$299',
    period: 'landing.professional_period',
    features: [
      'landing.professional_users',
      'landing.professional_sessions',
      'landing.professional_devices',
      'landing.professional_storage',
      'landing.professional_support',
      'landing.professional_backups',
      'landing.professional_monitoring',
      'landing.professional_branding',
      'landing.professional_api',
    ],
    recommended: true,
  },
  {
    nameKey: 'landing.enterprise_plan',
    users: '25,000',
    sessions: '5,000',
    devices: '200',
    storage: '500 GB',
    price: '$899',
    period: 'landing.enterprise_period',
    features: [
      'landing.enterprise_users',
      'landing.enterprise_sessions',
      'landing.enterprise_devices',
      'landing.enterprise_storage',
      'landing.enterprise_support',
      'landing.enterprise_backups',
      'landing.enterprise_monitoring',
      'landing.enterprise_branding',
      'landing.enterprise_api',
      'landing.enterprise_manager',
      'landing.enterprise_sla',
    ],
    recommended: false,
  },
];

const stats = [
  { value: '100+', labelKey: 'landing.providers_label' },
  { value: '500K+', labelKey: 'landing.users_label' },
  { value: '99.99%', labelKey: 'landing.uptime_label' },
  { value: '24/7', labelKey: 'landing.support_label' },
];

interface RegistrationFormProps {
  selectedPlan?: string;
  onClose?: () => void;
}

const RegistrationForm = ({ selectedPlan, onClose }: RegistrationFormProps) => {
  const { translate } = useLandingTranslate();
  const [submitted, setSubmitted] = useState(false);
  const [formData, setFormData] = useState({
    company_name: '',
    contact_name: '',
    email: '',
    phone: '',
    business_type: '',
    expected_users: '',
    message: '',
    selected_plan: selectedPlan || '',
  });

  useEffect(() => {
    if (selectedPlan) {
      setFormData(prev => ({ ...prev, selected_plan: selectedPlan }));
    }
  }, [selectedPlan]);

  const handleTextChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const target = e.target;
    setFormData({ ...formData, [target.name]: target.value });
  };

  const handleSelectChange = (e: SelectChangeEvent<string>) => {
    const target = e.target;
    setFormData({ ...formData, [target.name]: target.value });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      // Call the provider registration API
      const { fetchJson } = fetchUtils;
      const apiUrl = `${window.location.origin}/api/v1/providers/registrations`;

      // Map form data to API format
      const requestData = {
        company_name: formData.company_name,
        contact_name: formData.contact_name,
        email: formData.email,
        phone: formData.phone,
        address: '',
        business_type: formData.business_type,
        expected_users: parseInt(formData.expected_users) || 1000,
        expected_nas: 10,
        country: '',
        message: formData.message,
      };

      const options: RequestInit = {
        method: 'POST',
        body: JSON.stringify(requestData),
        headers: new Headers({
          'Content-Type': 'application/json',
        }),
      };

      await fetchJson(apiUrl, options);

      setSubmitted(true);
      // Close modal after 2 seconds
      setTimeout(() => {
        onClose?.();
      }, 2000);
    } catch (error) {
      console.error('Registration failed:', error);
      // Show error message to user
      alert(translate('landing.registration_error') || 'Registration failed. Please try again.');
    }
  };

  if (submitted) {
    return (
      <Alert severity="success" sx={{ mt: 3 }}>
        <AlertTitle>{translate('landing.registration_success')}</AlertTitle>
        {translate('landing.registration_success_desc')}
      </Alert>
    );
  }

  return (
    <Box
      component="form"
      onSubmit={handleSubmit}
      sx={{
        maxWidth: 600,
        mx: 'auto',
        p: 4,
      }}
    >
      <Typography variant="h5" sx={{ fontWeight: 700, mb: 3, textAlign: 'center' }}>
        {translate('landing.register_title')}
      </Typography>

      <Stack spacing={3}>
        <TextField
          fullWidth
          label={translate('landing.company_name')}
          name="company_name"
          value={formData.company_name}
          onChange={handleTextChange}
          required
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
            },
          }}
        />

        <TextField
          fullWidth
          label={translate('landing.contact_name')}
          name="contact_name"
          value={formData.contact_name}
          onChange={handleTextChange}
          required
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
            },
          }}
        />

        {selectedPlan && (
          <TextField
            fullWidth
            label={translate('landing.selected_plan')}
            name="selected_plan"
            value={translate(selectedPlan)}
            InputProps={{
              readOnly: true,
            }}
            sx={{
              '& .MuiOutlinedInput-root': {
                borderRadius: 2,
                backgroundColor: 'rgba(30, 58, 138, 0.04)',
              },
            }}
          />
        )}

        <TextField
          fullWidth
          label={translate('landing.email_address')}
          name="email"
          type="email"
          value={formData.email}
          onChange={handleTextChange}
          required
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
            },
          }}
        />

        <TextField
          fullWidth
          label={translate('landing.phone_number')}
          name="phone"
          value={formData.phone}
          onChange={handleTextChange}
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
            },
          }}
        />

        <FormControl fullWidth required>
          <InputLabel>{translate('landing.business_type')}</InputLabel>
          <Select
            label={translate('landing.business_type')}
            name="business_type"
            value={formData.business_type}
            onChange={handleSelectChange}
            sx={{ borderRadius: 2 }}
          >
            <MenuItem value="isp">{translate('landing.business_type_isp')}</MenuItem>
            <MenuItem value="wisp">{translate('landing.business_type_wisp')}</MenuItem>
            <MenuItem value="hotel">{translate('landing.business_type_hotel')}</MenuItem>
            <MenuItem value="enterprise">{translate('landing.business_type_enterprise')}</MenuItem>
            <MenuItem value="event">{translate('landing.business_type_venue')}</MenuItem>
            <MenuItem value="other">{translate('landing.business_type_other')}</MenuItem>
          </Select>
        </FormControl>

        <FormControl fullWidth required>
          <InputLabel>{translate('landing.expected_users')}</InputLabel>
          <Select
            label={translate('landing.expected_users')}
            name="expected_users"
            value={formData.expected_users}
            onChange={handleSelectChange}
            sx={{ borderRadius: 2 }}
          >
            <MenuItem value="1000">{translate('landing.users_range_100_500')}</MenuItem>
            <MenuItem value="5000">{translate('landing.users_range_500_1000')}</MenuItem>
            <MenuItem value="10000">{translate('landing.users_range_1000_5000')}</MenuItem>
            <MenuItem value="50000">{translate('landing.users_range_5000_10000')}</MenuItem>
          </Select>
        </FormControl>

        <TextField
          fullWidth
          label={translate('landing.additional_info')}
          name="message"
          value={formData.message}
          onChange={handleTextChange}
          multiline
          rows={4}
          placeholder={translate('landing.remark_placeholder')}
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
            },
          }}
        />

        <Button
          type="submit"
          variant="contained"
          size="large"
          fullWidth
          startIcon={<Send />}
          sx={{
            py: 1.5,
            borderRadius: 2,
            background: 'linear-gradient(135deg, #1e3a8a 0%, #1e40af 100%)',
            fontSize: 16,
            fontWeight: 600,
            textTransform: 'none',
            '&:hover': {
              background: 'linear-gradient(135deg, #1e40af 0%, #2563eb 100%)',
            },
          }}
        >
          {translate('landing.submit_registration')}
        </Button>
      </Stack>
    </Box>
  );
};

export const LandingPage = () => {
  const { translate } = useLandingTranslate();
  const navigate = useNavigate();
  const [selectedPlan, setSelectedPlan] = useState<string | null>(null);
  const [registrationOpen, setRegistrationOpen] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  // Check authentication status on mount
  useEffect(() => {
    const token = localStorage.getItem('token');
    setIsAuthenticated(!!token);
  }, []);

  const handlePlanClick = (planKey: string) => {
    setSelectedPlan(planKey);
  };

  // Use handlePlanClick to mark it as used
  void handlePlanClick;

  return (
    <Box sx={{ bgcolor: 'background.default' }}>
      {/* Navigation Bar */}
      <Box
        sx={{
          position: 'sticky',
          top: 0,
          zIndex: 1000,
          bgcolor: 'rgba(15, 23, 42, 0.95)',
          backdropFilter: 'blur(10px)',
          borderBottom: '1px solid rgba(255,255,255,0.1)',
        }}
      >
        <Container maxWidth="lg">
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', py: 2 }}>
            <Typography
              variant="h6"
              sx={{
                color: '#10b981',
                fontWeight: 700,
                fontSize: '1.5rem',
                cursor: 'pointer',
                textDecoration: 'none',
              }}
              onClick={() => navigate('/')}
            >
              {translate('app.title')}
            </Typography>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <LandingLanguageSwitcher />
              {isAuthenticated ? (
                <Button
                  variant="contained"
                  startIcon={<DashboardIcon />}
                  endIcon={
                    <Avatar
                      sx={{
                        width: 24,
                        height: 24,
                        bgcolor: 'rgba(255,255,255,0.2)',
                        fontSize: 12,
                      }}
                    >
                      <Person sx={{ fontSize: 14 }} />
                    </Avatar>
                  }
                  onClick={() => navigate('/')}
                  sx={{
                    borderRadius: 2,
                    textTransform: 'none',
                    fontWeight: 600,
                    background: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
                    '&:hover': {
                      background: 'linear-gradient(135deg, #059669 0%, #047857 100%)',
                    },
                  }}
                >
                  {translate('landing.dashboard')}
                </Button>
              ) : (
                <Button
                  variant="outlined"
                  startIcon={<Person />}
                  onClick={() => navigate('/login')}
                  sx={{
                    borderRadius: 2,
                    textTransform: 'none',
                    fontWeight: 600,
                    color: 'white',
                    borderColor: 'rgba(255,255,255,0.3)',
                    '&:hover': {
                      borderColor: 'white',
                      backgroundColor: 'rgba(255,255,255,0.05)',
                    },
                  }}
                >
                  {translate('landing.login')}
                </Button>
              )}
            </Box>
          </Box>
        </Container>
      </Box>

      {/* Hero Section */}
      <Box
        sx={{
          background: 'linear-gradient(135deg, #0f172a 0%, #1e3a8a 50%, #1e40af 100%)',
          color: 'white',
          py: 20,
          position: 'relative',
          overflow: 'hidden',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background:
              'radial-gradient(circle at 20% 50%, rgba(59, 130, 246, 0.1) 0%, transparent 50%), radial-gradient(circle at 80% 80%, rgba(16, 185, 129, 0.1) 0%, transparent 50%)',
          },
        }}
      >
        <Container maxWidth="xl" sx={{ position: 'relative', zIndex: 1 }}>
          {/* Title Section */}
          <Box sx={{ textAlign: 'center', mb: 8 }}>
            <Typography
              variant="overline"
              sx={{
                color: '#10b981',
                fontWeight: 600,
                letterSpacing: 2,
                textTransform: 'uppercase',
                mb: 2,
                display: 'block',
              }}
            >
              {translate('app.title')}
            </Typography>
            <Typography
              variant="h2"
              sx={{
                fontWeight: 800,
                fontSize: { xs: '2.5rem', md: '4rem' },
                mb: 3,
                lineHeight: 1.2,
              }}
            >
              {translate('landing.hero_title')}
            </Typography>
            <Typography
              variant="h5"
              sx={{
                color: '#10b981',
                fontWeight: 600,
                fontSize: { xs: '1.2rem', md: '1.5rem' },
                mb: 4,
              }}
            >
              {translate('landing.hero_subtitle')}
            </Typography>
            <Typography
              variant="h6"
              sx={{
                color: 'rgba(255,255,255,0.8)',
                mb: 6,
                fontSize: { xs: '1.1rem', md: '1.25rem' },
                maxWidth: 800,
                mx: 'auto',
              }}
            >
              {translate('app.subtitle')}
            </Typography>
            <Stack direction="row" spacing={3} justifyContent="center" sx={{ mb: 8 }}>
              <Button
                variant="contained"
                size="large"
                startIcon={<Rocket />}
                onClick={() => navigate('/login')}
                sx={{
                  px: 5,
                  py: 2,
                  borderRadius: 2,
                  background: '#10b981',
                  fontSize: 17,
                  fontWeight: 600,
                  textTransform: 'none',
                  '&:hover': { background: '#059669' },
                }}
              >
                {translate('landing.get_started')}
              </Button>
              <Button
                variant="outlined"
                size="large"
                onClick={() => document.getElementById('features')?.scrollIntoView({ behavior: 'smooth' })}
                sx={{
                  px: 5,
                  py: 2,
                  borderRadius: 2,
                  borderColor: 'rgba(255,255,255,0.3)',
                  color: 'white',
                  fontSize: 17,
                  fontWeight: 600,
                  textTransform: 'none',
                  '&:hover': {
                    borderColor: 'rgba(255,255,255,0.5)',
                    bgcolor: 'rgba(255,255,255,0.05)',
                  },
                }}
              >
                {translate('landing.learn_more')}
              </Button>
            </Stack>
          </Box>

          {/* Stats and Get Started Card in Same Row */}
          <Box sx={{ display: 'flex', flexDirection: { xs: 'column', md: 'row' }, gap: 4, alignItems: 'stretch' }}>
            {/* Stats */}
            <Box sx={{ flex: 7, display: 'flex', flexDirection: 'column' }}>
              <Stack direction="row" spacing={3} justifyContent="space-around" sx={{ flexWrap: 'wrap', height: '100%', alignItems: 'center' }}>
                {stats.map((stat, index) => (
                  <Box key={index} sx={{ textAlign: 'center', flex: 1, minWidth: { xs: '45%', md: 'auto' } }}>
                    <Typography
                      variant="h3"
                      sx={{
                        fontWeight: 800,
                        color: '#10b981',
                        fontSize: { xs: '2rem', md: '2.5rem' },
                        mb: 0.5,
                      }}
                    >
                      {stat.value}
                    </Typography>
                    <Typography
                      variant="body1"
                      sx={{ color: 'rgba(255,255,255,0.85)', fontSize: { xs: '0.9rem', md: '1rem' } }}
                    >
                      {translate(stat.labelKey)}
                    </Typography>
                  </Box>
                ))}
              </Stack>
            </Box>

            {/* Get Started Card */}
            <Box sx={{ flex: 5, display: 'flex', flexDirection: 'column' }}>
              <Box
                sx={{
                  background: 'rgba(255,255,255,0.05)',
                  backdropFilter: 'blur(10px)',
                  borderRadius: 3,
                  p: { xs: 3, md: 4 },
                  border: '1px solid rgba(255,255,255,0.1)',
                  height: '100%',
                  flexGrow: 1,
                }}
              >
                <Typography
                  variant="h5"
                  sx={{ fontWeight: 700, mb: 2, color: 'white', fontSize: { xs: '1.25rem', md: '1.5rem' } }}
                >
                  {translate('landing.get_started')}
                </Typography>
                <Typography
                  variant="body2"
                  sx={{ color: 'rgba(255,255,255,0.7)', mb: 3, fontSize: { xs: '0.9rem', md: '1rem' } }}
                >
                  {translate('landing.features_subtitle')}
                </Typography>
                <Stack spacing={2.5}>
                  <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 2 }}>
                    <CheckCircle sx={{ color: '#10b981', fontSize: 22, mt: 0.3, flexShrink: 0 }} />
                    <Typography variant="body2" sx={{ color: 'rgba(255,255,255,0.9)', flex: 1, fontSize: { xs: '0.9rem', md: '1rem' } }}>
                      {translate('landing.feature_monitoring')}
                    </Typography>
                  </Box>
                  <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 2 }}>
                    <CheckCircle sx={{ color: '#10b981', fontSize: 22, mt: 0.3, flexShrink: 0 }} />
                    <Typography variant="body2" sx={{ color: 'rgba(255,255,255,0.9)', flex: 1, fontSize: { xs: '0.9rem', md: '1rem' } }}>
                      {translate('landing.feature_multi_tenant')}
                    </Typography>
                  </Box>
                  <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 2 }}>
                    <CheckCircle sx={{ color: '#10b981', fontSize: 22, mt: 0.3, flexShrink: 0 }} />
                    <Typography variant="body2" sx={{ color: 'rgba(255,255,255,0.9)', flex: 1, fontSize: { xs: '0.9rem', md: '1rem' } }}>
                      {translate('landing.feature_security')}
                    </Typography>
                  </Box>
                </Stack>
              </Box>
            </Box>
          </Box>
        </Container>
      </Box>

      {/* Features Section */}
      <Box id="features" sx={{ py: { xs: 10, md: 15 }, bgcolor: 'background.paper' }}>
        <Container maxWidth="xl">
          <Typography
            variant="h4"
            sx={{
              fontWeight: 700,
              textAlign: 'center',
              mb: 2,
              fontSize: { xs: '1.75rem', md: '2.25rem' },
            }}
          >
            {translate('landing.features_title')}
          </Typography>
          <Typography
            variant="h6"
            sx={{
              color: 'text.secondary',
              textAlign: 'center',
              mb: { xs: 6, md: 10 },
              maxWidth: 700,
              mx: 'auto',
              fontSize: { xs: '1rem', md: '1.15rem' },
            }}
          >
            {translate('landing.features_subtitle')}
          </Typography>

          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 4, justifyContent: 'center' }}>
            {features.map((feature, index) => (
              <Box key={index} sx={{ flex: { xs: '1 1 100%', sm: '1 1 calc(50% - 16px)', lg: '1 1 calc(33.333% - 22px)' }, minWidth: { xs: '100%', sm: '280px', lg: '320px' } }}>
                <Card
                  sx={{
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column',
                    background: 'linear-gradient(135deg, rgba(255,255,255,0.98) 0%, rgba(248,250,252,0.98) 100%)',
                    border: '1px solid rgba(148, 163, 184, 0.15)',
                    borderRadius: 3,
                    transition: 'all 0.3s ease',
                    '&:hover': {
                      transform: 'translateY(-6px)',
                      boxShadow: '0 20px 40px -10px rgba(0, 0, 0, 0.15)',
                      borderColor: 'rgba(16, 185, 129, 0.3)',
                    },
                  }}
                >
                  <CardContent sx={{ p: { xs: 3, md: 4 }, flexGrow: 1, display: 'flex', flexDirection: 'column' }}>
                    <Box
                      sx={{
                        p: 2,
                        borderRadius: 2,
                        bgcolor: 'rgba(16, 185, 129, 0.1)',
                        color: '#10b981',
                        width: 'fit-content',
                        mb: 3,
                      }}
                    >
                      {feature.icon}
                    </Box>
                    <Typography
                      variant="h6"
                      sx={{ fontWeight: 700, mb: 2, fontSize: { xs: '1.15rem', md: '1.3rem' } }}
                    >
                      {translate(feature.titleKey)}
                    </Typography>
                    <Typography
                      variant="body2"
                      sx={{ color: 'text.secondary', lineHeight: 1.7, fontSize: { xs: '0.9rem', md: '1rem' }, flexGrow: 1 }}
                    >
                      {translate(feature.descriptionKey)}
                    </Typography>
                  </CardContent>
                </Card>
              </Box>
            ))}
          </Box>
        </Container>
      </Box>

      {/* Pricing Section */}
      <Box sx={{ py: { xs: 10, md: 15 }, bgcolor: 'background.default' }}>
        <Container maxWidth="xl">
          <Typography
            variant="h4"
            sx={{
              fontWeight: 700,
              textAlign: 'center',
              mb: 2,
              fontSize: { xs: '1.75rem', md: '2.25rem' },
            }}
          >
            {translate('landing.pricing_title')}
          </Typography>
          <Typography
            variant="h6"
            sx={{
              color: 'text.secondary',
              textAlign: 'center',
              mb: { xs: 6, md: 10 },
              maxWidth: 700,
              mx: 'auto',
              fontSize: { xs: '1rem', md: '1.15rem' },
            }}
          >
            {translate('landing.pricing_subtitle')}
          </Typography>

          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 4, justifyContent: 'center' }}>
            {pricingTiers.map((tier, index) => (
              <Box key={index} sx={{ flex: { xs: '1 1 100%', md: '1 1 calc(50% - 16px)', lg: '1 1 calc(33.333% - 22px)' }, minWidth: { xs: '100%', md: '350px', lg: '380px' } }}>
                <Card
                  raised={tier.recommended || selectedPlan === tier.nameKey}
                  onClick={() => setSelectedPlan(tier.nameKey)}
                  sx={{
                    height: '100%',
                    borderRadius: 3,
                    border: selectedPlan === tier.nameKey
                      ? '3px solid #1e3a8a'
                      : tier.recommended
                      ? '2px solid #10b981'
                      : '1px solid rgba(148, 163, 184, 0.1)',
                    position: 'relative',
                    transition: 'all 0.3s ease',
                    cursor: 'pointer',
                    '&:hover': {
                      transform: 'translateY(-4px)',
                      boxShadow: selectedPlan === tier.nameKey
                        ? '0 20px 40px -10px rgba(30, 58, 138, 0.4)'
                        : tier.recommended
                        ? '0 20px 40px -10px rgba(16, 185, 129, 0.3)'
                        : '0 12px 24px -10px rgba(0, 0, 0, 0.12)',
                    },
                  }}
                >
                  {tier.recommended && (
                    <Box
                      sx={{
                        position: 'absolute',
                        top: -12,
                        left: '50%',
                        transform: 'translateX(-50%)',
                        bgcolor: '#10b981',
                        color: 'white',
                        px: 3,
                        py: 1,
                        borderRadius: 20,
                        fontWeight: 600,
                        fontSize: 14,
                      }}
                    >
                      {translate('landing.professional_recommended')}
                    </Box>
                  )}
                  <CardContent sx={{ p: 4 }}>
                    <Typography
                      variant="h5"
                      sx={{ fontWeight: 700, mb: 1, textAlign: 'center' }}
                    >
                      {translate(tier.nameKey)}
                    </Typography>
                    <Box sx={{ textAlign: 'center', mb: 3 }}>
                      <Typography
                        variant="h3"
                        sx={{ fontWeight: 800, color: '#1e3a8a', fontSize: '2.5rem' }}
                      >
                        {tier.price}
                      </Typography>
                      <Typography
                        variant="body1"
                        sx={{ color: 'text.secondary', fontWeight: 500 }}
                      >
                        {translate(tier.period)}
                      </Typography>
                    </Box>
                    <Stack spacing={2} sx={{ mb: 4 }}>
                      <Typography
                        variant="body2"
                        sx={{ color: 'text.secondary', textAlign: 'center' }}
                      >
                        {translate('landing.pricing_users', { count: tier.users })} • {translate('landing.pricing_sessions', { count: tier.sessions })}
                      </Typography>
                      <Typography
                        variant="body2"
                        sx={{ color: 'text.secondary', textAlign: 'center' }}
                      >
                        {translate('landing.pricing_devices', { count: tier.devices })} • {tier.storage}
                      </Typography>
                    </Stack>
                    <Button
                      variant={selectedPlan === tier.nameKey ? 'contained' : tier.recommended ? 'contained' : 'outlined'}
                      color={selectedPlan === tier.nameKey ? 'primary' : tier.recommended ? 'success' : 'primary'}
                      fullWidth
                      sx={{
                        mb: 3,
                        py: 1.5,
                        borderRadius: 2,
                        fontWeight: 600,
                        textTransform: 'none',
                        bgcolor: selectedPlan === tier.nameKey ? '#1e3a8a' : tier.recommended ? '#10b981' : 'inherit',
                        '&:hover': {
                          bgcolor: selectedPlan === tier.nameKey ? '#1e40af' : tier.recommended ? '#059669' : 'rgba(30, 58, 138, 0.04)',
                        },
                      }}
                      onClick={(e) => {
                        e.stopPropagation();
                        setSelectedPlan(tier.nameKey);
                        setRegistrationOpen(true);
                      }}
                    >
                      {translate(tier.recommended ? 'landing.start_free_trial' : 'landing.get_started')}
                    </Button>
                    <Stack spacing={1.5}>
                      {tier.features.map((feature, featureIndex) => (
                        <Box
                          key={featureIndex}
                          sx={{ display: 'flex', alignItems: 'center', gap: 1 }}
                        >
                          <CheckCircle
                            sx={{ color: '#10b981', fontSize: 16, flexShrink: 0 }}
                          />
                          <Typography
                            variant="body2"
                            sx={{ color: 'text.secondary', fontSize: 14 }}
                          >
                            {translate(feature)}
                          </Typography>
                        </Box>
                      ))}
                    </Stack>
                  </CardContent>
                </Card>
              </Box>
            ))}
          </Box>
        </Container>
      </Box>

      {/* Registration Section */}
      <Box id="registration" sx={{ py: 15, bgcolor: 'background.paper' }}>
        <Container maxWidth="lg">
          <Box sx={{ maxWidth: 800, mx: 'auto', textAlign: 'center', mb: 8 }}>
            <Typography
              variant="h4"
              sx={{
                fontWeight: 700,
                mb: 2,
                fontSize: '2rem',
              }}
            >
              {translate('landing.register_title')}
            </Typography>
            <Typography
              variant="h6"
              sx={{ color: 'text.secondary' }}
            >
              {translate('landing.register_subtitle')}
            </Typography>
          </Box>

          {!selectedPlan ? (
            <Alert severity="info" sx={{ maxWidth: 600, mx: 'auto' }}>
              <AlertTitle>{translate('landing.select_plan_title')}</AlertTitle>
              {translate('landing.select_plan_desc')}
            </Alert>
          ) : (
            <Paper
              sx={{
                maxWidth: 800,
                mx: 'auto',
                borderRadius: 3,
                boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
              }}
            >
              <RegistrationForm selectedPlan={selectedPlan} />
            </Paper>
          )}
        </Container>
      </Box>

      {/* Footer */}
      <Box
        sx={{
          bgcolor: '#0f172a',
          color: 'white',
          py: 8,
        }}
      >
        <Container maxWidth="lg">
          <Typography variant="h6" sx={{ fontWeight: 700, mb: 2 }}>
            {translate('app.title')}
          </Typography>
          <Typography variant="body2" sx={{ color: 'rgba(255,255,255,0.6)' }}>
            {translate('app.rights')}
          </Typography>
        </Container>
      </Box>

      {/* Registration Modal */}
      <Dialog
        open={registrationOpen}
        onClose={() => setRegistrationOpen(false)}
        maxWidth="md"
        fullWidth
        PaperProps={{
          sx: {
            borderRadius: 3,
            maxHeight: '90vh',
          },
        }}
      >
        <DialogTitle sx={{ m: 0, p: 3, bgcolor: 'primary.main', color: 'white' }}>
          <Typography variant="h5" sx={{ fontWeight: 700 }}>
            {translate('landing.register_title')}
          </Typography>
          {selectedPlan && (
            <Typography variant="body2" sx={{ mt: 1, opacity: 0.9 }}>
              {translate('landing.selected_plan')}: {translate(selectedPlan)}
            </Typography>
          )}
          <IconButton
            aria-label="close"
            onClick={() => setRegistrationOpen(false)}
            sx={{
              position: 'absolute',
              right: 8,
              top: 8,
              color: 'white',
            }}
          >
            <CloseIcon />
          </IconButton>
        </DialogTitle>
        <DialogContent sx={{ p: 3 }}>
          <RegistrationForm selectedPlan={selectedPlan || undefined} onClose={() => setRegistrationOpen(false)} />
        </DialogContent>
      </Dialog>
    </Box>
  );
};

LandingPage.displayName = 'LandingPage';
