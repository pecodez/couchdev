import { setup } from '@storybook/vue3'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'

const vuetify = createVuetify({ components, directives, theme: { defaultTheme: 'dark' } })

setup((app) => {
  app.use(vuetify)
})

export const parameters = {
  backgrounds: {
    default: 'dark',
    values: [{ name: 'dark', value: '#121212' }],
  },
}
