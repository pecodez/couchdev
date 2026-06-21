<template>
  <v-app>
    <v-app-bar title="CouchDev" color="primary" flat>
      <template #append>
        <v-btn icon="mdi-key-variant" @click="tokenDialog = true" />
      </template>
    </v-app-bar>
    <v-main>
      <v-container>
        <ProjectList v-if="authed" />
        <v-card v-else class="mx-auto mt-8" max-width="400">
          <v-card-title>Set Bearer Token</v-card-title>
          <v-card-text>
            <v-text-field v-model="inputToken" label="Token" type="password" />
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" block @click="saveToken">Save</v-btn>
          </v-card-actions>
        </v-card>
      </v-container>
    </v-main>
    <v-dialog v-model="tokenDialog" max-width="400">
      <v-card>
        <v-card-title>Update Token</v-card-title>
        <v-card-text>
          <v-text-field v-model="inputToken" label="Token" type="password" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="tokenDialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="saveToken">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-app>
</template>

<script setup>
import { ref } from 'vue'
import { api } from './api.js'
import ProjectList from './components/ProjectList.vue'

const authed = ref(!!localStorage.getItem('couchdev_token'))
const tokenDialog = ref(false)
const inputToken = ref('')

function saveToken() {
  api.setToken(inputToken.value)
  authed.value = true
  tokenDialog.value = false
}
</script>
