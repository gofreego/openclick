import { useState, useCallback, useEffect } from 'react'
import {
  Typography, Box, Paper, Button, TextField, MenuItem, Select,
  FormControl, InputLabel, CircularProgress, Autocomplete
} from '@mui/material'
import { useTheme, alpha } from '@mui/material/styles'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import { analyticsService } from '../../services/analyticsService'
import { SaveToDashboardButton } from './SaveToDashboardButton'
import type { RetentionCohort } from '../../apis/proto/openclick/v1/analytics'
import { useNotification } from '@gofreego/tsutils'

export function RetentionTab({ projectId }: { projectId: string }) {
  const theme = useTheme()
  const [targetEvent, setTargetEvent] = useState('$pageview')
  const [returnEvent, setReturnEvent] = useState('$pageview')
  const [dateFrom, setDateFrom] = useState(() => {
    const d = new Date(); d.setDate(d.getDate() - 60); return d.toISOString().split('T')[0]
  })
  const [dateTo, setDateTo] = useState(() => new Date().toISOString().split('T')[0])
  const [period, setPeriod] = useState('Week')
  const [results, setResults] = useState<RetentionCohort[]>([])
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
          <Autocomplete
            freeSolo size="small" options={eventOptions}
            value={targetEvent}
            onInputChange={(_, val) => setTargetEvent(val)}
            sx={{ width: 200 }}
            renderInput={(params) => <TextField {...params} label="Target Event" />}
          />
          <Autocomplete
            freeSolo size="small" options={eventOptions}
            value={returnEvent}
            onInputChange={(_, val) => setReturnEvent(val)}
            sx={{ width: 200 }}
            renderInput={(params) => <TextField {...params} label="Return Event" />}
          />
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
          <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
            <Typography variant="subtitle2" fontWeight={600}>Retention Matrix</Typography>
            <SaveToDashboardButton projectId={projectId} type="retention"
              query={{ targetEvent: { id: targetEvent, name: targetEvent }, returnEvent: { id: returnEvent, name: returnEvent }, dateFrom, dateTo, period, retentionType: 'retention_first_time' }} />
          </Box>
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
                        backgroundColor: alpha(theme.palette.primary.main, opacity),
                        color: opacity > 0.5 ? theme.palette.primary.contrastText : theme.palette.text.primary,
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
