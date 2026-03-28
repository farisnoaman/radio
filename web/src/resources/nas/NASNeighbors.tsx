import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  CircularProgress,
  Alert,
  Button,
  Card,
  CardContent,
} from '@mui/material';
import { Refresh as RefreshIcon, Wifi as WifiIcon } from '@mui/icons-material';
import { useRecordContext, useNotify } from 'react-admin';

interface Neighbor {
  ip: string;
  mac?: string;
  interface: string;
  protocol: string;
  remote_id: string;
  state: string;
  device_type?: string;
}

export const NASNeighbors = () => {
  const record = useRecordContext();
  const notify = useNotify();
  const [neighbors, setNeighbors] = useState<Neighbor[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchNeighbors = async () => {
    if (!record) return;
    
    setLoading(true);
    setError(null);
    
    try {
      const apiUser = record.api_user || 'admin';
      const apiPass = record.api_pass || record.secret || '';
      
      const queryParams = new URLSearchParams({
        username: apiUser,
        password: apiPass,
      }).toString();
      
      const baseUrl = import.meta.env.VITE_API_BASE_URL || '';
      const response = await fetch(`${baseUrl}/api/v1/network/nas/${record.id}/neighbors?${queryParams}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const data = await response.json();
      // API returns { data: { count, neighbors: [] } }
      const neighborData = data?.data?.neighbors || [];
      setNeighbors(neighborData);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch neighbors');
      notify('Failed to fetch neighbors', { type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNeighbors();
  }, [record?.id]);

  const getProtocolColor = (protocol: string) => {
    switch (protocol?.toLowerCase()) {
      case 'ospf': return 'warning';
      case 'bgp': return 'info';
      case 'ppp': 
      case 'pppoe': 
      case 'hotspot': 
      case 'l2tp': 
      case 'pptp': 
      case 'sstp': 
        return 'success';
      case 'mndp': return 'primary';
      case 'static': return 'default';
      default: return 'default';
    }
  };

  const getProtocolLabel = (protocol: string) => {
    switch (protocol?.toLowerCase()) {
      case 'mndp': return 'Layer 2';
      case 'ospf': return 'OSPF';
      case 'bgp': return 'BGP';
      case 'pppoe': return 'PPPoE';
      case 'hotspot': return 'Hotspot';
      case 'l2tp': return 'L2TP';
      case 'pptp': return 'PPTP';
      case 'sstp': return 'SSTP';
      case 'ppp': return 'PPP';
      default: return protocol || 'Unknown';
    }
  };

  const getStateColor = (state: string) => {
    switch (state?.toLowerCase()) {
      case 'full': return 'success';
      case 'active': return 'success';
      case 'established': return 'success';
      default: return 'default';
    }
  };

  if (!record) {
    return null;
  }

  if (!record.api_user || !record.api_pass) {
    return (
      <Alert severity="warning" sx={{ m: 2 }}>
        API credentials not configured for this device. Please add API username and password in the device settings to discover neighbors.
      </Alert>
    );
  }

  return (
    <Box sx={{ p: 2 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h6" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <WifiIcon />
          Network Neighbors
        </Typography>
        <Button
          variant="outlined"
          startIcon={loading ? <CircularProgress size={16} /> : <RefreshIcon />}
          onClick={fetchNeighbors}
          disabled={loading}
        >
          Refresh
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
          <CircularProgress />
        </Box>
      ) : neighbors.length === 0 ? (
        <Card>
          <CardContent>
            <Typography variant="body1" color="text.secondary" textAlign="center">
              No neighbors discovered. Make sure the device is reachable and credentials are correct.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>IP Address</TableCell>
                <TableCell>MAC Address</TableCell>
                <TableCell>Interface</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Device/Identity</TableCell>
                <TableCell>State</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {neighbors.map((neighbor, index) => (
                <TableRow key={index}>
                  <TableCell sx={{ fontFamily: 'monospace' }}>{neighbor.ip}</TableCell>
                  <TableCell sx={{ fontFamily: 'monospace' }}>{neighbor.mac || '-'}</TableCell>
                  <TableCell>{neighbor.interface}</TableCell>
                  <TableCell>
                    <Chip 
                      label={getProtocolLabel(neighbor.protocol)} 
                      color={getProtocolColor(neighbor.protocol)} 
                      size="small" 
                    />
                  </TableCell>
                  <TableCell sx={{ fontFamily: 'monospace', fontSize: '0.8rem' }}>
                    {neighbor.remote_id || '-'}
                  </TableCell>
                  <TableCell>
                    <Chip 
                      label={neighbor.state || 'active'} 
                      color={getStateColor(neighbor.state || 'active')} 
                      size="small" 
                      variant="outlined"
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {!loading && neighbors.length > 0 && (
        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          Found {neighbors.length} neighbor(s)
        </Typography>
      )}
    </Box>
  );
};

export default NASNeighbors;
