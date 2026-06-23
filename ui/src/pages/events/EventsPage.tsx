import { useEffect, useState, useCallback } from 'react'
import {
  Typography, Box, Paper, Table, TableBody, TableCell, TableContainer,
  TableHead, TableRow, Button, TextField, InputAdornment, Tabs, Tab,
  TablePagination, Chip, IconButton, Tooltip, Dialog, DialogTitle,
  DialogContent, DialogActions, Drawer, Divider, LinearProgress
} from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import DeleteIcon from '@mui/icons-material/Delete'
import PlayCircleOutlineIcon from '@mui/icons-material/PlayCircleOutline'
import RefreshIcon from '@mui/icons-material/Refresh'
import { eventService } from '../../services/eventService'
import type { Event } from '../../services/eventService'
import { sessionService } from '../../services/sessionService'
import type { Session } from '../../services/sessionService'
import { SessionChunk } from '../../apis/proto/openclick/v1/analytics'
import { useCurrentProject } from '../../contexts/ProjectContext'
import { useNotification } from '@gofreego/tsutils'
import { PageHeader } from '../../components/PageHeader'
import { TabInfoButton } from '../../components/TabInfoButton'
import type { TabInfo } from '../../components/TabInfoButton'

const EVENTS_TAB_INFO: Record<string, TabInfo> = {
  events: {
    title: 'Events',
    meaning: 'Events are the raw actions captured from your users — page views, clicks, form submissions, and any custom events you track. Each event carries a distinct ID, timestamp, and a properties bag.',
    howToUse: 'Filter by event name or distinct ID to narrow the list. Click any row to open a detail panel showing the full properties JSON. Use pagination to browse older events.',
    example: 'To debug a sign-up issue: filter by event "$identify" and the user\'s distinct ID to inspect the exact properties sent during registration.'
  },
  replay: {
    title: 'Session Replays',
    meaning: 'Session Replays capture the sequence of rrweb chunks recorded during a user\'s browsing session. Each chunk is a timestamped snapshot of DOM mutations, mouse moves, and interactions.',
    howToUse: 'Search by distinct ID to find sessions for a specific user. Click the play icon to open the replay drawer and browse the recorded chunks for that session. Use the delete icon to remove sessions.',
    example: 'To reproduce a bug report: find the session by the user\'s distinct ID, open the replay, and inspect the chunk data around the reported timestamp to see what the user encountered.'
  }
}

function formatDuration(ms: number | string | bigint): string {
  const n = Number(ms)
  if (n < 1000) return `${n}ms`
  if (n < 60000) return `${(n / 1000).toFixed(1)}s`
  return `${Math.floor(n / 60000)}m ${Math.round((n % 60000) / 1000)}s`
}

function EventDetailDrawer({ event, open, onClose }: {
  event: Event | null
  open: boolean
  onClose: () => void
}) {
  return (
    <Drawer anchor="right" open={open} onClose={onClose}>
      <Box sx={{ width: 520, p: 3 }}>
        {event && (
          <>
            <Typography variant="h6" fontWeight={700} gutterBottom>Event Detail</Typography>
            <Divider sx={{ my: 2 }} />
            <Box mb={2}>
              <Chip label={event.event} color="primary" />
            </Box>
            <Box mb={1}>
              <Typography variant="caption" color="text.secondary">Distinct ID</Typography>
              <Typography variant="body2" fontFamily="monospace">{event.distinctId}</Typography>
            </Box>
            <Box mb={2}>
              <Typography variant="caption" color="text.secondary">Timestamp</Typography>
              <Typography variant="body2">{event.timestamp ? new Date(event.timestamp as any).toLocaleString() : '—'}</Typography>
            </Box>
            <Typography variant="subtitle2" fontWeight={600} gutterBottom>Properties</Typography>
            <Box component="pre" sx={{ m: 0, p: 2, fontSize: '0.8rem', whiteSpace: 'pre-wrap', wordBreak: 'break-all', lineHeight: 1.6, fontFamily: 'monospace', bgcolor: 'grey.900', color: 'grey.300', borderRadius: 1, border: 1, borderColor: 'divider', overflow: 'auto', maxHeight: 'calc(100vh - 280px)' }}>
              {JSON.stringify(event.properties, null, 2)}
            </Box>
          </>
        )}
      </Box>
    </Drawer>
  )
}

