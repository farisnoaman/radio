import React, { useMemo } from 'react';
import ReactECharts from 'echarts-for-react';
import { Card, CardContent, CardHeader, Typography, Box, CircularProgress, useTheme } from '@mui/material';
import { useApiQuery } from '../hooks/useApiQuery';

interface DataPoint {
    timestamp: string;
    value: number;
}

interface ForecastData {
    metric: string;
    forecast: DataPoint[];
    history: DataPoint[];
    confidence: number;
}

export const ForecastChart = () => {
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';

    const { data, isLoading, error } = useApiQuery<ForecastData>({
        path: '/analytics/forecast?metric=users&days=30&future=7',
        queryKey: ['analytics', 'forecast', 'users', '30'],
        staleTime: 60 * 60 * 1000, // 1 hour
    });

    const option = useMemo(() => {
        if (!data || (!data.history && !data.forecast)) return {};

        const allPoints = [...(data.history || []), ...(data.forecast || [])];
        const dates = allPoints.map(p => new Date(p.timestamp).toLocaleDateString());
        const historyValues = data.history?.map(p => Math.round(p.value)) || [];
        const forecastValues = data.forecast?.map(p => Math.round(p.value)) || [];

        // Fill history with nulls for forecast period and vice versa
        const historyData = [...historyValues, ...new Array(forecastValues.length).fill(null)];
        const forecastData = [...new Array(historyValues.length - 1).fill(null), historyValues[historyValues.length - 1], ...forecastValues];

        return {
            tooltip: {
                trigger: 'axis',
                backgroundColor: isDark ? '#333' : '#fff',
                textStyle: { color: isDark ? '#fff' : '#333' }
            },
            legend: {
                data: ['Actual', 'Forecast'],
                bottom: 0,
                textStyle: { color: isDark ? '#aaa' : '#666' }
            },
            grid: {
                left: '3%',
                right: '4%',
                bottom: '15%',
                containLabel: true
            },
            xAxis: {
                type: 'category',
                boundaryGap: false,
                data: dates,
                axisLine: { lineStyle: { color: isDark ? '#555' : '#ccc' } },
                axisLabel: { color: isDark ? '#aaa' : '#666', rotate: 45 }
            },
            yAxis: {
                type: 'value',
                axisLine: { show: false },
                splitLine: { lineStyle: { color: isDark ? '#333' : '#eee' } },
                axisLabel: { color: isDark ? '#aaa' : '#666' }
            },
            series: [
                {
                    name: 'Actual',
                    type: 'line',
                    smooth: true,
                    data: historyData,
                    lineStyle: { width: 3, color: theme.palette.primary.main },
                    itemStyle: { color: theme.palette.primary.main }
                },
                {
                    name: 'Forecast',
                    type: 'line',
                    smooth: true,
                    data: forecastData,
                    lineStyle: { width: 3, type: 'dashed', color: theme.palette.secondary.main },
                    itemStyle: { color: theme.palette.secondary.main }
                }
            ]
        };
    }, [data, isDark, theme]);

    if (isLoading) return <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}><CircularProgress /></Box>;
    if (error) return null;

    return (
        <Card sx={{ mt: 2, borderRadius: 4 }}>
            <CardHeader
                title={
                    <Box display="flex" alignItems="center" justifyContent="space-between">
                        <Typography variant="h6" sx={{ fontWeight: 700 }}>
                            {data?.metric === 'users' ? 'User Growth Forecast' : 'Traffic Forecast'}
                        </Typography>
                        {data?.confidence && (
                            <Typography variant="caption" color="textSecondary">
                                Accuracy: {(data.confidence * 100).toFixed(1)}%
                            </Typography>
                        )}
                    </Box>
                }
            />
            <CardContent>
                <ReactECharts option={option} style={{ height: 350 }} />
                <Typography variant="caption" color="textSecondary" sx={{ fontStyle: 'italic', display: 'block', textAlign: 'center', mt: 1 }}>
                    Forecast based on linear regression of historical data.
                </Typography>
            </CardContent>
        </Card>
    );
};
