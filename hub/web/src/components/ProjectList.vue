<template>
  <div>
    <v-row align="center" class="mb-4">
      <v-col>
        <div class="text-h6" style="color:#e8e8e8;font-weight:600;">Projects</div>
      </v-col>
      <v-col cols="auto">
        <v-btn prepend-icon="mdi-plus" variant="tonal" size="small" @click="dialog = true">
          New project
        </v-btn>
      </v-col>
    </v-row>

    <v-alert v-if="error" type="error" density="compact" class="mb-4">{{ error }}</v-alert>

    <div v-if="projects.length">
      <v-card v-for="p in projects" :key="p.id" class="project-card mb-3" :style="cardStyle(p)">

          <!-- Header row: name + registry badge -->
          <v-card-title class="pt-4 pb-1 d-flex align-center justify-space-between flex-wrap gap-2">
            <span class="project-name text-truncate">{{ p.name }}</span>
            <v-chip v-if="p.registry && p.registry !== 'custom'"
                    size="x-small" variant="outlined"
                    :prepend-icon="registryIcon(p.registry)"
                    style="opacity:.75;flex-shrink:0;">
              {{ p.registry }}
            </v-chip>
            <v-icon v-if="p.source_missing" size="16" color="warning" title="Source directory missing">
              mdi-alert-outline
            </v-icon>
          </v-card-title>

          <!-- Description -->
          <v-card-text class="pb-1 pt-0">
            <p class="description text-body-2" :class="p.description ? '' : 'text-medium-emphasis'">
              {{ p.description || (p.repo_url || p.repo_path) }}
            </p>
          </v-card-text>

          <!-- Language chips -->
          <div v-if="p.languages && p.languages.length" class="px-4 pb-3 d-flex flex-wrap gap-3">
            <span v-for="lang in p.languages" :key="lang"
                  class="lang-chip"
                  :style="langChipStyle(lang)">
              <span class="lang-dot" :style="{ background: langColor(lang) }"></span>
              {{ lang }}
            </span>
          </div>

          <v-divider style="opacity:.15;" />

          <!-- Sessions section -->
          <v-expansion-panels variant="accordion" flat>
            <v-expansion-panel>
              <v-expansion-panel-title class="px-4 py-2 sessions-toggle">
                <v-icon size="16" class="mr-2" style="color:#888;">mdi-console</v-icon>
                <span class="text-body-2" style="color:#aaa;">Sessions</span>
                <template #actions="{ expanded }">
                  <v-icon size="16" style="color:#666;">
                    {{ expanded ? 'mdi-chevron-up' : 'mdi-chevron-down' }}
                  </v-icon>
                </template>
              </v-expansion-panel-title>
              <v-expansion-panel-text class="pa-0">
                <SessionList :project="p.name" />
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>

        </v-card>
    </div>

    <div v-else-if="!error" class="text-medium-emphasis text-body-2 mt-4">
      No projects yet. Create one to get started.
    </div>

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

const LANG_COLORS = {
  'Go':              '#00ADD8',
  'JavaScript':      '#F1E05A',
  'TypeScript':      '#3178C6',
  'Vue':             '#4FC08D',
  'Python':          '#3572A5',
  'Rust':            '#DEA584',
  'Java':            '#B07219',
  'Kotlin':          '#A97BFF',
  'Shell':           '#89E051',
  'Bash':            '#89E051',
  'HTML':            '#E34C26',
  'CSS':             '#563D7C',
  'SCSS':            '#C6538C',
  'Ruby':            '#CC342D',
  'C':               '#555555',
  'C++':             '#F34B7D',
  'C#':              '#178600',
  'Swift':           '#F05138',
  'Dockerfile':      '#384D54',
  'YAML':            '#CB171E',
  'Markdown':        '#083FA1',
}

function langColor(lang) {
  return LANG_COLORS[lang] || '#8B8B8B'
}

function langChipStyle(lang) {
  const color = langColor(lang)
  return {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '4px',
    padding: '2px 7px',
    borderRadius: '10px',
    fontSize: '11px',
    fontWeight: '500',
    color: '#ccc',
    background: 'rgba(255,255,255,0.06)',
    border: '1px solid rgba(255,255,255,0.1)',
  }
}

function registryIcon(registry) {
  return registry === 'github' ? 'mdi-github' : registry === 'gitlab' ? 'mdi-gitlab' : 'mdi-git'
}

function cardStyle(p) {
  return {
    background: p.source_missing ? 'rgba(255,200,50,0.04)' : 'rgba(255,255,255,0.04)',
    border: '1px solid rgba(255,255,255,0.08)',
    transition: 'box-shadow .2s ease, border-color .2s ease',
  }
}
</script>

<style scoped>
.project-card:hover {
  border-color: rgba(255,255,255,0.18) !important;
  box-shadow: 0 4px 24px rgba(0,0,0,0.4) !important;
}
.project-name {
  font-size: 1rem;
  font-weight: 600;
  color: #e8e8e8;
}
.description {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.5;
  color: #aaa;
  min-height: 2.8em;
}
.lang-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.sessions-toggle {
  min-height: 40px !important;
}
</style>
