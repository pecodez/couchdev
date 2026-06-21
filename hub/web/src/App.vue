<template>
  <v-app>
    <v-app-bar title="CouchDev" color="primary" flat>
      <template #append>
        <v-btn icon="mdi-key-variant" @click="openTokenDialog" />
      </template>
    </v-app-bar>
    <v-main>
      <v-container>
        <ProjectList v-if="authed" />
        <v-card v-else class="mx-auto mt-8" max-width="400">
          <v-card-title>Set Bearer Token</v-card-title>
          <v-card-text>
            <v-text-field v-model="inputToken" label="Token" type="password"
                          :error-messages="tokenError" @keyup.enter="saveToken" />
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" block :loading="verifying" @click="saveToken">Connect</v-btn>
          </v-card-actions>
        </v-card>
      </v-container>
    </v-main>
    <v-dialog v-model="tokenDialog" max-width="400">
      <v-card>
        <v-card-title>Update Token</v-card-title>
        <v-card-text>
          <v-text-field v-model="inputToken" label="Token" type="password"
                        :error-messages="tokenError" @keyup.enter="saveToken" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="tokenDialog = false">Cancel</v-btn>
          <v-btn color="primary" :loading="verifying" @click="saveToken">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-app>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { api, onUnauthorized } from './api.js'
import ProjectList from './components/ProjectList.vue'

const authed = ref(false)
const tokenDialog = ref(false)
const inputToken = ref('')
const tokenError = ref('')
const verifying = ref(false)

function resetAuth() {
  authed.value = false
}

onUnauthorized(resetAuth)

onMounted(async () => {
  if (localStorage.getItem('couchdev_token')) {
    authed.value = await api.verify()
    if (!authed.value) api.clearToken()
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
    tokenError.value = 'Token rejected by hub — paste the token from "couchdev token generate", not the hash.'
  }
}
</script>
