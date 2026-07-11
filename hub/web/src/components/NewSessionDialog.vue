<template>
  <v-dialog :model-value="modelValue" max-width="480"
            @update:model-value="$emit('update:modelValue', $event)">
    <v-card>
      <v-card-title>New Session — {{ project }}</v-card-title>
      <v-card-text>
        <v-text-field v-model="sessionName" label="Session name" hint="e.g. auth-refactor" />
        <v-text-field v-model="cwd" label="Working directory" hint="Defaults to project repo path" />
        <template v-if="result">
          <v-alert v-if="result.warnings && result.warnings.length" type="warning" density="compact" class="mb-2">
            <div><strong>{{ result.canonical_name }}</strong> started — but remote control may not be active.</div>
            <div v-for="w in result.warnings" :key="w" class="mt-1 text-caption">{{ w }}</div>
          </v-alert>
          <v-alert v-else type="success" density="compact">
            <strong>{{ result.canonical_name }}</strong> created ({{ result.state }}) — find it in the Claude app.
          </v-alert>
        </template>
        <v-alert v-if="error" type="error" density="compact">{{ error }}</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="close">{{ result ? 'Done' : 'Cancel' }}</v-btn>
        <v-btn v-if="!result" color="primary" :loading="loading" @click="submit">Launch</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { api } from '../api.js'

const props = defineProps(['modelValue', 'project'])
const emit = defineEmits(['update:modelValue', 'created'])

const sessionName = ref('')
const cwd = ref('')
const error = ref('')
const result = ref(null)
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    result.value = await api.createSession(props.project, sessionName.value, cwd.value)
    emit('created')
  } catch (e) { error.value = e.message }
  finally { loading.value = false }
}

function close() {
  emit('update:modelValue', false)
  sessionName.value = ''
  cwd.value = ''
  result.value = null
  error.value = ''
}
</script>
