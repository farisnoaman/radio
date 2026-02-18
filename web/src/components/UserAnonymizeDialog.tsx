import { useState } from 'react';
import { useNotify, useRefresh } from 'react-admin';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button,
    Typography,
    Alert
} from '@mui/material';
import { httpClient } from '../utils/apiClient';

interface UserAnonymizeDialogProps {
    open: boolean;
    onClose: () => void;
    username: string;
}

const UserAnonymizeDialog = ({ open, onClose, username }: UserAnonymizeDialogProps) => {
    const [loading, setLoading] = useState(false);
    const notify = useNotify();
    const refresh = useRefresh();

    const handleConfirm = async () => {
        setLoading(true);
        try {
            await httpClient('/privacy/anonymize', {
                method: 'POST',
                body: JSON.stringify({ username }),
            });
            notify('User data anonymized successfully', { type: 'success' });
            refresh();
            onClose();
        } catch (error: any) {
            const msg = error?.body?.msg || 'Failed to anonymize user';
            notify(msg, { type: 'error' });
        } finally {
            setLoading(false);
        }
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>Anonymize User Data</DialogTitle>
            <DialogContent>
                <Alert severity="warning" sx={{ mb: 2 }}>
                    Warning: This action is irreversible!
                </Alert>
                <Typography variant="body1" gutterBottom>
                    You are about to anonymize all personal data for user: <strong>{username}</strong>.
                </Typography>
                <Typography variant="body2" color="textSecondary" paragraph>
                    This will mask the user's real name, email, phone number, and address.
                    This action is performed to comply with GDPR "Right to be Forgotten" requests.
                    Accounting records will be preserved but delinked from personal identity where possible.
                </Typography>
                <Typography variant="body2" color="textSecondary">
                    Are you sure you want to proceed?
                </Typography>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose} disabled={loading}>Cancel</Button>
                <Button
                    onClick={handleConfirm}
                    color="error"
                    variant="contained"
                    disabled={loading}
                >
                    {loading ? 'Processing...' : 'Confirm Anonymize'}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default UserAnonymizeDialog;
