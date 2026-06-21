import { useState, useCallback, useEffect } from 'react'
import {
  Typography, Box, Paper, Button, TextField, Tabs, Tab, Grid,
  Card, CardContent, CardActions, IconButton, Dialog, DialogTitle,
  DialogContent, DialogActions, Tooltip, Chip, MenuItem, Select,
  FormControl, InputLabel, CircularProgress, Divider, Drawer,
  List, ListItem, ListItemText, ListItemSecondaryAction
} from '@mui/material'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import DashboardIcon from '@mui/icons-material/Dashboard'
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip,
  Legend, ResponsiveContainer, BarChart, Bar, Cell
} from 'recharts'
import { analyticsService } from '../../services/analyticsService'
import { dashboardService } from '../../services/dashboardService'
import type { Dashboard, DashboardDetail } from '../../services/dashboardService'
import {
  TrendsSeries, FunnelStepResult, RetentionCohort, PathNode, PathLink
} from '../../apis/proto/openclick/v1/analytics'
import { useCurrentProject } from '../../contexts/ProjectContext'
import { useNotification } from '@gofreego/tsutils'
import { PageHeader } from '../../components/PageHeader'

const COLORS = ['#6366f1', '#f59e0b', '#10b981', '#ef4444', '#8b5cf6', '#06b6d4']

// ──────────────────────────────────────────────
// Trends Tab
// ──────────────────────────────────────────────
function TrendsTab({ projectId }: { projectId: string }) {
  const [eventName, setEventName] = useState('$pageview')
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date(); d.setDate(d.getDate() - 30); return d.toISOString().split('T')[0]
  })
  const [dateTo, setDateTo] = useState(() => new Date().toISOString().split('T')[0])
  const [interval, setInterval] = useState('day')
  const [results, setResults] = useState<TrendsSeries[]>([])
  const [loading, setLoading] = useState(false)
  const notify = useNotification()

  const run = useCallback(async () => {
    setLoading(true)
    try {
      const res = await analyticsService.queryTrends(projectId, {
        events: [{ id: eventName, name: eventName, math: 'total' }],
        dateFrom,
        dateTo,
        interval,
      })
      setResults(res.results || [])
    } catch {
      notify.error('Failed to query trends')
    } finally {
      setLoading(false)
    }
  }, [projectId, eventName, dateFrom, dateTo, interval])

  const chartData = results.length > 0
    ? results[0].days.map((day, i) => ({
        day,
        ...Object.fromEntries(results.map(s => [s.label || s.breakdownValue || 'Count', Number(s.data[i] || 0)]))
      }))
    : []

  return (
    <Box>
      <Paper sx={{ p: 2, mb: 3 }}>
        <Typography variant="subtitle2" fontWeight={600} gutterBottom>Configure Query</Typography>
        <Box display="flex" gap={2} flexWrap="wrap" alignItems="flex-end">
          <TextField size="small" label="Event Name" value={eventName}
            onChange={e => setEventName(e.target.value)} sx={{ width: 200 }} />
          <TextField size="small" type="date" label="From" value={dateFrom}
            onChange={e => setDateFrom(e.target.value)} InputLabelProps={{ shrink: true }} sx={{ width: 160 }} />
          <TextField size="small" type="date" label="To" value={dateTo}
            onChange={e => setDateTo(e.target.value)} InputLabelProps={{ shrink: true }} sx={{ width: 160 }} />
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <InputLabel>Interval</InputLabel>
            <Select value={interval} label="Interval" onChange={e => setInterval(e.target.value)}>
              <MenuItem value="hour">Hour</MenuItem>
              <MenuItem value="day">Day</MenuItem>
              <MenuItem value="week">Week</MenuItem>
              <MenuItem value="month">Month</MenuItem>
            </Select>
          </FormControl>
          <Button variant="contained" startIcon={<PlayArrowIcon />} onClick={run} disabled={loading}>
            {loading ? 'Running...' : 'Run Query'}
          </Button>
        </Box>
      </Paper>

      {loading && <Box display="flex" justifyContent="center" py={4}><CircularProgress /></Box>}

      {!loading && results.length > 0 && (
        <Paper sx={{ p: 2 }}>
          <Typography variant="subtitle2" fontWeight={600} gutterBottom>Results</Typography>
          <ResponsiveContainer width="100%" height={350}>
            <LineChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="day" tick={{ fontSize: 12 }} />
              <YAxis tick={{ fontSize: 12 }} />
              <RechartsTooltip />
              <Legend />
              {results.map((s, i) => (
                <Line key={i} type="monotone" dataKey={s.label || s.breakdownValue || 'Count'}
                  stroke={COLORS[i % COLORS.length]} strokeWidth={2} dot={false} />
              ))}
            </LineChart>
          </ResponsiveContainer>
        </Paper>
      )}

      {!loading && results.length === 0 && (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography color="text.secondary">Run a query to see trend data.</Typography>
        </Paper>
      )}
    </Box>
  )
}

