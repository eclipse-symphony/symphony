'use client';
import React, { createContext, useContext, useState, ReactNode } from 'react';
import { SymphonyObject } from '../app/types';

// Define the type for the global state
interface GlobalState {
  objects: SymphonyObject[];
  setObjects: React.Dispatch<React.SetStateAction<SymphonyObject[]>>;
}

// Create a context with a default undefined value
const GlobalStateContext = createContext<GlobalState | undefined>(undefined);

// Define the provider props
interface GlobalStateProviderProps {
  children: ReactNode;
}

// Create a provider component
export const GlobalStateProvider: React.FC<GlobalStateProviderProps> = ({ children }) => {
  const [objects, setObjects] = useState<SymphonyObject[]>([]); // State for storing Symphony objects

  return (
    <GlobalStateContext.Provider value={{ objects, setObjects }}>
      {children}
    </GlobalStateContext.Provider>
  );
};

// Custom hook to use the global state
export const useGlobalState = () => {
  const context = useContext(GlobalStateContext);
  if (!context) {
    throw new Error('useGlobalState must be used within a GlobalStateProvider');
  }
  return context;
};
