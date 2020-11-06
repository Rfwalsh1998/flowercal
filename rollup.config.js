import typescript from 'rollup-plugin-typescript2'
import vue from 'rollup-plugin-vue'
import alias from '@rollup/plugin-alias'
import commonjs from '@rollup/plugin-commonjs'
import autoExternal from 'rollup-plugin-auto-external'
import buble from '@rollup/plugin-buble'
import { terser } from 'rollup-plugin-terser'

export default {
  input: 'src/wrapper.js', // Path relative to package.json
  output: {
    name: 'NekoClient',
    exports: 'named',
    globals: {
      "vue": "Vue",
      "vue-property-decorator": "VueClassComponent",
      "vue-class-component": "VueClassComponent",
      "axios": "axios",
    },
  },
  plugins: [
    typescript({
      check: false,
    }),
    vue({
      css: true,
      compileTemplate: true,
    }),
    alias({
      entries: [
        { find:/^@\/(.+)/, replacement: './$1' }
      ]
    }),
    commonjs({
      include: 'node_modules/**',
    }),
    autoExternal(),
    buble({
      objectAssign: 'Object.assign',
    }),
    terser(),
  ],
  external: [
    "vue", 
    "vue-class-component",
    "axios"
  ],
};