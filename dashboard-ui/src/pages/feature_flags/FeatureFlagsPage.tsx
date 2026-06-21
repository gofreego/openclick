import { useEffect, useState } from 'react'
import { Typography, Box, Paper, Button, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Switch, Dialog, DialogTitle, DialogContent, DialogActions, TextField, IconButton } from '@mui/material'
import DeleteIcon from '@mui/icons-material/Delete'
import EditIcon from '@mui/icons-material/Edit'
import { featureFlagService } from '../../services/featureFlagService'
import type { FeatureFlag } from '../../services/featureFlagService'
import { useCurrentProject } from '../../contexts/ProjectContext'
import { useNotification } from '@gofreego/tsutils'
import { PageHeader } from '../../components/PageHeader'

export function FeatureFlagsPage() {
  const selectedProjectId = useCurrentProject()
  const [flags, setFlags] = useState<FeatureFlag[]>([])
  const [open, setOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [flagToDelete, setFlagToDelete] = useState<FeatureFlag | null>(null)
  const notify = useNotification()

  // form state
  const [editingFlagId, setEditingFlagId] = useState<string | null>(null)
  const [key, setKey] = useState('')
  const [name, setName] = useState('')
  const [rolloutPct, setRolloutPct] = useState(100)

  const handleOpenCreate = () => {
    setEditingFlagId(null)
    setKey('')
    setName('')
    setRolloutPct(100)
    setOpen(true)
  }

  const handleOpenEdit = (flag: FeatureFlag) => {
    setEditingFlagId(flag.id)
    setKey(flag.key)
    setName(flag.name)
    setRolloutPct(flag.rolloutPct)
    setOpen(true)
  }

  const handleClose = () => {
    setOpen(false)
    setEditingFlagId(null)
    setKey('')
    setName('')
    setRolloutPct(100)
  }

  const loadFlags = async (projectId: string) => {
    if (!projectId) return
    try {
      const res = await featureFlagService.list(projectId)
      setFlags(res.results || [])
    } catch (err: any) {
      notify.error('Failed to load feature flags')
    }
  }



  useEffect(() => {
    if (selectedProjectId) {
      loadFlags(selectedProjectId)
    }
  }, [selectedProjectId])

  const handleSave = async () => {
    if (!selectedProjectId) return
    try {
      if (editingFlagId) {
        await featureFlagService.update(selectedProjectId, editingFlagId, {
          key,
          name,
          rolloutPct
        })
        notify.success('Feature flag updated')
      } else {
        await featureFlagService.create(selectedProjectId, {
          key,
          name,
          active: true,
          rolloutPct
        })
        notify.success('Feature flag created')
      }
      handleClose()
      loadFlags(selectedProjectId)
    } catch (err: any) {
      notify.error(err.message || 'Failed to save feature flag')
    }
  }

  const toggleFlag = async (flag: FeatureFlag) => {
    try {
      await featureFlagService.update(selectedProjectId, flag.id, { active: !flag.active })
      loadFlags(selectedProjectId)
    } catch (err: any) {
      notify.error('Failed to update flag')
    }
  }

  const confirmDelete = (flag: FeatureFlag) => {
    setFlagToDelete(flag)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!selectedProjectId || !flagToDelete) return
    try {
      await featureFlagService.delete(selectedProjectId, flagToDelete.id)
      notify.success('Feature flag deleted')
      loadFlags(selectedProjectId)
      setDeleteDialogOpen(false)
      setFlagToDelete(null)
    } catch (err: any) {
      notify.error('Failed to delete flag')
    }
  }

  return (
    <Box sx={{ p: 3 }}>
      <PageHeader 
        title="Feature Flags" 
        infoTitle="About Feature Flags"
        infoDescription="Feature Flags allow you to safely deploy new features to your application without releasing them to all users immediately. You can control the rollout percentage, toggle features on and off instantly without a code deployment, and use them for A/B testing."
        action={
          <Button variant="contained" color="primary" onClick={handleOpenCreate} disabled={!selectedProjectId}>Create Flag</Button>
        }
      />

      {selectedProjectId ? (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Key</TableCell>
                <TableCell>Rollout %</TableCell>
                <TableCell>Active</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {flags.map((flag) => (
                <TableRow key={flag.id}>
                  <TableCell>{flag.name}</TableCell>
                  <TableCell><code>{flag.key}</code></TableCell>
                  <TableCell>{flag.rolloutPct}%</TableCell>
                  <TableCell>
                    <Switch
                      checked={flag.active}
                      onChange={() => toggleFlag(flag)}
                      color="primary"
                    />
                  </TableCell>
                  <TableCell align="right">
                    <IconButton size="small" color="primary" onClick={() => handleOpenEdit(flag)} title="Edit Flag">
                      <EditIcon />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => confirmDelete(flag)} title="Delete Flag">
                      <DeleteIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
              {flags.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} align="center">No feature flags found</TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      ) : (
        <Paper sx={{ p: 3, textAlign: 'center' }}>
          <Typography>Please create or select a project first.</Typography>
        </Paper>
      )}

      <Dialog open={open} onClose={handleClose}>
        <DialogTitle>{editingFlagId ? 'Edit Feature Flag' : 'Create Feature Flag'}</DialogTitle>
        <DialogContent>
          <TextField
            margin="dense"
            label="Name"
            fullWidth
            variant="outlined"
            value={name}
            onChange={(e) => setName(e.target.value)}
            sx={{ mb: 2, mt: 1 }}
          />
          <TextField
            margin="dense"
            label="Key (e.g. beta-feature)"
            fullWidth
            variant="outlined"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            disabled={!!editingFlagId}
            helperText={editingFlagId ? "The feature flag key cannot be changed after creation." : ""}
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Rollout Percentage (0-100)"
            type="number"
            fullWidth
            variant="outlined"
            value={rolloutPct}
            onChange={(e) => setRolloutPct(Number(e.target.value))}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>Cancel</Button>
          <Button onClick={handleSave} variant="contained" disabled={!name || !key}>Save</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
        <DialogTitle>Delete Feature Flag</DialogTitle>
        <DialogContent>
          <Typography>Are you sure you want to delete the feature flag "{flagToDelete?.name}"? This action cannot be undone.</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleDelete} variant="contained" color="error">Delete</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}
