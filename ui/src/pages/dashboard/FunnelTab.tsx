import { useState, useCallback, useEffect } from 'react'
import {
  Typography, Box, Paper, Button, TextField, Chip, IconButton, CircularProgress, Autocomplete
} from '@mui/material'
import { useTheme } from '@mui/material/styles'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import {
  BarChart, Bar, Cell, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip,
  ResponsiveContainer, LabelList
} from 'recharts'
import { analyticsService } from '../../services/analyticsService'
import type { FunnelStepResult } from '../../apis/proto/openclick/v1/analytics'
import { useNotification } from '@gofreego/tsutils'
import { COLORS } from './tabInfo'
import { SaveToDashboardButton } from './SaveToDashboardButton'

export function FunnelTab({ projectId }: { projectId: string }) {
  const theme = useTheme()
  const [steps, setSteps] = useState([{ event: '$pageview', name: 'Page View' }, { event: '$click', name: 'Click' }])
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date(); d.setDate(d.getDate() - 30); return d.toISOString().split('T')[0]
  })
  const [dateTo, setDateTo] = useState(() => new Date().toISOString().split('T')[0])
  const [conversionWindow, setConversionWindow] = useState(14)
  const [results, setResults] = useState<FunnelStepResult[]>([])
  const [loading, setLoading] = useState(false)
  const [eventOptions, setEventOptions] = useState<string[]>([])
  const notify = useNotification()

  useEffect(() => {
    analyticsService.listEventNames(projectId)
      .then(setEventOptions)
      .catch(() => {})
  }, [projectId])

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
            <Autocomplete
              freeSolo size="small" options={eventOptions}
              value={step.event}
              onInputChange={(_, val) => updateStep(i, 'event', val)}
              sx={{ width: 220 }}
              renderInput={(params) => <TextField {...params} label="Event" />}
            />
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
          <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
            <Typography variant="subtitle2" fontWeight={600}>Funnel Results</Typography>
            <SaveToDashboardButton projectId={projectId} type="funnel"
              query={{ steps, dateFrom, dateTo, conversionWindowDays: conversionWindow }} />
          </Box>
          <ResponsiveContainer width="100%" height={380}>
            <BarChart data={chartData} margin={{ top: 40, right: 30, left: 20, bottom: 60 }}>
              <CartesianGrid strokeDasharray="3 3" stroke={theme.palette.divider} />
              <XAxis
                dataKey="name"
                tick={{ fontSize: 13, fontWeight: 600, fill: theme.palette.text.primary }}
                angle={-30}
                textAnchor="end"
                interval={0}
                axisLine={{ stroke: theme.palette.divider }}
                tickLine={{ stroke: theme.palette.divider }}
              />
              <YAxis
                tick={{ fontSize: 12, fill: theme.palette.text.secondary }}
                axisLine={{ stroke: theme.palette.divider }}
                tickLine={{ stroke: theme.palette.divider }}
              />
              <RechartsTooltip
                contentStyle={{
                  backgroundColor: theme.palette.background.paper,
                  border: `1px solid ${theme.palette.divider}`,
                  color: theme.palette.text.primary,
                }}
                formatter={(value: any, name: any) =>
                  name === 'count' ? [value, 'Users'] : [`${value}%`, 'Conversion Rate']
                }
              />
              <Bar dataKey="count" name="count" radius={[4, 4, 0, 0]}>
                {chartData.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                <LabelList
                  dataKey="conversionRate"
                  position="top"
                  formatter={(v: any) => `${v}%`}
                  style={{ fontSize: 14, fontWeight: 700, fill: theme.palette.text.primary }}
                />
              </Bar>
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
