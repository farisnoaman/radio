import React, { useState, useRef } from 'react';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button,
    TextField,
    Select,
    MenuItem,
    FormControl,
    InputLabel,
    Box,
} from '@mui/material';
import PrintIcon from '@mui/icons-material/Print';
import { useDataProvider, useNotify } from 'react-admin';

interface VoucherPrintDialogProps {
    batchId: string | number;
    batchName: string;
    productName: string;
    productColor?: string;
    productValidity?: number;
    open: boolean;
    onClose: () => void;
}

const VoucherPrintDialog: React.FC<VoucherPrintDialogProps> = ({
    batchId,
    batchName,
    productName,
    productColor = '#000000',
    productValidity = 0,
    open,
    onClose
}) => {
    const [template, setTemplate] = useState('template1');
    const [loginLink, setLoginLink] = useState('');
    const [hotspotName, setHotspotName] = useState('WiFi Hotspot');
    const [loading, setLoading] = useState(false);
    const dataProvider = useDataProvider();
    const notify = useNotify();
    const printRef = useRef<HTMLDivElement>(null);
    const [vouchers, setVouchers] = useState<any[]>([]);

    const formatValidity = (seconds: number) => {
        if (!seconds) return 'Unlimited';
        if (seconds < 3600) return `${Math.floor(seconds / 60)} Mins`;
        if (seconds < 86400) return `${Math.floor(seconds / 3600)} Hours`;
        return `${Math.floor(seconds / 86400)} Days`;
    };

    const fetchVouchers = async () => {
        setLoading(true);
        try {
            const { data } = await dataProvider.getList('vouchers', {
                filter: { batch_id: batchId },
                pagination: { page: 1, perPage: 10000 },
                sort: { field: 'id', order: 'ASC' },
            });
            setVouchers(data);
            // Wait for render then print
            setTimeout(() => {
                handlePrint();
                setLoading(false);
            }, 500);
        } catch (error) {
            notify('Failed to fetch vouchers for printing', { type: 'error' });
            setLoading(false);
        }
    };

    const handlePrint = () => {
        const content = printRef.current;
        if (!content) return;

        const printWindow = window.open('', '', 'width=800,height=600');
        if (!printWindow) return;

        printWindow.document.write('<html><head><title>Print Vouchers</title>');
        printWindow.document.write('<style>');
        printWindow.document.write(`
            @page { margin: 0.5cm; }
            body { font-family: Arial, sans-serif; margin: 0; padding: 0; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
            .voucher-container { 
                display: flex; 
                flex-wrap: wrap; 
                justify-content: flex-start; 
                gap: 10px;
            }
            
            /* General Card Styles */
            .voucher-card {
                box-sizing: border-box;
                position: relative;
                page-break-inside: avoid;
                background: #fff;
            }

            /* Template 1: QR / Detailed (Image 1) */
            .template1 .voucher-card {
                width: 240px;
                height: 140px;
                border: 2px solid ${productColor};
                display: flex;
                flex-direction: column;
                padding: 0;
            }
            .template1 .header {
                background: ${productColor};
                color: #fff;
                padding: 5px;
                font-weight: bold;
                display: flex;
                justify-content: space-between;
                font-size: 14px;
            }
            .template1 .body {
                display: flex;
                flex: 1;
                padding: 5px;
            }
            .template1 .info-col {
                flex: 1;
                display: flex;
                flex-direction: column;
                justify-content: center;
            }
            .template1 .qr-col {
                width: 80px;
                display: flex;
                align-items: center;
                justify-content: center;
            }
            .template1 .qr-col img {
                width: 70px;
                height: 70px;
            }
            .template1 .code-label { font-size: 10px; color: #666; }
            .template1 .code { 
                font-size: 16px; 
                font-weight: bold; 
                color: ${productColor}; 
                margin-bottom: 5px; 
                font-family: monospace;
            }
            .template1 .validity { font-size: 12px; }
            .template1 .footer {
                background: ${productColor};
                color: #fff;
                font-size: 10px;
                text-align: center;
                padding: 2px;
            }

            /* Template 2: Simple Box (Image 2) */
            .template2 .voucher-card {
                width: 200px;
                height: 120px;
                border: 3px solid #000;
                display: flex;
                flex-direction: column;
                text-align: center;
            }
            .template2 .header {
                border-bottom: 2px solid #000;
                padding: 4px;
                font-weight: bold;
                font-size: 14px;
                background: #f0f0f0;
            }
            .template2 .body {
                flex: 1;
                display: flex;
                flex-direction: column;
                justify-content: center;
                align-items: center;
            }
            .template2 .code-label { font-size: 12px; margin-bottom: 4px; }
            .template2 .code { 
                font-size: 22px; 
                font-weight: bold; 
                font-family: monospace; 
                letter-spacing: 1px;
            }
            .template2 .footer {
                border-top: 2px solid #000;
                padding: 4px;
                font-size: 12px;
                font-weight: bold;
            }

            /* Template 3: Colorful Card */
            .template3 .voucher-card {
                width: 220px;
                height: 130px;
                border: 1px solid #ccc;
                border-left: 10px solid ${productColor};
                padding: 10px;
                box-shadow: 2px 2px 5px rgba(0,0,0,0.1);
            }
            .template3 .code {
                font-size: 20px;
                font-weight: bold;
                color: ${productColor};
                margin: 10px 0;
            }
        `);
        printWindow.document.write('</style></head><body>');
        printWindow.document.write(content.innerHTML);
        printWindow.document.write('</body></html>');
        printWindow.document.close();
        printWindow.print();
        printWindow.close();
    };

    const handleConfirm = () => {
        fetchVouchers();
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>Print Vouchers - {batchName}</DialogTitle>
            <DialogContent>
                <Box mt={2} mb={2}>
                    <Box mb={2}>
                        <FormControl fullWidth>
                            <InputLabel>Template</InputLabel>
                            <Select
                                value={template}
                                onChange={(e) => setTemplate(e.target.value as string)}
                                label="Template"
                            >
                                <MenuItem value="template1">QR Style (Detailed)</MenuItem>
                                <MenuItem value="template2">Box Style (Simple)</MenuItem>
                                <MenuItem value="template3">Card Style (Modern)</MenuItem>
                            </Select>
                        </FormControl>
                    </Box>
                    <Box mb={2}>
                        <TextField
                            fullWidth
                            label="Hotspot Name"
                            value={hotspotName}
                            onChange={(e) => setHotspotName(e.target.value)}
                        />
                    </Box>
                    <Box>
                        <TextField
                            fullWidth
                            label="Login Link (Optional)"
                            value={loginLink}
                            onChange={(e) => setLoginLink(e.target.value)}
                            helperText="e.g., http://wifi.login or IP address"
                        />
                    </Box>
                </Box>

                {/* Hidden Print Content */}
                <div style={{ display: 'none' }}>
                    <div ref={printRef} className={template}>
                        <div className="voucher-container">
                            {vouchers.map((voucher, index) => (
                                <React.Fragment key={voucher.id}>
                                    {template === 'template1' && (
                                        <div className="voucher-card">
                                            <div className="header">
                                                <span>{hotspotName}</span>
                                                <span>{voucher.price > 0 ? voucher.price : ''}</span>
                                            </div>
                                            <div className="body">
                                                <div className="info-col">
                                                    <div className="code-label">VOUCHER CODE</div>
                                                    <div className="code">{voucher.code}</div>
                                                    <div className="validity">
                                                        Active: {formatValidity(productValidity)}
                                                    </div>
                                                </div>
                                                <div className="qr-col">
                                                    <img src={`https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${voucher.code}`} alt="QR" />
                                                </div>
                                            </div>
                                            <div className="footer">
                                                {loginLink || 'Login to Hotspot'}
                                            </div>
                                        </div>
                                    )}

                                    {template === 'template2' && (
                                        <div className="voucher-card">
                                            <div className="header">
                                                {hotspotName} [{index + 1}]
                                            </div>
                                            <div className="body">
                                                <div className="code-label">Voucher Code</div>
                                                <div className="code">{voucher.code}</div>
                                            </div>
                                            <div className="footer">
                                                {formatValidity(productValidity)} {voucher.price > 0 ? `| ${voucher.price}` : ''}
                                            </div>
                                        </div>
                                    )}

                                    {template === 'template3' && (
                                        <div className="voucher-card">
                                            <div style={{ fontSize: '14px', fontWeight: 'bold' }}>{productName}</div>
                                            <div style={{ fontSize: '12px', color: '#666' }}>Voucher Code</div>
                                            <div className="code">{voucher.code}</div>
                                            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px' }}>
                                                <span>Price: {voucher.price}</span>
                                                <span>Valid: {formatValidity(productValidity)}</span>
                                            </div>
                                            {loginLink && <div style={{ fontSize: '10px', marginTop: '5px', color: '#999' }}>{loginLink}</div>}
                                        </div>
                                    )}
                                </React.Fragment>
                            ))}
                        </div>
                    </div>
                </div>

            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>Cancel</Button>
                <Button
                    onClick={handleConfirm}
                    variant="contained"
                    color="primary"
                    startIcon={<PrintIcon />}
                    disabled={loading}
                >
                    {loading ? 'Preparing...' : 'Print'}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default VoucherPrintDialog;
