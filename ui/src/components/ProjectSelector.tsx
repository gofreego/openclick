import { useState, useEffect, useRef, useCallback } from 'react';
import { FormControl, Select, MenuItem, Typography, Box, CircularProgress, Button } from '@mui/material';
import type { SelectChangeEvent } from '@mui/material';
import { projectService } from '../services/projectService';
import type { Project } from '../services/projectService';
import { useNotification } from '@gofreego/tsutils';
import { useCurrentProject, useSetCurrentProject } from '../contexts/ProjectContext';

export function ProjectSelector() {
  const [projects, setProjects] = useState<Project[]>([]);
  const selectedProjectId = useCurrentProject();
  const setSelectedProjectId = useSetCurrentProject();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const notify = useNotification();

  // Refs so the effect doesn't need these as deps and avoids re-fetch loops
  const notifyRef = useRef(notify);
  notifyRef.current = notify;
  const selectedProjectIdRef = useRef(selectedProjectId);
  selectedProjectIdRef.current = selectedProjectId;

  useEffect(() => {
    let mounted = true;
    setLoading(true);
    setError(false);

    const loadProjects = async () => {
      try {
        const res = await projectService.list();
        if (mounted) {
          const list = res.results || [];
          setProjects(list);
          if (list.length > 0 && !selectedProjectIdRef.current) {
            setSelectedProjectId(list[0].id);
          }
        }
      } catch (err: unknown) {
        if (mounted) {
          setError(true);
          notifyRef.current.error((err as Error).message || 'Failed to load projects');
        }
      } finally {
        if (mounted) {
          setLoading(false);
        }
      }
    };
    loadProjects();
    return () => {
      mounted = false;
    };
  // retryCount is the only intentional trigger for re-fetch beyond mount
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [setSelectedProjectId, retryCount]);

  const handleChange = (event: SelectChangeEvent) => {
    setSelectedProjectId(event.target.value as string);
  };

  const handleRetry = useCallback(() => setRetryCount(c => c + 1), []);

  if (loading) {
    return (
      <Box sx={{ p: 2, display: 'flex', justifyContent: 'center' }}>
        <CircularProgress size={24} />
      </Box>
    );
  }

  if (error) {
    return (
      <Box sx={{ p: 2, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 1 }}>
        <Typography variant="body2" color="error" align="center">
          Failed to load projects
        </Typography>
        <Button size="small" variant="outlined" onClick={handleRetry}>
          Retry
        </Button>
      </Box>
    );
  }

  if (projects.length === 0) {
    return (
      <Box sx={{ p: 2 }}>
        <Typography variant="body2" color="textSecondary" align="center">
          No projects found
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 2, borderTop: '1px solid rgba(255, 255, 255, 0.12)' }}>
      <Typography variant="caption" color="textSecondary" sx={{ mb: 1, display: 'block' }}>
        CURRENT PROJECT
      </Typography>
      <FormControl fullWidth size="small">
        <Select
          value={selectedProjectId}
          onChange={handleChange}
          displayEmpty
          sx={{
            '& .MuiSelect-select': {
              py: 1,
            }
          }}
        >
          {projects.map((p) => (
            <MenuItem key={p.id} value={p.id}>
              {p.name}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
    </Box>
  );
}
