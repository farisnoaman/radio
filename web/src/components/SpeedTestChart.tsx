import React from 'react';
import {
    Card,
    CardContent,
    Typography,
    Box,
} from '@mui/material';
import ReactECharts from 'echarts-for-react';

interface SpeedTestChartProps {
    results: Array<{
        created_at: string;
        upload_mbps: number;
        download_mbps: number;
    }>;
}

export const SpeedTestChart: React.FC<SpeedTestChartProps> = ({ results }) => {
    const chartData = results.map((result) => ({
        timestamp: new Date(result.created_at).getTime(),
        upload: result.upload_mbps,
        download: result.download_mbps,
    }));

    const option = {
        tooltip: {
            trigger: 'axis',
        },
        legend: {
            data: ['Download (Mbps)', 'Upload (Mbps)'],
        },
        xAxis: {
            type: 'time',
            name: 'Time',
        },
        yAxis: {
            type: 'value',
            name: 'Mbps',
        },
        series: [
            {
                name: 'Download (Mbps)',
                type: 'line',
                data: chartData.map(d => [d.timestamp, d.download]),
                smooth: true,
                color: '#4caf50',
            },
            {
                name: 'Upload (Mbps)',
                type: 'line',
                data: chartData.map(d => [d.timestamp, d.upload]),
                smooth: true,
                color: '#2196f3',
            },
        ],
    };

    return (
        <Card>
            <CardContent>
                <Typography variant="h6" gutterBottom>
                    Speed Test History
                </Typography>
                <Box sx={{ width: '100%', height: 300 }}>
                    <ReactECharts option={option} style={{ height: '100%', width: '100%' }} />
                </Box>
            </CardContent>
        </Card>
    );
};
