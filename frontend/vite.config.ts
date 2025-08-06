import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import postcssNesting from 'postcss-nesting'
import autoprefixer from 'autoprefixer'

// Custom plugin to handle &::before pseudo-elements
const fixPseudoElements = () => {
  return {
    postcssPlugin: 'fix-pseudo-elements',
    Rule(rule) {
      if (rule.selector.includes('&::')) {
        rule.selector = rule.selector.replace(/&::/g, '::')
      }
    }
  }
}
fixPseudoElements.postcss = true

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
        fixPseudoElements(),
        postcssNesting(),
        autoprefixer()
      ]
    }
  }
})