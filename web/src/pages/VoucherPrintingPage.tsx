import React, { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useNotify, Title, useTranslate, useLocaleState } from 'react-admin';
import {
    Box,
    Card,
    CardContent,
    Typography,
    TextField,
    Button,
    Checkbox,
    FormControlLabel,
    Divider,
    CircularProgress,
    Tabs,
    Tab,
    IconButton,
    Tooltip,
    Chip,
    useMediaQuery,
    Theme,
    Autocomplete,
} from '@mui/material';
import PrintIcon from '@mui/icons-material/Print';
import SaveIcon from '@mui/icons-material/Save';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import PublicIcon from '@mui/icons-material/Public';
import LockIcon from '@mui/icons-material/Lock';
import PreviewIcon from '@mui/icons-material/Preview';
import { apiRequest } from '../utils/apiClient';

// ─── Types ───────────────────────────────────────────────────────────────────
interface VoucherBatch {
    id: number;
    name: string;
    product_id: number;
    agent_id: number;
    count: number;
}

interface Product {
    id: number;
    name: string;
    price: number;
    color?: string;
    data_quota?: number;
    validity_seconds?: number;
}

interface VoucherData {
    id: number;
    code: string;
    price: number;
    status: string;
}

interface VoucherTemplate {
    id: number;
    name: string;
    content: string;
    owner_id: number;
    is_public: boolean;
    is_default: boolean;
    created_at?: string;
}

// ─── Built-in Template Definitions ───────────────────────────────────────────

