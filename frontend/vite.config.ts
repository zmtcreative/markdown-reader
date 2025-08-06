import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import postcssNested from 'postcss-nested'
import autoprefixer from 'autoprefixer'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  css: {
    preprocessorOptions: {
      scss: {
        api: 'legacy',
        silenceDeprecations: ['legacy-js-api']
      }
    },
    postcss: {
      plugins: [
        postcssNested({
          bubble: ['screen'],
          unwrap: ['supports']
        }),
        autoprefixer()
      ]
    }
  }
})