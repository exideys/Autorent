import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'

const normalizeBase = (value) => {
  if (!value || value === '/') {
    return '/'
  }

  const withLeadingSlash = value.startsWith('/') ? value : `/${value}`
  return withLeadingSlash.endsWith('/') ? withLeadingSlash : `${withLeadingSlash}/`
}

const repositoryName = process.env.GITHUB_REPOSITORY?.split('/')[1]
const pagesBase =
  repositoryName && repositoryName.endsWith('.github.io') ? '/' : normalizeBase(repositoryName || 'Autorent')
const base = normalizeBase(process.env.VITE_BASE_PATH || (process.env.GITHUB_PAGES === 'true' ? pagesBase : '/'))

// https://vite.dev/config/
export default defineConfig({
  base,
  plugins: [tailwindcss(), react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  },
})
