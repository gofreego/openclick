import { useState, useCallback, useEffect } from 'react'
import {
  Typography, Box, Paper, Button, TextField, Grid, Card, CardContent, CardActions,
  IconButton, Dialog, DialogTitle, DialogContent, DialogActions, Tooltip, Chip,
} from '@mui/material'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import CloseIcon from '@mui/icons-material/Close'
import DashboardIcon from '@mui/icons-material/Dashboard'
import { dashboardService } from '../../services/dashboardService'
import type { Dashboard, DashboardDetail, DashboardItem } from '../../services/dashboardService'
import { useNotification } from '@gofreego/tsutils'
import { DashboardItemWidget } from './DashboardItemWidget'

export function DashboardsTab({ projectId }: { projectId: string }) {
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
      notify.error('Failed to load dashboard')
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

  const handleDeleteItem = async (item: DashboardItem) => {
    if (!selectedDashboard) return
    try {
      await dashboardService.deleteItem(projectId, selectedDashboard.id, item.id)
      notify.success('Widget removed')
      const updated = await dashboardService.get(projectId, selectedDashboard.id)
      setSelectedDashboard(updated)
    } catch {
      notify.error('Failed to remove widget')
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
          <Grid key={d.id} size={{ xs: 12, sm: 6, md: 4 }}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" gap={1} mb={1}>
                  <DashboardIcon color="primary" />
                  <Typography variant="h6" fontWeight={600}>{d.name}</Typography>
                </Box>
                <Typography variant="body2" color="text.secondary">
                  {d.itemCount} widget{d.itemCount !== 1 ? 's' : ''}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Created: {d.createdAt ? new Date(d.createdAt as any).toLocaleDateString() : '—'}
                </Typography>
              </CardContent>
              <CardActions>
                <Button size="small" onClick={() => openDetail(d)}>Open Dashboard</Button>
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
            <Paper sx={{ p: 4, textAlign: 'center' }}>
              <Typography gutterBottom>No dashboards yet.</Typography>
              <Typography variant="body2" color="text.secondary">
                Create a dashboard, then use "Save to Dashboard" from Trends, Funnel, Retention, or Paths tabs to pin charts here.
              </Typography>
            </Paper>
          </Grid>
        )}
      </Grid>

      {/* Create dialog */}
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

      {/* Delete confirmation */}
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

      {/* Dashboard detail — full-width dialog with rendered widgets */}
      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)} maxWidth="xl" fullWidth>
        <DialogTitle>
          <Box display="flex" alignItems="center" justifyContent="space-between">
            <Box display="flex" alignItems="center" gap={1}>
              <DashboardIcon color="primary" />
              <Typography variant="h6" fontWeight={700}>{selectedDashboard?.name}</Typography>
              <Chip label={`${selectedDashboard?.items?.length || 0} widgets`} size="small" variant="outlined" />
            </Box>
            <IconButton onClick={() => setDetailOpen(false)} size="small"><CloseIcon /></IconButton>
          </Box>
        </DialogTitle>
        <DialogContent dividers>
          {selectedDashboard && (
            <>
              {(!selectedDashboard.items || selectedDashboard.items.length === 0) ? (
                <Paper sx={{ p: 4, textAlign: 'center' }}>
                  <Typography gutterBottom>No widgets yet.</Typography>
                  <Typography variant="body2" color="text.secondary">
                    Run a query in Trends, Funnel, Retention, or Paths and click "Save to Dashboard" to add widgets here.
                  </Typography>
                </Paper>
              ) : (
                <Grid container spacing={3}>
                  {selectedDashboard.items.map(item => (
                    <Grid key={item.id} size={{ xs: 12, md: 6, lg: 4 }}>
                      <Card variant="outlined" sx={{ height: '100%' }}>
                        <CardContent sx={{ pb: 0 }}>
                          <Box display="flex" alignItems="center" justifyContent="space-between" mb={1}>
                            <Box display="flex" alignItems="center" gap={1} minWidth={0}>
                              <Typography variant="subtitle2" fontWeight={600} noWrap>{item.name}</Typography>
                              <Chip label={item.type} size="small" color="primary" variant="outlined" sx={{ flexShrink: 0 }} />
                            </Box>
                            <Tooltip title="Remove widget">
                              <IconButton size="small" color="error" onClick={() => handleDeleteItem(item)} sx={{ flexShrink: 0 }}>
                                <DeleteIcon fontSize="small" />
                              </IconButton>
                            </Tooltip>
                          </Box>
                          <DashboardItemWidget projectId={projectId} item={item} />
                        </CardContent>
                      </Card>
                    </Grid>
                  ))}
                </Grid>
              )}
            </>
          )}
        </DialogContent>
      </Dialog>
    </Box>
  )
}
