import { config } from '@vue/test-utils'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

// jsdom stubs for browser APIs Vuetify depends on.
if (!window.visualViewport) {
  window.visualViewport = {
    width: 1280, height: 720,
    offsetLeft: 0, offsetTop: 0,
    pageLeft: 0, pageTop: 0,
    scale: 1,
    addEventListener: () => {},
    removeEventListener: () => {},
  }
}
if (!window.ResizeObserver) {
  window.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
}

// CSS skipped — behaviour tests don't need rendered styles.
const vuetify = createVuetify({ components, directives })
config.global.plugins = [vuetify]
