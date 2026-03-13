import React, { useState } from 'react';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button as MuiButton,
    Select,
    MenuItem,
    FormControl,
    InputLabel,
    Typography,
    Box,
    CircularProgress,
} from '@mui/material';
import { useNotify, useRefresh, useGetList, useTranslate } from 'react-admin';
import { httpClient } from '../utils/apiClient';

interface VoucherTransferDialogProps {
    open: boolean;
    onClose: () => void;
    batchId: string | number;
    batchName: string;
}

const VoucherTransferDialog: React.FC<VoucherTransferDialogProps> = ({ open, onClose, batchId, batchName }) => {
    const [targetAgentId, setTargetAgentId] = useState<string>('');
    const [loading, setLoading] = useState(false);
    const notify = useNotify();
    const refresh = useRefresh();
    const translate = useTranslate();

    const { data: agents, isLoading: agentsLoading } = useGetList('agents', {
        pagination: { page: 1, perPage: 100 },
        sort: { field: 'realname', order: 'ASC' },
    });

    const handleTransfer = async () => {
        if (!targetAgentId) {
            notify(translate('pages.voucher.dialogs.transfer.select_agent'), { type: 'warning' });
            return;
        }

        setLoading(true);
        try {
            await httpClient(`/voucher-batches/${batchId}/transfer`, {
                method: 'POST',
                body: JSON.stringify({
                    target_agent_id: targetAgentId,
                }),
            });
            notify(translate('pages.voucher.dialogs.transfer.success'), { type: 'success' });
            refresh();
            onClose();
        } catch (error: any) {
            const msg = error?.json?.msg || error?.message || translate('pages.voucher.dialogs.transfer.error');
            notify(msg, { type: 'error' });
        } finally {
            setLoading(false);
        }
    };

    return (
        <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
            <DialogTitle>{translate('pages.voucher.dialogs.transfer.title')}</DialogTitle>
            <DialogContent>
                <Box mb={2}>
                    <Typography variant="body1">
                        {translate('pages.voucher.dialogs.transfer.message', { batchName, batchId })}
                    </Typography>
                    <Typography variant="body2" color="textSecondary" mt={1}>
                        {translate('pages.voucher.dialogs.transfer.warning')}
                    </Typography>
                </Box>

                {agentsLoading ? (
                    <Box display="flex" justifyContent="center">
                        <CircularProgress />
                    </Box>
                ) : (
                    <FormControl fullWidth variant="outlined" margin="normal">
                        <InputLabel id="target-agent-label">{translate('pages.voucher.dialogs.transfer.target_agent')}</InputLabel>
                        <Select
                            labelId="target-agent-label"
                            value={targetAgentId}
                            onChange={(e) => setTargetAgentId(e.target.value as string)}
                            label={translate('pages.voucher.dialogs.transfer.target_agent')}
                        >
                            <MenuItem value="">
                                <em>{translate('common.none')}</em>
                            </MenuItem>
                            {agents?.map((agent: any) => (
                                <MenuItem key={agent.id} value={agent.id}>
                                    {agent.realname} ({agent.username})
                                </MenuItem>
                            ))}
                        </Select>
                    </FormControl>
                )}
            </DialogContent>
            <DialogActions>
                <MuiButton onClick={onClose} disabled={loading}>
                    {translate('common.cancel')}
                </MuiButton>
                <MuiButton
                    onClick={handleTransfer}
                    color="primary"
                    variant="contained"
                    disabled={loading || !targetAgentId}
                >
                    {loading ? <CircularProgress size={24} color="inherit" /> : translate('pages.voucher.dialogs.transfer.confirm')}
                </MuiButton>
            </DialogActions>
        </Dialog>
    );
};

export default VoucherTransferDialog;
