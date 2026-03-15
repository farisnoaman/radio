import { AuthProvider } from 'react-admin';
import { clearAuthStorage } from '../utils/storage';

export const portalAuthProvider: AuthProvider = {
  // 登录
  login: async ({ username, password }) => {
    const request = new Request('/api/v1/auth/portal/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
      headers: new Headers({ 'Content-Type': 'application/json' }),
    });

    try {
      const response = await fetch(request);
      let result;
      try {
        const text = await response.text();
        try {
          result = JSON.parse(text);
        } catch {
          if (response.status < 200 || response.status >= 300) {
            throw new Error(text || response.statusText || 'Server Error');
          }
          result = {};
        }
      } catch (e) {
        throw e;
      }

      if (response.status < 200 || response.status >= 300) {
        const errorMessage = result?.message || result?.error || response.statusText || 'Login Failed';
        throw new Error(errorMessage);
      }

      const auth = result.data || result;

      if (!auth.token) {
        throw new Error('Missing token in response');
      }

      localStorage.setItem('token', auth.token);
      localStorage.setItem('username', username);
      localStorage.setItem('permissions', JSON.stringify(auth.permissions || ['user']));

      if (auth.user) {
        localStorage.setItem('user', JSON.stringify(auth.user));
      }

      await new Promise(resolve => setTimeout(resolve, 0));
      return Promise.resolve();
    } catch (error) {
      console.error('Portal Login Error:', error);
      return Promise.reject(error);
    }
  },

  logout: async () => {
    clearAuthStorage();
    return Promise.resolve();
  },

  checkError: async (error) => {
    const status = error.status;
    if (status === 401) {
      clearAuthStorage();
      return Promise.reject({ message: 'Session expired, please login again' });
    }
    return Promise.resolve();
  },

  checkAuth: () => {
    const token = localStorage.getItem('token');
    if (!token || token.length < 10) {
      clearAuthStorage();
      return Promise.reject({ message: 'No token found', logoutUser: true });
    }
    return Promise.resolve();
  },

  getPermissions: async () => {
    const permissions = localStorage.getItem('permissions');
    return permissions ? Promise.resolve(JSON.parse(permissions)) : Promise.resolve(['user']);
  },

  getIdentity: async () => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      const user = JSON.parse(userStr);
      return Promise.resolve({
        id: user.id,
        fullName: user.realname || user.username,
        username: user.username,
        level: 'user',
        ...user,
      });
    }
    const username = localStorage.getItem('username');
    return Promise.resolve({
      id: username || 'anonymous',
      fullName: username || 'Anonymous',
      username: username || 'anonymous',
    });
  },
};
