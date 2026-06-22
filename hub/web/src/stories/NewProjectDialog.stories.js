import { ref, nextTick } from 'vue'
import NewProjectDialog from '../components/NewProjectDialog.vue'
import * as apiModule from '../api.js'

// Keep the dialog always open for stories and mock the api call.
function makeStory({ setupForm, mockCreate } = {}) {
  return {
    render: () => ({
      components: { NewProjectDialog },
      setup() {
        const open = ref(true)
        // Replace createProject on the exported api object so the component
        // picks it up without needing prop injection.
        apiModule.api.createProject = mockCreate ?? (() => new Promise(() => {}))
        return { open, setupForm }
      },
      mounted() {
        if (this.setupForm) this.setupForm(this.$refs.dialog)
      },
      template: `<NewProjectDialog ref="dialog" :model-value="open" @update:model-value="open = $event" />`,
    }),
  }
}

export default {
  title: 'Components/NewProjectDialog',
  parameters: {
    layout: 'centered',
  },
}

// 1. Default open state — source = clone, all fields empty.
export const CloneEmpty = {
  name: 'Clone — empty',
  ...makeStory(),
}

// 2. Clone with a URL pasted — name auto-populates.
export const CloneWithUrl = {
  name: 'Clone — URL entered, name auto-filled',
  render: () => ({
    components: { NewProjectDialog },
    setup() {
      const open = ref(true)
      apiModule.api.createProject = () => new Promise(() => {})
      return { open }
    },
    mounted() {
      // Drive the form to the filled state after mount.
      const dialog = this.$refs.dialog
      if (!dialog) return
      nextTick(() => {
        dialog.form.repo_url = 'git@github.com:org/my-project.git'
        dialog.form.name = 'my-project'
      })
    },
    template: `<NewProjectDialog ref="dialog" :model-value="open" @update:model-value="open = $event" />`,
  }),
}

// 3. New local repo — source switches, no Repo URL field shown.
export const GreenfieldEmpty = {
  name: 'New local repo — empty',
  render: () => ({
    components: { NewProjectDialog },
    setup() {
      const open = ref(true)
      apiModule.api.createProject = () => new Promise(() => {})
      return { open }
    },
    mounted() {
      nextTick(() => {
        if (this.$refs.dialog) this.$refs.dialog.form.source_type = 'greenfield'
      })
    },
    template: `<NewProjectDialog ref="dialog" :model-value="open" @update:model-value="open = $event" />`,
  }),
}

// 4. New local repo with a name typed.
export const GreenfieldNamed = {
  name: 'New local repo — name entered',
  render: () => ({
    components: { NewProjectDialog },
    setup() {
      const open = ref(true)
      apiModule.api.createProject = () => new Promise(() => {})
      return { open }
    },
    mounted() {
      nextTick(() => {
        const d = this.$refs.dialog
        if (!d) return
        d.form.source_type = 'greenfield'
        d.form.name = 'my-new-app'
      })
    },
    template: `<NewProjectDialog ref="dialog" :model-value="open" @update:model-value="open = $event" />`,
  }),
}

// 5. Invalid project name — shows inline error.
export const InvalidName = {
  name: 'Validation error — invalid name',
  render: () => ({
    components: { NewProjectDialog },
    setup() {
      const open = ref(true)
      apiModule.api.createProject = () => new Promise(() => {})
      return { open }
    },
    mounted() {
      nextTick(() => {
        const d = this.$refs.dialog
        if (!d) return
        d.form.source_type = 'greenfield'
        d.form.name = 'my project'
        d.nameError = 'Only letters, numbers, - and _ allowed'
      })
    },
    template: `<NewProjectDialog ref="dialog" :model-value="open" @update:model-value="open = $event" />`,
  }),
}

// 6. Submission in progress — button shows spinner.
export const Loading = {
  name: 'Loading — submit in progress',
  render: () => ({
    components: { NewProjectDialog },
    setup() {
      const open = ref(true)
      // Never resolves so the loading state persists.
      apiModule.api.createProject = () => new Promise(() => {})
      return { open }
    },
    mounted() {
      nextTick(() => {
        const d = this.$refs.dialog
        if (!d) return
        d.form.source_type = 'clone'
        d.form.repo_url = 'git@github.com:org/my-project.git'
        d.form.name = 'my-project'
        d.loading = true
      })
    },
    template: `<NewProjectDialog ref="dialog" :model-value="open" @update:model-value="open = $event" />`,
  }),
}

// 7. API error returned from server.
export const ApiError = {
  name: 'API error — project already exists',
  render: () => ({
    components: { NewProjectDialog },
    setup() {
      const open = ref(true)
      apiModule.api.createProject = () => new Promise(() => {})
      return { open }
    },
    mounted() {
      nextTick(() => {
        const d = this.$refs.dialog
        if (!d) return
        d.form.source_type = 'clone'
        d.form.repo_url = 'git@github.com:org/my-project.git'
        d.form.name = 'my-project'
        d.error = 'create project: UNIQUE constraint failed: projects.name'
      })
    },
    template: `<NewProjectDialog ref="dialog" :model-value="open" @update:model-value="open = $event" />`,
  }),
}