const BUILTIN_TEMPLATES: Record<string, { name: string; render: (vars: TemplateVars) => string }> = {
    template1: {
        name: 'QR Style (Detailed)',
        render: (v) => `
<div style="width:142px;height:120px;border:1px solid ${v.color};display:flex;flex-direction:column;margin:2px;box-sizing:border-box;page-break-inside:avoid;overflow:hidden;position:relative;direction:${v.rtl ? 'rtl' : 'ltr'};">
  <div style="background:${v.color};color:#fff;padding:2px;font-weight:bold;display:flex;justify-content:space-between;font-size:10px;">
    <span style="white-space:nowrap;overflow:hidden;text-overflow:ellipsis;flex:1;">${v.hotspot}</span>
    ${v.price ? `<span style="background:rgba(255,255,255,0.2);padding:0 2px;border-radius:2px;margin-${v.rtl ? 'right' : 'left'}:2px;">${v.price}</span>` : ''}
  </div>
  <div style="display:flex;flex:1;padding:2px;gap:2px;">
    <div style="flex:1;display:flex;flex-direction:column;justify-content:center;overflow:hidden;text-align:${v.rtl ? 'right' : 'left'};">
      <div style="font-size:8px;color:#666;">${v.t('pages.voucher.print.voucher_code')}</div>
      <div style="font-size:12px;font-weight:bold;color:${v.color};font-family:monospace;padding:1px 2px;border:1px solid ${v.color};border-radius:2px;display:inline-block;background:#fff;align-self:flex-start;">${v.code}</div>
      <div style="font-size:8px;margin-top:2px;font-weight:bold;color:#444;">${v.quota}</div>
      <div style="font-size:8px;color:#444;">${v.validity}</div>
    </div>
    ${v.showQR ? `<div style="width:50px;display:flex;align-items:center;justify-content:center;">${makeQR(v.link ? `${v.link}?username=${v.code}&password=${v.code}` : v.code, 50)}</div>` : ''}
  </div>
  <div style="background:${v.color};color:#fff;font-size:8px;text-align:center;padding:1px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${v.link || 'Hotspot Login'}</div>
  <div style="font-size:7px;font-family:monospace;position:absolute;bottom:1px;${v.rtl ? 'left' : 'right'}:1px;padding:0 2px;background:rgba(255,255,255,0.9);border:0.5px solid #ccc;border-radius:2px;display:flex;flex-direction:column;align-items:flex-end;">
    ${v.agent ? `<span style="font-size:6px;opacity:0.8;margin-bottom:1px;">${v.t('pages.voucher.print.agent_label')}: ${v.agent}</span>` : ''}
    <span>${v.t('pages.voucher.print.sn')}:${v.serial}</span>
  </div>
</div>`,
    },
    template2: {
        name: 'Box Style (Simple)',
        render: (v) => `
<div style="width:142px;height:100px;border:1px solid #000;display:flex;flex-direction:column;text-align:center;margin:2px;box-sizing:border-box;page-break-inside:avoid;position:relative;direction:${v.rtl ? 'rtl' : 'ltr'};">
  <div style="border-bottom:1px solid #000;padding:2px;font-weight:bold;font-size:10px;background:#f0f0f0;display:flex;justify-content:space-between;align-items:center;">
    <span style="white-space:nowrap;overflow:hidden;text-overflow:ellipsis;flex:1;text-align:${v.rtl ? 'right' : 'left'};">${v.hotspot}</span>
    <span style="font-size:9px;">${v.price}</span>
  </div>
  <div style="flex:1;display:flex;flex-direction:column;justify-content:center;padding:2px;">
    <div style="font-size:14px;font-weight:bold;letter-spacing:1px;margin:2px 0;">${v.code}</div>
    <div style="font-size:8px;color:#555;">${v.quota} | ${v.validity}</div>
  </div>
  <div style="font-size:7px;color:#888;padding:2px;border-top:1px dashed #ccc;display:flex;justify-content:space-between;">
    <span>${v.t('pages.voucher.print.sn')}:${v.serial}</span>
    ${v.agent ? `<span>${v.t('pages.voucher.print.agent_label')}: ${v.agent}</span>` : ''}
  </div>
</div>`,
    },
    template3: {
        name: 'Card Style (Modern)',
        render: (v) => `
<div style="width:142px;height:140px;border-radius:8px;background:linear-gradient(135deg, ${v.color} 0%, #2c3e50 100%);color:#fff;display:flex;flex-direction:column;margin:2px;box-sizing:border-box;page-break-inside:avoid;padding:8px;position:relative;box-shadow:0 2px 4px rgba(0,0,0,0.1);direction:${v.rtl ? 'rtl' : 'ltr'};">
  <div style="font-size:10px;font-weight:bold;margin-bottom:4px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;text-align:${v.rtl ? 'right' : 'left'};">${v.hotspot}</div>
  <div style="flex:1;display:flex;flex-direction:column;justify-content:center;align-items:center;background:rgba(255,255,255,0.1);border-radius:4px;margin:4px 0;padding:4px;">
    <div style="font-size:7px;opacity:0.8;text-transform:uppercase;">${v.t('pages.voucher.print.voucher_code')}</div>
    <div style="font-size:16px;font-weight:bold;letter-spacing:1px;color:#fff;">${v.code}</div>
  </div>
  <div style="font-size:9px;display:flex;justify-content:space-between;margin-top:2px;">
    <span>${v.quota}</span>
    <span>${v.price}</span>
  </div>
  <div style="font-size:8px;opacity:0.8;margin-top:2px;text-align:center;">${v.validity}</div>
  <div style="position:absolute;bottom:4px;${v.rtl ? 'left' : 'right'}:8px;font-size:7px;opacity:0.6;display:flex;flex-direction:column;align-items:flex-end;">
    ${v.agent ? `<span style="font-size:7px;font-weight:normal;opacity:0.7;">${v.agent}</span>` : ''}
    <span>${v.t('pages.voucher.print.sn')}:${v.serial}</span>
  </div>
</div>`,
    },
    template4: {
        name: 'Gradient Style',
        render: (v) => `
<div style="width:142px;height:120px;background:linear-gradient(to bottom, #f8f9fa, #e9ecef);border:1px solid #dee2e6;border-radius:6px;display:flex;flex-direction:column;margin:2px;box-sizing:border-box;page-break-inside:avoid;overflow:hidden;position:relative;color:#333;direction:${v.rtl ? 'rtl' : 'ltr'};">
  <div style="padding:4px;border-bottom:1px solid #dee2e6;background:#fff;display:flex;align-items:center;gap:4px;">
    <div style="width:8px;height:8px;border-radius:50%;background:${v.color};"></div>
    <div style="font-size:10px;font-weight:bold;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;flex:1;text-align:${v.rtl ? 'right' : 'left'};">${v.hotspot}</div>
  </div>
  <div style="flex:1;display:flex;flex-direction:column;justify-content:center;align-items:center;padding:4px;">
    <div style="font-size:18px;font-weight:bold;color:#000;margin-bottom:2px;">${v.code}</div>
    <div style="font-size:9px;color:#6c757d;display:flex;gap:4px;flex-wrap:nowrap;">
      <span>${v.quota}</span>
      <span>|</span>
      <span>${v.validity}</span>
      ${v.price ? `<span>|</span><span style="background:${v.color};color:#fff;padding:0 3px;border-radius:2px;">${v.price}</span>` : ''}
    </div>
  </div>
  <div style="padding:3px;background:rgba(0,0,0,0.03);font-size:7px;display:flex;justify-content:space-between;border-top:1px solid #dee2e6;color:#6c757d;">
    ${v.agent ? `<span style="font-size:6.5px;opacity:0.9;">${v.agent}</span>` : '<span></span>'}
    <span>${v.t('pages.voucher.print.sn')}:${v.serial}</span>
  </div>
</div>`,
    },
    template5: {
        name: 'Compact (5 per row)',
        render: (v) => `
<div style="width:142px;height:75px;border:1px solid #ccc;border-radius:4px;padding:4px;margin:2px;box-sizing:border-box;page-break-inside:avoid;display:flex;flex-direction:row;align-items:center;justify-content:space-between;overflow:hidden;background:#fff;direction:${v.rtl ? 'rtl' : 'ltr'};">
  <div style="flex:1;text-align:${v.rtl ? 'right' : 'left'};">
    <div style="font-size:8px;font-weight:bold;color:${v.color};margin-bottom:2px;">${v.hotspot}</div>
    <div style="font-size:12px;font-weight:bold;margin-bottom:1px;">${v.code}</div>
    <div style="font-size:7px;color:#666;">${v.quota} | ${v.validity}</div>
    <div style="font-size:6px;color:#999;margin-top:2px;">${v.t('pages.voucher.print.sn')}:${v.serial}</div>
  </div>
  <div style="display:flex;flex-direction:column;align-items:center;margin-${v.rtl ? 'right' : 'left'}:4px;">
    ${v.showQR ? makeQR(v.link ? `${v.link}?username=${v.code}&password=${v.code}` : v.code, 45) : ''}
    ${v.agent ? `<div style="font-size:7px;color:#888;margin-top:2px;text-align:center;max-width:50px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${v.agent}</div>` : ''}
  </div>
  <div style="clear:both;">  </div>
</div>`,
    },
};

interface TemplateVars {
    code: string;
    price: string;
    validity: string;
    quota: string;
    agent: string;
    hotspot: string;
    link: string;
    serial: string;
    qr: string;
    showQR: boolean;
    color: string;
    product: string;
    rtl: boolean;
    t: (key: string, options?: any) => string;
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

const formatValidity = (seconds: number): string => {
    if (!seconds) return 'Unlimited Time';
    if (seconds < 3600) return `${Math.floor(seconds / 60)} Mins`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)} Hours`;
    return `${Math.floor(seconds / 86400)} Days`;
};

const formatQuota = (mb?: number): string => {
    if (mb === undefined || mb === null) return '-';
    if (mb === 0) return 'Unlimited Data';
    if (mb >= 1024) {
      return `${(mb / 1024).toFixed(1)} GB`;
    }
    return `${mb} MB`;
};

const makeQR = (data: string, size = 80): string =>
    `<img src="https://api.qrserver.com/v1/create-qr-code/?size=${size}x${size}&data=${encodeURIComponent(data)}" width="${size}" height="${size}" alt="QR" style="display:block;" />`;

const replaceCustomVars = (html: string, vars: TemplateVars): string =>
    html
        .replace(/\{\{code\}\}/g, vars.code)
        .replace(/\{\{price\}\}/g, vars.price)
        .replace(/\{\{validity\}\}/g, vars.validity)
        .replace(/\{\{quota\}\}/g, vars.quota)
        .replace(/\{\{agent\}\}/g, vars.agent)
        .replace(/\{\{hotspot\}\}/g, vars.hotspot)
        .replace(/\{\{link\}\}/g, vars.link)
        .replace(/\{\{serial\}\}/g, vars.serial)
        .replace(/\{\{qr\}\}/g, vars.qr)
        .replace(/\{\{product\}\}/g, vars.product)
        .replace(/\{\{color\}\}/g, vars.color);

const getSampleTemplate = (t: (key: string, options?: any) => string, isRtl: boolean) => `
<div style="width:142px;height:120px;border:1px solid {{color}};margin:2px;padding:4px;box-sizing:border-box;direction:${isRtl ? 'rtl' : 'ltr'};display:flex;flex-direction:column;justify-content:space-between;page-break-inside:avoid;overflow:hidden;background:#fff;position:relative;">
  <div style="font-weight:bold;font-size:10px;text-align:${isRtl ? 'right' : 'left'};border-bottom:1px solid #ccc;padding-bottom:2px;">
    {{hotspot}}
  </div>
  <div style="font-size:14px;font-weight:bold;text-align:center;margin:4px 0;">
    {{code}}
  </div>
  <div style="font-size:8px;color:#555;">
    ${t('pages.voucher.print.quota_label')}: {{quota}} | ${t('pages.voucher.print.validity_label')}: {{validity}}
  </div>
  <div style="font-size:8px;">
    ${t('pages.voucher.print.price_label')}: <span style="font-weight:bold;">{{price}}</span>
  </div>
  <div style="font-size:8px;margin-top:2px;">
    ${t('pages.voucher.print.product_label')}: {{product}}
  </div>
  <div style="text-align:center;margin-top:2px;">
    {{qr}}
  </div>
  <div style="position:absolute;bottom:2px;${isRtl ? 'left' : 'right'}:2px;font-size:7px;color:#777;text-align:${isRtl ? 'left' : 'right'};">
    ${t('pages.voucher.print.agent_label')}: {{agent}} <br/>
    ${t('pages.voucher.print.sn')}: {{serial}}
  </div>
</div>`;

// ─── Main Page Component ─────────────────────────────────────────────────────

const VoucherPrintingPage: React.FC = () => {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const notify = useNotify();
    const isMobile = useMediaQuery((theme: Theme) => theme.breakpoints.down('md'));

    const translate = useTranslate();
    const [locale] = useLocaleState();
    const isRtl = locale === 'ar';

    // --- State ---
    const [batches, setBatches] = useState<VoucherBatch[]>([]);
    const [selectedBatchId, setSelectedBatchId] = useState<number | null>(null);
    const [products, setProducts] = useState<Record<number, Product>>({});
    const [hotspotName, setHotspotName] = useState('My Hotspot');
    const [agentName, setAgentName] = useState('');
    const [loginLink, setLoginLink] = useState('');
    const [printQR, setPrintQR] = useState(true);
    const [selectedTemplate, setSelectedTemplate] = useState('template1');
    const [customTemplates, setCustomTemplates] = useState<VoucherTemplate[]>([]);
    const [vouchers, setVouchers] = useState<VoucherData[]>([]);
    const [fetchingVouchers, setFetchingVouchers] = useState(false);
    const [loading, setLoading] = useState(false);
    const [loadingBatches, setLoadingBatches] = useState(true);

    // Template editor
    const [editorTab, setEditorTab] = useState(0); // 0 = builtin, 1 = custom, 2 = editor
    const [editorContent, setEditorContent] = useState('');
    const [editorName, setEditorName] = useState('');
    const [editorPublic, setEditorPublic] = useState(false);
    const [editingTemplateId, setEditingTemplateId] = useState<number | null>(null);

    // Auto-fill sample template when opening empty editor
    useEffect(() => {
        if (editorTab === 2 && !editingTemplateId && !editorContent) {
            setEditorContent(getSampleTemplate(translate, isRtl));
            setEditorName(translate('pages.voucher.print.editor.sample_name'));
        }
    }, [editorTab, editingTemplateId, editorContent, translate, isRtl]);


    const previewRef = useRef<HTMLIFrameElement>(null);

    // --- Load batches and templates on mount ---
    useEffect(() => {
        const loadData = async () => {
            setLoadingBatches(true);
            try {
                const [batchData, templateData] = await Promise.all([
                    apiRequest<any[]>('/voucher-batches?perPage=1000&sort=id&order=DESC'),
                    apiRequest<any[]>('/voucher-templates'),
                ]);
                setBatches(Array.isArray(batchData) ? batchData : []);
                setCustomTemplates(Array.isArray(templateData) ? templateData : []);

                // Pre-select batch from URL query param
                const batchParam = searchParams.get('batch');
                if (batchParam) {
                    setSelectedBatchId(parseInt(batchParam, 10));
                }
            } catch (err) {
                notify('Failed to load data', { type: 'error' });
            }
            setLoadingBatches(false);
        };
        loadData();
    }, []);

    // --- Fetch product info when batch changes ---
    useEffect(() => {
        if (!selectedBatchId) return;
        const batch = batches.find((b) => b.id === selectedBatchId);
        if (!batch) return;
        if (products[batch.product_id]) return;

        apiRequest<Product>(`/products/${batch.product_id}`)
            .then((p) => setProducts((prev) => ({ ...prev, [batch.product_id]: p })))
            .catch(() => { });
    }, [selectedBatchId, batches]);



    // --- Fetch vouchers for preview/print when batch changes ---
    useEffect(() => {
        if (!selectedBatchId) {
            setVouchers([]);
            return;
        }
        setFetchingVouchers(true);
        apiRequest<VoucherData[]>(`/voucher-batches/${selectedBatchId}/print`)
            .then((data) => {
                setVouchers(Array.isArray(data) ? data : []);
            })
            .catch((err) => {
                console.error('Failed to fetch vouchers:', err);
                notify('Failed to fetch batch vouchers', { type: 'error' });
            })
            .finally(() => setFetchingVouchers(false));
    }, [selectedBatchId, notify]);

    // --- Derived state ---
    const selectedBatch = batches.find((b) => b.id === selectedBatchId);
    const product = selectedBatch ? products[selectedBatch.product_id] : undefined;
    const productColor = product?.color || '#2563eb';
    const productValidity = product?.validity_seconds || 0;

    // --- Build template variables ---


    const makeVoucherVars = useCallback(
        (v: VoucherData): TemplateVars => ({
            code: v.code,
            price: v.price > 0 ? `${v.price}` : '',
            validity: formatValidity(productValidity),
            quota: formatQuota(product?.data_quota),
            agent: agentName,
            hotspot: hotspotName,
            link: loginLink,
            serial: `${selectedBatchId}-${v.id}`,
            qr: printQR
                ? makeQR(
                    loginLink
                        ? `${loginLink}?username=${v.code}&password=${v.code}`
                        : v.code
                )
                : '',
            showQR: printQR,
            color: productColor,
            product: product?.name || '',
            rtl: locale === 'ar',
            t: translate,
        }),
        [productValidity, hotspotName, loginLink, printQR, productColor, selectedBatchId, product, agentName, locale, translate]
    );

    // --- Render a single voucher with the selected template ---
    const renderVoucher = useCallback(
        (vars: TemplateVars): string => {
            if (selectedTemplate === '__editor__') {
                return replaceCustomVars(editorContent, vars);
            }
            if (selectedTemplate.startsWith('custom_')) {
                const tmplId = parseInt(selectedTemplate.replace('custom_', ''), 10);
                const tmpl = customTemplates.find((t) => t.id === tmplId);
                if (tmpl) return replaceCustomVars(tmpl.content, vars);
                return `<div style="color:red;">Template not found</div>`;
            }
            const builtin = BUILTIN_TEMPLATES[selectedTemplate];
            if (builtin) return builtin.render(vars);
            return `<div style="color:red;">Unknown template</div>`;
        },
        [selectedTemplate, customTemplates, editorContent]
    );

    // --- Preview HTML ---
    const previewHtml = useMemo(() => {
        if (fetchingVouchers) {
            return `<html><body style="display:flex;align-items:center;justify-content:center;height:100vh;font-family:sans-serif;color:#666;">Loading vouchers...</body></html>`;
        }
        if (!selectedBatchId) {
            return `<html><body style="display:flex;align-items:center;justify-content:center;height:100vh;font-family:sans-serif;color:#666;">Select a batch to see preview</body></html>`;
        }
        if (vouchers.length === 0) {
            return `<html><body style="display:flex;align-items:center;justify-content:center;height:100vh;font-family:sans-serif;color:#666;">No vouchers in this batch</body></html>`;
        }

        const cards = vouchers
            .map((v) => renderVoucher(makeVoucherVars(v)))
            .join('\n');

        return `<!DOCTYPE html>
<html dir="${isRtl ? 'rtl' : 'ltr'}"><head>
<meta charset="utf-8">
<style>
html,body{margin:0;padding:2px;font-family:Arial,sans-serif;font-size:12px;-webkit-print-color-adjust:exact;print-color-adjust:exact;}
.voucher-container{display:flex;flex-wrap:wrap;justify-content:flex-start;gap:2px;}
</style>
</head><body>
<div class="voucher-container">${cards}</div>
</body></html>`;
    }, [fetchingVouchers, selectedBatchId, vouchers, renderVoucher, makeVoucherVars, isRtl]);

    // --- Update preview iframe ---
    useEffect(() => {
        const iframe = previewRef.current;
        if (!iframe) return;
        const doc = iframe.contentDocument || iframe.contentWindow?.document;
        if (!doc) return;
        doc.open();
        doc.write(previewHtml);
        doc.close();
    }, [previewHtml]);



    // --- Template CRUD ---
    const handleSaveTemplate = async () => {
        if (!editorName.trim() || !editorContent.trim()) {
            notify('Name and content are required', { type: 'warning' });
            return;
        }
        try {
            if (editingTemplateId) {
                await apiRequest(`/voucher-templates/${editingTemplateId}`, {
                    method: 'PUT',
                    body: JSON.stringify({ name: editorName, content: editorContent, is_public: editorPublic }),
                });
                notify('Template updated', { type: 'success' });
            } else {
                await apiRequest('/voucher-templates', {
                    method: 'POST',
                    body: JSON.stringify({ name: editorName, content: editorContent, is_public: editorPublic }),
                });
                notify('Template saved', { type: 'success' });
            }
            // Reload templates
            const data = await apiRequest<VoucherTemplate[]>('/voucher-templates');
            setCustomTemplates(Array.isArray(data) ? data : []);

            setEditorName('');
            setEditorContent('');
            setEditingTemplateId(null);
        } catch (err) {
            notify('Failed to save template', { type: 'error' });
        }
    };

    const handleDeleteTemplate = async (id: number) => {
        if (!window.confirm('Delete this template?')) return;
        try {
            await apiRequest(`/voucher-templates/${id}`, { method: 'DELETE' });
            notify('Template deleted', { type: 'success' });
            setCustomTemplates((prev) => prev.filter((t) => t.id !== id));
            if (selectedTemplate === `custom_${id}`) setSelectedTemplate('template1');
        } catch (err) {
            notify('Failed to delete template', { type: 'error' });
        }
    };

    const handleEditTemplate = (tmpl: VoucherTemplate) => {
        setEditorName(tmpl.name);
        setEditorContent(tmpl.content);
        setEditorPublic(tmpl.is_public);
        setEditingTemplateId(tmpl.id);
        setEditorTab(2);
    };

    // --- Render ---
    return (
        <Box>
            <Title title="Voucher Printing" />

            {/* Header */}
            <Box display="flex" alignItems="center" gap={2} mb={3} flexWrap="wrap">
                <Button
                    startIcon={isRtl ? <ArrowBackIcon sx={{ transform: 'rotate(180deg)' }} /> : <ArrowBackIcon />}
                    onClick={() => navigate('/voucher-batches')}
                    size="small"
                >
                    {translate('pages.voucher.print.back_to_batches')}
                </Button>
                <Typography variant="h5" sx={{ fontWeight: 700, flex: 1 }}>
                    {translate('pages.voucher.print.studio')}
                </Typography>
            </Box>

            <Box display="flex" gap={3} flexDirection={isMobile ? 'column' : 'row'} dir={isRtl ? 'rtl' : 'ltr'}>
                {/* ─── Left: Configuration Panel ─── */}
                <Card sx={{ flex: isMobile ? '1' : '0 0 380px', maxWidth: isMobile ? '100%' : 400 }}>
                    <CardContent>
                        <Typography variant="h6" gutterBottom sx={{ fontWeight: 600 }}>
                            {translate('pages.voucher.print.config')}
                        </Typography>
                        <Divider sx={{ mb: 2 }} />

                        {/* Batch selector */}
                        <Box mb={2}>
                            {loadingBatches ? (
                                <CircularProgress size={24} />
                            ) : (
                                <Autocomplete
                                    options={batches}
                                    value={batches.find((b) => b.id === selectedBatchId) || null}
                                    getOptionLabel={(b) => `${b.name} (ID: ${b.id}, ${b.count} ${translate('resources.vouchers.name', { count: b.count })})`}
                                    onChange={(_, newVal) => {
                                        setSelectedBatchId(newVal?.id || null);
                                    }}
                                    renderInput={(params) => (
                                        <TextField {...params} label={translate('pages.voucher.print.select_batch')} size="small" fullWidth />
                                    )}
                                    size="small"
                                />
                            )}
                        </Box>

                        {/* Hotspot Name */}
                        <Box mb={2}>
                            <TextField
                                fullWidth
                                size="small"
                                label={translate('pages.voucher.print.hotspot_name')}
                                value={hotspotName}
                                onChange={(e) => setHotspotName(e.target.value)}
                                inputProps={{ style: { textAlign: isRtl ? 'right' : 'left', direction: isRtl ? 'rtl' : 'ltr' } }}
                                sx={isRtl ? { '& .MuiInputLabel-root': { left: 'auto', right: 28, transformOrigin: 'top right' }, '& .MuiOutlinedInput-notchedOutline': { textAlign: 'right' }, '& .MuiFormHelperText-root': { textAlign: 'right' } } : undefined}
                            />
                        </Box>

                        {/* Agent Name */}
                        <Box mb={2}>
                            <TextField
                                fullWidth
                                size="small"
                                label={translate('pages.voucher.print.agent_name')}
                                value={agentName}
                                onChange={(e) => setAgentName(e.target.value)}
                                helperText={translate('pages.voucher.print.agent_name_placeholder')}
                                inputProps={{ style: { textAlign: isRtl ? 'right' : 'left', direction: isRtl ? 'rtl' : 'ltr' } }}
                                sx={isRtl ? { '& .MuiInputLabel-root': { left: 'auto', right: 28, transformOrigin: 'top right' }, '& .MuiOutlinedInput-notchedOutline': { textAlign: 'right' }, '& .MuiFormHelperText-root': { textAlign: 'right', mr: 0 } } : undefined}
                            />
                        </Box>

                        {/* Login Link */}
                        <Box mb={2}>
                            <TextField
                                fullWidth
                                size="small"
                                label={translate('pages.voucher.print.login_link')}
                                value={loginLink}
                                onChange={(e) => setLoginLink(e.target.value)}
                                helperText={translate('pages.voucher.print.login_link_placeholder')}
                                inputProps={{ style: { textAlign: isRtl ? 'right' : 'left', direction: isRtl ? 'rtl' : 'ltr' } }}
                                sx={isRtl ? { '& .MuiInputLabel-root': { left: 'auto', right: 28, transformOrigin: 'top right' }, '& .MuiOutlinedInput-notchedOutline': { textAlign: 'right' }, '& .MuiFormHelperText-root': { textAlign: 'right', mr: 0 } } : undefined}
                            />
                        </Box>

                        {/* QR Toggle */}
                        <Box mb={2}>
                            <FormControlLabel
                                control={
                                    <Checkbox
                                        checked={printQR}
                                        onChange={(e) => setPrintQR(e.target.checked)}
                                        size="small"
                                    />
                                }
                                label={translate('pages.voucher.print.print_qr')}
                                sx={{ ml: isRtl ? -1.5 : 0, mr: isRtl ? 0 : -1.5 }}
                            />
                        </Box>

                        <Divider sx={{ mb: 2 }} />

                        {/* Template Selection */}
                        <Typography variant="subtitle2" gutterBottom sx={{ fontWeight: 600 }}>
                            {translate('pages.voucher.print.template')}
                        </Typography>
                        <Tabs
                            value={editorTab}
                            onChange={(_, v) => setEditorTab(v)}
                            variant="fullWidth"
                            sx={{ mb: 2, minHeight: 36 }}
                        >
                            <Tab label={translate('pages.voucher.print.tabs.builtin')} sx={{ minHeight: 36, py: 0 }} />
                            <Tab label={translate('pages.voucher.print.tabs.custom')} sx={{ minHeight: 36, py: 0 }} />
                            <Tab label={translate('pages.voucher.print.tabs.editor')} sx={{ minHeight: 36, py: 0 }} />
                        </Tabs>

                        {/* Tab 0: Built-in templates */}
                        {editorTab === 0 && (
                            <Box display="flex" flexDirection="column" gap={1}>
                                {Object.keys(BUILTIN_TEMPLATES).map((key) => (
                                    <Button
                                        key={key}
                                        variant={selectedTemplate === key ? 'contained' : 'outlined'}
                                        size="small"
                                        onClick={() => setSelectedTemplate(key)}
                                        fullWidth
                                        sx={{
                                            justifyContent: 'flex-start',
                                            textAlign: isRtl ? 'right' : 'left'
                                        }}
                                    >
                                        {translate(`pages.voucher.print.templates.${key === 'template1' ? 'qr_detailed' : key === 'template2' ? 'box_simple' : key === 'template3' ? 'card_modern' : key === 'template4' ? 'gradient' : 'compact'}`)}
                                    </Button>
                                ))}
                            </Box>
                        )}

                        {/* Tab 1: Custom templates */}
                        {editorTab === 1 && (
                            <Box>
                                {customTemplates.length === 0 ? (
                                    <Typography color="text.secondary" variant="body2" sx={{ mb: 2 }}>
                                        {translate('pages.voucher.print.no_custom_templates')}
                                    </Typography>
                                ) : (
                                    <Box display="flex" flexDirection="column" gap={1}>
                                        {customTemplates.map((tmpl) => (
                                            <Box
                                                key={tmpl.id}
                                                display="flex"
                                                alignItems="center"
                                                gap={1}
                                            >
                                                <Button
                                                    variant={
                                                        selectedTemplate === `custom_${tmpl.id}`
                                                            ? 'contained'
                                                            : 'outlined'
                                                    }
                                                    size="small"
                                                    onClick={() =>
                                                        setSelectedTemplate(`custom_${tmpl.id}`)
                                                    }
                                                    sx={{ flex: 1, justifyContent: 'flex-start', textTransform: 'none' }}
                                                >
                                                    <Box display="flex" alignItems="center" width="100%">
                                                        <Typography variant="body2" sx={{ fontWeight: 500 }}>{tmpl.name}</Typography>
                                                        {tmpl.is_public ? (
                                                            <Chip
                                                                icon={<PublicIcon />}
                                                                label={translate('pages.voucher.print.editor.public_chip')}
                                                                size="small"
                                                                sx={{ ml: isRtl ? 0 : 1, mr: isRtl ? 1 : 0, height: 20 }}
                                                            />
                                                        ) : (
                                                            <Chip
                                                                icon={<LockIcon />}
                                                                label={translate('pages.voucher.print.editor.private_chip')}
                                                                size="small"
                                                                sx={{ ml: isRtl ? 0 : 1, mr: isRtl ? 1 : 0, height: 20 }}
                                                            />
                                                        )}
                                                        <Box flex={1} />
                                                        {tmpl.created_at && (
                                                            <Typography variant="caption" sx={{ fontSize: '0.65rem', opacity: 0.7 }}>
                                                                {new Date(tmpl.created_at).toLocaleDateString(locale, { year: 'numeric', month: 'short', day: 'numeric' })}
                                                            </Typography>
                                                        )}
                                                    </Box>
                                                </Button>
                                                {!tmpl.is_default && (
                                                    <>
                                                        <Tooltip title={translate('pages.voucher.print.editor.edit_tooltip')}>
                                                            <IconButton
                                                                size="small"
                                                                onClick={() => handleEditTemplate(tmpl)}
                                                            >
                                                                <EditIcon fontSize="small" />
                                                            </IconButton>
                                                        </Tooltip>
                                                        <Tooltip title={translate('pages.voucher.print.editor.delete_tooltip')}>
                                                            <IconButton
                                                                size="small"
                                                                color="error"
                                                                onClick={() =>
                                                                    handleDeleteTemplate(tmpl.id)
                                                                }
                                                            >
                                                                <DeleteIcon fontSize="small" />
                                                            </IconButton>
                                                        </Tooltip>
                                                    </>
                                                )}
                                            </Box>
                                        ))}
                                    </Box>
                                )}
                            </Box>
                        )}

                        {/* Tab 2: Template Editor */}
                        {editorTab === 2 && (
                            <Box>
                                <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                                    {translate('pages.voucher.print.editor_hint')}
                                </Typography>
                                <TextField
                                    fullWidth
                                    size="small"
                                    label={translate('pages.voucher.print.editor.name_label')}
                                    value={editorName}
                                    onChange={(e) => setEditorName(e.target.value)}
                                    sx={{ mb: 1 }}
                                />
                                <TextField
                                    fullWidth
                                    multiline
                                    minRows={8}
                                    maxRows={16}
                                    size="small"
                                    label={translate('pages.voucher.print.editor.html_label')}
                                    value={editorContent}
                                    onChange={(e) => setEditorContent(e.target.value)}
                                    sx={{ mb: 1, '& textarea': { fontFamily: 'monospace', fontSize: 12 }, direction: 'ltr' }}
                                />
                                <FormControlLabel
                                    control={
                                        <Checkbox
                                            checked={editorPublic}
                                            onChange={(e) => setEditorPublic(e.target.checked)}
                                            size="small"
                                        />
                                    }
                                    label={translate('pages.voucher.print.editor.make_public')}
                                />
                                <Box display="flex" gap={1} mt={1}>
                                    <Button
                                        variant="contained"
                                        size="small"
                                        startIcon={<SaveIcon />}
                                        onClick={handleSaveTemplate}
                                    >
                                        {editingTemplateId ? translate('pages.voucher.print.editor.update_btn') : translate('pages.voucher.print.editor.save_btn')}
                                    </Button>
                                    {editorContent && (
                                        <Button
                                            variant="outlined"
                                            size="small"
                                            startIcon={<PreviewIcon />}
                                            onClick={() => {
                                                // Temporarily use editor content for preview
                                                setSelectedTemplate('__editor__');
                                            }}
                                        >
                                            {translate('pages.voucher.print.editor.preview_btn')}
                                        </Button>
                                    )}
                                </Box>
                            </Box>
                        )}

                        <Divider sx={{ my: 2 }} />

                        {/* Print Button */}
                        <Button
                            variant="contained"
                            color="primary"
                            fullWidth
                            size="large"
                            startIcon={<PrintIcon />}
                            onClick={async () => {
                                if (!selectedBatchId) {
                                    notify(translate('pages.voucher.print.editor.error_select_batch'), { type: 'warning' });
                                    return;
                                }
                                setLoading(true);
                                try {
                                    const data = await apiRequest<VoucherData[]>(
                                        `/voucher-batches/${selectedBatchId}/print`
                                    );
                                    const voucherList = Array.isArray(data) ? data : [];
                                    if (voucherList.length === 0) {
                                        notify('No vouchers to print', { type: 'warning' });
                                        setLoading(false);
                                        return;
                                    }

                                    // Determine which renderer to use
                                    let renderFn: (vars: TemplateVars) => string;
                                    if (selectedTemplate === '__editor__' && editorContent) {
                                        renderFn = (vars) => replaceCustomVars(editorContent, vars);
                                    } else if (selectedTemplate.startsWith('custom_')) {
                                        const tmplId = parseInt(selectedTemplate.replace('custom_', ''), 10);
                                        const tmpl = customTemplates.find((t) => t.id === tmplId);
                                        renderFn = tmpl
                                            ? (vars) => replaceCustomVars(tmpl.content, vars)
                                            : () => '<div>Template not found</div>';
                                    } else {
                                        const builtin = BUILTIN_TEMPLATES[selectedTemplate];
                                        renderFn = builtin
                                            ? builtin.render
                                            : () => '<div>Unknown template</div>';
                                    }

                                    const allCards = voucherList
                                        .map((v) => renderFn(makeVoucherVars(v)))
                                        .join('\n');

                                    const printWindow = window.open('', '', 'width=900,height=700');
                                    if (!printWindow) {
                                        notify(translate('pages.voucher.print.popup_blocked_error'), { type: 'error' });
                                        setLoading(false);
                                        return;
                                    }

                                    const html = `<!DOCTYPE html>
<html dir="${isRtl ? 'rtl' : 'ltr'}"><head><title>${translate('pages.voucher.print.title')}</title>
<meta charset="utf-8">
<style>
@page{size:A4 portrait;margin:0.1cm;}
html,body{margin:0;padding:0;font-family:Arial,sans-serif;font-size:12px;-webkit-print-color-adjust:exact;print-color-adjust:exact;}
.voucher-container{display:flex;flex-wrap:wrap;justify-content:flex-start;gap:2px;}
@media print{.no-print{display:none !important;}}
</style>
</head><body>
<div class="voucher-container">${allCards}</div>
<script>
var images=document.getElementsByTagName('img');
var total=images.length,loaded=0;
function tryPrint(){loaded++;if(loaded>=total)setTimeout(function(){window.print();},300);}
if(total===0){window.print();}
else{for(var i=0;i<total;i++){if(images[i].complete){tryPrint();}else{images[i].onload=tryPrint;images[i].onerror=tryPrint;}}}
<\/script>
</body></html>`;

                                    printWindow.document.write(html);
                                    printWindow.document.close();
                                } catch (err) {
                                    notify(translate('pages.voucher.print.print_failed_error'), { type: 'error' });
                                }
                                setLoading(false);
                            }}
                            disabled={loading || !selectedBatchId}
                            sx={{ py: 1.5 }}
                        >
                            {loading ? translate('pages.voucher.print.preparing') : translate('pages.voucher.print.print_btn')}
                        </Button>
                    </CardContent>
                </Card>

                {/* ─── Right: Live Preview ─── */}
                <Card sx={{ flex: 1, minHeight: 400 }}>
                    <CardContent sx={{ height: '100%' }}>
                        <Box display="flex" alignItems="center" justifyContent="space-between" mb={1}>
                            <Typography variant="h6" sx={{ fontWeight: 600 }}>
                                {translate('pages.voucher.print.preview')}
                            </Typography>
                            <Chip
                                label={
                                    selectedTemplate.startsWith('custom_')
                                        ? customTemplates.find(
                                            (t) =>
                                                t.id ===
                                                parseInt(selectedTemplate.replace('custom_', ''), 10)
                                        )?.name || 'Custom'
                                        : selectedTemplate === '__editor__'
                                            ? 'Editor Preview'
                                            : BUILTIN_TEMPLATES[selectedTemplate]?.name || selectedTemplate
                                }
                                size="small"
                                color="primary"
                                variant="outlined"
                            />
                        </Box>
                        <Divider sx={{ mb: 1 }} />
                        <Box
                            sx={{
                                border: '1px solid',
                                borderColor: 'divider',
                                borderRadius: 1,
                                overflow: 'hidden',
                                bgcolor: '#fff',
                                height: 'calc(100% - 60px)',
                                minHeight: 350,
                            }}
                        >
                            <iframe
                                ref={previewRef}
                                title="Voucher Preview"
                                style={{
                                    width: '100%',
                                    height: '100%',
                                    border: 'none',
                                    minHeight: 350,
                                }}
                                sandbox="allow-same-origin"
                            />
                        </Box>
                    </CardContent>
                </Card>
            </Box>
        </Box>
    );
};

export default VoucherPrintingPage;
