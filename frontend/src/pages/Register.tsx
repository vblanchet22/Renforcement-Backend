import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Mail, Lock, User, Phone, Home, ArrowRight, Check } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { Button, Input } from '../components/ui';

const features = [
  'Partagez les dépenses équitablement',
  'Suivez les soldes en temps réel',
  'Simplifiez les remboursements',
  'Prenez des décisions ensemble',
];

export function Register() {
  const navigate = useNavigate();
  const { register } = useAuth();
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    confirmPassword: '',
    nom: '',
    prenom: '',
    telephone: '',
  });
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (formData.password !== formData.confirmPassword) {
      setError('Les mots de passe ne correspondent pas');
      return;
    }

    if (formData.password.length < 8) {
      setError('Le mot de passe doit contenir au moins 8 caractères');
      return;
    }

    setIsLoading(true);

    try {
      await register({
        email: formData.email,
        password: formData.password,
        nom: formData.nom,
        prenom: formData.prenom,
        telephone: formData.telephone || undefined,
      });
      navigate('/dashboard');
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else if (typeof err === 'object' && err !== null && 'response' in err) {
        const axiosErr = err as { response?: { data?: { message?: string } } };
        setError(axiosErr.response?.data?.message || "Erreur lors de l'inscription");
      } else {
        setError("Erreur lors de l'inscription. Cet email est peut-être déjà utilisé.");
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50 grid lg:grid-cols-2">
      {/* Left — Form */}
      <div className="flex flex-col justify-center bg-white px-6 py-12 sm:px-12">
        <motion.div
          initial={{ opacity: 0, x: -20 }}
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
            <h2 className="font-display text-3xl text-slate-900 mb-2">Créer un compte</h2>
            <p className="text-slate-500">Rejoignez ColocApp et simplifiez votre quotidien</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <motion.div
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                className="p-3 rounded-xl bg-red-50 text-red-600 text-sm"
              >
                {error}
              </motion.div>
            )}

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <Input label="Prénom" name="prenom" value={formData.prenom} onChange={handleChange} placeholder="Jean" leftIcon={<User className="w-5 h-5" />} required />
              <Input label="Nom" name="nom" value={formData.nom} onChange={handleChange} placeholder="Dupont" leftIcon={<User className="w-5 h-5" />} required />
            </div>

            <Input label="Email" type="email" name="email" value={formData.email} onChange={handleChange} placeholder="votre@email.com" leftIcon={<Mail className="w-5 h-5" />} required />
            <Input label="Téléphone (optionnel)" type="tel" name="telephone" value={formData.telephone} onChange={handleChange} placeholder="06 12 34 56 78" leftIcon={<Phone className="w-5 h-5" />} />
            <Input label="Mot de passe" type="password" name="password" value={formData.password} onChange={handleChange} placeholder="••••••••" leftIcon={<Lock className="w-5 h-5" />} hint="Minimum 8 caractères" required />
            <Input label="Confirmer le mot de passe" type="password" name="confirmPassword" value={formData.confirmPassword} onChange={handleChange} placeholder="••••••••" leftIcon={<Lock className="w-5 h-5" />} required />

            <Button type="submit" isLoading={isLoading} className="w-full" size="lg">
              Créer mon compte
              <ArrowRight className="w-4 h-4" />
            </Button>
          </form>

          <p className="mt-6 text-center text-slate-500">
            Déjà un compte ?{' '}
            <Link to="/login" className="text-primary font-medium hover:underline">Se connecter</Link>
          </p>
        </motion.div>
      </div>

      {/* Right — Branding */}
      <div className="hidden lg:flex relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-bl from-accent via-orange-400 to-primary" />
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -bottom-24 -right-24 w-[500px] h-[500px] rounded-full bg-white/10 blur-3xl" />
          <div className="absolute top-0 left-0 w-80 h-80 rounded-full bg-primary/20 blur-3xl" />
        </div>

        <div className="relative z-10 flex flex-col justify-between p-12 text-white">
          <div className="flex items-center gap-3">
            <div className="w-11 h-11 rounded-xl bg-white/20 backdrop-blur-sm flex items-center justify-center">
              <Home className="w-5 h-5" />
            </div>
            <span className="font-display text-2xl">ColocApp</span>
          </div>

          <div className="space-y-8">
            <motion.h1 initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} className="font-display text-4xl leading-tight">
              Tout ce dont vous avez
              <br />besoin pour gérer
              <br />votre colocation
            </motion.h1>

            <div className="space-y-4">
              {features.map((feature, index) => (
                <motion.div key={feature} initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} transition={{ delay: 0.4 + index * 0.1 }} className="flex items-center gap-3">
                  <div className="w-6 h-6 rounded-full bg-white/20 flex items-center justify-center">
                    <Check className="w-4 h-4" />
                  </div>
                  <span className="text-white/90">{feature}</span>
                </motion.div>
              ))}
            </div>
          </div>

          <div className="text-sm text-white/50">
            Déjà plus de 1000 colocations gérées avec ColocApp
          </div>
        </div>
      </div>
    </div>
  );
}
