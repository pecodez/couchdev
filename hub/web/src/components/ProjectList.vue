<template>
  <div>
    <v-row align="center" class="mb-3">
      <v-col><div class="text-h6">Projects</div></v-col>
      <v-col cols="auto">
        <v-btn icon="mdi-plus" variant="tonal" size="small" @click="dialog = true" />
      </v-col>
    </v-row>
    <v-alert v-if="error" type="error" density="compact" class="mb-3">{{ error }}</v-alert>
    <v-expansion-panels>
      <v-expansion-panel v-for="p in projects" :key="p.id" :title="p.name">
        <v-expansion-panel-text>
          <div class="text-caption text-medium-emphasis mb-2">{{ p.repo_path }}</div>
          <SessionList :project="p.name" />
        </v-expansion-panel-text>
      </v-expansion-panel>
    </v-expansion-panels>
    <NewProjectDialog v-model="dialog" @created="load" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../api.js'
import SessionList from './SessionList.vue'
import NewProjectDialog from './NewProjectDialog.vue'

const projects = ref([])
const error = ref('')
const dialog = ref(false)

async function load() {
  try { projects.value = await api.listProjects() }
  catch (e) { error.value = e.message }
}
onMounted(load)
</script>
