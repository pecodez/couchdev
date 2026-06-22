<template>
  <v-dialog :model-value="modelValue" max-width="520"
            @update:model-value="$emit('update:modelValue', $event)">
    <v-card>
      <v-card-title>Add Project</v-card-title>
      <v-card-text>
        <v-select
          v-model="form.source_type"
          label="Source"
          :items="sourceOptions"
          item-title="label"
          item-value="value"
        />

        <v-text-field
          v-if="form.source_type === 'clone'"
          v-model="form.repo_url"
          label="Repo URL"
          placeholder="git@github.com:org/repo.git"
          @update:model-value="onUrlChange"
        />

        <v-text-field
          v-model="form.name"
          label="Project name"
          :error-messages="nameError"
          @update:model-value="nameError = ''"
        />

        <div v-if="form.name && !nameError" class="text-caption text-medium-emphasis mb-2">
          📁 {{ projectPath }}
        </div>

        <v-alert v-if="error" type="error" density="compact">{{ error }}</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="cancel">Cancel</v-btn>
        <v-btn color="primary" :loading="loading" @click="submit">Add Project</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import { api } from '../api.js'
import { extractRepoName, isValidProjectName } from '../utils.js'

defineProps(['modelValue'])
const emit = defineEmits(['update:modelValue', 'created'])

const sourceOptions = [
  { label: 'Clone existing repo', value: 'clone' },
  { label: 'New local repo', value: 'greenfield' },
]

const form = ref({ source_type: 'clone', repo_url: '', name: '' })
const nameError = ref('')
const error = ref('')
const loading = ref(false)

const projectPath = computed(() => form.value.name ? `/projects/${form.value.name}` : '')

function onUrlChange(url) {
  const extracted = extractRepoName(url)
  if (extracted) form.value.name = extracted
}

function validateName() {
  if (!form.value.name) {
    nameError.value = 'Project name is required'
    return false
  }
  if (!isValidProjectName(form.value.name)) {
    nameError.value = 'Only letters, numbers, - and _ allowed'
    return false
  }
  return true
}

defineExpose({ form, nameError, error, loading })

async function submit() {
  if (!validateName()) return
  error.value = ''
  loading.value = true
  try {
    await api.createProject(form.value)
    emit('update:modelValue', false)
    emit('created')
    reset()
  } catch (e) { error.value = e.message }
  finally { loading.value = false }
}

function cancel() {
  emit('update:modelValue', false)
  reset()
}

function reset() {
  form.value = { source_type: 'clone', repo_url: '', name: '' }
  nameError.value = ''
  error.value = ''
}
</script>
