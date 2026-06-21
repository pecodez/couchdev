<template>
  <div>
    <v-row align="center" class="mb-1">
      <v-col><span class="text-body-2 text-medium-emphasis">Sessions</span></v-col>
      <v-col cols="auto">
        <v-btn icon="mdi-plus" size="x-small" variant="text" @click="dialog = true" />
        <v-btn icon="mdi-refresh" size="x-small" variant="text" @click="load" />
      </v-col>
    </v-row>
    <v-list density="compact">
      <v-list-item v-for="s in sessions" :key="s.canonical_name">
        <template #title>
          <span class="text-body-2 font-weight-medium">{{ s.session }}</span>
          <v-chip :color="stateColor(s.state)" size="x-small" class="ml-2">{{ s.state }}</v-chip>
        </template>
        <template #subtitle>{{ s.cwd }}</template>
        <template #append>
          <v-btn icon="mdi-stop-circle-outline" size="x-small" variant="text"
                 :disabled="s.state === 'dead'" @click="teardown(s)" />
        </template>
      </v-list-item>
      <v-list-item v-if="sessions.length === 0">
        <v-list-item-subtitle>No sessions</v-list-item-subtitle>
      </v-list-item>
    </v-list>
    <NewSessionDialog v-model="dialog" :project="project" @created="load" />
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { api } from '../api.js'
import NewSessionDialog from './NewSessionDialog.vue'

const props = defineProps(['project'])
const sessions = ref([])
const dialog = ref(false)
const stateColor = s => ({ live: 'success', starting: 'warning', resumable: 'info', dead: 'default' }[s])

async function load() {
  const all = await api.listSessions()
  sessions.value = all.filter(s => s.canonical_name.startsWith(props.project + '/'))
}
async function teardown(s) {
  await api.deleteSession(props.project, s.session)
  await load()
}
watch(() => props.project, load)
onMounted(load)
</script>
