<template>
  <v-dialog :model-value="modelValue" max-width="480"
            @update:model-value="$emit('update:modelValue', $event)">
    <v-card>
      <v-card-title>Register Project</v-card-title>
      <v-card-text>
        <v-text-field v-model="form.name" label="Project name" />
        <v-text-field v-model="form.repo_path" label="Repo path on hub" />
        <v-alert v-if="error" type="error" density="compact">{{ error }}</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="$emit('update:modelValue', false)">Cancel</v-btn>
        <v-btn color="primary" :loading="loading" @click="submit">Register</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { api } from '../api.js'

defineProps(['modelValue'])
const emit = defineEmits(['update:modelValue', 'created'])

const form = ref({ name: '', repo_path: '' })
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await api.createProject(form.value.name, form.value.repo_path)
    emit('update:modelValue', false)
    emit('created')
    form.value = { name: '', repo_path: '' }
  } catch (e) { error.value = e.message }
  finally { loading.value = false }
}
</script>
