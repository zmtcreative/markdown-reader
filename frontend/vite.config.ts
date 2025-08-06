import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  css: {
    preprocessorOptions: {
      scss: {
        api: 'legacy', // Use legacy Sass API
        silenceDeprecations: ['legacy-js-api'],
        outputStyle: 'expanded'
      }
    },
    postcss: {
      plugins: [
        require('postcss-nesting')(),
        require('autoprefixer')()
      ]
    }
  }
})