// ──────────────────────────────────────────────
// Funnel Tab
// ──────────────────────────────────────────────
function FunnelTab({ projectId }: { projectId: string }) {
  const [steps, setSteps] = useState([{ event: '$pageview', name: 'Page View' }, { event: '$click', name: 'Click' }])
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date(); d.setDate(d.getDate() - 30); return d.toISOString().split('T')[0]
  })
  const [dateTo, setDateTo] = useState(() => new Date().toISOString().split('T')[0])
  const [conversionWindow, setConversionWindow] = useState(14)
  const [results, setResults] = useState<FunnelStepResult[]>([])
  const [loading, setLoading] = useState(false)
  const notify = useNotification()

  const run = useCallback(async () => {
    setLoading(true)
    try {
      const res = await analyticsService.queryFunnel(projectId, { steps, dateFrom, dateTo, conversionWindowDays: conversionWindow })
      setResults(res.result || [])
    } catch {
      notify.error('Failed to query funnel')
    } finally {
      setLoading(false)
    }
  }, [projectId, steps, dateFrom, dateTo, conversionWindow])

  const addStep = () => setSteps([...steps, { event: '', name: '' }])
  const removeStep = (i: number) => setSteps(steps.filter((_, idx) => idx !== i))
  const updateStep = (i: number, field: 'event' | 'name', value: string) => {
    const updated = [...steps]; updated[i] = { ...updated[i], [field]: value }; setSteps(updated)
  }

  const chartData = results.map(r => ({
    name: r.name || r.actionId,
    count: Number(r.count),
    conversionRate: Math.round(r.conversionRate * 100) / 100,
  }))

  return (
    <Box>
      <Paper sx={{ p: 2, mb: 3 }}>
        <Typography variant="subtitle2" fontWeight={600} gutterBottom>Funnel Steps</Typography>
        {steps.map((step, i) => (
          <Box key={i} display="flex" gap={1} mb={1} alignItems="center">
            <Chip label={`Step ${i + 1}`} size="small" color="primary" sx={{ minWidth: 60 }} />
            <TextField size="small" label="Event" value={step.event}
              onChange={e => updateStep(i, 'event', e.target.value)} sx={{ width: 200 }} />
            <TextField size="small" label="Label" value={step.name}
              onChange={e => updateStep(i, 'name', e.target.value)} sx={{ width: 200 }} />
            {steps.length > 2 && (
              <IconButton size="small" color="error" onClick={() => removeStep(i)}><DeleteIcon fontSize="small" /></IconButton>
            )}
          </Box>
        ))}
        <Button size="small" onClick={addStep} startIcon={<AddIcon />} sx={{ mb: 2 }}>Add Step</Button>

        <Box display="flex" gap={2} flexWrap="wrap" alignItems="flex-end">
          <TextField size="small" type="date" label="From" value={dateFrom}
            onChange={e => setDateFrom(e.target.value)} InputLabelProps={{ shrink: true }} sx={{ width: 160 }} />
          <TextField size="small" type="date" label="To" value={dateTo}
            onChange={e => setDateTo(e.target.value)} InputLabelProps={{ shrink: true }} sx={{ width: 160 }} />
          <TextField size="small" type="number" label="Conversion Window (days)" value={conversionWindow}
            onChange={e => setConversionWindow(Number(e.target.value))} sx={{ width: 200 }} />
          <Button variant="contained" startIcon={<PlayArrowIcon />} onClick={run} disabled={loading}>
            {loading ? 'Running...' : 'Run Funnel'}
          </Button>
        </Box>
      </Paper>

      {loading && <Box display="flex" justifyContent="center" py={4}><CircularProgress /></Box>}

      {!loading && results.length > 0 && (
        <Paper sx={{ p: 2 }}>
          <Typography variant="subtitle2" fontWeight={600} gutterBottom>Funnel Results</Typography>
          <ResponsiveContainer width="100%" height={350}>
            <BarChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" tick={{ fontSize: 12 }} />
              <YAxis yAxisId="left" tick={{ fontSize: 12 }} />
              <YAxis yAxisId="right" orientation="right" tickFormatter={(v: any) => `${v}%`} tick={{ fontSize: 12 }} />
              <RechartsTooltip formatter={(value: any, name: any) => name === 'conversionRate' ? `${value}%` : value} />
              <Legend />
              <Bar yAxisId="left" dataKey="count" name="Count" radius={[4, 4, 0, 0]}>
                {chartData.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
              </Bar>
              <Bar yAxisId="right" dataKey="conversionRate" name="Conversion Rate (%)" fill="#10b981" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </Paper>
      )}

      {!loading && results.length === 0 && (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography color="text.secondary">Configure steps and run the funnel query.</Typography>
        </Paper>
      )}
    </Box>
  )
}

