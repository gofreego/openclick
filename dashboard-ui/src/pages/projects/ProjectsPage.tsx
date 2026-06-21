import { useEffect, useState } from 'react'
import { Typography, Box, Paper, Button, Grid, Card, CardContent, CardActions, TextField, Dialog, DialogTitle, DialogContent, DialogActions } from '@mui/material'
import { projectService } from '../../services/projectService'
import type { Project } from '../../services/projectService'
import { useNotification } from '@gofreego/tsutils'
import { PageHeader } from '../../components/PageHeader'

export function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [open, setOpen] = useState(false)
  const [newProjectName, setNewProjectName] = useState('')
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
      await projectService.create({ name: newProjectName, timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC' })
      setOpen(false)
      setNewProjectName('')
      notify.success('Project created successfully')
      loadProjects()
    } catch (err: any) {
      notify.error(err.message || 'Failed to create project')
    }
  }

  return (
    <Box sx={{ p: 3 }}>
      <PageHeader 
        title="Projects" 
        infoTitle="About Projects"
        infoDescription="Projects are isolated environments within OpenClick. You can use different projects for different applications, or for different environments like development, staging, and production. Each project has its own API Key and settings."
        action={<Button variant="contained" color="primary" onClick={() => setOpen(true)}>Create Project</Button>} 
      />

      <Grid container spacing={3}>
        {projects.map(p => (
          <Grid size={{xs: 12, md: 4}} key={p.id}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>{p.name}</Typography>
                <Typography variant="body2" color="textSecondary">API Key: {p.apiKey}</Typography>
                <Typography variant="body2" color="textSecondary">Timezone: {p.timezone}</Typography>
              </CardContent>
              <CardActions>
                <Button size="small">Manage</Button>
              </CardActions>
            </Card>
          </Grid>
        ))}
        {projects.length === 0 && (
          <Grid size={{xs: 12}}>
            <Paper sx={{ p: 3, textAlign: 'center' }}>
              <Typography variant="body1">No projects found. Create one to get started.</Typography>
            </Paper>
          </Grid>
        )}
      </Grid>

      <Dialog open={open} onClose={() => setOpen(false)}>
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
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button onClick={handleCreate} variant="contained" disabled={!newProjectName}>Create</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}
