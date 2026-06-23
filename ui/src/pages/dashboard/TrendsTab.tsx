import { useState, useCallback, useEffect } from 'react'
import {
  Typography, Box, Paper, Button, TextField, MenuItem, Select,
  FormControl, InputLabel, CircularProgress, Autocomplete
} from '@mui/material'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip,
  Legend, ResponsiveContainer
} from 'recharts'
import { analyticsService } from '../../services/analyticsService'
import type { TrendsSeries } from '../../apis/proto/openclick/v1/analytics'
import { useNotification } from '@gofreego/tsutils'
import { COLORS } from './tabInfo'
import { SaveToDashboardButton } from './SaveToDashboardButton'

export function TrendsTab({ projectId }: { projectId: string }) {
  const [eventName, setEventName] = useState('$pageview')
  const [eventOptions, setEventOptions] = useState<string[]>([])
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date(); d.setDate(d.getDate() - 30); return d.toISOString().split('T')[0]
  })
  const [dateTo, setDateTo] = useState(() => new Date().toISOString().split('T')[0])
  const [interval, setInterval] = useState('day')
  const [results, setResults] = useState<TrendsSeries[]>([])
  const [loading, setLoading] = useState(false)
  const notify = useNotification()

  useEffect(() => {
    analyticsService.listEventNames(projectId)
      .then(setEventOptions)
      .catch(() => {})
  }, [projectId])

  const run = useCallback(async () => {
    setLoading(true)
    try {
      const res = await analyticsService.queryTrends(projectId, {
        events: [{ id: eventName, name: eventName, math: 'total' }],
        dateFrom, dateTo, interval,
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
          <Autocomplete
            freeSolo
            size="small"
            options={eventOptions}
            value={eventName}
            onInputChange={(_, val) => setEventName(val)}
            sx={{ width: 220 }}
            renderInput={(params) => <TextField {...params} label="Event Name" />}
          />
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
          <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
            <Typography variant="subtitle2" fontWeight={600}>Results</Typography>
            <SaveToDashboardButton projectId={projectId} type="trends"
              query={{ events: [{ id: eventName, name: eventName, math: 'total' }], dateFrom, dateTo, interval }} />
          </Box>
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