// ──────────────────────────────────────────────
// Retention Tab
// ──────────────────────────────────────────────
function RetentionTab({ projectId }: { projectId: string }) {
  const [targetEvent, setTargetEvent] = useState('$pageview')
  const [returnEvent, setReturnEvent] = useState('$pageview')
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date(); d.setDate(d.getDate() - 60); return d.toISOString().split('T')[0]
  })
  const [dateTo, setDateTo] = useState(() => new Date().toISOString().split('T')[0])
  const [period, setPeriod] = useState('Week')
  const [results, setResults] = useState<RetentionCohort[]>([])
  const [loading, setLoading] = useState(false)
  const notify = useNotification()

  const run = useCallback(async () => {
    setLoading(true)
    try {
      const res = await analyticsService.queryRetention(projectId, {
        targetEvent: { id: targetEvent, name: targetEvent },
        returnEvent: { id: returnEvent, name: returnEvent },
        dateFrom, dateTo, period,
        retentionType: 'retention_first_time',
      })
      setResults(res.result || [])
    } catch {
      notify.error('Failed to query retention')
    } finally {
      setLoading(false)
    }
  }, [projectId, targetEvent, returnEvent, dateFrom, dateTo, period])

  const maxPeriods = results.length > 0 ? Math.max(...results.map(r => r.values.length)) : 0

  return (
    <Box>
      <Paper sx={{ p: 2, mb: 3 }}>
        <Box display="flex" gap={2} flexWrap="wrap" alignItems="flex-end">
          <TextField size="small" label="Target Event" value={targetEvent}
            onChange={e => setTargetEvent(e.target.value)} sx={{ width: 180 }} />
          <TextField size="small" label="Return Event" value={returnEvent}
            onChange={e => setReturnEvent(e.target.value)} sx={{ width: 180 }} />
          <TextField size="small" type="date" label="From" value={dateFrom}
            onChange={e => setDateFrom(e.target.value)} InputLabelProps={{ shrink: true }} sx={{ width: 160 }} />
          <TextField size="small" type="date" label="To" value={dateTo}
            onChange={e => setDateTo(e.target.value)} InputLabelProps={{ shrink: true }} sx={{ width: 160 }} />
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <InputLabel>Period</InputLabel>
            <Select value={period} label="Period" onChange={e => setPeriod(e.target.value)}>
              <MenuItem value="Day">Day</MenuItem>
              <MenuItem value="Week">Week</MenuItem>
              <MenuItem value="Month">Month</MenuItem>
            </Select>
          </FormControl>
          <Button variant="contained" startIcon={<PlayArrowIcon />} onClick={run} disabled={loading}>
            {loading ? 'Running...' : 'Run Retention'}
          </Button>
        </Box>
      </Paper>

      {loading && <Box display="flex" justifyContent="center" py={4}><CircularProgress /></Box>}

      {!loading && results.length > 0 && (
        <Paper sx={{ p: 2, overflow: 'auto' }}>
          <Typography variant="subtitle2" fontWeight={600} gutterBottom>Retention Matrix</Typography>
          <Box component="table" sx={{ borderCollapse: 'collapse', fontSize: '0.8rem', minWidth: '100%' }}>
            <Box component="thead">
              <Box component="tr">
                <Box component="th" sx={{ p: 1, textAlign: 'left', borderBottom: '1px solid', borderColor: 'divider', whiteSpace: 'nowrap' }}>Cohort</Box>
                <Box component="th" sx={{ p: 1, textAlign: 'center', borderBottom: '1px solid', borderColor: 'divider' }}>Size</Box>
                {Array.from({ length: maxPeriods }, (_, i) => (
                  <Box key={i} component="th" sx={{ p: 1, textAlign: 'center', borderBottom: '1px solid', borderColor: 'divider', minWidth: 60 }}>
                    {period[0]}{i}
                  </Box>
                ))}
              </Box>
            </Box>
            <Box component="tbody">
              {results.map((cohort, ri) => (
                <Box key={ri} component="tr">
                  <Box component="td" sx={{ p: 1, borderBottom: '1px solid', borderColor: 'divider', whiteSpace: 'nowrap' }}>
                    {cohort.label || cohort.date}
                  </Box>
                  <Box component="td" sx={{ p: 1, textAlign: 'center', borderBottom: '1px solid', borderColor: 'divider' }}>
                    {cohort.cohortSize}
                  </Box>
                  {cohort.values.map((v, vi) => {
                    const pct = Math.round(v.percentage * 100) / 100
                    const opacity = 0.15 + (pct / 100) * 0.85
                    return (
                      <Box key={vi} component="td" sx={{
                        p: 1, textAlign: 'center', borderBottom: '1px solid', borderColor: 'divider',
                        backgroundColor: `rgba(99, 102, 241, ${opacity})`,
                        color: opacity > 0.5 ? 'white' : 'inherit',
                        fontWeight: 600,
                      }}>
                        {pct}%
                      </Box>
                    )
                  })}
                  {Array.from({ length: maxPeriods - cohort.values.length }, (_, i) => (
                    <Box key={`empty-${i}`} component="td" sx={{ p: 1, borderBottom: '1px solid', borderColor: 'divider', bgcolor: 'action.hover' }} />
                  ))}
                </Box>
              ))}
            </Box>
          </Box>
        </Paper>
      )}

      {!loading && results.length === 0 && (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography color="text.secondary">Configure and run the retention query to see the cohort matrix.</Typography>
        </Paper>
      )}
    </Box>
  )
}