function SessionReplayDrawer({ projectId, session, open, onClose }: {
  projectId: string
  session: Session | null
  open: boolean
  onClose: () => void
}) {
  const [chunks, setChunks] = useState<SessionChunk[]>([])
  const [totalChunks, setTotalChunks] = useState(0)
  const [loading, setLoading] = useState(false)
  const notify = useNotification()

  useEffect(() => {
    if (open && session) {
      setLoading(true)
      sessionService.getChunks(projectId, session.sessionId)
        .then((res) => {
          setChunks(res.chunks || [])
          setTotalChunks(Number(res.totalChunks) || 0)
        })
        .catch(() => notify.error('Failed to load session chunks'))
        .finally(() => setLoading(false))
    } else {
      setChunks([])
      setTotalChunks(0)
    }
  }, [open, session, projectId])

  return (
    <Drawer anchor="right" open={open} onClose={onClose}>
      <Box sx={{ width: 600, p: 3 }}>
        {session && (
          <>
            <Typography variant="h6" fontWeight={700} gutterBottom>Session Replay</Typography>
            <Typography variant="body2" fontFamily="monospace" color="text.secondary">{session.sessionId}</Typography>
            <Divider sx={{ my: 2 }} />

            <Box display="flex" flexWrap="wrap" gap={1} mb={2}>
              <Chip label={`Duration: ${formatDuration(session.durationMs)}`} size="small" />
              <Chip label={`Pages: ${session.pageCount}`} size="small" />
              <Chip label={`Clicks: ${session.clickCount}`} size="small" />
              {session.browser && <Chip label={session.browser} size="small" variant="outlined" />}
              {session.os && <Chip label={session.os} size="small" variant="outlined" />}
              {session.deviceType && <Chip label={session.deviceType} size="small" variant="outlined" />}
              {session.countryCode && <Chip label={`🌍 ${session.countryCode}`} size="small" variant="outlined" />}
            </Box>

            <Divider sx={{ mb: 2 }} />

            <Typography variant="subtitle2" fontWeight={600} gutterBottom>
              Replay Chunks ({totalChunks} total)
            </Typography>

            {loading && <LinearProgress sx={{ mb: 2 }} />}

            {!loading && chunks.length === 0 && (
              <Paper variant="outlined" sx={{ p: 2, textAlign: 'center' }}>
                <Typography variant="body2" color="text.secondary">No replay chunks recorded for this session.</Typography>
              </Paper>
            )}

            {!loading && chunks.length > 0 && (
              <Box sx={{ maxHeight: 500, overflow: 'auto' }}>
                {chunks.map((chunk) => (
                  <Paper key={chunk.chunkIndex} variant="outlined" sx={{ p: 2, mb: 1 }}>
                    <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                      <Chip label={`Chunk #${chunk.chunkIndex}`} size="small" color="primary" />
                      <Typography variant="caption" color="text.secondary">
                        {chunk.timestamp ? new Date(chunk.timestamp as any).toLocaleTimeString() : '—'}
                      </Typography>
                    </Box>
                    <Box component="pre" sx={{ m: 0, p: 1, maxHeight: 150, overflow: 'auto', fontSize: '0.7rem', fontFamily: 'monospace', whiteSpace: 'pre-wrap', wordBreak: 'break-all', lineHeight: 1.5, bgcolor: 'grey.900', color: 'grey.300', borderRadius: 1, border: 1, borderColor: 'divider' }}>
                      {chunk.data}
                    </Box>
                  </Paper>
                ))}
              </Box>
            )}
          </>
        )}
      </Box>
    </Drawer>
  )
}

