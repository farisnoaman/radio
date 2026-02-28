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
import { useNotify } from 'react-admin';

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
            const response = await fetch(`/api/v1/voucher-batches/${batchId}/print`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    'Content-Type': 'application/json',
                },
            });
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            const json = await response.json();
            setVouchers(json as any[]);
            // Wait for render then print
            setTimeout(() => {
                handlePrint();
                setLoading(false);
            }, 500);
        } catch (error) {
            console.error('Print fetch error:', error);
            notify('Failed to fetch vouchers for printing', { type: 'error' });
            setLoading(false);
        }
    };

    const handlePrint = () => {
        const printWindow = window.open('', '', 'width=900,height=700');
        if (!printWindow) return;

        if (template === 'template6') {
            // Generate exact self-contained HTML matching the provided MixRADIUS template
            // QR codes use api.qrserver.com (reliable, no external JS dependency)
            const voucherBoxes = vouchers.map((voucher: any) => {
                const qrData = loginLink
                    ? `${loginLink}?username=${voucher.code}&password=${voucher.code}`
                    : voucher.code;
                const qrSrc = `https://api.qrserver.com/v1/create-qr-code/?size=80x80&data=${encodeURIComponent(qrData)}`;
                return `
<div class="box">
    <div class="kiri1">
        <div class="user1">${voucher.code}</div>
        <div class="validity1">${formatValidity(productValidity)}</div>
        <div class="price1">${voucher.price > 0 ? 'Rp. ' + voucher.price : ''}</div>
        <div class="dns1">${hotspotName}</div>
        ${loginLink ? `<div style="margin-left:8px; font-size:8px; margin-top:2px; white-space:nowrap; overflow:hidden; width:85px; color:#555">${loginLink.replace(/^https?:\/\//, '')}</div>` : ''}
    </div>
    <div class="kanan">
        <div style="font-size:7px; font-family:monospace; color:#555; text-align:center; margin-bottom:2px; margin-top:-10px;">SN: ${batchId}-${voucher.id}</div>
        <img src="${qrSrc}" width="80" height="80" alt="QR" />
    </div>
    <div class="clear"></div>
</div>`;
            }).join('\n');

            const html = `<!DOCTYPE html>
<html>
<head>
<title>Print Voucher</title>
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
html, body {
    font-family: "Lucida Sans Unicode", "Lucida Grande", sans-serif;
    font-size: 12px;
    margin: 0;
    padding: 0;
}
@media print {
    @page {
        size: A4 portrait;
        margin: 0.5cm;
    }
    .no-print { display: none !important; }
}
.page {
    width: 100%;
}
.box {
    display: inline-block;
    width: 178px;
    height: 148px;
    margin: 2px;
    background: url(https://demo.mixradius.com:2143/theme/default/images/v20190415-1col.jpg) no-repeat;
    background-size: 100% 100%;
    vertical-align: top;
    border: 1px dashed #ccc;
    box-sizing: border-box;
}
.kiri1 {
    color: #444;
    float: left;
    width: 90px;
    font-family: "Courier New";
    font-size: 14px;
    font-weight: bold;
}
.user1 {
    margin-top: 18px;
    margin-left: 8px;
    display: inline-block;
    padding: 2px 4px;
    border: 2px solid #000;
    border-radius: 4px;
    background-color: #fff;
    font-size: 14px;
    font-weight: bold;
    color: #000;
}
.validity1 {
    margin-top: 8px;
    margin-left: 8px;
    font-size: 14px;
}
.price1 {
    margin-top: 2px;
    margin-left: 8px;
    font-size: 13px;
}
.dns1 {
    margin-top: 4px;
    margin-left: 8px;
    font-size: 11px;
}
.kanan {
    float: right;
    width: 80px;
    margin-top: 55px;
}
.kanan img {
    width: 80px;
    height: 80px;
    display: block;
}
.clear { clear: both; }
</style>
</head>
<body>
<div class="page">
${voucherBoxes}
</div>
<script>
// Wait for all QR images to load, then print
var images = document.getElementsByTagName('img');
var total = images.length;
var loaded = 0;
function tryPrint() {
    loaded++;
    if (loaded >= total) {
        setTimeout(function() { window.print(); }, 300);
    }
}
if (total === 0) {
    window.print();
} else {
    for (var i = 0; i < total; i++) {
        if (images[i].complete) {
            tryPrint();
        } else {
            images[i].onload = tryPrint;
            images[i].onerror = tryPrint;
        }
    }
}
<\/script>
</body>
</html>`;

            printWindow.document.write(html);
            printWindow.document.close();
            return;
        }

        // All other templates — use the existing React-rendered DOM approach
        const content = printRef.current;
        if (!content) return;

        printWindow.document.write('<html><head><title>Print Vouchers</title>');
        printWindow.document.write('<meta name="viewport" content="width=device-width, initial-scale=1.0">');
        printWindow.document.write('<style>');
        printWindow.document.write(`
            @page { size: A4 portrait; margin: 0.2cm; }
            body { font-family: Arial, sans-serif; margin: 0; padding: 0; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
            .voucher-container { 
                display: flex; 
                flex-wrap: wrap; 
                justify-content: flex-start; 
                gap: 5px;
            }
            
            /* General Card Styles */
            .voucher-card {
                box-sizing: border-box;
                position: relative;
                page-break-inside: avoid;
                background: #fff;
                border: 1px dashed #ccc;
                margin: 2px;
            }
            
            .serial {
                font-size: 9px;
                font-weight: bold;
                color: #000;
                font-family: monospace;
                position: absolute;
                bottom: 2px;
                right: 2px;
                padding: 2px 4px;
                border: 1px solid #000;
                background: #fff;
            }

            /* Template 1: QR / Detailed (Image 1) */
            .template1 .voucher-card {
                width: 220px;
                height: 120px;
                border: 2px solid ${productColor};
                display: flex;
                flex-direction: column;
                padding: 0;
            }
            .template1 .header {
                background: ${productColor};
                color: #fff;
                padding: 3px;
                font-weight: bold;
                display: flex;
                justify-content: space-between;
                font-size: 12px;
            }
            .template1 .body {
                display: flex;
                flex: 1;
                padding: 3px;
            }
            .template1 .info-col {
                flex: 1;
                display: flex;
                flex-direction: column;
                justify-content: center;
            }
            .template1 .qr-col {
                width: 70px;
                display: flex;
                align-items: center;
                justify-content: center;
            }
            .template1 .qr-col img {
                width: 60px;
                height: 60px;
            }
            .template1 .code-label { font-size: 9px; color: #666; }
            .template1 .code { 
                font-size: 14px; 
                font-weight: bold; 
                color: ${productColor}; 
                margin-bottom: 2px; 
                font-family: monospace;
                display: inline-block;
                padding: 2px 4px;
                border: 2px solid ${productColor};
                border-radius: 4px;
                background: #fff;
            }
            .template1 .validity { font-size: 10px; }
            .template1 .footer {
                background: ${productColor};
                color: #fff;
                font-size: 9px;
                text-align: center;
                padding: 2px;
            }

            /* Template 2: Simple Box (Image 2) */
            .template2 .voucher-card {
                width: 170px;
                height: 100px;
                border: 2px solid #000;
                display: flex;
                flex-direction: column;
                text-align: center;
            }
            .template2 .header {
                border-bottom: 1px solid #000;
                padding: 2px;
                font-weight: bold;
                font-size: 12px;
                background: #f0f0f0;
            }
            .template2 .body {
                flex: 1;
                display: flex;
                flex-direction: column;
                justify-content: center;
                align-items: center;
            }
            .template2 .code-label { font-size: 10px; margin-bottom: 2px; }
            .template2 .code { 
                font-size: 16px; 
                font-weight: bold; 
                font-family: monospace; 
                letter-spacing: 1px;
                display: inline-block;
                padding: 3px 6px;
                border: 2px solid #000;
                border-radius: 4px;
                background: #fff;
            }
            .template2 .footer {
                border-top: 1px solid #000;
                padding: 2px;
                font-size: 10px;
                font-weight: bold;
            }

            /* Template 3: Colorful Card */
            .template3 .voucher-card {
                width: 190px;
                height: 110px;
                border: 1px solid #ccc;
                border-left: 8px solid ${productColor};
                padding: 8px;
                box-shadow: 1px 1px 3px rgba(0,0,0,0.1);
            }
            .template3 .code {
                font-size: 16px;
                font-weight: bold;
                color: ${productColor};
                margin: 5px 0;
                display: inline-block;
                padding: 3px 6px;
                border: 2px solid ${productColor};
                border-radius: 4px;
                background: #fff;
            }

            /* Template 4: MixRADIUS Style */
            .template4 .voucher-card {
                width: 280px;
                height: 130px;
                padding: 0;
                position: relative;
                overflow: hidden;
            }
            .template4 .box {
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                padding: 0;
                width: 100%;
                height: 100%;
                display: flex;
                flex-direction: column;
            }
            .template4 .logo {
                padding: 3px 5px;
                color: #fff;
                font-weight: bold;
                font-size: 12px;
                display: flex;
                justify-content: space-between;
            }
            .template4 .kiri1 {
                flex: 1;
                padding: 5px 10px;
                display: flex;
                flex-direction: column;
                justify-content: center;
            }
            .template4 .user1 {
                font-size: 18px;
                font-weight: bold;
                color: #fff;
                font-family: monospace;
                text-shadow: 1px 1px 2px rgba(0,0,0,0.3);
                display: inline-block;
                padding: 3px 6px;
                border: 2px solid #fff;
                border-radius: 4px;
                background: rgba(0,0,0,0.2);
            }
            .template4 .validity1 {
                font-size: 12px;
                color: rgba(255,255,255,0.9);
                margin: 3px 0;
            }
            .template4 .price1 {
                font-size: 14px;
                font-weight: bold;
                color: #ffd700;
            }
            .template4 .dns1 {
                font-size: 10px;
                color: rgba(255,255,255,0.7);
                margin-top: 3px;
            }
            .template4 .kanan {
                width: 75px;
                display: flex;
                align-items: center;
                justify-content: center;
                padding: 5px;
            }
            .template4 .qrcode img {
                width: 65px;
                height: 65px;
                background: #fff;
                border-radius: 4px;
            }
            .template4 .clearboth {
                clear: both;
            }

            /* Template 5: MixRADIUS Compact - 4 Per Row A4 Portrait */
            .template5 .voucher-container {
                display: flex;
                flex-wrap: wrap;
                justify-content: flex-start;
                gap: 5px;
            }

            .template5 .voucher-card {
                display: inline-block;
                width: 178px;
                height: 148px;
                margin: 2px;
                background: url(https://demo.mixradius.com:2143/theme/default/images/v20190415-1col.jpg) no-repeat;
                background-size: 100% 100%;
                box-sizing: border-box;
                page-break-inside: avoid;
                position: relative;
                vertical-align: top;
                border: 1px dashed #ccc;
            }

            /* LEFT SIDE */
            .template5 .kiri1 {
                float: left;
                width: 90px;
                color: #444;
                font-family: "Courier New", monospace;
                font-size: 14px;
                font-weight: bold;
            }

            .template5 .user1 {
                margin-top: 18px;
                margin-left: 8px;
                display: inline-block;
                padding: 2px 4px;
                border: 2px solid #000;
                border-radius: 4px;
                background-color: #fff;
                font-size: 14px;
                font-weight: bold;
                color: #000;
            }

            .template5 .validity1 {
                margin-top: 8px;
                margin-left: 8px;
                font-size: 14px;
            }

            .template5 .price1 {
                margin-top: 2px;
                margin-left: 8px;
                font-size: 13px;
            }

            .template5 .dns1 {
                margin-top: 4px;
                margin-left: 8px;
                font-size: 11px;
            }

            /* RIGHT SIDE */
            .template5 .kanan {
                float: right;
                width: 80px;
                margin-top: 55px;
            }

            .template5 .qrcode img {
                width: 80px;
                height: 80px;
            }

            .template5 .clear {
                clear: both;
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
                                <MenuItem value="template4">MixRADIUS Style</MenuItem>
                                <MenuItem value="template5">MixRADIUS Compact</MenuItem>
                                <MenuItem value="template6">MixRADIUS Standard</MenuItem>
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
                                                    <img src={`https://api.qrserver.com/v1/create-qr-code/?size=100x100&data=${encodeURIComponent(loginLink ? loginLink + '?username=' + voucher.code + '&password=' + voucher.code : voucher.code)}`} alt="QR" />
                                                </div>
                                            </div>
                                            <div className="footer">
                                                {loginLink || 'Login to Hotspot'}
                                            </div>
                                            <div className="serial">SN: {batchId}-{voucher.id}</div>
                                        </div>
                                    )}

                                    {template === 'template2' && (
                                        <div className="voucher-card">
                                            <div className="header">
                                                {hotspotName}
                                            </div>
                                            <div className="body">
                                                <div className="code-label">Voucher Code</div>
                                                <div className="code">{voucher.code}</div>
                                                {loginLink && <div style={{ fontSize: '8px', color: '#666', marginTop: '2px' }}>{loginLink}</div>}
                                            </div>
                                            <div className="footer">
                                                {formatValidity(productValidity)} {voucher.price > 0 ? `| ${voucher.price}` : ''}
                                            </div>
                                            <div className="serial">SN: {batchId}-{voucher.id}</div>
                                        </div>
                                    )}

                                    {template === 'template3' && (
                                        <div className="voucher-card">
                                            <div style={{ fontSize: '13px', fontWeight: 'bold' }}>{productName}</div>
                                            <div style={{ fontSize: '11px', color: '#666' }}>Voucher Code</div>
                                            <div className="code">{voucher.code}</div>
                                            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px' }}>
                                                <span>Price: {voucher.price}</span>
                                                <span>Valid: {formatValidity(productValidity)}</span>
                                            </div>
                                            {loginLink && <div style={{ fontSize: '9px', marginTop: '5px', color: '#999' }}>{loginLink}</div>}
                                            <div className="serial">SN: {batchId}-{voucher.id}</div>
                                        </div>
                                    )}

                                    {template === 'template4' && (
                                        <div className="voucher-card">
                                            <div className="box">
                                                <div className="logo">
                                                    <span>VOUCHER</span>
                                                    {loginLink && <span style={{ fontSize: '8px', fontWeight: 'normal', opacity: 0.8 }}>{loginLink}</span>}
                                                </div>
                                                <div className="kiri1" style={{ float: 'left', width: '170px' }}>
                                                    <div className="user1">{voucher.code}</div>
                                                    <div className="validity1">
                                                        {formatValidity(productValidity)}
                                                    </div>
                                                    <div className="price1">{voucher.price > 0 ? voucher.price : ''}</div>
                                                    <div className="dns1">{hotspotName}</div>
                                                </div>
                                                <div className="kanan" style={{ position: 'absolute', right: '5px', top: '25px' }}>
                                                    <div className="qrcode">
                                                        <img src={`https://api.qrserver.com/v1/create-qr-code/?size=100x100&data=${encodeURIComponent(loginLink ? loginLink + '?username=' + voucher.code + '&password=' + voucher.code : voucher.code)}`} alt="QR" />
                                                    </div>
                                                </div>
                                            </div>
                                            <div className="serial" style={{ color: '#fff', border: '1px solid rgba(255,255,255,0.5)', background: 'transparent' }}>SN: {batchId}-{voucher.id}</div>
                                        </div>
                                    )}

                                    {template === 'template5' && (
                                        <div className="voucher-card">
                                            <div className="kiri1">
                                                <div className="user1">{voucher.code}</div>
                                                <div className="validity1">{formatValidity(productValidity)}</div>
                                                <div className="price1">{voucher.price > 0 ? 'Rp. ' + voucher.price : ''}</div>
                                                <div className="dns1" style={{ whiteSpace: 'nowrap', overflow: 'hidden' }}>{loginLink ? loginLink.replace(/^https?:\/\//, '') : hotspotName}</div>
                                            </div>
                                            <div className="kanan">
                                                <div className="qrcode">
                                                    <img src={`https://api.qrserver.com/v1/create-qr-code/?size=100x100&data=${encodeURIComponent(loginLink ? loginLink + '?username=' + voucher.code + '&password=' + voucher.code : voucher.code)}`} alt="QR" />
                                                </div>
                                            </div>
                                            <div className="clear"></div>
                                            <div className="serial">SN: {batchId}-{voucher.id}</div>
                                        </div>
                                    )}

                                    {template === 'template6' && (
                                        <div className="voucher-card">
                                            <div className="kiri1">
                                                <div className="user1">{voucher.code}</div>
                                                <div className="validity1">{formatValidity(productValidity)}</div>
                                                <div className="price1">{voucher.price > 0 ? 'Rp. ' + voucher.price : ''}</div>
                                                <div className="dns1" style={{ whiteSpace: 'nowrap', overflow: 'hidden' }}>{loginLink ? loginLink.replace(/^https?:\/\//, '') : hotspotName}</div>
                                            </div>
                                            <div className="kanan">
                                                <div className="qrcode" id={'qr' + voucher.code}></div>
                                            </div>
                                            <div className="clear"></div>
                                            <div className="serial">SN: {batchId}-{voucher.id}</div>
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
