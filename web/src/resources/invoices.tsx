import React from "react";
import {
    List,
    Datagrid,
    TextField,
    DateField,
    NumberField,
    Show,
    SimpleShowLayout,
    useNotify,
    useRefresh,
    useRecordContext,
    TopToolbar,
    useSidebarState,
    FunctionField,
    useListContext,
    RecordContextProvider,
    useTranslate,
} from "react-admin";
import { useMediaQuery, Theme } from "@mui/material";
import {
    Box,
    Card,
    CardContent,
    Typography,
    Chip,
    Stack,
    alpha,
    CardActions,
    IconButton,
    Tooltip,
    Button as MuiButton,
} from "@mui/material";
import PaymentsIcon from "@mui/icons-material/Payments";
import ReceiptIcon from "@mui/icons-material/Receipt";
import MenuIcon from "@mui/icons-material/Menu";
import WhatsAppIcon from "@mui/icons-material/WhatsApp";
import NoteIcon from "@mui/icons-material/Note";
import { httpClient } from "../utils/apiClient";

/* -------------------------------------------------------------------------- */
/*  Print‑only CSS (kept for the Show view)                                 */
/* -------------------------------------------------------------------------- */
const printStyles = `
    @media print {
        @page {
            margin: 8mm;
            size: A4 portrait;
        }
        html, body {
            height: 100% !important;
            overflow: hidden !important;
            background: white !important;
            margin: 0 !important;
            padding: 0 !important;
            font-size: 10pt !important;
        }
        #root {
            height: 100% !important;
            overflow: hidden !important;
        }
        .MuiDrawer-root,
        .MuiDrawer-paper,
        .RaSidebar-root,
        .RaSidebar-fixed,
        header.MuiAppBar-root,
        .RaAppBar-root,
        .no-print,
        .RaShow-actions {
            display: none !important;
            visibility: hidden !important;
        }
        .RaLayout-appFrame,
        .RaLayout-main,
        .RaLayout-content {
            margin: 0 !important;
            padding: 0 !important;
            width: 100% !important;
            height: 100% !important;
            overflow: hidden !important;
        }
        .invoice-content {
            margin: 0 !important;
            padding: 0 !important;
            width: 100% !important;
            height: 100% !important;
            max-width: 100% !important;
            overflow: hidden !important;
            display: flex !important;
            flex-direction: column !important;
        }
        .MuiCard-root {
            page-break-inside: avoid;
            margin-bottom: 6px !important;
        }
        .MuiCardContent-root {
            padding: 8px 12px !important;
        }
        .MuiTypography-h5 {
            font-size: 16pt !important;
        }
        .MuiTypography-h6 {
            font-size: 12pt !important;
        }
        .MuiTypography-body1 {
            font-size: 10pt !important;
        }
        .MuiTypography-body2 {
            font-size: 9pt !important;
        }
        .MuiTypography-caption {
            font-size: 8pt !important;
        }
        .MuiStack-root {
            gap: 6px !important;
        }
    }
`;
const PrintStyles = () => <style>{printStyles}</style>;

/* -------------------------------------------------------------------------- */
/*  Helper – lazily load html2canvas & jsPDF and return a blob URL           */
/* -------------------------------------------------------------------------- */
const generatePdfUrl = async (): Promise<string> => {
    // Lazy‑load the libraries only when needed
    const html2canvas = (await import("html2canvas")).default;
    const { jsPDF } = await import("jspdf");

    const element = document.querySelector(
        ".invoice-content"
    ) as HTMLElement;
    if (!element) throw new Error("Invoice content not found");

    // Render the invoice DOM to a canvas
    const canvas = await html2canvas(element, { scale: 2 });
    const imgData = canvas.toDataURL("image/png");

    // Create a PDF with the canvas image
    const pdf = new jsPDF("p", "mm", "a4");
    const imgProps = pdf.getImageProperties(imgData);
    const pdfWidth = pdf.internal.pageSize.getWidth();
    const pdfHeight = (imgProps.height * pdfWidth) / imgProps.width;
    pdf.addImage(imgData, "PNG", 0, 0, pdfWidth, pdfHeight);

    // Return a temporary object‑URL that can be shared
    const blob = pdf.output("blob");
    return URL.createObjectURL(blob);
};

