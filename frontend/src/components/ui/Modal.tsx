import type { ReactNode } from 'react';
import { useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X } from 'lucide-react';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: ReactNode;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  showCloseButton?: boolean;
}

const sizeClasses = { sm: 'max-w-sm', md: 'max-w-md', lg: 'max-w-lg', xl: 'max-w-xl' };

export function Modal({ isOpen, onClose, title, children, size = 'md', showCloseButton = true }: ModalProps) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    if (isOpen) {
      document.addEventListener('keydown', onKey);
      document.body.style.overflow = 'hidden';
    }
    return () => {
      document.removeEventListener('keydown', onKey);
      document.body.style.overflow = '';
    };
  }, [isOpen, onClose]);

  return (
    <AnimatePresence>
      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.15 }}
            className="absolute inset-0 bg-slate-900/40 backdrop-blur-sm"
            onClick={onClose}
          />
          <motion.div
            initial={{ opacity: 0, scale: 0.96, y: 8 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.96, y: 8 }}
            transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
            className={`relative w-full ${sizeClasses[size]} bg-white rounded-2xl shadow-xl max-h-[85vh] flex flex-col overflow-hidden`}
          >
            {(title || showCloseButton) && (
              <div className="flex items-center justify-between px-6 pt-6 pb-2">
                {title && <h2 className="font-display text-xl text-slate-900">{title}</h2>}
                {showCloseButton && (
                  <button onClick={onClose} className="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors">
                    <X className="w-4 h-4" />
                  </button>
                )}
              </div>
            )}
            <div className="px-6 pb-6 pt-2 overflow-y-auto">{children}</div>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  );
}
