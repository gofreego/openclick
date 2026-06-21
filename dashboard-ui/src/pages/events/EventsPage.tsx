import { useEffect, useState } from 'react'
import { Typography, Box, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Select, MenuItem, FormControl, InputLabel, Button } from '@mui/material'
import { eventService } from '../../services/eventService'
import type { Event } from '../../services/eventService'
import { projectService } from '../../services/projectService'
import type { Project } from '../../services/projectService'
import { useNotification } from '@gofreego/tsutils'
import { PageHeader } from '../../components/PageHeader'

export function EventsPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [selectedProjectId, setSelectedProjectId] = useState<string>('')
  const [events, setEvents] = useState<Event[]>([])
  const notify = useNotification()

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

  const loadEvents = async (projectId: string) => {
    if (!projectId) return
    try {
      const res = await eventService.queryEvents(projectId, { limit: 50 })
      setEvents(res.results || [])
    } catch (err: any) {
      notify.error('Failed to load events')
    }
  }

  useEffect(() => {
    loadProjects()
  }, [])

  useEffect(() => {
    if (selectedProjectId) {
      loadEvents(selectedProjectId)
    }
  }, [selectedProjectId])

  return (
    <Box sx={{ p: 3 }}>
      <PageHeader 
        title="Events & Replay" 
        infoTitle="About Events & Replay"
        infoDescription="This page displays the raw stream of events coming into your project. You can inspect event properties, view event sequences, and watch session replays to understand exactly how users are interacting with your application."
        action={
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
        }
      />

      {selectedProjectId ? (
        <TableContainer component={Paper}>
          <Box display="flex" justifyContent="flex-end" p={2}>
            <Button variant="outlined" onClick={() => loadEvents(selectedProjectId)}>Refresh</Button>
          </Box>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Event</TableCell>
                <TableCell>Distinct ID</TableCell>
                <TableCell>Properties</TableCell>
                <TableCell>Timestamp</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {events.map((event) => (
                <TableRow key={event.uuid}>
                  <TableCell><strong>{event.event}</strong></TableCell>
                  <TableCell><code>{event.distinctId}</code></TableCell>
                  <TableCell><pre style={{ margin: 0, fontSize: '0.8em', maxWidth: '300px', overflowX: 'auto' }}>{JSON.stringify(event.properties, null, 2)}</pre></TableCell>
                  <TableCell>{event.timestamp ? new Date(event.timestamp as any).toLocaleString() : ''}</TableCell>
                </TableRow>
              ))}
              {events.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} align="center">No events found in this project</TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      ) : (
        <Paper sx={{ p: 3, textAlign: 'center' }}>
          <Typography>Please select a project to view events.</Typography>
        </Paper>
      )}
    </Box>
  )
}