/* -------------------------------------------------------------------------- */
/*  Small reusable UI components                                             */
/* -------------------------------------------------------------------------- */
const StatusChip = ({
    label: _label,
    source: _source,
}: {
    label?: string;
    source?: string;
}) => {
    const record = useRecordContext();
    const translate = useTranslate();
    if (!record) return null;

    let color: "success" | "error" | "warning" | "default" = "default";
    const status = record.status || "unpaid";
    switch (status) {
        case "paid":
            color = "success";
            break;
        case "unpaid":
            color = "warning";
            break;
        case "overdue":
            color = "error";
            break;
    }

    return (
        <Chip
            label={translate(`resources.radius/invoices.status.${status}`)}
            color={color}
            size="small"
            sx={{ fontWeight: 600 }}
        />
    );
};

/* ---------------------------------------------------- */
/*  Pay button – Material‑UI (visible label)            */
/* ---------------------------------------------------- */
const PayButton = ({
    size = "small",
    fullWidth = false,
}: {
    size?: "small" | "medium" | "large";
    fullWidth?: boolean;
}) => {
    const record = useRecordContext();
    const notify = useNotify();
    const refresh = useRefresh();
    const translate = useTranslate();

    if (!record || record.status !== "unpaid") return null;

    const handlePay = async (e: React.MouseEvent) => {
        e.stopPropagation();
        try {
            await httpClient(`/radius/invoices/${record.id}/pay`, {
                method: "POST",
            });
            notify("resources.radius/invoices.notifications.paid", {
                type: "success",
            });
            refresh();
        } catch (error) {
            notify(
                "resources.radius/invoices.notifications.pay_error",
                { type: "error" }
            );
        }
    };

    return (
        <MuiButton
            variant="contained"
            color="primary"
            startIcon={<PaymentsIcon />}
            onClick={handlePay}
            size={size}
            fullWidth={fullWidth}
            sx={{
                boxShadow: 2,
                fontWeight: "bold",
                "&:hover": { boxShadow: 4 },
            }}
        >
            {translate("resources.radius/invoices.actions.pay")}
        </MuiButton>
    );
};

/* ---------------------------------------------------- */
/*  Icon‑only Pay button for the mobile grid             */
/* ---------------------------------------------------- */
const PayIconButton = () => {
    const record = useRecordContext();
    const notify = useNotify();
    const refresh = useRefresh();
    const translate = useTranslate();

    if (!record || record.status !== "unpaid") return null;

    const handlePay = async (e: React.MouseEvent) => {
        e.stopPropagation();
        try {
            await httpClient(`/radius/invoices/${record.id}/pay`, {
                method: "POST",
            });
            notify("resources.radius/invoices.notifications.paid", {
                type: "success",
            });
            refresh();
        } catch (error) {
            notify(
                "resources.radius/invoices.notifications.pay_error",
                { type: "error" }
            );
        }
    };

    return (
        <Tooltip title={translate("resources.radius/invoices.actions.pay")}>
            <IconButton
                onClick={handlePay}
                size="small"
                sx={{
                    backgroundColor: "primary.main",
                    color: "white",
                    "&:hover": { backgroundColor: "primary.dark" },
                    boxShadow: 2,
                    width: 32,
                    height: 32,
                }}
            >
                <PaymentsIcon fontSize="small" />
            </IconButton>
        </Tooltip>
    );
};

