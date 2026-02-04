import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import type { ReactNode } from 'react';
import { colocationApi } from '../api';
import type { Colocation, ColocationWithMembers } from '../types';
import { useAuth } from './AuthContext';

interface ColocationContextType {
  colocations: Colocation[];
  currentColocation: ColocationWithMembers | null;
  isLoading: boolean;
  error: string | null;
  setCurrentColocation: (coloc: ColocationWithMembers | null) => void;
  selectColocation: (id: string) => Promise<void>;
  refreshColocations: () => Promise<void>;
  refreshCurrentColocation: () => Promise<void>;
}

const ColocationContext = createContext<ColocationContextType | undefined>(undefined);

const CURRENT_COLOC_KEY = 'current_colocation_id';

export function ColocationProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth();
  const [colocations, setColocations] = useState<Colocation[]>([]);
  const [currentColocation, setCurrentColocation] = useState<ColocationWithMembers | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refreshColocations = useCallback(async () => {
    if (!isAuthenticated) return;

    setIsLoading(true);
    setError(null);
    try {
      const data = await colocationApi.list();
      setColocations(data);

      // Auto-select saved colocation or first one
      const savedId = localStorage.getItem(CURRENT_COLOC_KEY);
      if (savedId && data.some((c) => c.id === savedId)) {
        const full = await colocationApi.get(savedId);
        setCurrentColocation(full);
      } else if (data.length > 0 && !currentColocation) {
        const full = await colocationApi.get(data[0].id);
        setCurrentColocation(full);
        localStorage.setItem(CURRENT_COLOC_KEY, data[0].id);
      }
    } catch (err) {
      setError('Erreur lors du chargement des colocations');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated, currentColocation]);

  const selectColocation = async (id: string) => {
    setIsLoading(true);
    try {
      const full = await colocationApi.get(id);
      setCurrentColocation(full);
      localStorage.setItem(CURRENT_COLOC_KEY, id);
    } catch (err) {
      setError('Erreur lors de la sélection de la colocation');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  const refreshCurrentColocation = async () => {
    if (!currentColocation) return;
    try {
      const full = await colocationApi.get(currentColocation.id);
      setCurrentColocation(full);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    if (isAuthenticated) {
      refreshColocations();
    } else {
      setColocations([]);
      setCurrentColocation(null);
    }
  }, [isAuthenticated]);

  return (
    <ColocationContext.Provider
      value={{
        colocations,
        currentColocation,
        isLoading,
        error,
        setCurrentColocation,
        selectColocation,
        refreshColocations,
        refreshCurrentColocation,
      }}
    >
      {children}
    </ColocationContext.Provider>
  );
}

export function useColocation() {
  const context = useContext(ColocationContext);
  if (context === undefined) {
    throw new Error('useColocation must be used within a ColocationProvider');
  }
  return context;
}
