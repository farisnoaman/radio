import { useState } from 'react';
import { useNotify, useRefresh } from 'react-admin';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button,
    TextField,
    Box,
    Typography
} from '@mui/material';
import { httpClient } from '../utils/apiClient';

interface VoucherExtensionDialogProps {
    open: boolean;
    onClose: () => void;
    voucherCode: string;
    currentExpiry: string;
}

const VoucherExtensionDialog = ({ open, onClose, voucherCode, currentExpiry }: VoucherExtensionDialogProps) => {
    const [days, setDays] = useState<number>(30);
    const [loading, setLoading] = useState(false);
    const notify = useNotify();
    const refresh = useRefresh();

    const handleConfirm = async () => {
        setLoading(true);
        try {
            await httpClient(`/vouchers/${voucherCode}/extend`, {
                method: 'POST',
                body: JSON.stringify({ validity_days: Number(days) }),
            });
            notify('Voucher extended successfully', { type: 'success' });
            refresh();
            onClose();
        } catch (error: any) {
            const msg = error?.body?.msg || 'Failed to extend voucher';
            notify(msg, { type: 'error' });
        } finally {
            setLoading(false);
        }
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>Extend Voucher Validity</DialogTitle>
            <DialogContent>
                <Box sx={{ mt: 2 }}>
                    <Typography variant="body1" gutterBottom>
                        Voucher Code: <strong>{voucherCode}</strong>
                    </Typography>
                    <Typography variant="body2" color="textSecondary" gutterBottom>
                        Current Expiry: {new Date(currentExpiry).toLocaleString()}
                    </Typography>

                    <Box sx={{ mt: 3 }}>
                        <TextField
                            label="Extend by Days"
                            type="number"
                            value={days}
                            onChange={(e) => setDays(parseInt(e.target.value))}
                            fullWidth
                            InputProps={{ inputProps: { min: 1 } }}
                            helperText="Enter the number of days to add to the current expiration date."
                        />
                    </Box>
                </Box>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose} disabled={loading}>Cancel</Button>
                <Button
                    onClick={handleConfirm}
                    color="primary"
                    variant="contained"
                    disabled={loading || days < 1}
                >
                    {loading ? 'Extending...' : 'Extend Validity'}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default VoucherExtensionDialog;
