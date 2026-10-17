import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { login } from '../services/authService';
import './LoginPage.css';

export default function LoginPage({ onLogin, initialError = '' }) {
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState(initialError);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      await login(username.trim(), password);
      onLogin();
      navigate('/create', { replace: true });
    } catch (requestError) {
      setError(requestError.message || 'Unable to sign in. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <main className="login-page">
      <section className="login-card" aria-labelledby="login-title">
        <div className="login-brand-mark" aria-hidden="true">D</div>
        <p className="login-eyebrow">DENTIST CLINIC</p>
        <h1 id="login-title">Welcome back</h1>
        <p className="login-description">Sign in to continue to your clinic workspace.</p>

        <form className="login-form" onSubmit={handleSubmit}>
          <div className="login-field">
            <label htmlFor="login-username">Username</label>
            <input
              id="login-username"
              name="username"
              type="text"
              autoComplete="username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              placeholder="Enter your username"
              required
              autoFocus
            />
          </div>

          <div className="login-field">
            <label htmlFor="login-password">Password</label>
            <div className="login-password-input">
              <input
                id="login-password"
                name="password"
                type={showPassword ? 'text' : 'password'}
                autoComplete="current-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="Enter your password"
                required
              />
              <button
                className="password-visibility"
                type="button"
                onClick={() => setShowPassword((visible) => !visible)}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? 'Hide' : 'Show'}
              </button>
            </div>
          </div>

          {error && <p className="login-error" role="alert">{error}</p>}

          <button className="login-submit" type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Signing in...' : 'Sign in'}
          </button>
        </form>

        <p className="login-security-note">Your sign-in is protected with a secure session.</p>
      </section>
    </main>
  );
}
