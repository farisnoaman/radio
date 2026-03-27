import { useEffect, useState } from 'react';
import { Box, Card, CardContent, Typography, Stack, LinearProgress, alpha, useTheme, Button, Tooltip, CircularProgress } from '@mui/material';
import { useNotify } from 'react-admin';
import WorkspacePremiumIcon from '@mui/icons-material/WorkspacePremium';
import RedeemIcon from '@mui/icons-material/Redeem';
import StarsIcon from '@mui/icons-material/Stars';
import EmojiEventsIcon from '@mui/icons-material/EmojiEvents';

export const LoyaltyStatusCard = () => {
    const notify = useNotify();
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';
    
    const [loyaltyData, setLoyaltyData] = useState<any>(null);
    const [loading, setLoading] = useState(true);
    const [redeeming, setRedeeming] = useState(false);

    const fetchLoyalty = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('/api/v1/portal/loyalty', {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            const result = await response.json();
            if (result.data) {
                setLoyaltyData(result.data);
            }
        } catch (error) {
            console.error('Failed to fetch loyalty status:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchLoyalty();
        const interval = setInterval(fetchLoyalty, 60000); // refresh every minute
        return () => clearInterval(interval);
    }, []);

    const handleRedeem = async (points: number) => {
        setRedeeming(true);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('/api/v1/portal/loyalty/redeem', {
                method: 'POST',
                headers: { 
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ points })
            });
            const result = await response.json();
            
            if (response.ok) {
                notify(result.data?.message || 'Points redeemed perfectly!', { type: 'success' });
                fetchLoyalty();
            } else {
                notify(result.message || 'Failed to redeem points', { type: 'error' });
            }
        } catch (error) {
            notify('Failed to connect to server', { type: 'error' });
        } finally {
            setRedeeming(false);
        }
    };

    if (loading) {
        return (
            <Card sx={{ borderRadius: 6, mb: 3, border: '1px solid', borderColor: 'divider', bgcolor: 'background.paper' }}>
                <CardContent sx={{ p: 3, display: 'flex', justifyContent: 'center' }}>
                    <CircularProgress size={24} />
                </CardContent>
            </Card>
        );
    }

    if (!loyaltyData || !loyaltyData.profile) {
        return null;
    }

    const { profile, rules } = loyaltyData;
    
    // Determine the next goal based on the first rule for simplicity, or hardcoded for the demo
    // The main rule (points threshold)
    const nextRule = rules && rules.length > 0 ? rules[0] : null;
    let dataProgress = 0;
    if (nextRule && nextRule.data_threshold > 0) {
        dataProgress = Math.min(100, (profile.milestone_data_used / nextRule.data_threshold) * 100);
    }

    // Badge formatting
    let badgeColor = theme.palette.text.secondary;
    let BadgeIcon = StarsIcon;
    let bgGradient = 'transparent';

    switch (profile.badge) {
        case 'Gold':
            badgeColor = '#FFD700';
            BadgeIcon = WorkspacePremiumIcon;
            bgGradient = isDark ? 'linear-gradient(135deg, rgba(255, 215, 0, 0.1) 0%, rgba(255, 215, 0, 0.02) 100%)' : 'linear-gradient(135deg, rgba(255, 215, 0, 0.2) 0%, rgba(255, 215, 0, 0.05) 100%)';
            break;
        case 'Silver':
            badgeColor = '#C0C0C0';
            BadgeIcon = WorkspacePremiumIcon;
            bgGradient = isDark ? 'linear-gradient(135deg, rgba(192, 192, 192, 0.1) 0%, rgba(192, 192, 192, 0.02) 100%)' : 'linear-gradient(135deg, rgba(192, 192, 192, 0.2) 0%, rgba(192, 192, 192, 0.05) 100%)';
            break;
        case 'Bronze':
             badgeColor = '#cd7f32';
             BadgeIcon = WorkspacePremiumIcon;
             bgGradient = isDark ? 'linear-gradient(135deg, rgba(205, 127, 50, 0.1) 0%, rgba(205, 127, 50, 0.02) 100%)' : 'linear-gradient(135deg, rgba(205, 127, 50, 0.2) 0%, rgba(205, 127, 50, 0.05) 100%)';
             break;
        default:
             badgeColor = theme.palette.text.secondary;
             bgGradient = isDark ? alpha(theme.palette.primary.main, 0.05) : alpha(theme.palette.primary.main, 0.02);
    }

    return (
        <Card sx={{ 
            borderRadius: 6, 
            mb: 3, 
            border: '1px solid', 
            borderColor: profile.badge !== 'None' ? badgeColor : 'divider', 
            background: bgGradient,
            boxShadow: profile.badge !== 'None' && !isDark ? `0 4px 20px -5px ${alpha(badgeColor, 0.4)}` : 'none',
            position: 'relative',
            overflow: 'hidden'
        }}>
            {profile.badge !== 'None' && (
                <Box sx={{ position: 'absolute', top: -15, right: -15, opacity: 0.1 }}>
                    <EmojiEventsIcon sx={{ fontSize: 120, color: badgeColor }} />
                </Box>
            )}

            <CardContent sx={{ p: 3, position: 'relative', zIndex: 1 }}>
                <Stack direction="row" spacing={2} alignItems="center" sx={{ mb: 3 }}>
                    <Box sx={{ p: 1, borderRadius: 2, bgcolor: alpha(badgeColor, 0.15) }}>
                        <BadgeIcon sx={{ color: badgeColor }} />
                    </Box>
                    <Typography variant="h6" fontWeight={800}>
                        Loyalty Status
                    </Typography>
                </Stack>

                <Stack spacing={3}>
                    <Box display="flex" justifyContent="space-between" alignItems="center">
                        <Box>
                            <Typography variant="caption" color="text.secondary" fontWeight={600} display="block">Current Badge</Typography>
                            <Typography variant="h5" fontWeight={800} sx={{ color: badgeColor }}>
                                {profile.badge === 'None' ? 'Member' : profile.badge}
                            </Typography>
                        </Box>
                        <Box textAlign="right">
                            <Typography variant="caption" color="text.secondary" fontWeight={600} display="block">Available Points</Typography>
                            <Typography variant="h4" fontWeight={800} color="primary.main">
                                {profile.points} <span style={{ fontSize: '0.5em', color: theme.palette.text.secondary }}>pts</span>
                            </Typography>
                        </Box>
                    </Box>

                    {/* Progress to next Reward */}
                    {nextRule && (
                        <Box>
                            <Stack direction="row" justifyContent="space-between" mb={1}>
                                <Typography variant="caption" fontWeight={600}>Next Reward Drop ({nextRule.points_awarded} pts)</Typography>
                                <Typography variant="caption" fontWeight={700} color="primary.main">{Math.round(dataProgress)}%</Typography>
                            </Stack>
                            <LinearProgress 
                                variant="determinate" 
                                value={dataProgress} 
                                sx={{ 
                                    height: 8, 
                                    borderRadius: 4,
                                    bgcolor: alpha(theme.palette.divider, 0.1),
                                    '& .MuiLinearProgress-bar': {
                                        borderRadius: 4,
                                        background: isDark 
                                            ? `linear-gradient(90deg, ${theme.palette.primary.main} 0%, ${theme.palette.secondary.main} 100%)`
                                            : 'linear-gradient(90deg, #3b82f6 0%, #8b5cf6 100%)'
                                    }
                                }} 
                            />
                        </Box>
                    )}

                    {/* Redemption Action */}
                    <Box pt={1}>
                        <Tooltip title={profile.points < 20 ? "You need at least 20 points to redeem a 10GB Data reward" : ""}>
                            <span>
                                <Button 
                                    variant="contained" 
                                    color="primary" 
                                    fullWidth 
                                    disabled={profile.points < 20 || redeeming}
                                    startIcon={<RedeemIcon />}
                                    onClick={() => handleRedeem(20)}
                                    sx={{ 
                                        borderRadius: 3, 
                                        py: 1.2, 
                                        fontWeight: 700,
                                        textTransform: 'none',
                                        background: profile.points >= 20 ? `linear-gradient(45deg, ${theme.palette.primary.main}, ${theme.palette.secondary.main})` : undefined,
                                        boxShadow: profile.points >= 20 && !isDark ? '0 4px 14px 0 rgba(0,118,255,0.39)' : 'none'
                                    }}
                                >
                                    {redeeming ? 'Redeeming...' : 'Redeem 20 pts for 10 GB Data'}
                                </Button>
                            </span>
                        </Tooltip>
                    </Box>
                </Stack>
            </CardContent>
        </Card>
    );
};