/* ---------------------------------------------------- */
/*  WhatsApp share – includes a PDF link                */
/* ---------------------------------------------------- */
const WhatsAppShareButton = ({
    fullWidth = false,
}: {
    fullWidth?: boolean;
}) => {
    const record = useRecordContext();
    const translate = useTranslate();
    if (!record) return null;

    const handleShare = async () => {
        try {
            const pdfUrl = await generatePdfUrl();
            const message = encodeURIComponent(
                translate("resources.radius/invoices.whatsapp.invoice_label", {
                    id: record.id,
                }) +
                    "\n\n" +
                    translate("resources.radius/invoices.whatsapp.user_label", {
                        username: record.username,
                    }) +
                    "\n" +
                    translate(
                        "resources.radius/invoices.whatsapp.amount_label",
                        { amount: Number(record.amount).toFixed(2) }
                    ) +
                    "\n" +
                    translate(
                        "resources.radius/invoices.whatsapp.base_fee_label",
                        { baseFee: Number(record.base_amount || 0).toFixed(2) }
                    ) +
                    "\n" +
                    translate(
                        "resources.radius/invoices.whatsapp.usage_label",
                        { usage: Number(record.usage_gb || 0).toFixed(2) }
                    ) +
                    "\n" +
                    translate(
                        "resources.radius/invoices.whatsapp.status_label",
                        {
                            status: translate(
                                `resources.radius/invoices.status.${record.status}`
                            ).toUpperCase(),
                        }
                    ) +
                    "\n" +
                    translate(
                        "resources.radius/invoices.whatsapp.due_date_label",
                        {
                            dueDate: new Date(
                                record.due_date
                            ).toLocaleDateString(),
                        }
                    ) +
                    "\n\n" +
                    translate("resources.radius/invoices.whatsapp.pdf_label", {
                        pdfUrl,
                    }) +
                    "\n\n" +
                    translate(
                        "resources.radius/invoices.whatsapp.payment_reminder"
                    )
            );
            window.open(`https://wa.me/?text=${message}`, "_blank");
        } catch (e) {
            console.error(e);
        }
    };

    return (
        <MuiButton
            variant="contained"
            color="success"
            startIcon={<WhatsAppIcon />}
            onClick={handleShare}
            fullWidth={fullWidth}
            size={fullWidth ? "medium" : "small"}
            sx={{
                boxShadow: 2,
                fontWeight: "bold",
                "&:hover": { boxShadow: 4 },
            }}
        >
            WhatsApp
        </MuiButton>
    );
};

/* ---------------------------------------------------- */
/*  WhatsApp icon‑only button for list cards (mobile)   */
/* ---------------------------------------------------- */
const WhatsAppListButton = () => {
    const record = useRecordContext();
    const translate = useTranslate();
    if (!record) return null;

    const handleShare = async (e: React.MouseEvent) => {
        e.stopPropagation();
        try {
            const pdfUrl = await generatePdfUrl();
            const message = encodeURIComponent(
                translate("resources.radius/invoices.whatsapp.invoice_label", {
                    id: record.id,
                }) +
                    "\n\n" +
                    translate("resources.radius/invoices.whatsapp.user_label", {
                        username: record.username,
                    }) +
                    "\n" +
                    translate(
                        "resources.radius/invoices.whatsapp.amount_label",
                        { amount: Number(record.amount).toFixed(2) }
                    ) +
                    "\n" +
                    translate(
                        "resources.radius/invoices.whatsapp.status_label",
                        {
                            status: translate(
                                `resources.radius/invoices.status.${record.status}`
                            ).toUpperCase(),
                        }
                    ) +
                    "\n" +
                    translate(
                        "resources.radius/invoices.whatsapp.due_date_label",
                        {
                            dueDate: new Date(
                                record.due_date
                            ).toLocaleDateString(),
                        }
                    ) +
                    "\n\n" +
                    translate("resources.radius/invoices.whatsapp.pdf_label", {
                        pdfUrl,
                    }) +
                    "\n\n" +
                    translate(
                        "resources.radius/invoices.whatsapp.payment_reminder"
                    )
            );
            window.open(`https://wa.me/?text=${message}`, "_blank");
        } catch (e) {
            console.error(e);
        }
    };

    return (
        <Tooltip title={translate("appbar.language.whatsapp") || "WhatsApp"}>
            <IconButton
                onClick={handleShare}
                size="small"
                sx={{
                    backgroundColor: "success.main",
                    color: "white",
                    "&:hover": { backgroundColor: "success.dark" },
                    boxShadow: 2,
                    width: 32,
                    height: 32,
                }}
            >
                <WhatsAppIcon fontSize="small" />
            </IconButton>
        </Tooltip>
    );
};

