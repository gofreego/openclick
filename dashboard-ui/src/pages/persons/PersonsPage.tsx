import { useEffect, useState, useCallback } from 'react'
import {
  Typography, Box, Paper, Table, TableBody, TableCell, TableContainer,
  TableHead, TableRow, IconButton, Tooltip, Chip, TextField, Button,
  Dialog, DialogTitle, DialogContent, DialogActions, Drawer, Divider,
  Tabs, Tab, TablePagination, InputAdornment, CircularProgress
} from '@mui/material'
import DeleteIcon from '@mui/icons-material/Delete'
import SearchIcon from '@mui/icons-material/Search'
import AddIcon from '@mui/icons-material/Add'
import { personService } from '../../services/personService'
import type { Person, PersonDetail } from '../../services/personService'
import { cohortService } from '../../services/cohortService'
import type { Cohort } from '../../services/cohortService'
import { useCurrentProject } from '../../contexts/ProjectContext'
import { useNotification } from '@gofreego/tsutils'
import { PageHeader } from '../../components/PageHeader'

function PersonDetailDrawer({ projectId, person, open, onClose, onDeleted }: {
  projectId: string
  person: Person | null
  open: boolean
  onClose: () => void
  onDeleted: () => void
}) {
  const [detail, setDetail] = useState<PersonDetail | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const notify = useNotification()

  useEffect(() => {
    if (open && person) {
      setLoading(true)
      personService.get(projectId, person.distinctId)
        .then(setDetail)
        .catch(() => notify.error('Failed to load person details'))
        .finally(() => setLoading(false))
    } else {
      setDetail(null)
    }
  }, [open, person, projectId])

  const handleDelete = async () => {
    if (!person) return
    try {
      await personService.delete(projectId, person.distinctId)
      notify.success('Person deleted')
      setDeleteOpen(false)
      onClose()
      onDeleted()
    } catch {
      notify.error('Failed to delete person')
    }
  }

  return (
    <>
      <Drawer anchor="right" open={open} onClose={onClose}>
        <Box sx={{ width: 520, p: 3 }}>
          {loading && <Box display="flex" justifyContent="center" pt={4}><CircularProgress /></Box>}
          {detail && !loading && (
            <>
              <Box display="flex" justifyContent="space-between" alignItems="flex-start">
                <Box>
                  <Typography variant="h6" fontWeight={700}>Person Detail</Typography>
                  <Typography variant="body2" fontFamily="monospace" color="text.secondary">{detail.person?.distinctId}</Typography>
                </Box>
                <Tooltip title="Delete person">
                  <IconButton color="error" onClick={() => setDeleteOpen(true)}>
                    <DeleteIcon />
                  </IconButton>
                </Tooltip>
              </Box>

              <Divider sx={{ my: 2 }} />

              <Typography variant="subtitle2" fontWeight={600} gutterBottom>Properties</Typography>
              <Paper variant="outlined" sx={{ p: 2, mb: 3, maxHeight: 200, overflow: 'auto', bgcolor: 'grey.50' }}>
                <pre style={{ margin: 0, fontSize: '0.8rem', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                  {JSON.stringify(detail.person?.properties, null, 2)}
                </pre>
              </Paper>

              <Typography variant="subtitle2" fontWeight={600} gutterBottom>
                Recent Events ({detail.recentEvents?.length || 0})
              </Typography>
              <TableContainer component={Paper} variant="outlined" sx={{ maxHeight: 350 }}>
                <Table size="small" stickyHeader>
                  <TableHead>
                    <TableRow>
                      <TableCell>Event</TableCell>
                      <TableCell>Timestamp</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {(detail.recentEvents || []).map((e, i) => (
                      <TableRow key={e.uuid || i}>
                        <TableCell><strong>{e.event}</strong></TableCell>
                        <TableCell>{e.timestamp ? new Date(e.timestamp as any).toLocaleString() : '—'}</TableCell>
                      </TableRow>
                    ))}
                    {(!detail.recentEvents || detail.recentEvents.length === 0) && (
                      <TableRow>
                        <TableCell colSpan={2} align="center">No recent events</TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </TableContainer>
            </>
          )}
        </Box>
      </Drawer>

      <Dialog open={deleteOpen} onClose={() => setDeleteOpen(false)}>
        <DialogTitle>Delete Person</DialogTitle>
        <DialogContent>
          <Typography>Delete person <strong>{person?.distinctId}</strong>? All associated data will be removed. This cannot be undone.</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteOpen(false)}>Cancel</Button>
          <Button onClick={handleDelete} variant="contained" color="error">Delete</Button>
        </DialogActions>
      </Dialog>
    </>
  )
}

function CohortsTab({ projectId }: { projectId: string }) {
  const [cohorts, setCohorts] = useState<Cohort[]>([])
  const [createOpen, setCreateOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deletingCohort, setDeletingCohort] = useState<Cohort | null>(null)
  const [cohortName, setCohortName] = useState('')
  const [cohortFilters, setCohortFilters] = useState('{}')
  const [filtersError, setFiltersError] = useState('')
  const notify = useNotification()

  const loadCohorts = useCallback(async () => {
    try {
      const res = await cohortService.list(projectId)
      setCohorts(res.results || [])
    } catch {
      notify.error('Failed to load cohorts')
    }
  }, [projectId])

  useEffect(() => { loadCohorts() }, [loadCohorts])

  const handleCreate = async () => {
    let filters: Record<string, any> = {}
    try {
      filters = JSON.parse(cohortFilters)
    } catch {
      setFiltersError('Invalid JSON')
      return
    }
    try {
      await cohortService.create(projectId, { name: cohortName, filters })
      notify.success('Cohort created')
      setCreateOpen(false)
      setCohortName('')
      setCohortFilters('{}')
      loadCohorts()
    } catch (err: any) {
      notify.error(err.message || 'Failed to create cohort')
    }
  }

  const confirmDelete = (c: Cohort) => { setDeletingCohort(c); setDeleteOpen(true) }
  const handleDelete = async () => {
    if (!deletingCohort) return
    try {
      await cohortService.delete(projectId, deletingCohort.id)
      notify.success('Cohort deleted')
      setDeleteOpen(false)
      setDeletingCohort(null)
      loadCohorts()
    } catch {
      notify.error('Failed to delete cohort')
    }
  }

  return (
    <Box>
      <Box display="flex" justifyContent="flex-end" mb={2}>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setCreateOpen(true)}>
          Create Cohort
        </Button>
      </Box>
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Person Count</TableCell>
              <TableCell>Created</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {cohorts.map((c) => (
              <TableRow key={c.id} hover>
                <TableCell><strong>{c.name}</strong></TableCell>
                <TableCell>
                  <Chip label={c.personCount?.toString() || '0'} size="small" />
                </TableCell>
                <TableCell>{c.createdAt ? new Date(c.createdAt as any).toLocaleDateString() : '—'}</TableCell>
                <TableCell align="right">
                  <Tooltip title="Delete cohort">
                    <IconButton size="small" color="error" onClick={() => confirmDelete(c)}>
                      <DeleteIcon />
                    </IconButton>
                  </Tooltip>
                </TableCell>
              </TableRow>
            ))}
            {cohorts.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} align="center">No cohorts found. Create one to segment your users.</TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Create Cohort</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus margin="dense" label="Cohort Name" fullWidth variant="outlined"
            value={cohortName} onChange={(e) => setCohortName(e.target.value)} sx={{ mb: 2, mt: 1 }}
          />
          <TextField
            margin="dense" label="Filters (JSON)" fullWidth multiline rows={5} variant="outlined"
            value={cohortFilters}
            onChange={(e) => { setCohortFilters(e.target.value); setFiltersError('') }}
            error={!!filtersError} helperText={filtersError || 'e.g. { "country": "US" }'}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOpen(false)}>Cancel</Button>
          <Button onClick={handleCreate} variant="contained" disabled={!cohortName}>Create</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={deleteOpen} onClose={() => setDeleteOpen(false)}>
        <DialogTitle>Delete Cohort</DialogTitle>
        <DialogContent>
          <Typography>Delete cohort <strong>{deletingCohort?.name}</strong>? This cannot be undone.</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteOpen(false)}>Cancel</Button>
          <Button onClick={handleDelete} variant="contained" color="error">Delete</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}

export function PersonsPage() {
  const selectedProjectId = useCurrentProject()
  const [activeTab, setActiveTab] = useState(0)
  const [persons, setPersons] = useState<Person[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(25)
  const [search, setSearch] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [selectedPerson, setSelectedPerson] = useState<Person | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const notify = useNotification()

  const loadPersons = useCallback(async (projectId: string) => {
    try {
      const res = await personService.list(projectId, {
        search: search || undefined,
        limit: rowsPerPage,
        offset: page * rowsPerPage,
      })
      setPersons(res.results || [])
      setTotal(Number(res.total) || 0)
    } catch {
      notify.error('Failed to load persons')
    }
  }, [search, rowsPerPage, page])

  useEffect(() => {
    if (selectedProjectId) loadPersons(selectedProjectId)
  }, [selectedProjectId, loadPersons])

  const handleSearch = () => {
    setSearch(searchInput)
    setPage(0)
  }

  const openDetail = (person: Person) => {
    setSelectedPerson(person)
    setDetailOpen(true)
  }

  return (
    <Box sx={{ p: 3 }}>
      <PageHeader
        title="Persons & Cohorts"
        infoTitle="About Persons & Cohorts"
        infoDescription="Persons represent the unique users of your application. Track their properties, event history, and group them into cohorts."
      />

      {selectedProjectId ? (
        <>
          <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)} sx={{ mb: 2 }}>
            <Tab label="Persons" />
            <Tab label="Cohorts" />
          </Tabs>

          {activeTab === 0 && (
            <Box>
              <Box display="flex" gap={1} mb={2}>
                <TextField
                  size="small"
                  placeholder="Search by distinct ID or property..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                  InputProps={{
                    startAdornment: <InputAdornment position="start"><SearchIcon fontSize="small" /></InputAdornment>
                  }}
                  sx={{ width: 360 }}
                />
                <Button variant="outlined" onClick={handleSearch}>Search</Button>
                {search && <Button variant="text" onClick={() => { setSearch(''); setSearchInput(''); setPage(0) }}>Clear</Button>}
              </Box>

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
                      <TableRow
                        key={person.id}
                        hover
                        sx={{ cursor: 'pointer' }}
                        onClick={() => openDetail(person)}
                      >
                        <TableCell>
                          <Typography variant="body2" fontFamily="monospace" color="primary.main">
                            {person.distinctId}
                          </Typography>
                        </TableCell>
                        <TableCell>
                          <Typography variant="body2" color="text.secondary" sx={{ maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                            {JSON.stringify(person.properties)}
                          </Typography>
                        </TableCell>
                        <TableCell>{person.createdAt ? new Date(person.createdAt as any).toLocaleString() : '—'}</TableCell>
                      </TableRow>
                    ))}
                    {persons.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={3} align="center">No persons tracked yet</TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
                <TablePagination
                  component="div"
                  count={total}
                  page={page}
                  onPageChange={(_, p) => setPage(p)}
                  rowsPerPage={rowsPerPage}
                  onRowsPerPageChange={(e) => { setRowsPerPage(parseInt(e.target.value, 10)); setPage(0) }}
                  rowsPerPageOptions={[10, 25, 50, 100]}
                />
              </TableContainer>
            </Box>
          )}

          {activeTab === 1 && <CohortsTab projectId={selectedProjectId} />}
        </>
      ) : (
        <Paper sx={{ p: 3, textAlign: 'center' }}>
          <Typography>Please select a project to view persons.</Typography>
        </Paper>
      )}

      <PersonDetailDrawer
        projectId={selectedProjectId || ''}
        person={selectedPerson}
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        onDeleted={() => selectedProjectId && loadPersons(selectedProjectId)}
      />
    </Box>
  )
}
