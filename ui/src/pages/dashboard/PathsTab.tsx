import { useState, useCallback } from 'react'
import {
  Typography, Box, Paper, Button, TextField, Chip, CircularProgress, Grid
} from '@mui/material'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import { analyticsService } from '../../services/analyticsService'
import type { PathNode, PathLink } from '../../apis/proto/openclick/v1/analytics'
import { useNotification } from '@gofreego/tsutils'

export function PathsTab({ projectId }: { projectId: string }) {
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
