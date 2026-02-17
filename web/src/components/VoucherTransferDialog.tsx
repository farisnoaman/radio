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
import { useNotify, useRefresh, useGetList } from 'react-admin';
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

    const { data: agents, isLoading: agentsLoading } = useGetList('agents', {
        pagination: { page: 1, perPage: 100 },
        sort: { field: 'realname', order: 'ASC' },
    });

    const handleTransfer = async () => {
        if (!targetAgentId) {
            notify('Please select a target agent', { type: 'warning' });
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
            notify('Batch transferred successfully', { type: 'success' });
            refresh();
            onClose();
        } catch (error: any) {
            const msg = error?.json?.msg || error?.message || 'Failed to transfer batch';
            notify(msg, { type: 'error' });
        } finally {
            setLoading(false);
        }
    };

    return (
        <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
            <DialogTitle>Transfer Voucher Batch</DialogTitle>
            <DialogContent>
                <Box mb={2}>
                    <Typography variant="body1">
                        You are about to transfer batch <strong>{batchName}</strong> (ID: {batchId}) to another agent.
                    </Typography>
                    <Typography variant="body2" color="textSecondary" mt={1}>
                        This will move all vouchers in this batch to the target agent's ownership.
                    </Typography>
                </Box>

                {agentsLoading ? (
                    <Box display="flex" justifyContent="center">
                        <CircularProgress />
                    </Box>
                ) : (
                    <FormControl fullWidth variant="outlined" margin="normal">
                        <InputLabel id="target-agent-label">Target Agent</InputLabel>
                        <Select
                            labelId="target-agent-label"
                            value={targetAgentId}
                            onChange={(e) => setTargetAgentId(e.target.value as string)}
                            label="Target Agent"
                        >
                            <MenuItem value="">
                                <em>None</em>
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
                    Cancel
                </MuiButton>
                <MuiButton
                    onClick={handleTransfer}
                    color="primary"
                    variant="contained"
                    disabled={loading || !targetAgentId}
                >
                    {loading ? <CircularProgress size={24} color="inherit" /> : 'Confirm Transfer'}
                </MuiButton>
            </DialogActions>
        </Dialog>
    );
};

export default VoucherTransferDialog;
