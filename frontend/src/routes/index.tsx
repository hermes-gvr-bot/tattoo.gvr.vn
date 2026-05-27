import { createFileRoute, redirect } from '@tanstack/react-router';
import { isAuthenticated } from '../lib/auth';

export const Route = createFileRoute('/')({
  loader: () => {
    if (isAuthenticated()) {
      throw redirect({ to: '/dashboard' });
    }
    throw redirect({ to: '/login' });
  },
});
