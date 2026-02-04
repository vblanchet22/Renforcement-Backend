import type { ReactNode } from 'react';
import { Sidebar } from './Sidebar';

interface AppLayoutProps {
  children: ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  return (
    <div className="min-h-screen bg-slate-100 flex">
      <Sidebar />
      <main className="flex-1 min-h-screen overflow-auto">
        <div className="p-8 md:p-12 lg:p-16 xl:p-20">
          <div className="max-w-6xl">
            {children}
          </div>
        </div>
      </main>
    </div>
  );
}
