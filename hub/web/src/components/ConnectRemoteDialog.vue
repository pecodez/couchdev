<template>
  <v-dialog :model-value="modelValue" max-width="520"
            @update:model-value="$emit('update:modelValue', $event)">
    <v-card>
      <v-card-title>Connect Remote</v-card-title>
      <v-card-text>
        <p class="text-body-2 text-medium-emphasis mb-4">
          Attach an existing remote repo to <strong>{{ project?.name }}</strong> and push its
          default branch to it.
        </p>

        <v-text-field
          v-model="repoUrl"
          label="Repo URL"
          placeholder="git@github.com:org/repo.git"
          :error-messages="urlError"
          @update:model-value="urlError = ''"
        />

        <v-alert v-if="error" type="error" density="compact">{{ error }}</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="cancel">Cancel</v-btn>
        <v-btn color="primary" :loading="loading" @click="submit">Connect</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { api } from '../api.js'

const props = defineProps(['modelValue', 'project'])
const emit = defineEmits(['update:modelValue', 'connected'])

const repoUrl = ref('')
const urlError = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  if (!repoUrl.value) {
    urlError.value = 'Repo URL is required'
    return
  }
  error.value = ''
  loading.value = true
  try {
    const result = await api.connectRemote(props.project.name, repoUrl.value)
    emit('update:modelValue', false)
    emit('connected', result.warning || '')
    reset()
  } catch (e) { error.value = e.message }
  finally { loading.value = false }
}

function cancel() {
  emit('update:modelValue', false)
  reset()
}

function reset() {
  repoUrl.value = ''
  urlError.value = ''
  error.value = ''
}
</script>
