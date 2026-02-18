import { useState, useEffect, useRef } from 'react';

interface DashboardStats {
    online_count: number;
    today_income: number;
    active_vouchers: number;
    auth_requests: number;
    cpu_usage: number;
    memory_usage: number;
    [key: string]: any;
}

export const useDashboardWebSocket = () => {
    const [realtimeStats, setRealtimeStats] = useState<DashboardStats | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout>>();

    const connect = () => {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host;
        const wsUrl = `${protocol}//${host}/api/v1/dashboard/ws`;

        const ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            console.log('Dashboard WebSocket connected');
            setIsConnected(true);
            // Authenticate if needed, or backend handles it via cookie/header if possible (WS usually cookies)
        };

        ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                if (message.type === 'stats') {
                    setRealtimeStats(message.data);
                }
            } catch (e) {
                console.error('Failed to parse WebSocket message', e);
            }
        };

        ws.onclose = () => {
            console.log('Dashboard WebSocket disconnected');
            setIsConnected(false);
            // Attempt reconnect after 5 seconds
            reconnectTimeoutRef.current = setTimeout(connect, 5000);
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            ws.close();
        };

        wsRef.current = ws;
    };

    useEffect(() => {
        connect();
        return () => {
            if (wsRef.current) {
                wsRef.current.close();
            }
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
        };
    }, []);

    return { realtimeStats, isConnected };
};
