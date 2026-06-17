export type AuthenticatedUser = {
  sub: string;
  email: string;
  role: 'ADMIN' | 'USER';
};