/* ---------------------------------------------------- */
/*  Responsive grid for mobile list view                */
/* ---------------------------------------------------- */
const InvoiceGrid = () => {
    const { data, isLoading } = useListContext();
    const translate = useTranslate();
    if (isLoading || !data) return null;

    return (
        <Box
            display="grid"
            gridTemplateColumns={{
                xs: "1fr",
                sm: "1fr 1fr",
                md: "repeat(3, 1fr)",
            }}
            gap={2}
            p={2}
            sx={{
                bgcolor: (theme) =>
                    theme.palette.mode === "dark"
                        ? "transparent"
                        : "rgba(0,0,0,0.02)",
            }}
        >
            {data.map((record) => (
                <RecordContextProvider value={record} key={record.id}>
                    <Card
                        elevation={0}
                        sx={{
                            borderRadius: 3,
                            border: (theme) =>
                                `1px solid ${theme.palette.divider}`,
                            cursor: "pointer",
                            transition: "box-shadow 0.2s",
                            "&:hover": { boxShadow: 4 },
                        }}
                        onClick={() => {
                            window.location.href = `#/radius/invoices/${record.id}/show`;
                        }}
                    >
                        <CardContent sx={{ pb: 1 }}>
                            {/* Header */}
                            <Box
                                display="flex"
                                justifyContent="space-between"
                                alignItems="flex-start"
                                mb={2}
                            >
                                <Box>
                                    <Typography
                                        variant="subtitle1"
                                        component="div"
                                        sx={{
                                            fontWeight: 700,
                                            lineHeight: 1.2,
                                        }}
                                    >
                                        {translate(
                                            "resources.radius/invoices.list.invoice_id",
                                            { id: record.id }
                                        )}
                                    </Typography>
                                    <Typography
                                        variant="caption"
                                        color="text.secondary"
                                        sx={{ fontFamily: "monospace" }}
                                    >
                                        {translate(
                                            "resources.radius/invoices.list.user_label",
                                            { username: record.username }
                                        )}
                                    </Typography>
                                </Box>
                                <StatusChip />
                            </Box>

                            {/* Summary */}
                            <Box
                                sx={{
                                    bgcolor: (theme) =>
                                        alpha(
                                            theme.palette.grey[500],
                                            0.05
                                        ),
                                    p: 1.5,
                                    borderRadius: 2,
                                    mb: 2,
                                }}
                            >
                                <Box
                                    display="flex"
                                    justifyContent="space-between"
                                    mb={1}
                                >
                                    <Typography
                                        variant="body2"
                                        color="text.secondary"
                                    >
                                        {translate(
                                            "resources.radius/invoices.list.usage_label"
                                        )}
                                    </Typography>
                                    <Typography
                                        variant="body2"
                                        sx={{ fontWeight: "bold" }}
                                    >
                                        {Number(record.usage_gb).toFixed(2)} GB
                                    </Typography>
                                </Box>
                                <Box
                                    display="flex"
                                    justifyContent="space-between"
                                    mb={1}
                                >
                                    <Typography
                                        variant="body2"
                                        color="text.secondary"
                                    >
                                        {translate(
                                            "resources.radius/invoices.list.amount_label"
                                        )}
                                    </Typography>
                                    <Typography
                                        variant="body2"
                                        sx={{
                                            fontWeight: "bold",
                                            color: "success.main",
                                        }}
                                    >
                                        ${Number(record.amount).toFixed(2)}
                                    </Typography>
                                </Box>
                                <Box
                                    display="flex"
                                    justifyContent="space-between"
                                >
                                    <Typography
                                        variant="body2"
                                        color="text.secondary"
                                    >
                                        {translate(
                                            "resources.radius/invoices.list.issued_label"
                                        )}
                                    </Typography>
                                    <Typography
                                        variant="body2"
                                        sx={{ fontFamily: "monospace" }}
                                    >
                                        {new Date(
                                            record.issue_date
                                        ).toLocaleDateString()}
                                    </Typography>
                                </Box>
                            </Box>
                        </CardContent>

                        {/* Action buttons – **no print button** */}
                        <CardActions
                            sx={{
                                justifyContent: "flex-end",
                                borderTop: (theme) =>
                                    `1px solid ${theme.palette.divider}`,
                                px: 2,
                                py: 1.5,
                                gap: 1,
                            }}
                        >
                            <Box onClick={(e) => e.stopPropagation()}>
                                <PayIconButton />
                            </Box>

                            <Box onClick={(e) => e.stopPropagation()}>
                                <WhatsAppListButton />
                            </Box>
                        </CardActions>
                    </Card>
                </RecordContextProvider>
            ))}
        </Box>
    );
};

