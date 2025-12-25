<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1 class="logo">CodeEcho</h1>
      <p class="tagline">Whisper code. Echo clarity.</p>

      <div class="tabs">
        <button :class="['tab', { active: mode === 'login' }]" @click="mode = 'login'">Login</button>
        <button :class="['tab', { active: mode === 'register' }]" @click="mode = 'register'">Register</button>
      </div>

      <form @submit.prevent="handleSubmit" class="auth-form">
        <div class="form-group">
          <label>Username</label>
          <input
            type="text"
            v-model="username"
            placeholder="Enter username"
            required
            minlength="3"
          />
        </div>

        <div class="form-group">
          <label>Password</label>
          <input
            type="password"
            v-model="password"
            placeholder="Enter password"
            required
            minlength="6"
          />
        </div>

        <div class="form-group" v-if="mode === 'register'">
          <label>Confirm Password</label>
          <input
            type="password"
            v-model="confirmPassword"
            placeholder="Confirm password"
            required
          />
        </div>

        <p class="error" v-if="error">{{ error }}</p>

        <button type="submit" class="submit-btn" :disabled="isLoading">
          {{ isLoading ? 'Loading...' : (mode === 'login' ? 'Login' : 'Register') }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { authApi } from '../api/auth'

const router = useRouter()
const userStore = useUserStore()

const mode = ref<'login' | 'register'>('login')
const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const isLoading = ref(false)

const handleSubmit = async () => {
  error.value = ''

  if (mode.value === 'register' && password.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }

  isLoading.value = true

  try {
    if (mode.value === 'register') {
      await authApi.register(username.value, password.value)
      // Auto-login after registration
      mode.value = 'login'
    }

    const { token, user } = await authApi.login(username.value, password.value)
    userStore.setToken(token)
    userStore.setUser(user)
    router.push('/')
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Something went wrong'
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  padding: 20px;
}

.auth-card {
  width: 100%;
  max-width: 400px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 40px;
  text-align: center;
}

.logo {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.tagline {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 8px 0 32px;
}

.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
}

.tab {
  flex: 1;
  padding: 10px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: transparent;
  color: var(--text-secondary);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.tab:hover {
  background: var(--bg-hover);
}

.tab.active {
  background: var(--accent-color);
  border-color: var(--accent-color);
  color: white;
}

.auth-form {
  text-align: left;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.form-group input {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 0.9375rem;
  outline: none;
  transition: border-color 0.2s;
}

.form-group input:focus {
  border-color: var(--accent-color);
}

.error {
  color: var(--error-color);
  font-size: 0.875rem;
  margin-bottom: 16px;
}

.submit-btn {
  width: 100%;
  padding: 14px;
  background: var(--accent-color);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}

.submit-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
