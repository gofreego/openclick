import { useEffect, useState } from 'react'
import {
  Typography, Box, Paper, Button, Grid, Card, CardContent, CardActions,
  TextField, Dialog, DialogTitle, DialogContent, DialogActions,
  IconButton, Chip, Divider, Tooltip, Drawer, List, ListItem,
  ListItemText, ListItemSecondaryAction, Select, MenuItem, FormControl,
  InputLabel, InputAdornment
} from '@mui/material'
import DeleteIcon from '@mui/icons-material/Delete'
import EditIcon from '@mui/icons-material/Edit'
import SettingsIcon from '@mui/icons-material/Settings'
import ContentCopyIcon from '@mui/icons-material/ContentCopy'
import PersonAddIcon from '@mui/icons-material/PersonAdd'
import { projectService } from '../../services/projectService'
import type { Project, ProjectDetail } from '../../services/projectService'
import { useNotification } from '@gofreego/tsutils'
import { PageHeader } from '../../components/PageHeader'

const ROLES = ['admin', 'member', 'viewer']

export function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [detailOpen, setDetailOpen] = useState(false)
  const [addMemberOpen, setAddMemberOpen] = useState(false)
  const [newProjectName, setNewProjectName] = useState('')
  const [newProjectTimezone, setNewProjectTimezone] = useState(Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC')
  const [editingProject, setEditingProject] = useState<Project | null>(null)
  const [editName, setEditName] = useState('')
  const [editTimezone, setEditTimezone] = useState('')
  const [deletingProject, setDeletingProject] = useState<Project | null>(null)
  const [selectedProject, setSelectedProject] = useState<ProjectDetail | null>(null)
  const [newMemberUserId, setNewMemberUserId] = useState('')
  const [newMemberRole, setNewMemberRole] = useState('member')
  const notify = useNotification()

  const loadProjects = async () => {
    try {
      const res = await projectService.list()
      setProjects(res.results || [])
    } catch (err: any) {
      notify.error(err.message || 'Failed to load projects')
    }
  }

  useEffect(() => {
    loadProjects()
  }, [])

  const handleCreate = async () => {
    try {
      await projectService.create({ name: newProjectName, timezone: newProjectTimezone })
      setCreateOpen(false)
      setNewProjectName('')
      notify.success('Project created successfully')
      loadProjects()
    } catch (err: any) {
      notify.error(err.message || 'Failed to create project')
    }
  }

  const openEdit = (p: Project) => {
    setEditingProject(p)
    setEditName(p.name)
    setEditTimezone(p.timezone)
    setEditOpen(true)
  }

  const handleUpdate = async () => {
    if (!editingProject) return
    try {
      await projectService.update(editingProject.id, { name: editName, timezone: editTimezone })
      notify.success('Project updated')
      setEditOpen(false)
      setEditingProject(null)
      loadProjects()
    } catch (err: any) {
      notify.error(err.message || 'Failed to update project')
    }
  }

  const openDelete = (p: Project) => {
    setDeletingProject(p)
    setDeleteOpen(true)
  }

  const handleDelete = async () => {
    if (!deletingProject) return
    try {
      await projectService.delete(deletingProject.id)
      notify.success('Project deleted')
      setDeleteOpen(false)
      setDeletingProject(null)
      loadProjects()
    } catch (err: any) {
      notify.error(err.message || 'Failed to delete project')
    }
  }

  const openDetail = async (p: Project) => {
    try {
      const detail = await projectService.getById(p.id)
      setSelectedProject(detail)
      setDetailOpen(true)
    } catch (err: any) {
      notify.error('Failed to load project details')
    }
  }

  const handleAddMember = async () => {
    if (!selectedProject) return
    try {
      await projectService.addMember(selectedProject.id, newMemberUserId, newMemberRole)
      notify.success('Member added')
      setAddMemberOpen(false)
      setNewMemberUserId('')
      setNewMemberRole('member')
      const detail = await projectService.getById(selectedProject.id)
      setSelectedProject(detail)
    } catch (err: any) {
      notify.error(err.message || 'Failed to add member')
    }
  }

  const handleRemoveMember = async (userId: string) => {
    if (!selectedProject) return
    try {
      await projectService.removeMember(selectedProject.id, userId)
      notify.success('Member removed')
      const detail = await projectService.getById(selectedProject.id)
      setSelectedProject(detail)
    } catch (err: any) {
      notify.error('Failed to remove member')
    }
  }

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    notify.success(`${label} copied to clipboard`)
  }

  return (
    <Box sx={{ p: 3 }}>
      <PageHeader
        title="Projects"
        infoTitle="About Projects"
        infoDescription="Projects are isolated environments within OpenClick. Each project has its own API Key and settings."
        action={<Button variant="contained" color="primary" onClick={() => setCreateOpen(true)}>Create Project</Button>}
      />

      <Grid container spacing={3}>
        {projects.map(p => (
          <Grid size={{ xs: 12, md: 4 }} key={p.id}>
            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1 }}>
                <Typography variant="h6" gutterBottom fontWeight={600}>{p.name}</Typography>
                <Box sx={{ mb: 1 }}>
                  <Typography variant="caption" color="text.secondary">API Key</Typography>
                  <Box display="flex" alignItems="center" gap={0.5} mt={0.25}>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.75rem', wordBreak: 'break-all' }}>
                      {p.apiKey}
                    </Typography>
                    <Tooltip title="Copy API Key">
                      <IconButton size="small" onClick={() => copyToClipboard(p.apiKey, 'API Key')}>
                        <ContentCopyIcon sx={{ fontSize: 14 }} />
                      </IconButton>
                    </Tooltip>
                  </Box>
                </Box>
                <Chip label={p.timezone} size="small" variant="outlined" sx={{ mt: 0.5 }} />
              </CardContent>
              <CardActions sx={{ justifyContent: 'space-between', px: 2, pb: 2 }}>
                <Button size="small" startIcon={<SettingsIcon />} onClick={() => openDetail(p)}>
                  Manage
                </Button>
                <Box>
                  <Tooltip title="Edit project">
                    <IconButton size="small" color="primary" onClick={() => openEdit(p)}>
                      <EditIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Delete project">
                    <IconButton size="small" color="error" onClick={() => openDelete(p)}>
                      <DeleteIcon />
                    </IconButton>
                  </Tooltip>
                </Box>
              </CardActions>
            </Card>
          </Grid>
        ))}
        {projects.length === 0 && (
          <Grid size={{ xs: 12 }}>
            <Paper sx={{ p: 3, textAlign: 'center' }}>
              <Typography variant="body1">No projects found. Create one to get started.</Typography>
            </Paper>
          </Grid>
        )}
      </Grid>

      {/* Create Dialog */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Create New Project</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Project Name"
            fullWidth
            variant="outlined"
            value={newProjectName}
            onChange={(e) => setNewProjectName(e.target.value)}
            sx={{ mb: 2, mt: 1 }}
          />
          <TextField
            margin="dense"
            label="Timezone"
            fullWidth
            variant="outlined"
            value={newProjectTimezone}
            onChange={(e) => setNewProjectTimezone(e.target.value)}
            helperText="e.g. America/New_York, Asia/Kolkata"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOpen(false)}>Cancel</Button>
          <Button onClick={handleCreate} variant="contained" disabled={!newProjectName}>Create</Button>
        </DialogActions>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={editOpen} onClose={() => setEditOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Edit Project</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Project Name"
            fullWidth
            variant="outlined"
            value={editName}
            onChange={(e) => setEditName(e.target.value)}
            sx={{ mb: 2, mt: 1 }}
          />
          <TextField
            margin="dense"
            label="Timezone"
            fullWidth
            variant="outlined"
            value={editTimezone}
            onChange={(e) => setEditTimezone(e.target.value)}
            helperText="e.g. America/New_York, Asia/Kolkata"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEditOpen(false)}>Cancel</Button>
          <Button onClick={handleUpdate} variant="contained" disabled={!editName}>Save</Button>
        </DialogActions>
      </Dialog>

      {/* Delete Confirmation */}
      <Dialog open={deleteOpen} onClose={() => setDeleteOpen(false)}>
        <DialogTitle>Delete Project</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete <strong>{deletingProject?.name}</strong>? This will permanently delete all data including events, persons, and feature flags.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteOpen(false)}>Cancel</Button>
          <Button onClick={handleDelete} variant="contained" color="error">Delete</Button>
        </DialogActions>
      </Dialog>

      {/* Project Detail Drawer */}
      <Drawer anchor="right" open={detailOpen} onClose={() => setDetailOpen(false)}>
        <Box sx={{ width: 480, p: 3 }}>
          {selectedProject && (
            <>
              <Typography variant="h5" fontWeight={700} gutterBottom>{selectedProject.name}</Typography>
              <Typography variant="caption" color="text.secondary">
                Created: {selectedProject.createdAt ? new Date(selectedProject.createdAt as any).toLocaleDateString() : '—'}
              </Typography>

              <Divider sx={{ my: 2 }} />

              <Typography variant="subtitle2" fontWeight={600} gutterBottom>API Key</Typography>
              <Box display="flex" alignItems="center" gap={1} mb={2}>
                <TextField
                  size="small"
                  fullWidth
                  value={selectedProject.apiKey}
                  InputProps={{
                    readOnly: true,
                    sx: { fontFamily: 'monospace', fontSize: '0.8rem' },
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton size="small" onClick={() => copyToClipboard(selectedProject.apiKey, 'API Key')}>
                          <ContentCopyIcon fontSize="small" />
                        </IconButton>
                      </InputAdornment>
                    )
                  }}
                />
              </Box>

              <Typography variant="subtitle2" fontWeight={600} gutterBottom>Secret Key</Typography>
              <Box display="flex" alignItems="center" gap={1} mb={2}>
                <TextField
                  size="small"
                  fullWidth
                  type="password"
                  value={selectedProject.secretKey}
                  InputProps={{
                    readOnly: true,
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton size="small" onClick={() => copyToClipboard(selectedProject.secretKey, 'Secret Key')}>
                          <ContentCopyIcon fontSize="small" />
                        </IconButton>
                      </InputAdornment>
                    )
                  }}
                />
              </Box>

              <Typography variant="subtitle2" fontWeight={600} gutterBottom>Timezone</Typography>
              <Typography variant="body2" mb={2}>{selectedProject.timezone}</Typography>

              <Divider sx={{ my: 2 }} />

              <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                <Typography variant="subtitle2" fontWeight={600}>Team Members ({selectedProject.members?.length || 0})</Typography>
                <Button size="small" startIcon={<PersonAddIcon />} onClick={() => setAddMemberOpen(true)}>
                  Add Member
                </Button>
              </Box>

              <List dense>
                {(selectedProject.members || []).map((m) => (
                  <ListItem key={m.userId} divider>
                    <ListItemText
                      primary={
                        <Box display="flex" alignItems="center" gap={1.5}>
                          <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 500 }}>
                            {m.userId}
                          </Typography>
                          <Chip
                            label={m.role}
                            size="small"
                            color={m.role === 'admin' ? 'primary' : (m.role === 'owner' ? 'error' : 'default')}
                          />
                        </Box>
                      }
                      primaryTypographyProps={{ component: 'div' }}
                    />
                    <ListItemSecondaryAction>
                      <Tooltip title="Remove member">
                        <IconButton size="small" color="error" onClick={() => handleRemoveMember(m.userId)}>
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </ListItemSecondaryAction>
                  </ListItem>
                ))}
                {(!selectedProject.members || selectedProject.members.length === 0) && (
                  <Typography variant="body2" color="text.secondary" sx={{ py: 1 }}>
                    No members found.
                  </Typography>
                )}
              </List>
            </>
          )}
        </Box>
      </Drawer>

      {/* Add Member Dialog */}
      <Dialog open={addMemberOpen} onClose={() => setAddMemberOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Team Member</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="User ID"
            fullWidth
            variant="outlined"
            value={newMemberUserId}
            onChange={(e) => setNewMemberUserId(e.target.value)}
            sx={{ mb: 2, mt: 1 }}
          />
          <FormControl fullWidth>
            <InputLabel>Role</InputLabel>
            <Select
              value={newMemberRole}
              label="Role"
              onChange={(e) => setNewMemberRole(e.target.value)}
            >
              {ROLES.map(r => (
                <MenuItem key={r} value={r}>{r.charAt(0).toUpperCase() + r.slice(1)}</MenuItem>
              ))}
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setAddMemberOpen(false)}>Cancel</Button>
          <Button onClick={handleAddMember} variant="contained" disabled={!newMemberUserId}>Add</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}