/* -------------------------------------------------------------------------- */
/*  LIST VIEW                                                                */
/* -------------------------------------------------------------------------- */
export const InvoiceList = () => {
    const isMobile = useMediaQuery(
        (theme: Theme) => theme.breakpoints.down("sm")
    );

    return (
        <List sort={{ field: "id", order: "DESC" }}>
            {isMobile ? (
                <InvoiceGrid />
            ) : (
                <Datagrid rowClick="show">
                    <TextField source="id" />
                    <TextField source="username" />
                    <NumberField
                        source="usage_gb"
                        options={{ maximumFractionDigits: 2 }}
                    />
                    <NumberField
                        source="amount"
                        options={{ style: "currency", currency: "USD" }}
                    />
                    <StatusChip source="status" />
                    <DateField source="issue_date" />
                    <PayButton />
                    <WhatsAppListButton />
                </Datagrid>
            )}
        </List>
    );
};

/* -------------------------------------------------------------------------- */
/*  SHOW VIEW – toolbar (hamburger menu, no back, no print)                 */
/* -------------------------------------------------------------------------- */
const InvoiceShowActions = () => {
    const [, setSidebarOpen] = useSidebarState();

    return (
        <TopToolbar
            sx={{
                justifyContent: "space-between",
                flexWrap: "wrap",
                gap: 1,
            }}
        >
            {/* Hamburger menu – opens the sidebar */}
            <IconButton onClick={() => setSidebarOpen(true)} size="large">
                <MenuIcon />
            </IconButton>

            {/* WhatsApp share (PDF attached) */}
            <Box sx={{ display: "flex", gap: 1, flexWrap: "wrap" }}>
                <WhatsAppShareButton />
            </Box>
        </TopToolbar>
    );
};

