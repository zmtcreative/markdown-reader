import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import postcssNested from 'postcss-nested'
import autoprefixer from 'autoprefixer'

// More targeted pseudo-element processor
const processPseudoElements = () => {
  return {
    postcssPlugin: 'process-pseudo-elements',
    Once(root) {
      root.walkRules(rule => {
        if (rule.selector.includes('&::')) {
          // Split selectors and process each one
          const selectors = rule.selector.split(',').map(sel => {
            return sel.trim().replace(/([^&]+)&::/g, '$1::')
          })
          rule.selector = selectors.join(', ')
        }
      })
    }
  }
}
processPseudoElements.postcss = true

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
        processPseudoElements(),
        postcssNested({
          bubble: ['screen'],
          unwrap: ['supports']
        }),
        autoprefixer()
      ]
    }
  }
})