import { useEffect, useState } from 'react'
import { Typography, Box, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material'
import { personService } from '../../services/personService'
import type { Person } from '../../services/personService'
import { useCurrentProject } from '../../hooks/useCurrentProject'
import { useNotification } from '@gofreego/tsutils'
import { PageHeader } from '../../components/PageHeader'

export function PersonsPage() {
  const selectedProjectId = useCurrentProject()
  const [persons, setPersons] = useState<Person[]>([])
  const notify = useNotification()

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
    if (selectedProjectId) {
      loadPersons(selectedProjectId)
    }
  }, [selectedProjectId])

  return (
    <Box sx={{ p: 3 }}>
      <PageHeader 
        title="Persons & Cohorts" 
        infoTitle="About Persons & Cohorts"
        infoDescription="Persons represent the unique users of your application. You can track their properties, see their complete event history, and group them into cohorts based on their behavior or attributes."
      />

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
                  <TableCell><code>{person.distinctId}</code></TableCell>
                  <TableCell><pre style={{ margin: 0, fontSize: '0.8em' }}>{JSON.stringify(person.properties, null, 2)}</pre></TableCell>
                  <TableCell>{person.createdAt ? new Date(person.createdAt as any).toLocaleString() : ''}</TableCell>
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