/* -------------------------------------------------------------------------- */
/*  SHOW VIEW – main layout                                                */
/* -------------------------------------------------------------------------- */
export const InvoiceShow = () => {
    const isMobile = useMediaQuery(
        (theme: Theme) => theme.breakpoints.down("sm")
    );
    const translate = useTranslate();

    return (
        <Show actions={<InvoiceShowActions />}>
            <PrintStyles />
            <SimpleShowLayout>
                <Box sx={{ width: "100%" }} className="invoice-content">
                    <Stack
                        spacing={isMobile ? 1 : 2}
                        sx={{ height: "100%" }}
                    >
                        {/* -------------------------------------------------- */}
                        {/* Header summary card                               */}
                        {/* -------------------------------------------------- */}
                        <Card
                            elevation={3}
                            sx={{
                                borderRadius: 2,
                                background:
                                    "linear-gradient(135deg, #1e3a8a 0%, #3b82f6 100%)",
                                color: "white",
                                flexShrink: 0,
                                mx: isMobile ? 0 : 2,
                                mt: isMobile ? 0 : 2,
                            }}
                        >
                            <CardContent sx={{ p: isMobile ? 2 : 3 }}>
                                <Box
                                    display="flex"
                                    justifyContent="space-between"
                                    alignItems="center"
                                >
                                    <Box>
                                        <Typography
                                            variant={isMobile ? "h6" : "h5"}
                                            fontWeight={700}
                                        >
                                            {translate(
                                                "resources.radius/invoices.show.invoice_title"
                                            )}
                                        </Typography>
                                        <Typography
                                            variant="body2"
                                            sx={{ opacity: 0.9, mt: 0.5 }}
                                        >
                                            {translate(
                                                "resources.radius/invoices.show.user_label"
                                            )}
                                            :{" "}
                                            <TextField
                                                source="username"
                                                sx={{
                                                    fontWeight: 600,
                                                    color: "inherit",
                                                }}
                                            />
                                        </Typography>
                                    </Box>

                                    <Box textAlign="right">
                                        <StatusChip />
                                        <Typography
                                            variant={isMobile ? "h5" : "h4"}
                                            fontWeight={800}
                                            sx={{ mt: 1 }}
                                        >
                                            <NumberField
                                                source="amount"
                                                options={{
                                                    style: "currency",
                                                    currency: "USD",
                                                }}
                                            />
                                        </Typography>
                                        <Typography
                                            variant="caption"
                                            sx={{
                                                opacity: 0.8,
                                                display: "block",
                                            }}
                                        >
                                            {translate(
                                                "resources.radius/invoices.show.total_due"
                                            )}
                                        </Typography>
                                    </Box>
                                </Box>
                            </CardContent>
                        </Card>

                        {/* -------------------------------------------------- */}
                        {/* Main two‑column grid                           */}
                        {/* -------------------------------------------------- */}
                        <Box
                            display="grid"
                            gridTemplateColumns={{
                                xs: "1fr",
                                md: "2fr 1fr",
                            }}
                            gap={isMobile ? 1 : 2}
                            sx={{
                                flex: 1,
                                minHeight: 0,
                                px: isMobile ? 0 : 2,
                            }}
                        >
                            {/* -------------------- LEFT COLUMN -------------------- */}
                            <Stack spacing={isMobile ? 1 : 2}>
                                {/* Billing breakdown */}
                                <Card elevation={2} sx={{ borderRadius: 2 }}>
                                    <CardContent sx={{ p: isMobile ? 2 : 3 }}>
                                        <Typography
                                            variant="h6"
                                            fontWeight={600}
                                            gutterBottom
                                            sx={{
                                                display: "flex",
                                                alignItems: "center",
                                                gap: 1,
                                                mb: 2,
                                            }}
                                        >
                                            <ReceiptIcon color="primary" />
                                            {translate(
                                                "resources.radius/invoices.show.billing_breakdown"
                                            )}
                                        </Typography>

                                        <Box>
                                            {/* Base monthly fee */}
                                            <Box
                                                display="flex"
                                                justifyContent="space-between"
                                                py={1}
                                                borderBottom={(theme) => `1px solid ${theme.palette.divider}`}
                                            >
                                                <Typography color="textSecondary">
                                                    {translate(
                                                        "resources.radius/invoices.show.base_monthly_fee"
                                                    )}
                                                </Typography>
                                                <Typography fontWeight={600}>
                                                    <NumberField
                                                        source="base_amount"
                                                        options={{
                                                            style: "currency",
                                                            currency: "USD",
                                                        }}
                                                    />
                                                </Typography>
                                            </Box>

                                            {/* Data usage */}
                                            <Box
                                                display="flex"
                                                justifyContent="space-between"
                                                py={1}
                                                borderBottom={(theme) => `1px solid ${theme.palette.divider}`}
                                            >
                                                <Typography color="textSecondary">
                                                    {translate(
                                                        "resources.radius/invoices.show.data_usage"
                                                    )}
                                                </Typography>
                                                <Typography fontWeight={600}>
                                                    <NumberField
                                                        source="usage_gb"
                                                        options={{
                                                            maximumFractionDigits: 2,
                                                        }}
                                                    />{" "}
                                                    {translate(
                                                        "resources.radius/invoices.gb"
                                                    )}
                                                </Typography>
                                            </Box>

                                            {/* Price per GB */}
                                            <Box
                                                display="flex"
                                                justifyContent="space-between"
                                                py={1}
                                                borderBottom={(theme) => `1px solid ${theme.palette.divider}`}
                                            >
                                                <Typography color="textSecondary">
                                                    {translate(
                                                        "resources.radius/invoices.show.price_per_gb"
                                                    )}
                                                </Typography>
                                                <Typography fontWeight={600}>
                                                    <NumberField
                                                        source="price_per_gb"
                                                        options={{
                                                            style: "currency",
                                                            currency: "USD",
                                                        }}
                                                    />
                                                </Typography>
                                            </Box>

                                            {/* Usage charge – highlighted */}
                                            <Box
                                                display="flex"
                                                justifyContent="space-between"
                                                py={1}
                                                borderBottom={(theme) => `1px solid ${theme.palette.divider}`}
                                                bgcolor={(theme) => alpha(theme.palette.primary.main, 0.04)}
                                                px={1}
                                            >
                                                <Typography
                                                    fontWeight={600}
                                                    color="primary"
                                                >
                                                    {translate(
                                                        "resources.radius/invoices.show.usage_charge"
                                                    )}
                                                </Typography>
                                                <FunctionField
                                                    render={(record) => {
                                                        const consumption =
                                                            (record.usage_gb ||
                                                                0) *
                                                            (record.price_per_gb ||
                                                                0);
                                                        return (
                                                            <Typography
                                                                fontWeight={700}
                                                                color="primary"
                                                            >
                                                                {consumption.toLocaleString(
                                                                    "en-US",
                                                                    {
                                                                        style: "currency",
                                                                        currency:
                                                                            "USD",
                                                                    }
                                                                )}
                                                            </Typography>
                                                        );
                                                    }}
                                                />
                                            </Box>

                                            {/* Total amount */}
                                            <Box
                                                display="flex"
                                                justifyContent="space-between"
                                                py={1.5}
                                            >
                                                <Typography
                                                    variant="h6"
                                                    fontWeight={700}
                                                >
                                                    {translate(
                                                        "resources.radius/invoices.show.total_amount"
                                                    )}
                                                </Typography>
                                                <Typography
                                                    variant="h6"
                                                    fontWeight={800}
                                                    color="primary.main"
                                                >
                                                    <NumberField
                                                        source="amount"
                                                        options={{
                                                            style: "currency",
                                                            currency: "USD",
                                                        }}
                                                    />
                                                </Typography>
                                            </Box>
                                        </Box>
                                    </CardContent>
                                </Card>

                                {/* Usage statistics */}
                                <Card elevation={2} sx={{ borderRadius: 2 }}>
                                    <CardContent sx={{ p: isMobile ? 2 : 3 }}>
                                        <Typography
                                            variant="h6"
                                            fontWeight={600}
                                            gutterBottom
                                            sx={{
                                                display: "flex",
                                                alignItems: "center",
                                                gap: 1,
                                                mb: 2,
                                            }}
                                        >
                                            <PaymentsIcon color="primary" />
                                            {translate(
                                                "resources.radius/invoices.show.usage_statistics"
                                            )}
                                        </Typography>

                                        <Box
                                            display="grid"
                                            gridTemplateColumns="1fr 1fr"
                                            gap={2}
                                        >
                                            <Box
                                                p={2}
                                                bgcolor={(theme) => alpha(theme.palette.action.hover, 0.5)}
                                                borderRadius={2}
                                            >
                                                <Typography
                                                    variant="body2"
                                                    color="textSecondary"
                                                >
                                                    {translate(
                                                        "resources.radius/invoices.show.total_sessions"
                                                    )}
                                                </Typography>
                                                <Typography
                                                    variant="h5"
                                                    fontWeight={700}
                                                >
                                                    <NumberField
                                                        source="session_count"
                                                    />
                                                </Typography>
                                            </Box>

                                            <Box
                                                p={2}
                                                bgcolor={(theme) => alpha(theme.palette.action.hover, 0.5)}
                                                borderRadius={2}
                                            >
                                                <Typography
                                                    variant="body2"
                                                    color="textSecondary"
                                                >
                                                    {translate(
                                                        "resources.radius/invoices.show.data_consumed"
                                                    )}
                                                </Typography>
                                                <Typography
                                                    variant="h5"
                                                    fontWeight={700}
                                                >
                                                    <NumberField
                                                        source="usage_gb"
                                                        options={{
                                                            maximumFractionDigits: 2,
                                                        }}
                                                    />{" "}
                                                    {translate(
                                                        "resources.radius/invoices.gb"
                                                    )}
                                                </Typography>
                                            </Box>
                                        </Box>
                                    </CardContent>
                                </Card>
                            </Stack>

                            {/* -------------------- RIGHT COLUMN -------------------- */}
                            <Stack spacing={isMobile ? 1 : 2}>
                                {/* Dates & references */}
                                <Card elevation={2} sx={{ borderRadius: 2 }}>
                                    <CardContent sx={{ p: isMobile ? 2 : 3 }}>
                                        <Typography
                                            variant="subtitle2"
                                            color="textSecondary"
                                            gutterBottom
                                        >
                                            {translate(
                                                "resources.radius/invoices.show.invoice_reference"
                                            )}
                                        </Typography>
                                        <Typography
                                            variant="h6"
                                            fontWeight={600}
                                            gutterBottom
                                        >
                                            #<TextField source="id" />
                                        </Typography>

                                        {/* Issue date */}
                                        <Box mt={1}>
                                            <Typography
                                                variant="caption"
                                                color="textSecondary"
                                                display="block"
                                            >
                                                {translate(
                                                    "resources.radius/invoices.show.issue_date"
                                                )}
                                            </Typography>
                                            <Typography fontWeight={500}>
                                                <DateField
                                                    source="issue_date"
                                                    showTime
                                                />
                                            </Typography>
                                        </Box>

                                        {/* Due date – highlighted */}
                                        <Box
                                            mt={2}
                                            p={2}
                                            borderRadius={2}
                                            bgcolor={alpha("#ef4444", 0.05)}
                                            border={`1px solid ${alpha(
                                                "#ef4444",
                                                0.2
                                            )}`}
                                        >
                                            <Typography
                                                variant="body2"
                                                color="error"
                                                display="block"
                                                fontWeight={600}
                                            >
                                                {translate(
                                                    "resources.radius/invoices.show.due_date"
                                                )}
                                            </Typography>
                                            <Typography
                                                variant="h6"
                                                fontWeight={700}
                                                color="error.main"
                                            >
                                                <DateField source="due_date" />
                                            </Typography>
                                        </Box>

                                        {/* Billing period */}
                                        <Box mt={2}>
                                            <Typography
                                                variant="caption"
                                                color="textSecondary"
                                                display="block"
                                            >
                                                {translate(
                                                    "resources.radius/invoices.show.billing_period"
                                                )}
                                            </Typography>
                                            <Typography fontWeight={500}>
                                                <DateField
                                                    source="billing_period_start"
                                                />{" "}
                                                -{" "}
                                                <DateField
                                                    source="billing_period_end"
                                                />
                                            </Typography>
                                        </Box>

                                        {/* Paid‑at (optional) */}
                                        {useRecordContext()?.paid_at && (
                                            <Box
                                                mt={2}
                                                p={2}
                                                borderRadius={2}
                                                bgcolor={alpha("#10b981", 0.05)}
                                                border={`1px solid ${alpha(
                                                    "#10b981",
                                                    0.2
                                                )}`}
                                            >
                                                <Typography
                                                    variant="body2"
                                                    color="success.main"
                                                    display="block"
                                                    fontWeight={600}
                                                >
                                                    {translate(
                                                        "resources.radius/invoices.show.paid_on"
                                                    )}
                                                </Typography>
                                                <Typography
                                                    variant="h6"
                                                    fontWeight={700}
                                                    color="success.main"
                                                >
                                                    <DateField
                                                        source="paid_at"
                                                        showTime
                                                    />
                                                </Typography>
                                            </Box>
                                        )}
                                    </CardContent>
                                </Card>

                                {/* ------------------- QUICK ACTIONS ------------------- */}
                                <Card
                                    elevation={3}
                                    sx={{
                                        borderRadius: 2,
                                        bgcolor: "background.paper",
                                        border: "1px solid",
                                        borderColor: "divider",
                                    }}
                                    className="no-print"
                                >
                                    <CardContent
                                        sx={{ p: isMobile ? 2 : 3 }}
                                    >
                                        <Typography
                                            variant="h6"
                                            gutterBottom
                                            sx={{
                                                fontWeight: 600,
                                                mb: 2,
                                            }}
                                        >
                                            {translate(
                                                "resources.radius/invoices.show.quick_actions"
                                            )}
                                        </Typography>

                                        <Box
                                            display="flex"
                                            flexDirection="column"
                                            gap={1.5}
                                        >
                                            {/* Pay */}
                                            <PayButton
                                                size={
                                                    isMobile ? "medium" : "small"
                                                }
                                                fullWidth={isMobile}
                                            />

                                            {/* WhatsApp (PDF attached) */}
                                            <WhatsAppShareButton
                                                fullWidth={isMobile}
                                            />
                                        </Box>
                                    </CardContent>
                                </Card>
                            </Stack>
                        </Box>

                        {/* ------------------- Remarks (non‑print) ------------------- */}
                        <Card
                            elevation={1}
                            sx={{
                                borderRadius: 2,
                                bgcolor: "background.paper",
                                border: (theme) => `1px dashed ${theme.palette.divider}`,
                                flexShrink: 0,
                                mx: isMobile ? 0 : 2,
                                mb: isMobile ? 0 : 2,
                            }}
                            className="no-print"
                        >
                            <CardContent sx={{ p: isMobile ? 2 : 3 }}>
                                <Typography
                                    variant="subtitle1"
                                    color="textSecondary"
                                    gutterBottom
                                >
                                    <NoteIcon
                                        sx={{
                                            fontSize: 20,
                                            verticalAlign: "middle",
                                            mr: 0.5,
                                        }}
                                    />{" "}
                                    {translate(
                                        "resources.radius/invoices.show.internal_remarks"
                                    )}
                                </Typography>
                                <Typography variant="body1">
                                    <TextField
                                        source="remark"
                                        emptyText={translate(
                                            "resources.radius/invoices.show.no_remarks"
                                        )}
                                    />
                                </Typography>
                            </CardContent>
                        </Card>
                    </Stack>
                </Box>
            </SimpleShowLayout>
        </Show>
    );
};