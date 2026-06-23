import { useState } from 'react'
import {
  Button, Dialog, DialogTitle, DialogContent, DialogActions,
  TextField, MenuItem, Select, FormControl, InputLabel, Typography, CircularProgress, Box
} from '@mui/material'
import BookmarkAddIcon from '@mui/icons-material/BookmarkAdd'
import { dashboardService } from '../../services/dashboardService'
import type { Dashboard } from '../../services/dashboardService'
import { useNotification } from '@gofreego/tsutils'

interface Props {
  projectId: string
  type: string
  query: Record<string, any>
  disabled?: boolean
}

export function SaveToDashboardButton({ projectId, type, query, disabled }: Props) {
  const [open, setOpen] = useState(false)
  const [dashboards, setDashboards] = useState<Dashboard[]>([])
  const [dashboardId, setDashboardId] = useState('')
  const [name, setName] = useState('')
  const [saving, setSaving] = useState(false)
  const notify = useNotification()

  const openDialog = async () => {
    try {
      const res = await dashboardService.list(projectId)
      const list = res.results || []
      setDashboards(list)
      setDashboardId(list[0]?.id || '')
    } catch {
      notify.error('Failed to load dashboards')
      return
    }
    setOpen(true)
  }

  const save = async () => {
    if (!dashboardId || !name) return
    setSaving(true)
    try {
      await dashboardService.createItem(projectId, dashboardId, { name, type, query })
      notify.success('Saved to dashboard')
      setOpen(false)
      setName('')
    } catch {
      notify.error('Failed to save to dashboard')
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <Button
        variant="outlined"
        size="small"
        startIcon={<BookmarkAddIcon />}
        onClick={openDialog}
        disabled={disabled}
      >
        Save to Dashboard
      </Button>

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>Save to Dashboard</DialogTitle>
        <DialogContent>
          <Box display="flex" flexDirection="column" gap={2} mt={1}>
            <TextField
              label="Widget Name"
              size="small"
              fullWidth
              value={name}
              onChange={e => setName(e.target.value)}
              autoFocus
              placeholder={`My ${type} chart`}
            />
            {dashboards.length > 0 ? (
              <FormControl size="small" fullWidth>
                <InputLabel>Dashboard</InputLabel>
                <Select value={dashboardId} label="Dashboard" onChange={e => setDashboardId(e.target.value)}>
                  {dashboards.map(d => (
                    <MenuItem key={d.id} value={d.id}>{d.name}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            ) : (
              <Typography variant="body2" color="text.secondary">
                No dashboards found. Create one first from the Dashboards tab.
              </Typography>
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button
            onClick={save}
            variant="contained"
            disabled={!name || !dashboardId || saving}
            startIcon={saving ? <CircularProgress size={14} /> : null}
          >
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </>
  )
}
