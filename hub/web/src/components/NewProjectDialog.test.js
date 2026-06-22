import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import NewProjectDialog from './NewProjectDialog.vue'
import * as apiModule from '../api.js'

function mountDialog() {
  return mount(NewProjectDialog, {
    props: { modelValue: true },
    attachTo: document.body,
  })
}

beforeEach(() => {
  vi.restoreAllMocks()
})

// v-dialog teleports its content to document.body, so DOM queries must use body.
const bodyQuery = (sel) => document.body.querySelector(sel)

describe('NewProjectDialog source type', () => {
  it('shows Repo URL field when source is clone', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    expect(bodyQuery('[placeholder="git@github.com:org/repo.git"]')).not.toBeNull()
    wrapper.unmount()
  })

  it('hides Repo URL field when source is greenfield', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    wrapper.vm.form.source_type = 'greenfield'
    await flushPromises()
    expect(bodyQuery('[placeholder="git@github.com:org/repo.git"]')).toBeNull()
    wrapper.unmount()
  })
})

describe('NewProjectDialog URL auto-fill', () => {
  it('populates project name from a valid SSH URL', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    wrapper.vm.form.repo_url = 'git@github.com:org/my-project.git'
    wrapper.vm.onUrlChange('git@github.com:org/my-project.git')
    await flushPromises()
    expect(wrapper.vm.form.name).toBe('my-project')
    wrapper.unmount()
  })

  it('populates project name from a valid HTTPS URL', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    wrapper.vm.onUrlChange('https://github.com/org/another-repo.git')
    await flushPromises()
    expect(wrapper.vm.form.name).toBe('another-repo')
    wrapper.unmount()
  })
})

describe('NewProjectDialog path preview', () => {
  it('shows path preview when name is valid', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    wrapper.vm.form.source_type = 'greenfield'
    wrapper.vm.form.name = 'my-app'
    await flushPromises()
    expect(document.body.textContent).toContain('/projects/my-app')
    wrapper.unmount()
  })

  it('hides path preview when name is invalid', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    wrapper.vm.form.source_type = 'greenfield'
    wrapper.vm.form.name = 'my app'
    wrapper.vm.nameError = 'Only letters, numbers, - and _ allowed'
    await flushPromises()
    expect(document.body.textContent).not.toContain('/projects/')
    wrapper.unmount()
  })
})

describe('NewProjectDialog cancel', () => {
  it('resets form to initial state', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    wrapper.vm.form.name = 'something'
    wrapper.vm.cancel()
    await flushPromises()
    expect(wrapper.vm.form).toEqual({ source_type: 'clone', repo_url: '', name: '' })
    wrapper.unmount()
  })
})

describe('NewProjectDialog submit', () => {
  it('calls api.createProject with form payload', async () => {
    const mock = vi.spyOn(apiModule.api, 'createProject').mockResolvedValue({ id: 1 })
    const wrapper = mountDialog()
    await flushPromises()
    wrapper.vm.form.source_type = 'greenfield'
    wrapper.vm.form.name = 'my-app'
    await wrapper.vm.submit()
    expect(mock).toHaveBeenCalledWith({ source_type: 'greenfield', repo_url: '', name: 'my-app' })
    wrapper.unmount()
  })

  it('blocks submit and shows error for invalid name', async () => {
    const mock = vi.spyOn(apiModule.api, 'createProject').mockResolvedValue({})
    const wrapper = mountDialog()
    await flushPromises()
    wrapper.vm.form.name = 'bad name'
    await wrapper.vm.submit()
    expect(mock).not.toHaveBeenCalled()
    expect(wrapper.vm.nameError).toBe('Only letters, numbers, - and _ allowed')
    wrapper.unmount()
  })

  it('surfaces API error on rejection', async () => {
    vi.spyOn(apiModule.api, 'createProject').mockRejectedValue(new Error('conflict'))
    const wrapper = mountDialog()
    await flushPromises()
    wrapper.vm.form.name = 'my-app'
    await wrapper.vm.submit()
    await flushPromises()
    expect(wrapper.vm.error).toBe('conflict')
    wrapper.unmount()
  })
})
