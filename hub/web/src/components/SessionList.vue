<template>
  <div>
    <v-row align="center" class="mb-1">
      <v-col><span class="text-body-2 text-medium-emphasis">Sessions</span></v-col>
      <v-col cols="auto" style="min-width: 150px">
        <v-select v-model="filter" :items="filterOptions" item-title="label" item-value="value"
                  density="compact" variant="outlined" hide-details clearable
                  placeholder="Live + Resumable" />
      </v-col>
      <v-col cols="auto">
        <v-btn icon="mdi-plus" size="x-small" variant="text" @click="dialog = true" />
        <v-btn icon="mdi-refresh" size="x-small" variant="text" @click="load" />
      </v-col>
    </v-row>
    <v-expansion-panels variant="accordion" flat>
      <v-expansion-panel v-for="s in sessions" :key="s.canonical_name"
                         @group:selected="onExpand(s, $event)">
        <v-expansion-panel-title class="py-1 px-2">
          <div class="d-flex align-center gap-2 flex-wrap">
            <span class="text-body-2 font-weight-medium">{{ s.session }}</span>
            <v-chip :color="stateColor(s.state)" size="x-small">{{ s.state }}</v-chip>
            <v-chip v-if="s.branch" size="x-small" variant="outlined"
                    prepend-icon="mdi-source-branch">{{ s.branch }}</v-chip>
            <v-chip v-if="s.warnings && s.warnings.length" color="warning" size="x-small"
                    prepend-icon="mdi-alert-outline" :title="s.warnings.join('\n')">RC warning</v-chip>
          </div>
          <template #actions>
            <template v-if="s.state === 'resumable'">
              <v-btn icon="mdi-play-circle-outline" size="x-small" variant="text"
                     title="Resume" @click.stop="resume(s)" />
              <v-btn icon="mdi-delete-outline" size="x-small" variant="text"
                     title="Delete" @click.stop="requestDelete(s)" />
            </template>
            <v-btn v-else icon="mdi-stop-circle-outline" size="x-small" variant="text"
                   :disabled="s.state === 'starting'" @click.stop="requestDelete(s)" />
          </template>
        </v-expansion-panel-title>
        <v-expansion-panel-text>
          <div class="text-caption text-medium-emphasis mb-2">{{ s.cwd }}</div>
          <div v-if="changes[s.canonical_name] === undefined" class="text-caption text-medium-emphasis">
            Loading changes…
          </div>
          <div v-else-if="changes[s.canonical_name] === null" class="text-caption text-error">
            Could not load changes
          </div>
          <template v-else>
            <div class="text-caption mb-1">
              {{ changes[s.canonical_name].ahead }} commit{{ changes[s.canonical_name].ahead === 1 ? '' : 's' }} ahead
            </div>
            <div v-if="changes[s.canonical_name].changed_files.length === 0"
                 class="text-caption text-medium-emphasis">No uncommitted changes</div>
            <v-list v-else density="compact" class="pa-0">
              <v-list-item v-for="f in changes[s.canonical_name].changed_files" :key="f"
                           class="pa-0 text-caption font-weight-mono">
                <v-icon size="x-small" class="mr-1">mdi-file-edit-outline</v-icon>{{ f }}
              </v-list-item>
            </v-list>
          </template>
        </v-expansion-panel-text>
      </v-expansion-panel>
    </v-expansion-panels>
    <div v-if="sessions.length === 0" class="text-caption text-medium-emphasis px-2">No sessions</div>
    <NewSessionDialog v-model="dialog" :project="project" @created="load" />

    <v-dialog :model-value="!!confirmTarget" max-width="420"
              @update:model-value="!$event && closeConfirm()">
      <v-card v-if="confirmTarget">
        <v-card-title>Delete {{ confirmTarget.session }}?</v-card-title>
        <v-card-text>
          <template v-if="blockedReason">
            <v-alert type="warning" density="compact" class="mb-2">{{ blockedReason }}</v-alert>
            <div class="text-caption text-medium-emphasis">
              The worktree has been preserved so you can merge it later. Force delete will
              stop the session and discard this work permanently.
            </div>
          </template>
          <template v-else>
            This will stop the session and remove its worktree.
          </template>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="closeConfirm">Cancel</v-btn>
          <v-btn v-if="blockedReason" color="error" :loading="deleting"
                 @click="teardown(confirmTarget, true)">Force delete anyway</v-btn>
          <v-btn v-else color="error" :loading="deleting"
                 @click="teardown(confirmTarget, false)">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { api } from '../api.js'
import NewSessionDialog from './NewSessionDialog.vue'

const props = defineProps(['project'])
const sessions = ref([])
const dialog = ref(false)
const changes = ref({})
const filter = ref(null) // null (live+resumable, today's default) | 'live' | 'resumable' | 'all'
const confirmTarget = ref(null)
const blockedReason = ref('')
const deleting = ref(false)

const filterOptions = [
  { label: 'Live', value: 'live' },
  { label: 'Resumable', value: 'resumable' },
  { label: 'All', value: 'all' },
]

const stateColor = s => ({ live: 'success', starting: 'warning', resumable: 'info', dead: 'default' }[s])

async function load() {
  const all = await api.listSessions()
  sessions.value = all.filter(s => {
    if (!s.canonical_name.startsWith(props.project + '/')) return false
    if (filter.value === 'live') return s.state === 'live' || s.state === 'starting'
    if (filter.value === 'resumable') return s.state === 'resumable'
    if (filter.value === 'all') return true
    return s.state !== 'dead'
  })
}

function requestDelete(s) {
  confirmTarget.value = s
  blockedReason.value = ''
}

function closeConfirm() {
  confirmTarget.value = null
  blockedReason.value = ''
}

async function teardown(s, force = false) {
  deleting.value = true
  try {
    await api.deleteSession(props.project, s.session, force)
    closeConfirm()
    await load()
  } catch (e) {
    if (e.status === 409) {
      blockedReason.value = e.body?.reason || e.message
    } else {
      throw e
    }
  } finally {
    deleting.value = false
  }
}

async function resume(s) {
  await api.resumeSession(props.project, s.session)
  await load()
}

async function onExpand(s, { value }) {
  if (!value) return
  const key = s.canonical_name
  if (changes.value[key] !== undefined) return
  changes.value[key] = undefined // trigger "Loading…"
  try {
    changes.value[key] = await api.getChanges(props.project, s.session)
  } catch {
    changes.value[key] = null
  }
}

watch(() => props.project, () => { changes.value = {}; load() })
watch(filter, load)
onMounted(load)
</script>
