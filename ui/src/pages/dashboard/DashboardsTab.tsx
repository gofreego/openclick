import { useState, useCallback, useEffect } from 'react'
import {
  Typography, Box, Paper, Button, TextField, Grid, Card, CardContent, CardActions,
  IconButton, Dialog, DialogTitle, DialogContent, DialogActions, Tooltip, Chip,
  Divider, Drawer, List, ListItem, ListItemText, ListItemSecondaryAction
} from '@mui/material'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import DashboardIcon from '@mui/icons-material/Dashboard'
import { dashboardService } from '../../services/dashboardService'
import type { Dashboard, DashboardDetail } from '../../services/dashboardService'
import { useNotification } from '@gofreego/tsutils'

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
