import { useState, useEffect } from 'react';
import { FormControl, Select, MenuItem, Typography, Box, CircularProgress } from '@mui/material';
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
  const notify = useNotification();

  useEffect(() => {
    let mounted = true;
    const loadProjects = async () => {
      try {
        const res = await projectService.list();
        if (mounted) {
          const list = res.results || [];
          setProjects(list);
          // Set first project as selected by default if we don't have one selected
          if (list.length > 0 && !selectedProjectId) {
            const defaultId = list[0].id;
            setSelectedProjectId(defaultId);
          }
        }
      } catch (err: unknown) {
        if (mounted) {
          notify.error((err as Error).message || 'Failed to load projects');
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
  }, [notify, selectedProjectId]);

  const handleChange = (event: SelectChangeEvent) => {
    const val = event.target.value as string;
    setSelectedProjectId(val);
  };



  if (loading) {
    return (
      <Box sx={{ p: 2, display: 'flex', justifyContent: 'center' }}>
        <CircularProgress size={24} />
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
