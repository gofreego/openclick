import { useEffect, useState } from 'react'
import { Typography, Box, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Select, MenuItem, FormControl, InputLabel } from '@mui/material'
import { personService } from '../../services/personService'
import type { Person } from '../../services/personService'
import { projectService } from '../../services/projectService'
import type { Project } from '../../services/projectService'
import { useNotification } from '@gofreego/tsutils'

export function PersonsPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [selectedProjectId, setSelectedProjectId] = useState<string>('')
  const [persons, setPersons] = useState<Person[]>([])
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

  const loadPersons = async (projectId: string) => {
    if (!projectId) return
    try {
      const res = await personService.list(projectId)
      setPersons(res.results || [])
    } catch (err: any) {
      notify.error('Failed to load persons')
    }
  }

  useEffect(() => {
    loadProjects()
  }, [])

  useEffect(() => {
    if (selectedProjectId) {
      loadPersons(selectedProjectId)
    }
  }, [selectedProjectId])

  return (
    <Box sx={{ p: 3 }}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Persons & Cohorts</Typography>
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
      </Box>

      {selectedProjectId ? (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Distinct ID</TableCell>
                <TableCell>Properties</TableCell>
                <TableCell>First Seen</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {persons.map((person) => (
                <TableRow key={person.id}>
                  <TableCell><code>{person.distinct_id}</code></TableCell>
                  <TableCell><pre style={{ margin: 0, fontSize: '0.8em' }}>{JSON.stringify(person.properties, null, 2)}</pre></TableCell>
                  <TableCell>{new Date(person.created_at).toLocaleString()}</TableCell>
                </TableRow>
              ))}
              {persons.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} align="center">No persons tracked yet</TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      ) : (
        <Paper sx={{ p: 3, textAlign: 'center' }}>
          <Typography>Please select a project to view persons.</Typography>
        </Paper>
      )}
    </Box>
  )
}
