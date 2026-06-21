import { useState, useEffect } from 'react';

export function useCurrentProject() {
  const [projectId, setProjectId] = useState<string>(() => localStorage.getItem('selectedProjectId') || '');

  useEffect(() => {
    const handleProjectChange = () => {
      const current = localStorage.getItem('selectedProjectId') || '';
      if (current !== projectId) {
        setProjectId(current);
      }
    };

    window.addEventListener('projectChanged', handleProjectChange);
    // Also listen to storage events for cross-tab synchronization
    window.addEventListener('storage', handleProjectChange);

    // Initial check just in case it changed between initial render and effect
    handleProjectChange();

    return () => {
      window.removeEventListener('projectChanged', handleProjectChange);
      window.removeEventListener('storage', handleProjectChange);
    };
  }, [projectId]);

  return projectId;
}