// ──────────────────────────────────────────────
// Paths Tab
// ──────────────────────────────────────────────
function PathsTab({ projectId }: { projectId: string }) {
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date(); d.setDate(d.getDate() - 30); return d.toISOString().split('T')[0]
  })
  const [dateTo, setDateTo] = useState(() => new Date().toISOString().split('T')[0])
  const [startPoint, setStartPoint] = useState('')
  const [endPoint, setEndPoint] = useState('')
  const [stepLimit, setStepLimit] = useState(5)
  const [nodes, setNodes] = useState<PathNode[]>([])
  const [links, setLinks] = useState<PathLink[]>([])
  const [loading, setLoading] = useState(false)
  const notify = useNotification()

  const run = useCallback(async () => {
    setLoading(true)
    try {
      const res = await analyticsService.queryPaths(projectId, {
        dateFrom, dateTo,
        startPoint: startPoint || undefined,
        endPoint: endPoint || undefined,
        stepLimit,
        pathType: 'url',
        minEdgeWeight: 1,
      } as any)
      setNodes(res.nodes || [])
      setLinks(res.links || [])
    } catch {
      notify.error('Failed to query paths')
    } finally {
      setLoading(false)
    }
  }, [projectId, dateFrom, dateTo, startPoint, endPoint, stepLimit])

  // Simple top-links table since full sankey requires a dedicated library
  const sortedLinks = [...links].sort((a, b) => Number(b.value) - Number(a.value)).slice(0, 20)

  return (
    <Box>
      <Paper sx={{ p: 2, mb: 3 }}>
        <Box display="flex" gap={2} flexWrap="wrap" alignItems="flex-end">
          <TextField size="small" type="date" label="From" value={dateFrom}
            onChange={e => setDateFrom(e.target.value)} InputLabelProps={{ shrink: true }} sx={{ width: 160 }} />
          <TextField size="small" type="date" label="To" value={dateTo}
            onChange={e => setDateTo(e.target.value)} InputLabelProps={{ shrink: true }} sx={{ width: 160 }} />
          <TextField size="small" label="Start Point (optional)" value={startPoint}
            onChange={e => setStartPoint(e.target.value)} sx={{ width: 220 }} />
          <TextField size="small" label="End Point (optional)" value={endPoint}
            onChange={e => setEndPoint(e.target.value)} sx={{ width: 220 }} />
          <TextField size="small" type="number" label="Step Limit" value={stepLimit}
            onChange={e => setStepLimit(Number(e.target.value))} sx={{ width: 120 }} />
          <Button variant="contained" startIcon={<PlayArrowIcon />} onClick={run} disabled={loading}>
            {loading ? 'Running...' : 'Run Paths'}
          </Button>
        </Box>
      </Paper>

      {loading && <Box display="flex" justifyContent="center" py={4}><CircularProgress /></Box>}

      {!loading && (nodes.length > 0 || links.length > 0) && (
        <Grid container spacing={2}>
          <Grid size={{ xs: 12, md: 4 }}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="subtitle2" fontWeight={600} gutterBottom>Nodes ({nodes.length})</Typography>
              <Box sx={{ maxHeight: 400, overflow: 'auto' }}>
                {nodes.map((n, i) => (
                  <Chip key={i} label={n.name || n.id} size="small" sx={{ m: 0.25 }} variant="outlined" />
                ))}
              </Box>
            </Paper>
          </Grid>
          <Grid size={{ xs: 12, md: 8 }}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="subtitle2" fontWeight={600} gutterBottom>Top Path Flows</Typography>
              <Box component="table" sx={{ borderCollapse: 'collapse', width: '100%', fontSize: '0.8rem' }}>
                <Box component="thead">
                  <Box component="tr">
                    <Box component="th" sx={{ p: 1, textAlign: 'left', borderBottom: '1px solid', borderColor: 'divider' }}>Source</Box>
                    <Box component="th" sx={{ p: 1, textAlign: 'left', borderBottom: '1px solid', borderColor: 'divider' }}>Target</Box>
                    <Box component="th" sx={{ p: 1, textAlign: 'right', borderBottom: '1px solid', borderColor: 'divider' }}>Count</Box>
                  </Box>
                </Box>
                <Box component="tbody">
                  {sortedLinks.map((l, i) => (
                    <Box key={i} component="tr" sx={{ '&:hover': { bgcolor: 'action.hover' } }}>
                      <Box component="td" sx={{ p: 1, borderBottom: '1px solid', borderColor: 'divider', maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{l.source}</Box>
                      <Box component="td" sx={{ p: 1, borderBottom: '1px solid', borderColor: 'divider', maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{l.target}</Box>
                      <Box component="td" sx={{ p: 1, textAlign: 'right', borderBottom: '1px solid', borderColor: 'divider', fontWeight: 600 }}>{l.value}</Box>
                    </Box>
                  ))}
                </Box>
              </Box>
            </Paper>
          </Grid>
        </Grid>
      )}

      {!loading && nodes.length === 0 && links.length === 0 && (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography color="text.secondary">Configure and run the paths query to see user journey data.</Typography>
        </Paper>
      )}
    </Box>
  )
}

// ──────────────────────────────────────────────
// Dashboards Tab
// ──────────────────────────────────────────────
function DashboardsTab({ projectId }: { projectId: string }) {
  const [dashboards, setDashboards] = useState<Dashboard[]>([])
  const [loaded, setLoaded] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deletingDashboard, setDeletingDashboard] = useState<Dashboard | null>(null)
  const [newName, setNewName] = useState('')
  const [selectedDashboard, setSelectedDashboard] = useState<DashboardDetail | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const notify = useNotification()

  const load = useCallback(async () => {
    try {
      const res = await dashboardService.list(projectId)
      setDashboards(res.results || [])
      setLoaded(true)
    } catch {
      notify.error('Failed to load dashboards')
    }
  }, [projectId])

  // Lazy load on mount
  useEffect(() => { load() }, [load])

  const handleCreate = async () => {
    try {
      await dashboardService.create(projectId, newName)
      notify.success('Dashboard created')
      setCreateOpen(false)
      setNewName('')
      load()
    } catch (err: any) {
      notify.error(err.message || 'Failed to create dashboard')
    }
  }

  const openDetail = async (d: Dashboard) => {
    try {
      const detail = await dashboardService.get(projectId, d.id)
      setSelectedDashboard(detail)
      setDetailOpen(true)
    } catch {
      notify.error('Failed to load dashboard details')
    }
  }

  const confirmDelete = (d: Dashboard) => { setDeletingDashboard(d); setDeleteOpen(true) }
  const handleDelete = async () => {
    if (!deletingDashboard) return
    try {
      await dashboardService.delete(projectId, deletingDashboard.id)
      notify.success('Dashboard deleted')
      setDeleteOpen(false)
      setDeletingDashboard(null)
      load()
    } catch {
      notify.error('Failed to delete dashboard')
    }
  }

  const handleDeleteItem = async (itemId: string) => {
    if (!selectedDashboard) return
    try {
      await dashboardService.deleteItem(projectId, selectedDashboard.id, itemId)
      notify.success('Item removed')
      const updated = await dashboardService.get(projectId, selectedDashboard.id)
      setSelectedDashboard(updated)
    } catch {
      notify.error('Failed to delete item')
    }
  }

  return (
    <Box>
      <Box display="flex" justifyContent="flex-end" mb={2}>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setCreateOpen(true)}>
          Create Dashboard
        </Button>
      </Box>

      <Grid container spacing={3}>
        {dashboards.map(d => (
          <Grid key={d.id} size={{ xs: 12, md: 4 }}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" gap={1} mb={1}>
                  <DashboardIcon color="primary" />
                  <Typography variant="h6" fontWeight={600}>{d.name}</Typography>
                </Box>
                <Typography variant="body2" color="text.secondary">
                  {d.itemCount} item{d.itemCount !== 1 ? 's' : ''}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Created: {d.createdAt ? new Date(d.createdAt as any).toLocaleDateString() : '—'}
                </Typography>
              </CardContent>
              <CardActions>
                <Button size="small" onClick={() => openDetail(d)}>View Items</Button>
                <Tooltip title="Delete dashboard">
                  <IconButton size="small" color="error" onClick={() => confirmDelete(d)} sx={{ ml: 'auto' }}>
                    <DeleteIcon />
                  </IconButton>
                </Tooltip>
              </CardActions>
            </Card>
          </Grid>
        ))}
        {loaded && dashboards.length === 0 && (
          <Grid size={{ xs: 12 }}>
            <Paper sx={{ p: 3, textAlign: 'center' }}>
              <Typography>No dashboards yet. Create one to organize your analytics charts.</Typography>
            </Paper>
          </Grid>
        )}
      </Grid>

      {/* Create Dialog */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Create Dashboard</DialogTitle>
        <DialogContent>
          <TextField autoFocus margin="dense" label="Dashboard Name" fullWidth variant="outlined"
            value={newName} onChange={e => setNewName(e.target.value)} sx={{ mt: 1 }} />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOpen(false)}>Cancel</Button>
          <Button onClick={handleCreate} variant="contained" disabled={!newName}>Create</Button>
        </DialogActions>
      </Dialog>

      {/* Delete Confirmation */}
      <Dialog open={deleteOpen} onClose={() => setDeleteOpen(false)}>
        <DialogTitle>Delete Dashboard</DialogTitle>
        <DialogContent>
          <Typography>Delete <strong>{deletingDashboard?.name}</strong>? This cannot be undone.</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteOpen(false)}>Cancel</Button>
          <Button onClick={handleDelete} variant="contained" color="error">Delete</Button>
        </DialogActions>
      </Dialog>

      {/* Dashboard Detail Drawer */}
      <Drawer anchor="right" open={detailOpen} onClose={() => setDetailOpen(false)}>
        <Box sx={{ width: 460, p: 3 }}>
          {selectedDashboard && (
            <>
              <Typography variant="h6" fontWeight={700}>{selectedDashboard.name}</Typography>
              <Typography variant="caption" color="text.secondary">
                {selectedDashboard.createdAt ? new Date(selectedDashboard.createdAt as any).toLocaleDateString() : ''}
              </Typography>
              <Divider sx={{ my: 2 }} />
              <Typography variant="subtitle2" fontWeight={600} gutterBottom>
                Items ({selectedDashboard.items?.length || 0})
              </Typography>
              <List dense>
                {(selectedDashboard.items || []).map(item => (
                  <ListItem key={item.id} divider>
                    <ListItemText
                      primary={item.name}
                      secondary={<Chip label={item.type} size="small" variant="outlined" />}
                    />
                    <ListItemSecondaryAction>
                      <Tooltip title="Remove item">
                        <IconButton size="small" color="error" onClick={() => handleDeleteItem(item.id)}>
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </ListItemSecondaryAction>
                  </ListItem>
                ))}
                {(!selectedDashboard.items || selectedDashboard.items.length === 0) && (
                  <Typography variant="body2" color="text.secondary" sx={{ py: 1 }}>
                    No items in this dashboard.
                  </Typography>
                )}
              </List>
            </>
          )}
        </Box>
      </Drawer>
    </Box>
  )
}

