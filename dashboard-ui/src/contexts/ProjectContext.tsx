import { createContext, useContext, useState } from 'react';
import type { ReactNode } from 'react';

interface ProjectContextType {
  projectId: string;
  setProjectId: (id: string) => void;
}

const ProjectContext = createContext<ProjectContextType | undefined>(undefined);

export function ProjectProvider({ children }: { children: ReactNode }) {
  const [projectId, setProjectId] = useState<string>('');

  return (
    <ProjectContext.Provider value={{ projectId, setProjectId }}>
      {children}
    </ProjectContext.Provider>
  );
}

export function useCurrentProject() {
  const context = useContext(ProjectContext);
  if (context === undefined) {
    throw new Error('useCurrentProject must be used within a ProjectProvider');
  }
  return context.projectId;
}

export function useSetCurrentProject() {
  const context = useContext(ProjectContext);
  if (context === undefined) {
    throw new Error('useSetCurrentProject must be used within a ProjectProvider');
  }
  return context.setProjectId;
}
