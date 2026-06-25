import { useState, useEffect, useCallback } from 'react'
import {
  Box, Typography, Table, TableBody, TableCell, TableHead, TableRow,
  Paper, Chip, CircularProgress, Drawer, IconButton, Divider,
  TablePagination, TextField, InputAdornment,
} from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import { useTheme } from '@mui/material/styles'
import CloseIcon from '@mui/icons-material/Close'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip,
  ResponsiveContainer, Cell,
} from 'recharts'
import { deviceService } from '../../services/deviceService'
import type { DeviceResponse, GetDeviceStatsResponse } from '../../apis/proto/openclick/v1/person'
import { useNotification } from '@gofreego/tsutils'
import { COLORS } from '../dashboard/tabInfo'

const PROP_LABELS: Record<string, string> = {
  $browser: 'Browser', $browser_version: 'Version',
  $os: 'OS', $os_version: 'OS Ver',
  $device_type: 'Type',
  $lib: 'Library', $lib_version: 'Lib Ver',
  $screen_height: 'Screen H', $screen_width: 'Screen W',
  $viewport_height: 'Viewport H', $viewport_width: 'Viewport W',
  $referrer: 'Referrer',
  $user_agent: 'User Agent',
}

function StatChart({ title, data }: { title: string; data: { value: string; count: string }[] }) {
  const theme = useTheme()
  const chartData = data.map(d => ({ name: d.value || '(unknown)', count: Number(d.count) }))
  if (!chartData.length) return null
  return (
    <Paper sx={{ p: 2 }}>
      <Typography variant="subtitle2" fontWeight={600} gutterBottom>{title}</Typography>
      <ResponsiveContainer width="100%" height={160}>
        <BarChart data={chartData} layout="vertical" margin={{ top: 0, right: 20, left: 0, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke={theme.palette.divider} horizontal={false} />
          <XAxis type="number" tick={{ fontSize: 11, fill: theme.palette.text.secondary }}
            axisLine={{ stroke: theme.palette.divider }} tickLine={false} />
          <YAxis type="category" dataKey="name" width={100} tick={{ fontSize: 11, fill: theme.palette.text.primary }}
            axisLine={false} tickLine={false} />
          <RechartsTooltip contentStyle={{
            backgroundColor: theme.palette.background.paper,
            border: `1px solid ${theme.palette.divider}`,
            color: theme.palette.text.primary,
          }} />
          <Bar dataKey="count" name="Devices" radius={[0, 4, 4, 0]}>
            {chartData.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </Paper>
  )
}

function DeviceDrawer({ device, onClose }: { device: DeviceResponse | null; onClose: () => void }) {
  if (!device) return null
  const props = device.properties ?? {}
  return (
    <Drawer anchor="right" open={!!device} onClose={onClose} PaperProps={{ sx: { width: 480, p: 3 } }}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
        <Typography variant="h6" fontWeight={600}>Device Details</Typography>
        <IconButton onClick={onClose}><CloseIcon /></IconButton>
      </Box>
      <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace', wordBreak: 'break-all' }}>
        {device.id}
      </Typography>
      <Divider sx={{ my: 2 }} />
      <Typography variant="subtitle2" fontWeight={600} gutterBottom>Properties</Typography>
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
        {Object.entries(props).map(([key, val]) => (
          <Box key={key} display="flex" justifyContent="space-between" alignItems="flex-start" gap={1}>
            <Typography variant="body2" color="text.secondary" sx={{ minWidth: 130, fontFamily: 'monospace', fontSize: 12 }}>
              {PROP_LABELS[key] ?? key}
            </Typography>
            <Typography variant="body2" sx={{ textAlign: 'right', wordBreak: 'break-all', fontSize: 12 }}>
              {val != null ? String(val) : ''}
            </Typography>
          </Box>
        ))}
      </Box>
      <Divider sx={{ my: 2 }} />
      <Typography variant="caption" color="text.secondary">
        First seen: {device.createdAt ? device.createdAt.toLocaleString() : '—'}
      </Typography>
      <br />
      <Typography variant="caption" color="text.secondary">
        Last seen: {device.updatedAt ? device.updatedAt.toLocaleString() : '—'}
      </Typography>
    </Drawer>
  )
}

export function DevicesTab({ projectId }: { projectId: string }) {
  const [devices, setDevices] = useState<DeviceResponse[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(25)
  const [stats, setStats] = useState<GetDeviceStatsResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [selected, setSelected] = useState<DeviceResponse | null>(null)
  const [searchInput, setSearchInput] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const notify = useNotification()

  const loadDevices = useCallback(async () => {
    setLoading(true)
    try {
      const res = await deviceService.list(projectId, {
        deviceId: searchQuery || undefined,
        limit: rowsPerPage,
        offset: page * rowsPerPage,
      })
      setDevices(res.results ?? [])
      setTotal(Number(res.total ?? 0))
    } catch {
      notify.error('Failed to load devices')
    } finally {
      setLoading(false)
    }
  }, [projectId, page, rowsPerPage, searchQuery])

  const loadStats = useCallback(async () => {
    try {
      const res = await deviceService.getStats(projectId)
      setStats(res)
    } catch {}
  }, [projectId])

  useEffect(() => { loadDevices() }, [loadDevices])
  useEffect(() => { loadStats() }, [loadStats])

  const getProp = (d: DeviceResponse, key: string): string => {
    const v = d.properties?.[key]
    return v != null ? String(v) : ''
  }

  return (
    <Box>
      <Box mb={2}>
        <TextField
          size="small"
          placeholder="Search by device ID (exact match)..."
          value={searchInput}
          onChange={e => setSearchInput(e.target.value)}
          onKeyDown={e => {
            if (e.key === 'Enter') { setSearchQuery(searchInput); setPage(0) }
          }}
          sx={{ width: 380 }}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon fontSize="small" />
              </InputAdornment>
            ),
          }}
        />
      </Box>

      {stats && (
        <Box display="grid" gridTemplateColumns="repeat(auto-fill, minmax(260px, 1fr))" gap={2} mb={3}>
          <StatChart title="Browsers" data={stats.browsers} />
          <StatChart title="Operating Systems" data={stats.osList} />
          <StatChart title="Device Types" data={stats.deviceTypes} />
          <StatChart title="Libraries" data={stats.libs} />
        </Box>
      )}

      <Paper>
        {loading ? (
          <Box display="flex" justifyContent="center" py={6}><CircularProgress /></Box>
        ) : (
          <>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Device ID</TableCell>
                  <TableCell>Browser</TableCell>
                  <TableCell>OS</TableCell>
                  <TableCell>Type</TableCell>
                  <TableCell>Screen</TableCell>
                  <TableCell>Library</TableCell>
                  <TableCell>First Seen</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {devices.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} align="center" sx={{ py: 4 }}>
                      <Typography color="text.secondary">No devices registered yet.</Typography>
                    </TableCell>
                  </TableRow>
                ) : devices.map(d => (
                  <TableRow key={d.id} hover sx={{ cursor: 'pointer' }} onClick={() => setSelected(d)}>
                    <TableCell sx={{ fontFamily: 'monospace', fontSize: 11, maxWidth: 160, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {d.id}
                    </TableCell>
                    <TableCell>
                      {getProp(d, '$browser') && (
                        <Chip label={getProp(d, '$browser')} size="small" variant="outlined" />
                      )}
                    </TableCell>
                    <TableCell>
                      {getProp(d, '$os') && (
                        <Chip label={getProp(d, '$os')} size="small" variant="outlined" color="secondary" />
                      )}
                    </TableCell>
                    <TableCell>
                      {getProp(d, '$device_type') && (
                        <Chip label={getProp(d, '$device_type')} size="small" />
                      )}
                    </TableCell>
                    <TableCell sx={{ fontSize: 12, color: 'text.secondary' }}>
                      {getProp(d, '$screen_width') && getProp(d, '$screen_height')
                        ? `${getProp(d, '$screen_width')}×${getProp(d, '$screen_height')}`
                        : '—'}
                    </TableCell>
                    <TableCell sx={{ fontSize: 12 }}>{getProp(d, '$lib') || '—'}</TableCell>
                    <TableCell sx={{ fontSize: 12, color: 'text.secondary' }}>
                      {d.createdAt ? d.createdAt.toLocaleDateString() : '—'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            <TablePagination
              component="div"
              count={total}
              page={page}
              onPageChange={(_, p) => setPage(p)}
              rowsPerPage={rowsPerPage}
              onRowsPerPageChange={e => { setRowsPerPage(Number(e.target.value)); setPage(0) }}
              rowsPerPageOptions={[10, 25, 50, 100]}
            />
          </>
        )}
      </Paper>

      <DeviceDrawer device={selected} onClose={() => setSelected(null)} />
    </Box>
  )
}