// ──────────────────────────────────────────────
// Main DashboardPage
// ──────────────────────────────────────────────
export function DashboardPage() {
  const selectedProjectId = useCurrentProject()
  const [activeTab, setActiveTab] = useState(0)

  return (
    <Box sx={{ p: 3 }}>
      <PageHeader
        title="Analytics Dashboard"
        infoTitle="About Analytics"
        infoDescription="Run Trends, Funnel, Retention, and Paths queries against your event data. Create and manage dashboards to organize your key metrics."
      />

      {selectedProjectId ? (
        <>
          <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)} sx={{ mb: 3 }}>
            <Tab label="Trends" />
            <Tab label="Funnel" />
            <Tab label="Retention" />
            <Tab label="Paths" />
            <Tab label="Dashboards" />
          </Tabs>
          {activeTab === 0 && <TrendsTab projectId={selectedProjectId} />}
          {activeTab === 1 && <FunnelTab projectId={selectedProjectId} />}
          {activeTab === 2 && <RetentionTab projectId={selectedProjectId} />}
          {activeTab === 3 && <PathsTab projectId={selectedProjectId} />}
          {activeTab === 4 && <DashboardsTab projectId={selectedProjectId} />}
        </>
      ) : (
        <Paper sx={{ p: 3, textAlign: 'center', mt: 2 }}>
          <Typography variant="body1">
            Please select a project to view analytics.
          </Typography>
        </Paper>
      )}
    </Box>
  )
}
