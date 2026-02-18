import { useMemo } from 'react';
import ReactECharts from 'echarts-for-react';
import { Card, CardContent, CardHeader, Typography, Box, CircularProgress, useTheme } from '@mui/material';
import { useApiQuery } from '../hooks/useApiQuery';

interface ForecastData {
    period: string; // "next_30_days", "next_90_days"
    predictions: {
        date: string;
        value: number;
        lower_bound: number;
        upper_bound: number;
    }[];
    confidence_score: number;
    model: string; // "linear_regression"
}

export const ForecastChart = () => {
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';

    const { data, isLoading, error } = useApiQuery<ForecastData>({
        path: '/analytics/forecast?period=30', // Default to 30 days
        queryKey: ['analytics', 'forecast', '30'],
        staleTime: 24 * 60 * 60 * 1000, // Cache for 24 hours
    });

    const option = useMemo(() => {
        if (!data || !data.predictions) return {};

        const dates = data.predictions.map(p => p.date);
        const values = data.predictions.map(p => Math.round(p.value));
        // const lowerBounds = data.predictions.map(p => Math.round(p.lower_bound));
        // const upperBounds = data.predictions.map(p => Math.round(p.upper_bound));

        return {
            tooltip: {
                trigger: 'axis',
                backgroundColor: isDark ? '#333' : '#fff',
                textStyle: { color: isDark ? '#fff' : '#333' }
            },
            grid: {
                left: '3%',
                right: '4%',
                bottom: '3%',
                containLabel: true
            },
            xAxis: {
                type: 'category',
                boundaryGap: false,
                data: dates,
                axisLine: { lineStyle: { color: isDark ? '#555' : '#ccc' } },
                axisLabel: { color: isDark ? '#aaa' : '#666' }
            },
            yAxis: {
                type: 'value',
                axisLine: { show: false },
                splitLine: { lineStyle: { color: isDark ? '#333' : '#eee' } },
                axisLabel: { color: isDark ? '#aaa' : '#666' }
            },
            series: [
                {
                    name: 'Projected Users',
                    type: 'line',
                    smooth: true,
                    data: values,
                    lineStyle: { width: 3, color: theme.palette.primary.main },
                    areaStyle: {
                        color: {
                            type: 'linear',
                            x: 0, y: 0, x2: 0, y2: 1,
                            colorStops: [
                                { offset: 0, color: theme.palette.primary.main },
                                { offset: 1, color: theme.palette.primary.light + '00' } // Transparent
                            ]
                        },
                        opacity: 0.2
                    },
                    itemStyle: { color: theme.palette.primary.main }
                }
            ]
        };
    }, [data, isDark, theme]);

    if (isLoading) return <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}><CircularProgress /></Box>;
    if (error) return null; // Hide if error (server might not support analytics yet)

    return (
        <Card sx={{ mt: 2 }}>
            <CardHeader
                title={
                    <Box display="flex" alignItems="center" justifyContent="space-between">
                        <Typography variant="h6">User Growth Forecast (Next 30 Days)</Typography>
                        {data?.confidence_score && (
                            <Typography variant="caption" color="textSecondary">
                                Confidence: {(data.confidence_score * 100).toFixed(1)}%
                            </Typography>
                        )}
                    </Box>
                }
            />
            <CardContent>
                <ReactECharts option={option} style={{ height: 300 }} />
                <Typography variant="caption" color="textSecondary" sx={{ fontStyle: 'italic', display: 'block', textAlign: 'center', mt: 1 }}>
                    Based on historical data (Model: {data?.model || 'Linear Regression'})
                </Typography>
            </CardContent>
        </Card>
    );
};
