import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import postcssNesting from 'postcss-nesting'
import autoprefixer from 'autoprefixer'



// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'jsdom',
    globals: false,
    setupFiles: './vitest.setup.ts',
    css: true,
    restoreMocks: true,
    clearMocks: true,
  },
  css: {
    preprocessorOptions: {
      scss: {
        api: 'modern',
        silenceDeprecations: ['legacy-js-api']
      }
    },
    postcss: {
      plugins: [
        // Use postcss-nesting to handle all nesting, including pseudo-elements.
        // It correctly transforms `&::before` and other nested syntax.
        postcssNesting(),
        autoprefixer()
      ]
    }
  }
})