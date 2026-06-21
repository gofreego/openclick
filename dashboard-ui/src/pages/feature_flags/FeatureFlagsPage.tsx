import { useEffect, useState } from 'react'
import { Typography, Box, Paper, Button, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Switch, Dialog, DialogTitle, DialogContent, DialogActions, TextField, Select, MenuItem, FormControl, InputLabel } from '@mui/material'
import { featureFlagService } from '../../services/featureFlagService'
import type { FeatureFlag } from '../../services/featureFlagService'
import { projectService } from '../../services/projectService'
import type { Project } from '../../services/projectService'
import { useNotification } from '@gofreego/tsutils'

export function FeatureFlagsPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [selectedProjectId, setSelectedProjectId] = useState<string>('')
  const [flags, setFlags] = useState<FeatureFlag[]>([])
  const [open, setOpen] = useState(false)
  const notify = useNotification()

  // form state
  const [key, setKey] = useState('')
  const [name, setName] = useState('')
  const [rolloutPct, setRolloutPct] = useState(100)

  const loadProjects = async () => {
    try {
      const res = await projectService.list()
      setProjects(res.results || [])
      if (res.results && res.results.length > 0) {
        setSelectedProjectId(res.results[0].id)
      }
    } catch (err: any) {
      notify.error('Failed to load projects')
    }
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
    loadProjects()
  }, [])

  useEffect(() => {
    if (selectedProjectId) {
      loadFlags(selectedProjectId)
    }
  }, [selectedProjectId])

  const handleCreate = async () => {
    if (!selectedProjectId) return
    try {
      await featureFlagService.create(selectedProjectId, {
        key,
        name,
        active: true,
        rolloutPct: rolloutPct
      })
      notify.success('Feature flag created')
      setOpen(false)
      loadFlags(selectedProjectId)
    } catch (err: any) {
      notify.error(err.message || 'Failed to create feature flag')
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

  return (
    <Box sx={{ p: 3 }}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Feature Flags</Typography>
        <Box display="flex" gap={2}>
          <FormControl size="small" sx={{ minWidth: 200 }}>
            <InputLabel>Project</InputLabel>
            <Select
              label="Project"
              value={selectedProjectId}
              onChange={(e) => setSelectedProjectId(e.target.value)}
            >
              {projects.map(p => (
                <MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>
              ))}
            </Select>
          </FormControl>
          <Button variant="contained" color="primary" onClick={() => setOpen(true)} disabled={!selectedProjectId}>Create Flag</Button>
        </Box>
      </Box>

      {selectedProjectId ? (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Key</TableCell>
                <TableCell>Rollout %</TableCell>
                <TableCell>Active</TableCell>
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
                </TableRow>
              ))}
              {flags.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} align="center">No feature flags found</TableCell>
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

      <Dialog open={open} onClose={() => setOpen(false)}>
        <DialogTitle>Create Feature Flag</DialogTitle>
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
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button onClick={handleCreate} variant="contained" disabled={!name || !key}>Create</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}
