import React from 'react';
import {
    Card,
    CardContent,
    Typography,
    Box,
} from '@mui/material';
import {
    Chart,
    ArgumentAxis,
    ValueAxis,
    LineSeries,
} from '@mui/x-charts';

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

    return (
        <Card>
            <CardContent>
                <Typography variant="h6" gutterBottom>
                    Speed Test History
                </Typography>
                <Box sx={{ width: '100%', height: 300 }}>
                    <Chart
                        data={chartData}
                        margin={{ top: 20, right: 30, left: 40, bottom: 30 }}
                    >
                        <ArgumentAxis />
                        <ValueAxis />
                        <LineSeries
                            label="Download (Mbps)"
                            valueKey="download"
                            color="#4caf50"
                        />
                        <LineSeries
                            label="Upload (Mbps)"
                            valueKey="upload"
                            color="#2196f3"
                        />
                    </Chart>
                </Box>
            </CardContent>
        </Card>
    );
};
