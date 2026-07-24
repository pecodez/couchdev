<template>
  <v-app>
    <v-app-bar flat height="56" style="border-bottom:1px solid rgba(255,255,255,0.08);">
      <!-- Logo + wordmark -->
      <div class="d-flex align-center ml-3 gap-3">
        <CouchdevLogo />
        <span style="font-family:monospace;font-size:1.1rem;font-weight:700;letter-spacing:.04em;color:#e8e8e8;">
          couchdev
        </span>
      </div>

      <v-spacer />

      <!-- Powered by Claude badge -->
      <PoweredByClaude class="mr-3" />

      <!-- Token button -->
      <v-btn icon size="small" class="mr-1" @click="openTokenDialog">
        <v-icon size="18">mdi-key-variant</v-icon>
      </v-btn>
    </v-app-bar>

    <v-main>
      <v-container fluid class="pa-4">
        <div style="max-width:860px;margin:0 auto;width:100%;">
          <ProjectList v-if="authed" />
          <v-row v-else justify="center">
            <v-col cols="12" sm="8" md="5" lg="4">
              <v-card class="mt-8">
                <v-card-title class="pt-5">Connect to hub</v-card-title>
                <v-card-text>
                  <v-text-field v-model="inputToken" label="Bearer token" type="password"
                                variant="outlined" density="comfortable"
                                :error-messages="tokenError" @keyup.enter="saveToken" />
                </v-card-text>
                <v-card-actions class="px-4 pb-4">
                  <v-btn color="primary" variant="flat" block :loading="verifying" @click="saveToken">
                    Connect
                  </v-btn>
                </v-card-actions>
              </v-card>
            </v-col>
          </v-row>
        </div>
      </v-container>
    </v-main>

    <v-footer app style="background:transparent;border-top:1px solid rgba(255,255,255,0.06);padding:16px 0 14px;">
      <div style="width:100%;text-align:center;font-family:sans-serif;line-height:1.7;">
        <div style="font-size:0.75rem;color:#888;"><strong>couchdev</strong> — remote control hub for Claude Code</div>
        <div style="font-size:0.75rem;color:#888;font-style:italic;">build from the comfort of your couch</div>
        <div style="font-size:0.65rem;color:#666;">
          {{ version ? version + ' - ' : '' }}brought to you by
          <a href="https://github.com/pecodez" target="_blank" rel="noopener"
             style="color:#666;text-decoration:underline;">pecodez</a>
        </div>
      </div>
    </v-footer>

    <v-dialog v-model="tokenDialog" max-width="400">
      <v-card>
        <v-card-title class="pt-5">Update token</v-card-title>
        <v-card-text>
          <v-text-field v-model="inputToken" label="Bearer token" type="password"
                        variant="outlined" density="comfortable"
                        :error-messages="tokenError" @keyup.enter="saveToken" />
        </v-card-text>
        <v-card-actions class="px-4 pb-4">
          <v-spacer />
          <v-btn variant="text" @click="tokenDialog = false">Cancel</v-btn>
          <v-btn color="primary" variant="flat" :loading="verifying" @click="saveToken">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-app>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { api, onUnauthorized } from './api.js'
import ProjectList from './components/ProjectList.vue'
import CouchdevLogo from './components/CouchdevLogo.vue'
import PoweredByClaude from './components/PoweredByClaude.vue'

const authed = ref(false)
const tokenDialog = ref(false)
const inputToken = ref('')
const tokenError = ref('')
const verifying = ref(false)
const version = ref('')

onUnauthorized(() => { authed.value = false })

onMounted(async () => {
  authed.value = await api.verify()
  if (!authed.value) api.clearToken()

  try {
    const res = await fetch('/api/version')
    version.value = (await res.json()).version
  } catch {
    // non-critical, leave footer without a version
  }
})

function openTokenDialog() {
  tokenError.value = ''
  tokenDialog.value = true
}

async function saveToken() {
  tokenError.value = ''
  verifying.value = true
  api.setToken(inputToken.value)
  const ok = await api.verify()
  verifying.value = false
  if (ok) {
    authed.value = true
    tokenDialog.value = false
    inputToken.value = ''
  } else {
    api.clearToken()
    tokenError.value = 'Token rejected — paste the raw token from "couchdev token generate", not the hash.'
  }
}
</script>
