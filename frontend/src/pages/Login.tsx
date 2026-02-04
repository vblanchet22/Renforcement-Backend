import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Mail, Lock, Home, ArrowRight } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { Button, Input } from '../components/ui';

export function Login() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      await login({ email, password });
      navigate('/dashboard');
    } catch (err) {
      setError('Email ou mot de passe incorrect');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50 grid lg:grid-cols-2">
      {/* Left — Branding */}
      <div className="hidden lg:flex relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-primary via-blue-500 to-accent" />
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -top-24 -left-24 w-[500px] h-[500px] rounded-full bg-white/10 blur-3xl" />
          <div className="absolute bottom-0 right-0 w-80 h-80 rounded-full bg-accent/30 blur-3xl" />
        </div>

        <div className="relative z-10 flex flex-col justify-between p-12 text-white">
          <div className="flex items-center gap-3">
            <div className="w-11 h-11 rounded-xl bg-white/20 backdrop-blur-sm flex items-center justify-center">
              <Home className="w-5 h-5" />
            </div>
            <span className="font-display text-2xl">ColocApp</span>
          </div>

          <div className="space-y-6">
            <motion.h1
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="font-display text-5xl leading-[1.1]"
            >
              Simplifiez la gestion
              <br />
              de votre colocation
            </motion.h1>
            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.3 }}
              className="text-lg text-white/80 max-w-md"
            >
              Partagez les dépenses, suivez les soldes et prenez des décisions ensemble.
            </motion.p>
          </div>

          <div className="flex items-center gap-6 text-sm text-white/50">
            <span>Dépenses partagées</span>
            <span className="w-1 h-1 rounded-full bg-white/30" />
            <span>Soldes en temps réel</span>
            <span className="w-1 h-1 rounded-full bg-white/30" />
            <span>Décisions collectives</span>
          </div>
        </div>
      </div>

      {/* Right — Form */}
      <div className="flex items-center justify-center bg-white px-6 py-12 sm:px-12">
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.4 }}
          className="w-full max-w-md mx-auto"
        >
          <div className="lg:hidden flex items-center gap-3 mb-8">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary to-accent flex items-center justify-center">
              <Home className="w-5 h-5 text-white" />
            </div>
            <span className="font-display text-xl text-slate-900">ColocApp</span>
          </div>

          <div className="mb-8">
            <h2 className="font-display text-3xl text-slate-900 mb-2">Bon retour !</h2>
            <p className="text-slate-500">Connectez-vous pour accéder à votre colocation</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <motion.div
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                className="p-3 rounded-xl bg-red-50 text-red-600 text-sm"
              >
                {error}
              </motion.div>
            )}

            <Input
              label="Email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="votre@email.com"
              leftIcon={<Mail className="w-5 h-5" />}
              required
            />

            <Input
              label="Mot de passe"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              leftIcon={<Lock className="w-5 h-5" />}
              required
            />

            <div className="flex items-center justify-between text-sm">
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" className="w-4 h-4 rounded border-slate-300 text-primary focus:ring-primary" />
                <span className="text-slate-500">Se souvenir de moi</span>
              </label>
              <a href="#" className="text-primary hover:underline font-medium">
                Mot de passe oublié ?
              </a>
            </div>

            <Button type="submit" isLoading={isLoading} className="w-full" size="lg">
              Se connecter
              <ArrowRight className="w-4 h-4" />
            </Button>
          </form>

          <p className="mt-8 text-center text-slate-500">
            Pas encore de compte ?{' '}
            <Link to="/register" className="text-primary font-medium hover:underline">
              Créer un compte
            </Link>
          </p>
        </motion.div>
      </div>
    </div>
  );
}