function SessionsTab({ projectId }: { projectId: string }) {
  const [sessions, setSessions] = useState<Session[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(25)
  const [search, setSearch] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [replaySession, setReplaySession] = useState<Session | null>(null)
  const [replayOpen, setReplayOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deletingSession, setDeletingSession] = useState<Session | null>(null)
  const notify = useNotification()

  const loadSessions = useCallback(async () => {
    try {
      const res = await sessionService.list(projectId, {
        search: search || undefined,
        limit: rowsPerPage,
        offset: page * rowsPerPage,
      })
      setSessions(res.results || [])
      setTotal(Number(res.total) || 0)
    } catch {
      notify.error('Failed to load sessions')
    }
  }, [projectId, search, page, rowsPerPage])

  useEffect(() => { loadSessions() }, [loadSessions])

  const openReplay = (s: Session) => { setReplaySession(s); setReplayOpen(true) }
  const confirmDelete = (s: Session) => { setDeletingSession(s); setDeleteOpen(true) }
  const handleDelete = async () => {
    if (!deletingSession) return
    try {
      await sessionService.delete(projectId, deletingSession.sessionId)
      notify.success('Session deleted')
      setDeleteOpen(false)
      setDeletingSession(null)
      loadSessions()
    } catch {
      notify.error('Failed to delete session')
    }
  }

  return (
    <Box>
      <Box display="flex" gap={1} mb={2} justifyContent="space-between">
        <Box display="flex" gap={1}>
          <TextField
            size="small"
            placeholder="Search by distinct ID..."
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && (setSearch(searchInput), setPage(0))}
            InputProps={{ startAdornment: <InputAdornment position="start"><SearchIcon fontSize="small" /></InputAdornment> }}
            sx={{ width: 320 }}
          />
          <Button variant="outlined" onClick={() => { setSearch(searchInput); setPage(0) }}>Search</Button>
          {search && <Button variant="text" onClick={() => { setSearch(''); setSearchInput(''); setPage(0) }}>Clear</Button>}
        </Box>
        <Button startIcon={<RefreshIcon />} onClick={loadSessions}>Refresh</Button>
      </Box>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Distinct ID</TableCell>
              <TableCell>Duration</TableCell>
              <TableCell>Pages</TableCell>
              <TableCell>Clicks</TableCell>
              <TableCell>Browser</TableCell>
              <TableCell>OS</TableCell>
              <TableCell>Country</TableCell>
              <TableCell>Start Time</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {sessions.map((s) => (
              <TableRow key={s.sessionId} hover>
                <TableCell>
                  <Typography variant="body2" fontFamily="monospace" color="primary.main">{s.distinctId}</Typography>
                </TableCell>
                <TableCell>{formatDuration(s.durationMs)}</TableCell>
                <TableCell>{s.pageCount}</TableCell>
                <TableCell>{s.clickCount}</TableCell>
                <TableCell>{s.browser || '—'}</TableCell>
                <TableCell>{s.os || '—'}</TableCell>
                <TableCell>{s.countryCode || '—'}</TableCell>
                <TableCell>{s.startTime ? new Date(s.startTime as any).toLocaleString() : '—'}</TableCell>
                <TableCell align="right">
                  <Tooltip title="View Replay">
                    <IconButton size="small" color="primary" onClick={() => openReplay(s)}>
                      <PlayCircleOutlineIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Delete session">
                    <IconButton size="small" color="error" onClick={() => confirmDelete(s)}>
                      <DeleteIcon />
                    </IconButton>
                  </Tooltip>
                </TableCell>
              </TableRow>
            ))}
            {sessions.length === 0 && (
              <TableRow>
                <TableCell colSpan={9} align="center">No sessions found</TableCell>
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
          rowsPerPageOptions={[10, 25, 50]}
        />
      </TableContainer>

      <SessionReplayDrawer
        projectId={projectId}
        session={replaySession}
        open={replayOpen}
        onClose={() => setReplayOpen(false)}
      />

      <Dialog open={deleteOpen} onClose={() => setDeleteOpen(false)}>
        <DialogTitle>Delete Session</DialogTitle>
        <DialogContent>
          <Typography>Delete session <strong>{deletingSession?.sessionId}</strong>? This action cannot be undone.</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteOpen(false)}>Cancel</Button>
          <Button onClick={handleDelete} variant="contained" color="error">Delete</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}

function EventsTab({ projectId }: { projectId: string }) {
  const [events, setEvents] = useState<Event[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(50)
  const [eventFilter, setEventFilter] = useState('')
  const [eventFilterInput, setEventFilterInput] = useState('')
  const [distinctIdFilter, setDistinctIdFilter] = useState('')
  const [distinctIdInput, setDistinctIdInput] = useState('')
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null)
  const [eventDetailOpen, setEventDetailOpen] = useState(false)
  const notify = useNotification()

  const loadEvents = useCallback(async () => {
    try {
      const res = await eventService.queryEvents(projectId, {
        event: eventFilter || undefined,
        distinctId: distinctIdFilter || undefined,
        limit: rowsPerPage,
        offset: page * rowsPerPage,
      } as any)
      setEvents(res.results || [])
      setTotal(Number(res.total) || 0)
    } catch {
      notify.error('Failed to load events')
    }
  }, [projectId, eventFilter, distinctIdFilter, page, rowsPerPage])

  useEffect(() => { loadEvents() }, [loadEvents])

  const applyFilters = () => {
    setEventFilter(eventFilterInput)
    setDistinctIdFilter(distinctIdInput)
    setPage(0)
  }

  const clearFilters = () => {
    setEventFilter(''); setEventFilterInput('')
    setDistinctIdFilter(''); setDistinctIdInput('')
    setPage(0)
  }

  return (
    <Box>
      <Box display="flex" gap={1} mb={2} flexWrap="wrap" justifyContent="space-between">
        <Box display="flex" gap={1} flexWrap="wrap">
          <TextField
            size="small" placeholder="Event name..." value={eventFilterInput}
            onChange={(e) => setEventFilterInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && applyFilters()}
            sx={{ width: 200 }}
          />
          <TextField
            size="small" placeholder="Distinct ID..." value={distinctIdInput}
            onChange={(e) => setDistinctIdInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && applyFilters()}
            sx={{ width: 200 }}
          />
          <Button variant="outlined" onClick={applyFilters}>Filter</Button>
          {(eventFilter || distinctIdFilter) && (
            <Button variant="text" onClick={clearFilters}>Clear</Button>
          )}
        </Box>
        <Button startIcon={<RefreshIcon />} onClick={loadEvents}>Refresh</Button>
      </Box>

      <TableContainer component={Paper}>
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
              <TableRow key={event.uuid} hover sx={{ cursor: 'pointer' }} onClick={() => { setSelectedEvent(event); setEventDetailOpen(true) }}>
                <TableCell><Chip label={event.event} size="small" /></TableCell>
                <TableCell>
                  <Typography variant="body2" fontFamily="monospace">{event.distinctId}</Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontFamily: 'monospace', fontSize: '0.75rem' }}>
                    {JSON.stringify(event.properties)}
                  </Typography>
                </TableCell>
                <TableCell>{event.timestamp ? new Date(event.timestamp as any).toLocaleString() : '—'}</TableCell>
              </TableRow>
            ))}
            {events.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} align="center">No events found in this project</TableCell>
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
          rowsPerPageOptions={[25, 50, 100]}
        />
      </TableContainer>

      <EventDetailDrawer
        event={selectedEvent}
        open={eventDetailOpen}
        onClose={() => setEventDetailOpen(false)}
      />
    </Box>
  )
}

export function EventsPage() {
  const selectedProjectId = useCurrentProject()
  const [activeTab, setActiveTab] = useState(0)

  return (
    <Box sx={{ p: 3 }}>
      <PageHeader
        title="Events & Replay"
        infoTitle="About Events & Replay"
        infoDescription="View the raw stream of events and session replays. Filter by event name or distinct ID, paginate through results, and watch session recordings."
      />

      {selectedProjectId ? (
        <>
          <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)} sx={{ mb: 2 }}>
            <Tab label={<Box display="flex" alignItems="center" gap={0.5}>Events<TabInfoButton info={EVENTS_TAB_INFO.events} /></Box>} />
            <Tab label={<Box display="flex" alignItems="center" gap={0.5}>Session Replays<TabInfoButton info={EVENTS_TAB_INFO.replay} /></Box>} />
          </Tabs>
          {activeTab === 0 && <EventsTab projectId={selectedProjectId} />}
          {activeTab === 1 && <SessionsTab projectId={selectedProjectId} />}
        </>
      ) : (
        <Paper sx={{ p: 3, textAlign: 'center' }}>
          <Typography>Please select a project to view events.</Typography>
        </Paper>
      )}
    </Box>
  )
}
